package session

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"

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

func (repo *sessionRepository) Create(ctx context.Context, input *servicedto.StartSessionInput) (sessionID string, err error) {
	const op = "sessionRepository.Create"

	if err := repo.db.QueryRowContext(ctx, queryInsertSession,
		input.SurveyID,
		session.CreatedStatus,
		"{}",
	).Scan(&sessionID); err != nil {
		return "", fmt.Errorf("%s: failed to insert session: %w", op, err)
	}

	return sessionID, nil
}

func (repo *sessionRepository) CurrentQuestion(ctx context.Context, sessionID string) (*question.Question, error) {
	const op = "sessionRepository.GetCurrentQuestion"

	var q question.Question

	err := repo.db.GetContext(ctx, &q, queryGetCurrentQuestionBySessionID, sessionID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("%s: %w", op, domain.ErrNotFound)
		}
		return nil, fmt.Errorf("%s: failed to get current question: %w", op, err)
	}

	return &q, nil
}

func (repo *sessionRepository) SaveResponseAndUpdateState(ctx context.Context, data session.ResponseUpdate) error {
	const op = "sessionRepository.SaveResponseAndUpdateState"

	tx, err := repo.db.BeginTxx(ctx, nil)

	defer func() {
		if rollBackErr := tx.Rollback(); rollBackErr != nil {
			err = rollBackErr
		}
	}()

	if err != nil {
		return fmt.Errorf("%s: failed to start transaction: %w", op, err)
	}

	status := session.InProgressStatus
	if data.IsFinished {
		status = session.CompletedStatus
	}

	res, err := tx.ExecContext(ctx, queryInsertSession, data.SessionID, data.QuestionID, data.AnswerID)

	if err != nil {
		return err
	}

	rows, _ := res.RowsAffected()
	if rows == 0 {
		// Если 0 строк, значит либо сессии нет, либо на этот вопрос уже ответили
		// Возвращаем кастомную ошибку, чтобы сервис понял: это дубликат
		return domain.ErrConflict
	}

	if _, err := tx.ExecContext(ctx, queryUpdateSession, data.NextQuestionID, status, data.IsFinished, data.SessionID); err != nil {
		return err
	}

	if err := tx.Commit(); err != nil {
		return err
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
