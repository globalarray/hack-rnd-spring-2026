package session

import "sourcecraft.dev/benzo/testengine/internal/domain/models/survey"

type State struct {
	Session
	SettingsSurvey survey.SurveySettingMap
}
