package grpc

import (
	"sourcecraft.dev/benzo/testengine/internal/gen/pb"
	"sourcecraft.dev/benzo/testengine/internal/service/survey/dto"
)

func mapCreateSurveyRequestToInput(req *pb.CreateSurveyRequest) *dto.CreateSurveyInput {
	questions := make([]dto.QuestionInput, len(req.Questions))

	for i, q := range req.Questions {
		answers := make([]dto.AnswerInput, len(q.Answers))

		for j, a := range q.Answers {
			answers[j] = dto.AnswerInput{
				Text:        a.Text,
				Weight:      a.Weight,
				CategoryTag: a.CategoryTag,
			}
		}

		questions[i] = dto.QuestionInput{
			OrderNum:   q.OrderNum,
			Type:       int(q.Type),
			Text:       q.Text,
			LogicRules: q.LogicRulesJson,
		}
	}

	return &dto.CreateSurveyInput{
		PsychologistID: req.PsychologistId,
		Title:          req.Title,
		Description:    req.Description,
		Settings:       req.SettingsJson,
	}
}
