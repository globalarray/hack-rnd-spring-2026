package dto

type QuestionWithAnswerRecord struct {
	ID          string  `db:"id"`
	OrderNumber int     `db:"order_num"`
	Type        int     `db:"type"`
	Text        string  `db:"text"`
	LogicRules  string  `db:"logic_rules"`
	AnswerID    string  `db:"answer_id"`
	AnswerText  string  `db:"answer_text"`
	Weight      float64 `db:"weight"`
	CategoryTag string  `db:"category_tag"`
}
