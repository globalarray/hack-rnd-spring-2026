package dto

type CreateSurveyInput struct {
	PsychologistID string
	Title          string
	Description    string
	Settings       string
	Questions      []QuestionInput
}

type QuestionInput struct {
	OrderNum   uint32
	Type       int
	Text       string
	LogicRules string
	Answers    []AnswerInput
}

type AnswerInput struct {
	ID          string
	Text        string
	Weight      float64
	CategoryTag string
}
