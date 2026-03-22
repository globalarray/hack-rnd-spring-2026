package logic_rules

import (
	"encoding/json"
	"fmt"
	"strings"
)

type IterAlgorithmType string

const (
	LinearAlgorithm IterAlgorithmType = "linear"
)

const (
	JumpAction   = "JMP"
	FinishAction = "FINISH"
)

type LogicRuleRecord struct {
	Action string  `db:"action"`
	Next   *string `db:"next"`
}

type LogicRulesStorageRecord struct {
	Rules          map[string]LogicRuleRecord `json:"rules"`
	DefaultNextAlg IterAlgorithmType          `json:"default_next"`
}

type legacyLogicRuleRecord struct {
	AnswerID       string  `json:"answerId"`
	Action         string  `json:"action"`
	NextQuestionID *string `json:"nextQuestionId"`
}

func (r *LogicRulesStorageRecord) UnmarshalJSON(data []byte) error {
	type alias LogicRulesStorageRecord
	var direct alias
	if err := json.Unmarshal(data, &direct); err == nil && direct.Rules != nil {
		*r = LogicRulesStorageRecord(direct)
		return nil
	}

	var legacyRules []legacyLogicRuleRecord
	if err := json.Unmarshal(data, &legacyRules); err == nil {
		r.DefaultNextAlg = LinearAlgorithm
		rules, err := mapLegacyRules(legacyRules)
		if err != nil {
			return err
		}
		r.Rules = rules
		return nil
	}

	var payload struct {
		Rules          json.RawMessage   `json:"rules"`
		DefaultNextAlg IterAlgorithmType `json:"default_next"`
	}
	if err := json.Unmarshal(data, &payload); err != nil {
		return err
	}

	r.DefaultNextAlg = payload.DefaultNextAlg
	if len(payload.Rules) == 0 || string(payload.Rules) == "null" {
		r.Rules = map[string]LogicRuleRecord{}
		if r.DefaultNextAlg == "" {
			r.DefaultNextAlg = LinearAlgorithm
		}
		return nil
	}

	var mappedRules map[string]LogicRuleRecord
	if err := json.Unmarshal(payload.Rules, &mappedRules); err == nil {
		r.Rules = mappedRules
		if r.Rules == nil {
			r.Rules = map[string]LogicRuleRecord{}
		}
		if r.DefaultNextAlg == "" {
			r.DefaultNextAlg = LinearAlgorithm
		}
		return nil
	}

	if err := json.Unmarshal(payload.Rules, &legacyRules); err != nil {
		return fmt.Errorf("unsupported legacy rules payload: %w", err)
	}

	rules, err := mapLegacyRules(legacyRules)
	if err != nil {
		return err
	}
	r.Rules = rules
	if r.DefaultNextAlg == "" {
		r.DefaultNextAlg = LinearAlgorithm
	}
	return nil
}

func mapLegacyRules(legacyRules []legacyLogicRuleRecord) (map[string]LogicRuleRecord, error) {
	rules := make(map[string]LogicRuleRecord, len(legacyRules))

	for _, rule := range legacyRules {
		answerID := strings.TrimSpace(rule.AnswerID)
		if answerID == "" {
			continue
		}

		switch strings.ToLower(strings.TrimSpace(rule.Action)) {
		case "", string(LinearAlgorithm):
			continue
		case "jump", strings.ToLower(JumpAction):
			if rule.NextQuestionID == nil || strings.TrimSpace(*rule.NextQuestionID) == "" {
				return nil, fmt.Errorf("jump rule for '%s' missing nextQuestionId", answerID)
			}

			nextQuestionID := strings.TrimSpace(*rule.NextQuestionID)
			rules[answerID] = LogicRuleRecord{
				Action: JumpAction,
				Next:   &nextQuestionID,
			}
		case "finish", strings.ToLower(FinishAction):
			rules[answerID] = LogicRuleRecord{Action: FinishAction}
		default:
			return nil, fmt.Errorf("unknown legacy logic rule action %q for answer %s", rule.Action, answerID)
		}
	}

	return rules, nil
}
