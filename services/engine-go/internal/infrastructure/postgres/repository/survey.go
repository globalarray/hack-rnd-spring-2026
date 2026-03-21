package repository

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/jmoiron/sqlx"
	"sourcecraft.dev/benzo/testengine/internal/service/survey/dto"
)

type surveyRepository struct {
	log *slog.Logger
	db  *sqlx.DB
}

const timeout = time.Second * 5

func NewSurveyRepository(log *slog.Logger, db *sqlx.DB) *surveyRepository {
	return &surveyRepository{
		log: log,
		db:  db,
	}
}

func (r *surveyRepository) CreateFull(ctx context.Context, in *dto.CreateSurveyInput) (surveyUUID string, err error) {
	ctx, cancel := context.WithTimeout(ctx, timeout)

	defer cancel()

	tx, err := r.db.BeginTxx(ctx, nil)

	if err != nil {
		return
	}

	defer func() {
		if err := tx.Rollback(); err != nil {
			r.log.With(slog.Any("error", err)).Error("rollback transaction")
		}
	}()

	if err := tx.QueryRowxContext(ctx, queryInsertSurvey, in.PsychologistID, in.Title, in.Description, in.Settings).Scan(&surveyUUID); err != nil {
		return surveyUUID, fmt.Errorf("insert survey: %w", err)
	}

	for _, q := range in.Questions {
		var qUUID string

		if err := tx.QueryRowxContext(ctx, queryInsertQuestion, surveyUUID, q.OrderNum, q.Type, q.Text, q.LogicRules).Scan(&qUUID); err != nil {
			return surveyUUID, fmt.Errorf("insert question: %w", err)
		}

		for _, a := range q.Answers {
			if _, err := tx.ExecContext(ctx, queryInsertAnswer, surveyUUID, qUUID, a.Text, a.Weight, a.CategoryTag); err != nil {
				return surveyUUID, fmt.Errorf("insert answer: %w", err)
			}
		}
	}

	if err := tx.Commit(); err != nil {
		return surveyUUID, fmt.Errorf("survey commit: %w", err)
	}

	return surveyUUID, nil
}
