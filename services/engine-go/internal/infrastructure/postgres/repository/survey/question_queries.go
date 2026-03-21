package survey

import _ "embed"

//go:embed sql/question/insert_question.sql
var queryInsertQuestion string

//go:embed sql/question/select_with_answers.sql
var querySelectQuestionWithAnswers string

//go:embed sql/question/select_next_order_num.sql
var querySelectNextOrderNum string

//go:embed sql/question/get_question_with_answer.sql
var querySelectQuestionWithAnswer string
