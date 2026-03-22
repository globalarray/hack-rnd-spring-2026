package domain

type SurveyDraft struct {
	PsychologistID string
	Title          string
	Description    string
	Settings       map[string]any
	Questions      []SurveyQuestionDraft
}

type SurveyQuestionDraft struct {
	OrderNum   uint32
	Type       QuestionType
	Text       string
	LogicRules map[string]any
	Answers    []SurveyAnswerDraft
}

type SurveyAnswerDraft struct {
	ID          string
	Text        string
	Weight      float64
	CategoryTag string
}

type SurveySummary struct {
	SurveyID         string
	Title            string
	CompletionsCount int32
}
