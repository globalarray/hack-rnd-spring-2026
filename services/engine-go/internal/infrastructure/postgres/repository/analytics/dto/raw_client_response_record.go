package dto

type RawClientResponseRecord struct {
	QuestionID     string  `db:"question_id"`
	QuestionType   int32   `db:"question_type"`
	QuestionText   string  `db:"question_text"`
	SelectedWeight float64 `db:"selected_weight"`
	CategoryTag    string  `db:"category_tag"`
	RawText        string  `db:"raw_text"`
}
