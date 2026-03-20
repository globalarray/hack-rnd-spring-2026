package service

import (
	"context"

	"github.com/google/uuid"
)

type surveyRepository interface {
	Get()
}

type surveyService struct {
	repo surveyRepository
}

func NewSurvey(repo surveyRepository) *surveyService {
	return &surveyService{repo: repo}
}

func (s *surveyService) Create(ctx context.Context, CreateSurveyInput) error {

}
