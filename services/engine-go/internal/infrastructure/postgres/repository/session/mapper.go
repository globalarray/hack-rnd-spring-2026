package session

import (
	"encoding/json"
	"fmt"

	"sourcecraft.dev/benzo/testengine/internal/domain/models/answer"
	"sourcecraft.dev/benzo/testengine/internal/domain/models/question"
	"sourcecraft.dev/benzo/testengine/internal/domain/models/session"
	"sourcecraft.dev/benzo/testengine/internal/domain/models/survey"
	"sourcecraft.dev/benzo/testengine/internal/infrastructure/postgres/repository/session/dto"
	surveydto "sourcecraft.dev/benzo/testengine/internal/infrastructure/postgres/repository/survey/dto"
	questiondto "sourcecraft.dev/benzo/testengine/internal/infrastructure/postgres/repository/survey/dto/logic_rules"
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

func mapCurrentQuestionRecordToDomain(record *dto.CurrentQuestionRecord) (*question.Question, error) {
	var storage questiondto.LogicRulesStorageRecord
	var answerRecords []surveydto.AnswerRecord

	if err := json.Unmarshal([]byte(record.LogicRules), &storage); err != nil {
		return nil, fmt.Errorf("cannot unmarshal logic rules: %w", err)
	}
	if err := json.Unmarshal([]byte(record.AnswersJSON), &answerRecords); err != nil {
		return nil, fmt.Errorf("cannot unmarshal answers: %w", err)
	}

	domainRules := make(map[string]question.LogicRule, len(storage.Rules))
	defaultNext := question.LinearIterAnswers

	for cond, rule := range storage.Rules {
		switch rule.Action {
		case questiondto.FinishAction:
			domainRules[cond] = question.FinishRule{}
		case questiondto.JumpAction:
			if rule.Next == nil {
				return nil, fmt.Errorf("jump rule for '%s' missing 'next' field", cond)
			}
			domainRules[cond] = question.JumpRule{NextQuestionID: *rule.Next}
		}
	}

	answers := make([]answer.Answer, len(answerRecords))
	for i, a := range answerRecords {
		answers[i] = a.ToDomain()
	}

	return &question.Question{
		ID:          record.ID,
		Type:        record.Type,
		OrderNumber: record.OrderNumber,
		Title:       record.Title,
		LogicRules:  domainRules,
		DefaultNext: defaultNext,
		Answers:     answers,
	}, nil
}
