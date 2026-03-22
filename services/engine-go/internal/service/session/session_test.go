package session

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"testing"

	"sourcecraft.dev/benzo/testengine/internal/domain"
	"sourcecraft.dev/benzo/testengine/internal/domain/models/answer"
	"sourcecraft.dev/benzo/testengine/internal/domain/models/question"
	sessionmodel "sourcecraft.dev/benzo/testengine/internal/domain/models/session"
	"sourcecraft.dev/benzo/testengine/internal/service/session/dto"
)

func TestServiceStartRejectsCompletedShareLinkSession(t *testing.T) {
	t.Parallel()

	repo := &sessionRepoStub{
		hasCompletedShareLinkSession: true,
	}

	service := NewSessionService(
		slog.New(slog.NewTextHandler(io.Discard, nil)),
		repo,
		surveyRepoStub{},
	)

	_, err := service.Start(context.Background(), &dto.StartSessionInput{
		SurveyID:          "survey-id",
		ClientMetadataRaw: `{"email":"candidate@example.com","__shareLinkId":"b590f5ab-f58e-4c1e-a995-d6d23a8228e3"}`,
		ShareLinkID:       "b590f5ab-f58e-4c1e-a995-d6d23a8228e3",
	})
	if !errors.Is(err, domain.ErrShareLinkUsed) {
		t.Fatalf("expected ErrShareLinkUsed, got %v", err)
	}

	if repo.createCalled {
		t.Fatal("expected Create not to be called for an already used share link")
	}
}

type sessionRepoStub struct {
	hasActiveSession                bool
	hasActiveSessionErr             error
	hasCompletedShareLinkSession    bool
	hasCompletedShareLinkSessionErr error
	createCalled                    bool
}

func (s *sessionRepoStub) Create(context.Context, *dto.StartSessionInput) (string, error) {
	s.createCalled = true
	return "session-id", nil
}

func (s *sessionRepoStub) HasActiveSession(context.Context, *dto.StartSessionInput) (bool, error) {
	return s.hasActiveSession, s.hasActiveSessionErr
}

func (s *sessionRepoStub) HasCompletedShareLinkSession(context.Context, *dto.StartSessionInput) (bool, error) {
	return s.hasCompletedShareLinkSession, s.hasCompletedShareLinkSessionErr
}

func (s *sessionRepoStub) CurrentQuestion(context.Context, string) (*question.Question, error) {
	return nil, nil
}

func (s *sessionRepoStub) SessionState(context.Context, string) (*sessionmodel.State, error) {
	return nil, nil
}

func (s *sessionRepoStub) Close(context.Context, string, sessionmodel.SessionStatus) error {
	return nil
}

func (s *sessionRepoStub) SaveResponseAndUpdateState(context.Context, sessionmodel.ResponseUpdate) error {
	return nil
}

type surveyRepoStub struct{}

func (surveyRepoStub) QuestionWithAnswer(context.Context, string, string) (*question.Question, *answer.Answer, error) {
	return nil, nil, nil
}

func (surveyRepoStub) QuestionByID(context.Context, string) (*question.Question, error) {
	return nil, nil
}

func (surveyRepoStub) NextQuestionByOrder(context.Context, string, int) (string, error) {
	return "", nil
}
