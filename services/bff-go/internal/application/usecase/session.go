package usecase

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"github.com/google/uuid"
	"sourcecraft.dev/benzo/bff/internal/application/ports"
	"sourcecraft.dev/benzo/bff/internal/domain"
)

type SessionUseCase struct {
	log                 *slog.Logger
	engine              ports.EngineGateway
	analytics           ports.AnalyticsGateway
	mailer              ports.Mailer
	defaultReportFormat domain.ReportFormat
}

func NewSessionUseCase(
	log *slog.Logger,
	engine ports.EngineGateway,
	analytics ports.AnalyticsGateway,
	mailer ports.Mailer,
	defaultReportFormat string,
) *SessionUseCase {
	return &SessionUseCase{
		log:                 log,
		engine:              engine,
		analytics:           analytics,
		mailer:              mailer,
		defaultReportFormat: domain.ParseReportFormat(defaultReportFormat),
	}
}

func (uc *SessionUseCase) StartSession(ctx context.Context, surveyID, shareLinkID string, metadata domain.ClientMetadata) (string, *domain.Question, error) {
	if err := validateUUID("surveyId", surveyID); err != nil {
		return "", nil, err
	}
	if strings.TrimSpace(shareLinkID) != "" {
		if err := validateUUID("shareLinkId", shareLinkID); err != nil {
			return "", nil, err
		}

		values := metadata.Values()
		values["__shareLinkId"] = strings.TrimSpace(shareLinkID)
		metadata = domain.NewClientMetadata(values)
	}

	if _, err := metadata.Email(); err != nil {
		return "", nil, err
	}

	clientMetadataJSON, err := metadata.JSON()
	if err != nil {
		return "", nil, err
	}

	return uc.engine.StartSession(ctx, surveyID, clientMetadataJSON)
}

func (uc *SessionUseCase) GetCurrentQuestion(ctx context.Context, sessionID string) (*domain.Question, error) {
	if err := validateUUID("sessionId", sessionID); err != nil {
		return nil, err
	}

	return uc.engine.GetCurrentQuestion(ctx, sessionID)
}

func (uc *SessionUseCase) SubmitAnswer(ctx context.Context, input ports.SubmitAnswerInput) (*domain.SubmitAnswerResult, error) {
	if err := validateUUID("sessionId", input.SessionID); err != nil {
		return nil, err
	}

	if err := validateUUID("questionId", input.QuestionID); err != nil {
		return nil, err
	}

	if err := validateAnswerPayload(input); err != nil {
		return nil, err
	}

	nextQuestionID, isFinished, err := uc.engine.SubmitAnswer(ctx, input)
	if err != nil {
		return nil, err
	}

	result := &domain.SubmitAnswerResult{
		NextQuestionID: nextQuestionID,
		IsFinished:     isFinished,
	}

	if !isFinished {
		nextQuestion, err := uc.engine.GetCurrentQuestion(ctx, input.SessionID)
		if err != nil {
			return nil, err
		}

		if nextQuestion != nil && strings.TrimSpace(result.NextQuestionID) == "" {
			result.NextQuestionID = nextQuestion.ID
		}

		result.NextQuestion = nextQuestion
		return result, nil
	}

	delivery, err := uc.SendClientReport(ctx, input.SessionID, uc.defaultReportFormat)
	if err != nil {
		uc.log.Warn("report delivery failed after session completion",
			slog.String("session_id", input.SessionID),
			slog.Any("error", err),
		)

		result.ReportDelivery = &domain.ReportDelivery{
			Status:       domain.ReportDeliveryFailed,
			ErrorMessage: err.Error(),
		}
		return result, nil
	}

	result.ReportDelivery = delivery
	return result, nil
}

func (uc *SessionUseCase) GetSessionAnalytics(ctx context.Context, sessionID string) (*domain.SessionAnalytics, error) {
	if err := validateUUID("sessionId", sessionID); err != nil {
		return nil, err
	}

	return uc.engine.GetSessionAnalytics(ctx, sessionID)
}

func (uc *SessionUseCase) ListSurveySessions(ctx context.Context, surveyID string) ([]domain.SurveySessionSummary, error) {
	if err := validateUUID("surveyId", surveyID); err != nil {
		return nil, err
	}

	return uc.engine.ListSurveySessions(ctx, surveyID)
}

func (uc *SessionUseCase) SendClientReport(ctx context.Context, sessionID string, format domain.ReportFormat) (*domain.ReportDelivery, error) {
	if err := validateUUID("sessionId", sessionID); err != nil {
		return nil, err
	}

	analyticsData, err := uc.engine.GetSessionAnalytics(ctx, sessionID)
	if err != nil {
		return nil, err
	}

	email, err := analyticsData.ClientMetadata.Email()
	if err != nil {
		return nil, err
	}

	report, err := uc.analytics.GenerateReport(ctx, *analyticsData, format)
	if err != nil {
		return nil, err
	}

	message := ports.ReportEmail{
		To:          email,
		Subject:     "ProfDNA report",
		Body:        "Your ProfDNA report is attached to this email.",
		FileName:    report.FileName,
		ContentType: report.ContentType,
		Attachment:  report.Content,
	}

	if err := uc.mailer.SendReport(ctx, message); err != nil {
		return nil, err
	}

	return &domain.ReportDelivery{
		Status:      domain.ReportDeliverySent,
		Email:       email,
		FileName:    report.FileName,
		ContentType: report.ContentType,
	}, nil
}

func validateUUID(fieldName, value string) error {
	if _, err := uuid.Parse(strings.TrimSpace(value)); err != nil {
		return fmt.Errorf("%w: %s must be a valid UUID", domain.ErrInvalidInput, fieldName)
	}
	return nil
}

func validateAnswerPayload(input ports.SubmitAnswerInput) error {
	filled := 0
	if strings.TrimSpace(input.AnswerID) != "" {
		filled++
	}
	if strings.TrimSpace(input.RawText) != "" {
		filled++
	}
	if len(input.AnswerIDs) > 0 {
		filled++
	}

	if filled == 0 {
		return fmt.Errorf("%w: one of answerId, rawText or answerIds must be provided", domain.ErrInvalidInput)
	}

	if filled > 1 {
		return fmt.Errorf("%w: provide only one answer payload", domain.ErrInvalidInput)
	}

	if answerID := strings.TrimSpace(input.AnswerID); answerID != "" {
		if err := validateUUID("answerId", answerID); err != nil {
			return err
		}
	}

	if rawText := strings.TrimSpace(input.RawText); input.RawText != "" && rawText == "" {
		return fmt.Errorf("%w: rawText must not be empty", domain.ErrInvalidInput)
	}

	if len(input.AnswerIDs) > 0 {
		seen := make(map[string]struct{}, len(input.AnswerIDs))
		for _, answerID := range input.AnswerIDs {
			if err := validateUUID("answerIds[]", answerID); err != nil {
				return err
			}

			normalized := strings.TrimSpace(answerID)
			if _, exists := seen[normalized]; exists {
				return fmt.Errorf("%w: answerIds must be unique", domain.ErrInvalidInput)
			}
			seen[normalized] = struct{}{}
		}
	}

	return nil
}
