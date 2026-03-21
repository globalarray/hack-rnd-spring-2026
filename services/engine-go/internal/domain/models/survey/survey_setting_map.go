package survey

import (
	"fmt"
	"regexp"
)

var hexRegex = regexp.MustCompile(`^#([A-Fa-f0-9]{6}|[A-Fa-f0-9]{3})$`)
var validThemes = map[string]struct{}{"light": {}, "dark": {}, "cyberpunk": {}}

type SurveySettingMap map[SurveyGroup]map[string]any

func NewSurveySettingMap() SurveySettingMap {
	return make(SurveySettingMap)
}

func (s SurveySettingMap) GetString(group SurveyGroup, setting string, defaultValue string) string {
	if v, ok := s[group]; ok {
		if v, ok := v[setting].(string); ok {
			return v
		}
	}

	return defaultValue
}

func (s SurveySettingMap) GetFloat64(group SurveyGroup, setting string, defaultValue float64) float64 {
	if v, ok := s[group]; ok {
		if v, ok := v[setting].(float64); ok {
			return v
		}
	}

	return defaultValue
}

func (s SurveySettingMap) Validate() error {
	if err := s.validateLimits(); err != nil {
		return err
	}

	if err := s.validateUI(); err != nil {
		return err
	}

	return nil
}

func (s SurveySettingMap) validateLimits() error {
	group, ok := s[LimitsGroup]
	if !ok {
		return nil
	}

	if v, ok := group[LimitTimeLimit].(float64); ok {
		if v < 0 || v > 86400 {
			return fmt.Errorf("%s.%s must be between 0 and 86400 seconds", LimitsGroup, LimitTimeLimit)
		}
	}
	return nil
}

func (s SurveySettingMap) validateUI() error {
	group, ok := s[UIGroup]
	if !ok {
		return nil
	}

	if theme, ok := group[UITheme].(string); ok {
		if _, ok := validThemes[theme]; !ok {
			return fmt.Errorf("invalid theme: %s", theme)
		}
	}

	if color, ok := group[UIAccentColor].(string); ok {
		if !hexRegex.MatchString(color) {
			return fmt.Errorf("invalid hex color format: %s", color)
		}
	}

	return nil
}
