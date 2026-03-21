package dto

import "sourcecraft.dev/benzo/testengine/internal/domain/models/answer"

type AnswerRecord struct {
	ID          string  `json:"id"`
	Text        string  `json:"text"`
	Weight      float64 `json:"weight"`
	CategoryTag string  `json:"category_tag"`
}

func (r AnswerRecord) ToDomain() answer.Answer {
	return answer.Answer{
		ID:          r.ID,
		Text:        r.Text,
		Weight:      r.Weight,
		CategoryTag: r.CategoryTag,
	}
}
