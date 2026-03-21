package domain

import (
	"encoding/json"
	"fmt"

	"sourcecraft.dev/benzo/testengine/internal/domain/models/survey"
)

func ParseSettings(settings string) (survey.SurveySettingMap, error) {
	if settings == "" {
		return make(survey.SurveySettingMap), nil
	}

	var m survey.SurveySettingMap
	if err := json.Unmarshal([]byte(settings), &m); err != nil {
		return nil, fmt.Errorf("failed to unmarshal survey settings: %w", err)
	}

	return m, nil
}
