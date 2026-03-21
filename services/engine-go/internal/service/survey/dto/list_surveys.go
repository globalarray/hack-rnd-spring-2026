package dto

type ListSurveysInput struct {
	PsychologistID string
}

type SurveySummary struct {
	SurveyID         string
	Title            string
	CompletionsCount int32
}

type ListSurveysOutput struct {
	Surveys []SurveySummary
}
