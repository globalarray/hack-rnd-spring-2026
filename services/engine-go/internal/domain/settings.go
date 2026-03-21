package domain

import (
	"encoding/json"

	"sourcecraft.dev/benzo/testengine/internal/domain/models/survey"
)

func ParseSettings(settings string) (m survey.SurveySettingMap, err error) {
	if err := json.Unmarshal([]byte(settings), &m); err != nil {
		return nil, err
	}

	return m, nil
}
