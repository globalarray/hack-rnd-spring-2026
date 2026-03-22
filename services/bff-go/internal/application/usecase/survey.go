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

	return uc.engine.CreateSurvey(ctx, draft)
}

func (uc *SurveyUseCase) ListSurveys(ctx context.Context, psychologistID string) ([]domain.SurveySummary, error) {
	if _, err := uuid.Parse(strings.TrimSpace(psychologistID)); err != nil {
		return nil, fmt.Errorf("%w: psychologistId must be a valid UUID", domain.ErrInvalidInput)
	}

	return uc.engine.ListSurveys(ctx, psychologistID)
}
