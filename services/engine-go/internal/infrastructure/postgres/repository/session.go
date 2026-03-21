package repository

import (
	"context"
	"log/slog"

	"github.com/jmoiron/sqlx"
	"sourcecraft.dev/benzo/testengine/internal/domain/models/question"
)

type sessionRepository struct {
	log *slog.Logger
	db  *sqlx.DB
}

func NewSessionRepository(log *slog.Logger, db *sqlx.DB) *sessionRepository {
	return &sessionRepository{log: log, db: db}
}

func (repo *sessionRepository) Create(ctx context.Context, surveyID string) error {
	
}
