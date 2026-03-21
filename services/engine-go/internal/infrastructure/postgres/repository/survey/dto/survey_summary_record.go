package dto

type SurveySummaryRecord struct {
	SurveyID         string `db:"survey_id"`
	Title            string `db:"title"`
	CompletionsCount int32  `db:"completions_count"`
}
