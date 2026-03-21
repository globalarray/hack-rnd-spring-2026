package dto

type SessionDataRecord struct {
	SurveyID           string `db:"survey_id"`
	SessionID          string `db:"session_id"`
	ClientMetadataJSON string `db:"client_metadata_json"`
}
