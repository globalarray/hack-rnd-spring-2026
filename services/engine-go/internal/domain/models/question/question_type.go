package question

type QuestionType int

var (
	UnspecifiedType  QuestionType = 0
	SingleChoiceType QuestionType = 1
	MultiChoiceType  QuestionType = 2
	ScaleType        QuestionType = 3
	TextType         QuestionType = 4
)

func (t QuestionType) String() string {
	switch t {
	case SingleChoiceType:
		return "SINGLE_CHOICE"
	case MultiChoiceType:
		return "MULTI_CHOICE"
	case ScaleType:
		return "SCALE"
	case TextType:
		return "TEXT"
	}
	return "UNSPECIFIED"
}
