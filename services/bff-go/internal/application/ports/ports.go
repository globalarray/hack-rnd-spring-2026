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

type AuthGateway interface {
	Login(ctx context.Context, email, password string) (*domain.AuthTokens, error)
	RefreshToken(ctx context.Context, refreshToken string) (*domain.AuthTokens, error)
	Register(ctx context.Context, token, password string) (*domain.AuthTokens, error)
	GetProfile(ctx context.Context, authorization string) (*domain.UserProfile, error)
	UpdateProfile(ctx context.Context, authorization string, input domain.ProfileUpdate) (*domain.UserProfile, error)
	GetPublicProfile(ctx context.Context, userID string) (*domain.PublicProfile, error)
	CreateInvitation(ctx context.Context, authorization string, input domain.InvitationDraft) (string, error)
	BlockUser(ctx context.Context, authorization, email string) error
	UnblockUser(ctx context.Context, authorization, email string) error
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
