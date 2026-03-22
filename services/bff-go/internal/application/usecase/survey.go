package usecase

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"sourcecraft.dev/benzo/bff/internal/application/ports"
	"sourcecraft.dev/benzo/bff/internal/domain"
)

type SurveyUseCase struct {
	engine ports.EngineGateway
}

func NewSurveyUseCase(engine ports.EngineGateway) *SurveyUseCase {
	return &SurveyUseCase{engine: engine}
}

func (uc *SurveyUseCase) CreateSurvey(ctx context.Context, draft domain.SurveyDraft) (string, error) {
	if _, err := uuid.Parse(strings.TrimSpace(draft.PsychologistID)); err != nil {
		return "", fmt.Errorf("%w: psychologistId must be a valid UUID", domain.ErrInvalidInput)
	}

	if strings.TrimSpace(draft.Title) == "" {
		return "", fmt.Errorf("%w: title is required", domain.ErrInvalidInput)
	}

	if len(draft.Questions) == 0 {
		return "", fmt.Errorf("%w: at least one question is required", domain.ErrInvalidInput)
	}

	orderNums := make(map[uint32]struct{}, len(draft.Questions))
	answerIDs := make(map[string]struct{})

	for index, question := range draft.Questions {
		if question.OrderNum == 0 {
			return "", fmt.Errorf("%w: questions[%d].orderNum must be greater than zero", domain.ErrInvalidInput, index)
		}
		if _, exists := orderNums[question.OrderNum]; exists {
			return "", fmt.Errorf("%w: duplicate question orderNum %d", domain.ErrInvalidInput, question.OrderNum)
		}
		orderNums[question.OrderNum] = struct{}{}

		if strings.TrimSpace(question.Text) == "" {
			return "", fmt.Errorf("%w: questions[%d].text is required", domain.ErrInvalidInput, index)
		}

		switch question.Type {
		case domain.QuestionTypeSingleChoice, domain.QuestionTypeMultipleChoice, domain.QuestionTypeScale:
			if len(question.Answers) == 0 {
				return "", fmt.Errorf("%w: questions[%d] must contain answers", domain.ErrInvalidInput, index)
			}
		}

		for answerIndex, answer := range question.Answers {
			id := strings.TrimSpace(answer.ID)
			if _, err := uuid.Parse(id); err != nil {
				return "", fmt.Errorf("%w: questions[%d].answers[%d].id must be a valid UUID", domain.ErrInvalidInput, index, answerIndex)
			}
			if _, exists := answerIDs[id]; exists {
				return "", fmt.Errorf("%w: answer ids must be globally unique", domain.ErrInvalidInput)
			}
			answerIDs[id] = struct{}{}

			if strings.TrimSpace(answer.Text) == "" {
				return "", fmt.Errorf("%w: questions[%d].answers[%d].text is required", domain.ErrInvalidInput, index, answerIndex)
			}
		}
	}

	return uc.engine.CreateSurvey(ctx, draft)
}

func (uc *SurveyUseCase) ListSurveys(ctx context.Context, psychologistID string) ([]domain.SurveySummary, error) {
	if _, err := uuid.Parse(strings.TrimSpace(psychologistID)); err != nil {
		return nil, fmt.Errorf("%w: psychologistId must be a valid UUID", domain.ErrInvalidInput)
	}

	return uc.engine.ListSurveys(ctx, psychologistID)
}
