package survey

import _ "embed"

//go:embed sql/answer/insert_answer.sql
var queryInsertAnswer string
