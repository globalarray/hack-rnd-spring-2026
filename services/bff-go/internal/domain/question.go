package domain

import (
	"fmt"
	"strings"
)

type QuestionType string

const (
	QuestionTypeSingleChoice   QuestionType = "single_choice"
	QuestionTypeMultipleChoice QuestionType = "multiple_choice"
	QuestionTypeScale          QuestionType = "scale"
	QuestionTypeText           QuestionType = "text"
)

type Question struct {
	ID      string
	Type    QuestionType
	Text    string
	Answers []AnswerOption
}

type AnswerOption struct {
	ID   string
	Text string
}

func ParseQuestionType(value string) (QuestionType, error) {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "single_choice", "single-choice", "singlechoice":
		return QuestionTypeSingleChoice, nil
	case "multiple_choice", "multiple-choice", "multiplechoice":
		return QuestionTypeMultipleChoice, nil
	case "scale":
		return QuestionTypeScale, nil
	case "text":
		return QuestionTypeText, nil
	default:
		return "", fmt.Errorf("%w: unsupported question type %q", ErrInvalidInput, value)
	}
}
