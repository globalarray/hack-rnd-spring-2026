package usecase

import (
	"context"
	"io"
	"log/slog"
	"testing"

	"sourcecraft.dev/benzo/bff/internal/application/ports"
	"sourcecraft.dev/benzo/bff/internal/domain"
)

func TestSessionUseCaseSubmitAnswerFetchesNextQuestionWithoutNextQuestionID(t *testing.T) {
	t.Parallel()

	engine := &engineGatewayStub{
		submitAnswerNextQuestionID: "",
		submitAnswerIsFinished:     false,
		currentQuestion: &domain.Question{
			ID:   "next-question-id",
			Type: domain.QuestionTypeSingleChoice,
			Text: "Next question",
			Answers: []domain.AnswerOption{
				{ID: "answer-1", Text: "Option 1"},
			},
		},
	}

	uc := NewSessionUseCase(
		slog.New(slog.NewTextHandler(io.Discard, nil)),
		engine,
		analyticsGatewayStub{},
		mailerStub{},
		"",
	)

	result, err := uc.SubmitAnswer(context.Background(), ports.SubmitAnswerInput{
		SessionID:  "b8910caa-941b-4a70-857e-09400ef57aef",
		QuestionID: "3ea0e14c-589a-4dbe-8d40-2372ed9afb1e",
		AnswerID:   "123e4567-e89b-42d3-a456-426614174002",
	})
	if err != nil {
		t.Fatalf("SubmitAnswer returned error: %v", err)
	}

	if !engine.getCurrentQuestionCalled {
		t.Fatal("expected GetCurrentQuestion to be called")
	}

	if result.NextQuestion == nil {
		t.Fatal("expected next question to be populated")
	}

	if result.NextQuestionID != engine.currentQuestion.ID {
		t.Fatalf("expected nextQuestionId %q, got %q", engine.currentQuestion.ID, result.NextQuestionID)
	}

	if result.IsFinished {
		t.Fatal("expected unfinished session")
	}
}

type engineGatewayStub struct {
	submitAnswerNextQuestionID string
	submitAnswerIsFinished     bool
	submitAnswerErr            error
	currentQuestion            *domain.Question
	currentQuestionErr         error
	getCurrentQuestionCalled   bool
}

func (s *engineGatewayStub) CreateSurvey(context.Context, domain.SurveyDraft) (string, error) {
	panic("unexpected call to CreateSurvey")
}

func (s *engineGatewayStub) ListSurveys(context.Context, string) ([]domain.SurveySummary, error) {
	panic("unexpected call to ListSurveys")
}

func (s *engineGatewayStub) StartSession(context.Context, string, string) (string, *domain.Question, error) {
	panic("unexpected call to StartSession")
}

func (s *engineGatewayStub) GetCurrentQuestion(_ context.Context, _ string) (*domain.Question, error) {
	s.getCurrentQuestionCalled = true
	return s.currentQuestion, s.currentQuestionErr
}

func (s *engineGatewayStub) SubmitAnswer(_ context.Context, _ ports.SubmitAnswerInput) (string, bool, error) {
	return s.submitAnswerNextQuestionID, s.submitAnswerIsFinished, s.submitAnswerErr
}

func (s *engineGatewayStub) GetSessionAnalytics(context.Context, string) (*domain.SessionAnalytics, error) {
	panic("unexpected call to GetSessionAnalytics")
}

func (s *engineGatewayStub) ListSurveySessions(context.Context, string) ([]domain.SurveySessionSummary, error) {
	panic("unexpected call to ListSurveySessions")
}

type analyticsGatewayStub struct{}

func (analyticsGatewayStub) GenerateReport(context.Context, domain.SessionAnalytics, domain.ReportFormat) (*domain.GeneratedReport, error) {
	panic("unexpected call to GenerateReport")
}

type mailerStub struct{}

func (mailerStub) SendReport(context.Context, ports.ReportEmail) error {
	panic("unexpected call to SendReport")
}
