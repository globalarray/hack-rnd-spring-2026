package survey

import _ "embed"

//go:embed sql/survey/insert_survey.sql
var queryInsertSurvey string

//go:embed sql/survey/list_surveys.sql
var queryListSurveys string
