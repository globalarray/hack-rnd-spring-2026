package session

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"sourcecraft.dev/benzo/testengine/internal/domain"
	"sourcecraft.dev/benzo/testengine/internal/domain/models/answer"
	"sourcecraft.dev/benzo/testengine/internal/domain/models/question"
	"sourcecraft.dev/benzo/testengine/internal/domain/models/session"
	"sourcecraft.dev/benzo/testengine/internal/domain/models/survey"
	"sourcecraft.dev/benzo/testengine/internal/service/session/dto"
)

type sessionRepo interface {
	Create(ctx context.Context, input *dto.StartSessionInput) (string, error)
	HasActiveSession(ctx context.Context, input *dto.StartSessionInput) (bool, error)
	HasCompletedShareLinkSession(ctx context.Context, input *dto.StartSessionInput) (bool, error)
	CurrentQuestion(ctx context.Context, sessionID string) (*question.Question, error)
	SessionState(ctx context.Context, sessionID string) (*session.State, error)
	Close(ctx context.Context, sessionID string, status session.SessionStatus) error
	SaveResponseAndUpdateState(ctx context.Context, data session.ResponseUpdate) error
}

type surveyRepo interface {
	QuestionWithAnswer(ctx context.Context, questionID, answerID string) (*question.Question, *answer.Answer, error)
	QuestionByID(ctx context.Context, questionID string) (*question.Question, error)
	NextQuestionByOrder(ctx context.Context, serveyID string, currentOrder int) (string, error)
}

type service struct {
	log        *slog.Logger
	repo       sessionRepo
	surveyRepo surveyRepo
}

func NewSessionService(log *slog.Logger, repo sessionRepo, surveyRepo surveyRepo) *service {
	return &service{log: log, repo: repo, surveyRepo: surveyRepo}
}

func (s *service) Start(ctx context.Context, input *dto.StartSessionInput) (*dto.StartSessionOutput, error) {
	const op = "session.Start"

	s.log.Info("starting new survey session",
		slog.String("op", op),
		slog.String("survey_id", input.SurveyID),
	)

	hasActiveSession, err := s.repo.HasActiveSession(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("%s: check active session: %w", op, err)
	}
	if hasActiveSession {
		return nil, fmt.Errorf("%s: %w", op, domain.ErrConflict)
	}

	hasCompletedShareLinkSession, err := s.repo.HasCompletedShareLinkSession(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("%s: check completed share link session: %w", op, err)
	}
	if hasCompletedShareLinkSession {
		return nil, fmt.Errorf("%s: %w", op, domain.ErrShareLinkUsed)
	}

	sessionID, err := s.repo.Create(ctx, input)
	if err != nil {
		s.log.Error("failed to create session",
			slog.String("op", op),
			slog.Any("error", err),
		)
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	q, err := s.repo.CurrentQuestion(ctx, sessionID)
	if err != nil {
		s.log.Error("failed to get first question",
			slog.String("op", op),
			slog.String("session_id", sessionID),
			slog.Any("error", err),
		)
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &dto.StartSessionOutput{
		SessionID:     sessionID,
		FirstQuestion: q,
	}, nil
}

func (s *service) CurrentQuestion(ctx context.Context, sessionID string) (*question.Question, error) {
	const op = "session.GetCurrentQuestion"

	q, err := s.repo.CurrentQuestion(ctx, sessionID)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return q, nil
}

func (s *service) resolveNextStep(
	currentQ *question.Question,
	answerID string,
	nextByOrder string,
) (nextID string, isFinished bool) {
	if rule, ok := currentQ.LogicRules[answerID]; ok {
		switch rule.Action() {
		case question.ActionFinish:
			return "", true
		case question.ActionJump:
			switch jumpRule := rule.(type) {
			case question.JumpRule:
				return jumpRule.NextQuestionID, false
			case *question.JumpRule:
				return jumpRule.NextQuestionID, false
			}
		}
	}

	if nextByOrder == "" {
		return "", true // Вопросов больше нет
	}

	return nextByOrder, false
}

func (s *service) SubmitAnswer(ctx context.Context, input dto.SubmitAnswerInput) (*dto.SubmitAnswerOutput, error) {
	const op = "session.SubmitAnswer"

	sess, err := s.repo.SessionState(ctx, input.SessionID)
	if err != nil {
		return nil, fmt.Errorf("%s: get session state: %w", op, err)
	}
	timeLimit := sess.SettingsSurvey.GetFloat64(survey.LimitsGroup, survey.LimitTimeLimit, survey.DefaultTimeLimit)
	if timeLimit > 0 && time.Since(sess.StartedAt).Seconds() > timeLimit {
		_ = s.repo.Close(ctx, input.SessionID, session.RevokedStatus)
		return nil, fmt.Errorf("%s: %w", op, domain.ErrTimeLimitExceeded)
	}

	var (
		q   *question.Question
		ans *answer.Answer
	)

	switch {
	case input.AnswerID != "":
		q, ans, err = s.surveyRepo.QuestionWithAnswer(ctx, input.QuestionID, input.AnswerID)
		if err != nil {
			return nil, fmt.Errorf("%s: validate question and answer: %w", op, err)
		}
	case input.RawText != "":
		q, err = s.surveyRepo.QuestionByID(ctx, input.QuestionID)
		if err != nil {
			return nil, fmt.Errorf("%s: validate question: %w", op, err)
		}
		ans = &answer.Answer{}
	default:
		return nil, fmt.Errorf("%s: empty answer payload", op)
	}

	nextLinearID, err := s.surveyRepo.NextQuestionByOrder(ctx, sess.SurveyID, q.OrderNumber)
	if err != nil {
		// Ошибка тут не критична, если это просто конец списка, репозиторий вернет nil
		s.log.Debug("no linear next question", slog.Any("err", err))
	}

	// 5. Вычисляем переход
	nextID, isFinished := s.resolveNextStep(q, input.AnswerID, nextLinearID)

	// 6. Подготавливаем обновление (Optimistic Lock)
	// Используем указатели, как ты просил
	update := session.ResponseUpdate{
		SessionID:                 input.SessionID,
		ExpectedCurrentQuestionID: input.QuestionID, // Проверяем, что юзер не "перескочил" вопрос
		QuestionID:                input.QuestionID,
		Weight:                    ans.Weight,
		IsFinished:                isFinished,
	}

	if input.AnswerID != "" {
		update.AnswerID = &input.AnswerID
	}
	if input.RawText != "" {
		update.RawText = &input.RawText
	}

	if nextID != "" {
		update.NextQuestionID = &nextID
	}

	// 7. Атомарное сохранение в БД
	err = s.repo.SaveResponseAndUpdateState(ctx, update)
	if err != nil {
		if errors.Is(err, domain.ErrConflict) {
			s.log.Warn("conflict: answer already submitted", slog.String("session_id", input.SessionID))
			// Можно либо вернуть ошибку, либо текущее состояние
			return nil, fmt.Errorf("%s: %w", op, domain.ErrConflict)
		}
		return nil, fmt.Errorf("%s: save state: %w", op, err)
	}

	return &dto.SubmitAnswerOutput{
		NextQuestionID: nextID,
		IsFinished:     isFinished,
	}, nil
}

func (s *service) Close(ctx context.Context, sessionID string, status session.SessionStatus) error {
	const op = "session.Close"

	s.log.Info("closing session",
		slog.String("op", op),
		slog.String("session_id", sessionID),
		slog.String("status", string(status)),
	)

	err := s.repo.Close(ctx, sessionID, status)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			s.log.Warn("attempted to close non-existent or already closed session",
				slog.String("op", op),
				slog.String("session_id", sessionID),
			)
			return fmt.Errorf("%s: %w", op, domain.ErrNotFound)
		}

		s.log.Error("failed to close session",
			slog.String("op", op),
			slog.Any("error", err),
		)
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}
