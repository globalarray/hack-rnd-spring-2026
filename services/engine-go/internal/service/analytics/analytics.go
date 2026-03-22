package analytics

import (
	"context"
	"fmt"

	"sourcecraft.dev/benzo/testengine/internal/service/analytics/dto"
)

type analyticsRepository interface {
	GetSessionData(ctx context.Context, sessionID string) (*dto.GetSessionDataOutput, error)
	ListSurveySessions(ctx context.Context, surveyID string) (*dto.ListSurveySessionsOutput, error)
}

type service struct {
	repo analyticsRepository
}

func NewAnalyticsService(repo analyticsRepository) *service {
	return &service{repo: repo}
}

func (s *service) GetSessionData(ctx context.Context, input *dto.GetSessionDataInput) (*dto.GetSessionDataOutput, error) {
	data, err := s.repo.GetSessionData(ctx, input.SessionID)
	if err != nil {
		return nil, fmt.Errorf("analytics.GetSessionData: %w", err)
	}

	return data, nil
}

func (s *service) ListSurveySessions(ctx context.Context, input *dto.ListSurveySessionsInput) (*dto.ListSurveySessionsOutput, error) {
	data, err := s.repo.ListSurveySessions(ctx, input.SurveyID)
	if err != nil {
		return nil, fmt.Errorf("analytics.ListSurveySessions: %w", err)
	}

	return data, nil
}
