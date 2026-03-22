package dto

import (
	"database/sql"
	"time"
)

type SessionRecord struct {
	ID               string       `db:"id"`
	CurrentRequestID string       `db:"current_question_id"`
	SurveyID         string       `db:"survey_id"`
	Metadata         string       `db:"client_metadata"`
	Status           string       `db:"status"`
	StartedAt        time.Time    `db:"started_at"`
	CompletedAt      sql.NullTime `db:"completed_at"`
}
