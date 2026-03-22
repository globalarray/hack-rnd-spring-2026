package usecase

import (
	"context"
	"encoding/json"
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

func TestSessionUseCaseStartSessionInjectsShareLinkIDIntoMetadata(t *testing.T) {
	t.Parallel()

	engine := &engineGatewayStub{
		startSessionID: "67d77769-eb9e-44ce-999e-84d34b7379fd",
		startSessionQuestion: &domain.Question{
			ID:   "3e4b60ec-c80d-4ae0-97ea-0a2c7dd7c8d2",
			Type: domain.QuestionTypeSingleChoice,
			Text: "First question",
		},
	}

	uc := NewSessionUseCase(
		slog.New(slog.NewTextHandler(io.Discard, nil)),
		engine,
		analyticsGatewayStub{},
		mailerStub{},
		"",
	)

	sessionID, _, err := uc.StartSession(
		context.Background(),
		"f6c3201d-0619-4a26-a07d-f7c819540b99",
		"234b4e61-9389-4474-aab6-76d0831f8c53",
		domain.NewClientMetadata(map[string]any{
			"email":    "candidate@example.com",
			"fullName": "Иван Иванов",
		}),
	)
	if err != nil {
		t.Fatalf("StartSession returned error: %v", err)
	}

	if sessionID != engine.startSessionID {
		t.Fatalf("expected session id %q, got %q", engine.startSessionID, sessionID)
	}

	if engine.startSessionSurveyID != "f6c3201d-0619-4a26-a07d-f7c819540b99" {
		t.Fatalf("unexpected survey id passed to engine: %q", engine.startSessionSurveyID)
	}

	var payload map[string]any
	if err := json.Unmarshal([]byte(engine.startSessionClientMetadataJSON), &payload); err != nil {
		t.Fatalf("expected valid metadata json, got %v", err)
	}

	if payload["__shareLinkId"] != "234b4e61-9389-4474-aab6-76d0831f8c53" {
		t.Fatalf("expected __shareLinkId to be injected, got %#v", payload["__shareLinkId"])
	}
}

type engineGatewayStub struct {
	submitAnswerNextQuestionID     string
	submitAnswerIsFinished         bool
	submitAnswerErr                error
	startSessionID                 string
	startSessionSurveyID           string
	startSessionClientMetadataJSON string
	startSessionQuestion           *domain.Question
	startSessionErr                error
	currentQuestion                *domain.Question
	currentQuestionErr             error
	getCurrentQuestionCalled       bool
}

func (s *engineGatewayStub) CreateSurvey(context.Context, domain.SurveyDraft) (string, error) {
	panic("unexpected call to CreateSurvey")
}

func (s *engineGatewayStub) ListSurveys(context.Context, string) ([]domain.SurveySummary, error) {
	panic("unexpected call to ListSurveys")
}

func (s *engineGatewayStub) StartSession(_ context.Context, surveyID, clientMetadataJSON string) (string, *domain.Question, error) {
	s.startSessionSurveyID = surveyID
	s.startSessionClientMetadataJSON = clientMetadataJSON
	return s.startSessionID, s.startSessionQuestion, s.startSessionErr
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
