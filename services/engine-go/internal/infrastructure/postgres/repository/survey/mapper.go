package survey

import (
	"encoding/json"
	"fmt"

	"sourcecraft.dev/benzo/testengine/internal/domain/models/question"
	"sourcecraft.dev/benzo/testengine/internal/infrastructure/postgres/repository/survey/dto"
	"sourcecraft.dev/benzo/testengine/internal/infrastructure/postgres/repository/survey/dto/logic_rules"
)

func mapQuestionRecordToQuestion(record dto.QuestionRecord) (*question.Question, error) {
	var storage logic_rules.LogicRulesStorageRecord

	if err := json.Unmarshal([]byte(record.LogicRules), &storage); err != nil {
		return nil, fmt.Errorf("cannot unmarshal logicRulesStorage: %w", err)
	}

	domainRules := make(map[string]question.LogicRule, len(storage.Rules))

	var domainAlgType question.IterAnswersAlgorithm

	switch storage.DefaultNextAlg {
	case logic_rules.LinearAlgorithm:
	default:
		domainAlgType = question.LinearIterAnswers
		break
	}

	for cond, rule := range storage.Rules {
		switch rule.Action {
		case logic_rules.FinishAction:
			domainRules[cond] = question.FinishRule{}

		case logic_rules.JumpAction:
			if rule.Next == nil {
				return nil, fmt.Errorf("jump rule for '%s' missing 'next' field", cond)
			}
			domainRules[cond] = question.JumpRule{
				NextQuestionID: *rule.Next,
			}
		}
	}

	return &question.Question{
		OrderNumber: record.OrderNumber,
		Title:       record.Text,
		LogicRules:  domainRules,
		DefaultNext: domainAlgType,
	}, nil
}
