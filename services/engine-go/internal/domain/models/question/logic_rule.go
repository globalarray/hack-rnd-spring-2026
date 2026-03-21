package question

type IterAnswersAlgorithm string

const (
	LinearIterAnswers IterAnswersAlgorithm = "linear"
)

// LogicRule определяет поведение при ответе
type LogicRule interface {
	GetAction() string
}

type JumpRule struct {
	NextQuestionID string // или OrderNumber, смотря что хранишь в БД
}

func (r JumpRule) GetAction() string { return "JMP" }

type FinishRule struct{}

func (r FinishRule) GetAction() string { return "FINISH" }
