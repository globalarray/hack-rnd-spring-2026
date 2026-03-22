package dto

import (
	"database/sql"
	"time"
)

type SurveySessionRecord struct {
	SurveyID           string       `db:"survey_id"`
	SessionID          string       `db:"session_id"`
	ClientMetadataJSON string       `db:"client_metadata_json"`
	Status             string       `db:"status"`
	ResponsesCount     int32        `db:"responses_count"`
	StartedAt          time.Time    `db:"started_at"`
	CompletedAt        sql.NullTime `db:"completed_at"`
}
