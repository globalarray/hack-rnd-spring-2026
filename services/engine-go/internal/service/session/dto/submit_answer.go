package dto

type SubmitAnswerInput struct {
	SessionID  string `json:"session_id" validate:"required,uuid4"`
	QuestionID string `json:"question_id" validate:"required,uuid4"`

	AnswerID  string   `json:"answer_id,omitempty"`
	RawText   string   `json:"raw_text,omitempty"`
	AnswerIDs []string `json:"answer_ids,omitempty"`
}

type SubmitAnswerOutput struct {
	NextQuestionID string `json:"next_question_id,omitempty"`
	IsFinished     bool   `json:"is_finished"`
}
