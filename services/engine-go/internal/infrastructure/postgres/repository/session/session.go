package session

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"strings"

	"github.com/jmoiron/sqlx"
	"sourcecraft.dev/benzo/testengine/internal/domain"
	"sourcecraft.dev/benzo/testengine/internal/domain/models/question"
	"sourcecraft.dev/benzo/testengine/internal/domain/models/session"
	"sourcecraft.dev/benzo/testengine/internal/infrastructure/postgres/repository/session/dto"
	servicedto "sourcecraft.dev/benzo/testengine/internal/service/session/dto"
)

type sessionRepository struct {
	log *slog.Logger
	db  *sqlx.DB
}

func NewSessionRepository(log *slog.Logger, db *sqlx.DB) *sessionRepository {
	return &sessionRepository{log: log, db: db}
}

func normalizeClientMetadata(raw string) string {
	if raw == "" {
		return "{}"
	}

	return raw
}

func extractShareLinkID(raw string) string {
	payload := strings.TrimSpace(raw)
	if payload == "" {
		return ""
	}

	var metadata map[string]any
	if err := json.Unmarshal([]byte(payload), &metadata); err != nil {
		return ""
	}

	shareLinkID, _ := metadata["__shareLinkId"].(string)
	return strings.TrimSpace(shareLinkID)
}

func (repo *sessionRepository) Create(ctx context.Context, input *servicedto.StartSessionInput) (sessionID string, err error) {
	const op = "sessionRepository.Create"

	clientMetadata := normalizeClientMetadata(input.ClientMetadataRaw)

	if err := repo.db.QueryRowContext(ctx, queryInsertSession,
		input.SurveyID,
		session.CreatedStatus,
		clientMetadata,
	).Scan(&sessionID); err != nil {
		return "", fmt.Errorf("%s: failed to insert session: %w", op, err)
	}

	return sessionID, nil
}

func (repo *sessionRepository) HasActiveSession(ctx context.Context, input *servicedto.StartSessionInput) (bool, error) {
	const op = "sessionRepository.HasActiveSession"

	var exists bool

	if err := repo.db.GetContext(ctx, &exists, queryHasActiveSession, input.SurveyID, normalizeClientMetadata(input.ClientMetadataRaw)); err != nil {
		return false, fmt.Errorf("%s: check active session: %w", op, err)
	}

	return exists, nil
}

func (repo *sessionRepository) HasCompletedShareLinkSession(ctx context.Context, input *servicedto.StartSessionInput) (bool, error) {
	const op = "sessionRepository.HasCompletedShareLinkSession"

	shareLinkID := strings.TrimSpace(input.ShareLinkID)
	if shareLinkID == "" {
		shareLinkID = extractShareLinkID(input.ClientMetadataRaw)
	}
	if shareLinkID == "" {
		return false, nil
	}

	var exists bool
	if err := repo.db.GetContext(ctx, &exists, queryHasCompletedShareLinkSession, input.SurveyID, shareLinkID); err != nil {
		return false, fmt.Errorf("%s: check completed share link session: %w", op, err)
	}

	return exists, nil
}

func (repo *sessionRepository) CurrentQuestion(ctx context.Context, sessionID string) (*question.Question, error) {
	const op = "sessionRepository.GetCurrentQuestion"

	var record dto.CurrentQuestionRecord

	err := repo.db.GetContext(ctx, &record, queryGetCurrentQuestionBySessionID, sessionID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("%s: %w", op, domain.ErrNotFound)
		}
		return nil, fmt.Errorf("%s: failed to get current question: %w", op, err)
	}

	q, err := mapCurrentQuestionRecordToDomain(&record)
	if err != nil {
		return nil, fmt.Errorf("%s: map current question: %w", op, err)
	}

	return q, nil
}

func (repo *sessionRepository) SaveResponseAndUpdateState(ctx context.Context, data session.ResponseUpdate) error {
	const op = "sessionRepository.SaveResponseAndUpdateState"

	tx, err := repo.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("%s: failed to start transaction: %w", op, err)
	}

	defer func() {
		if rollBackErr := tx.Rollback(); rollBackErr != nil && !errors.Is(rollBackErr, sql.ErrTxDone) {
			err = rollBackErr
		}
	}()

	if _, err := tx.ExecContext(ctx, queryInsertResponse, data.SessionID, data.QuestionID, data.AnswerID, data.RawText); err != nil {
		return fmt.Errorf("%s: insert response: %w", op, err)
	}

	status := session.InProgressStatus
	if data.IsFinished {
		status = session.CompletedStatus
	}

	res, err := tx.ExecContext(ctx, queryUpdateSession, data.NextQuestionID, status, data.IsFinished, data.SessionID, data.ExpectedCurrentQuestionID)
	if err != nil {
		return fmt.Errorf("%s: update session: %w", op, err)
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("%s: rows affected: %w", op, err)
	}
	if rows == 0 {
		return domain.ErrConflict
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("%s: commit transaction: %w", op, err)
	}

	return nil
}

func (repo *sessionRepository) SessionState(ctx context.Context, sessionID string) (*session.State, error) {
	const op = "sessionRepository.GetSessionState"

	var record = &dto.SessionStateRecord{}

	if err := repo.db.GetContext(ctx, record, queryGetSessionState, sessionID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("%s: session not found: %w", op, domain.ErrNotFound)
		}
		return nil, fmt.Errorf("%s: execute query: %w", op, err)
	}

	return mapSessionStateRecordToDomain(record)
}

func (repo *sessionRepository) Close(ctx context.Context, sessionID string, status session.SessionStatus) error {
	const op = "sessionRepository.Close"

	result, err := repo.db.ExecContext(ctx, queryCloseSession, status, sessionID)
	if err != nil {
		return fmt.Errorf("%s: execute update: %w", op, err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("%s: get rows affected: %w", op, err)
	}

	if rows == 0 {
		return fmt.Errorf("%s: session not found or already closed: %w", op, domain.ErrNotFound)
	}

	return nil
}
