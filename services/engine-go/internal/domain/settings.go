package domain

import (
	"encoding/json"

	"sourcecraft.dev/benzo/testengine/internal/domain/models"
)

func ParseSettings(settings string) (m models.SurveySettingMap, err error) {
	if err := json.Unmarshal([]byte(settings), &m); err != nil {
		return nil, err
	}

	return m, nil
}
