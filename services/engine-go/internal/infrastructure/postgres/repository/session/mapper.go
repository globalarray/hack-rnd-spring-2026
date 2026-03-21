package session

import (
	"fmt"

	"sourcecraft.dev/benzo/testengine/internal/domain/models/session"
	"sourcecraft.dev/benzo/testengine/internal/domain/models/survey"
	"sourcecraft.dev/benzo/testengine/internal/infrastructure/postgres/repository/session/dto"
)

func mapSessionStateRecordToDomain(record *dto.SessionStateRecord) (*session.State, error) {
	settings, err := survey.ParseSettings(record.SettingsSurvey)

	if err != nil {
		return nil, fmt.Errorf(`invalid settings: %w`, err)
	}

	return &session.State{
		Session:        *mapSessionRecordToDomain(&record.SessionRecord),
		SettingsSurvey: settings,
	}, nil
}

func mapSessionRecordToDomain(record *dto.SessionRecord) *session.Session {
	return &session.Session{
		ID:                record.ID,
		CurrentQuestionID: record.CurrentRequestID,
		SurveyID:          record.SurveyID,
		Metadata:          record.Metadata,
		Status:            session.SessionStatus(record.Status),
		StartedAt:         record.StartedAt,
		CompletedAt:       record.CompletedAt,
	}
}
