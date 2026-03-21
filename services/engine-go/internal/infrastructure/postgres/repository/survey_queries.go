package repository

import _ "embed"

//go:embed sql/survey/insert_survey.sql
var queryInsertSurvey string
