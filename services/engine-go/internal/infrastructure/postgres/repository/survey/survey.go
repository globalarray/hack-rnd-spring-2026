package survey

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"sourcecraft.dev/benzo/testengine/internal/domain"
	"sourcecraft.dev/benzo/testengine/internal/domain/models/answer"
	"sourcecraft.dev/benzo/testengine/internal/domain/models/question"
	"sourcecraft.dev/benzo/testengine/internal/infrastructure/postgres/repository/survey/dto"
	servicedto "sourcecraft.dev/benzo/testengine/internal/service/survey/dto"
)

type surveyRepository struct {
	log *slog.Logger
	db  *sqlx.DB
}

func NewSurveyRepository(log *slog.Logger, db *sqlx.DB) *surveyRepository {
	return &surveyRepository{
		log: log,
		db:  db,
	}
}

func (r *surveyRepository) SaveFull(ctx context.Context, in *servicedto.CreateSurveyInput) (surveyUUID string, err error) {
	const op = "surveyRepository.SaveFull"

	tx, err := r.db.BeginTxx(ctx, nil)

	if err != nil {
		return
	}

	defer func() {
		if rollBackErr := tx.Rollback(); rollBackErr != nil && !errors.Is(rollBackErr, sql.ErrTxDone) {
			r.log.With(slog.Any("error", rollBackErr)).Error("rollback transaction")

			if err == nil {
				err = rollBackErr
			}
		}
	}()

	if err := tx.QueryRowxContext(ctx, queryInsertSurvey, in.PsychologistID, in.Title, in.Description, in.Settings).Scan(&surveyUUID); err != nil {
		return surveyUUID, fmt.Errorf("%s: insert survey: %w", op, err)
	}

	for _, q := range in.Questions {
		var qUUID string

		if err := tx.QueryRowxContext(ctx, queryInsertQuestion, surveyUUID, q.OrderNum, q.Type, q.Text, q.LogicRules).Scan(&qUUID); err != nil {
			return surveyUUID, fmt.Errorf("%s: insert question: %w", op, err)
		}

		for _, a := range q.Answers {
			if _, err := tx.ExecContext(ctx, queryInsertAnswer, a.ID, qUUID, a.Text, a.Weight, a.CategoryTag); err != nil {
				return surveyUUID, fmt.Errorf("%s: insert answer: %w", op, err)
			}
		}
	}

	if err := tx.Commit(); err != nil {
		return surveyUUID, fmt.Errorf("%s: survey commit: %w", op, err)
	}

	return surveyUUID, nil
}

func (r *surveyRepository) GetQuestionByOrderAndSurvey(ctx context.Context, surveyID uuid.UUID, orderNumber int) (*question.Question, error) {
	const op = "surveyRepository.GetQuestionByOrderAndSurvey"

	var questionRow dto.QuestionRecord

	if err := r.db.GetContext(ctx, &questionRow, querySelectQuestionWithAnswers, surveyID, orderNumber); err != nil {
		return nil, fmt.Errorf("%s: select question: %w", op, err)
	}

	return mapQuestionRecordToQuestion(questionRow)
}

func (r *surveyRepository) NextQuestionByOrder(ctx context.Context, surveyID string, currentOrder int) (nextID string, err error) {
	if err := r.db.GetContext(ctx, &nextID, querySelectNextOrderNum, surveyID, currentOrder); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", nil
		}
	}
	return nextID, err
}

func (r *surveyRepository) QuestionWithAnswer(ctx context.Context, qID, aID string) (*question.Question, *answer.Answer, error) {
	const op = "surveyRepository.QuestionWithAnswer"

	var res struct {
		Question question.Question `db:"q"`
		Answer   answer.Answer     `db:"a"`
	}

	err := r.db.GetContext(ctx, &res, querySelectQuestionWithAnswer, qID, aID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil, fmt.Errorf("%s: pair not found: %w", op, domain.ErrNotFound)
		}
		return nil, nil, fmt.Errorf("%s: execute query: %w", op, err)
	}

	return &res.Question, &res.Answer, nil
}
