package dto

import "github.com/google/uuid"

type QuestionRecord struct {
	ID          uuid.UUID `db:"id"`
	SurveyID    uuid.UUID `db:"survey_id"`
	OrderNumber int       `db:"order_number"`
	Type        string    `db:"type"`
	Text        string    `db:"text"`
	LogicRules  string    `db:"logic_rules"`
}
