package survey

import (
	"context"
	"fmt"

	"sourcecraft.dev/benzo/testengine/internal/domain/models/survey"
	"sourcecraft.dev/benzo/testengine/internal/service/survey/dto"
)

type surveyRepository interface {
	SaveFull(ctx context.Context, in *dto.CreateSurveyInput) (string, error)
	ListByPsychologist(ctx context.Context, psychologistID string) ([]dto.SurveySummary, error)
}

type service struct {
	repo surveyRepository
}

func NewSurveyService(repo surveyRepository) *service {
	return &service{repo: repo}
}

func (s *service) Create(ctx context.Context, input *dto.CreateSurveyInput) (uuid string, err error) {
	settingsMap, err := survey.ParseSettings(input.Settings)

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

func (s *service) List(ctx context.Context, input *dto.ListSurveysInput) (*dto.ListSurveysOutput, error) {
	surveys, err := s.repo.ListByPsychologist(ctx, input.PsychologistID)
	if err != nil {
		return nil, fmt.Errorf("cannot list surveys: %w", err)
	}

	return &dto.ListSurveysOutput{Surveys: surveys}, nil
}
