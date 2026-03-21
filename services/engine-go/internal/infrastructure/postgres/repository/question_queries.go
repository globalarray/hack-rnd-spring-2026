package repository

import _ "embed"

//go:embed sql/question/insert_question.sql
var queryInsertQuestion string

//go:embed sql/question/select_with_answers.sql
var querySelectQuestionWithAnswers string
