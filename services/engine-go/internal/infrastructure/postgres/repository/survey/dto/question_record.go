package dto

import "github.com/google/uuid"

type QuestionRecord struct {
	ID          uuid.UUID `db:"id"`
	SurveyID    uuid.UUID `db:"survey_id"`
	OrderNumber int       `db:"order_num"`
	Type        int       `db:"type"`
	Text        string    `db:"text"`
	LogicRules  string    `db:"logic_rules"`
	AnswersJSON string    `db:"answers_json"`
}
