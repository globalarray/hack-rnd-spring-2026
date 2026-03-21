package dto

import "time"

type SessionRecord struct {
	ID               string
	CurrentRequestID string
	SurveyID         string
	Metadata         string
	Status           string
	StartedAt        time.Time
	CompletedAt      time.Time
}
