package question

type IterAnswersAlgorithm string

const (
	LinearIterAnswers IterAnswersAlgorithm = "linear"
)

// LogicRule определяет поведение при ответе
type LogicRule interface {
	Action() LogicAction
}

type JumpRule struct {
	NextQuestionID string
}

func (r JumpRule) Action() LogicAction { return ActionJump }

type FinishRule struct{}

func (r FinishRule) Action() LogicAction { return ActionFinish }
