package ports

import (
	"context"

	"sourcecraft.dev/benzo/bff/internal/domain"
)

type EngineGateway interface {
	CreateSurvey(ctx context.Context, draft domain.SurveyDraft) (string, error)
	ListSurveys(ctx context.Context, psychologistID string) ([]domain.SurveySummary, error)
	StartSession(ctx context.Context, surveyID, clientMetadataJSON string) (string, *domain.Question, error)
	GetCurrentQuestion(ctx context.Context, sessionID string) (*domain.Question, error)
	SubmitAnswer(ctx context.Context, input SubmitAnswerInput) (string, bool, error)
	GetSessionAnalytics(ctx context.Context, sessionID string) (*domain.SessionAnalytics, error)
}

type AnalyticsGateway interface {
	GenerateReport(ctx context.Context, analytics domain.SessionAnalytics, format domain.ReportFormat) (*domain.GeneratedReport, error)
}

type Mailer interface {
	SendReport(ctx context.Context, message ReportEmail) error
}

type SubmitAnswerInput struct {
	SessionID  string
	QuestionID string
	AnswerID   string
	RawText    string
	AnswerIDs  []string
}

type ReportEmail struct {
	To          string
	Subject     string
	Body        string
	FileName    string
	ContentType string
	Attachment  []byte
}
