package session

import "time"

type Session struct {
	ID                string
	CurrentQuestionID string
	SurveyID          string
	Metadata          string
	Status            SessionStatus
	StartedAt         time.Time
	CompletedAt       time.Time
}
