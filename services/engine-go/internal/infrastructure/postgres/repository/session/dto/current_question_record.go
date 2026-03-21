package dto

type CurrentQuestionRecord struct {
	ID          string `db:"id"`
	OrderNumber int    `db:"order_num"`
	Type        int    `db:"type"`
	Title       string `db:"text"`
	LogicRules  string `db:"logic_rules"`
	AnswersJSON string `db:"answers_json"`
}
