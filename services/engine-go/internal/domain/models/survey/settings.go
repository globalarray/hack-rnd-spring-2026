package survey

import (
	"encoding/json"
	"fmt"
)

func ParseSettings(settings string) (SurveySettingMap, error) {
	if settings == "" {
		return make(SurveySettingMap), nil
	}

	var m SurveySettingMap
	if err := json.Unmarshal([]byte(settings), &m); err != nil {
		return nil, fmt.Errorf("failed to unmarshal survey settings: %w", err)
	}

	return m, nil
}
