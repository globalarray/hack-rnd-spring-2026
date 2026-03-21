package repository

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"sourcecraft.dev/benzo/testengine/internal/domain/models/question"
	"sourcecraft.dev/benzo/testengine/internal/infrastructure/postgres/repository/dto"
	servicedto "sourcecraft.dev/benzo/testengine/internal/service/survey/dto"
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

func (r *surveyRepository) SaveFull(ctx context.Context, in *servicedto.CreateSurveyInput) (surveyUUID string, err error) {
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
			if _, err := tx.ExecContext(ctx, queryInsertAnswer, a.ID, qUUID, a.Text, a.Weight, a.CategoryTag); err != nil {
				return surveyUUID, fmt.Errorf("insert answer: %w", err)
			}
		}
	}

	if err := tx.Commit(); err != nil {
		return surveyUUID, fmt.Errorf("survey commit: %w", err)
	}

	return surveyUUID, nil
}

func (r *surveyRepository) GetQuestionByOrderAndSurvey(ctx context.Context, surveyID uuid.UUID, orderNumber int) (q question.Question, err error) {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	var questionRow dto.QuestionRecord

	if err := r.db.GetContext(ctx, &questionRow, querySelectQuestionWithAnswers, surveyID, orderNumber); err != nil {
		return q, fmt.Errorf("select question: %w", err)
	}

	return mapQuestionRecordToQuestion(questionRow)
}
