package survey

import (
	"context"
	"fmt"

	"sourcecraft.dev/benzo/testengine/internal/domain"
	"sourcecraft.dev/benzo/testengine/internal/service/survey/dto"
)

type surveyRepository interface {
	SaveFull(ctx context.Context, in *dto.CreateSurveyInput) (string, error)
}

type surveyService struct {
	repo surveyRepository
}

func NewSurvey(repo surveyRepository) *surveyService {
	return &surveyService{repo: repo}
}

func (s *surveyService) CreateFull(ctx context.Context, input *dto.CreateSurveyInput) (uuid string, err error) {
	settingsMap, err := domain.ParseSettings(input.Settings)

	if err != nil {
		return uuid, fmt.Errorf("cannot parse settingsJson: %w", err)
	}
	if err := settingsMap.Validate(); err != nil {
		return uuid, fmt.Errorf("invalid settings: %w", err)
	}

	uuid, err = s.repo.SaveFull(ctx, input)

	if err != nil {
		return uuid, fmt.Errorf("cannot save survey: %w", err)
	}

	return uuid, nil
}
