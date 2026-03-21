package repository

import _ "embed"

//go:embed sql/question/insert_question.sql
var queryInsertQuestion string
