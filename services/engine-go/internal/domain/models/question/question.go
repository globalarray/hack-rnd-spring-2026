package question

import "sourcecraft.dev/benzo/testengine/internal/domain/models/answer"

type Question struct {
	ID          string
	Type        int
	OrderNumber int
	Title       string
	LogicRules  map[string]LogicRule
	DefaultNext IterAnswersAlgorithm
	Answers     []answer.Answer
}
