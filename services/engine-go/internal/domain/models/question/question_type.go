package question

type QuestionType int

var (
	UnspecifiedType  QuestionType = 0
	SingleChoiceType QuestionType = 1
	MultiChoiceType  QuestionType = 2
	ScaleType        QuestionType = 3
	TextType         QuestionType = 4
)
