package survey

import (
	"fmt"

	"sourcecraft.dev/benzo/testengine/internal/gen/pb"
	"sourcecraft.dev/benzo/testengine/internal/service/survey/dto"
)

func mapCreateSurveyRequestToInput(req *pb.CreateSurveyRequest) (*dto.CreateSurveyInput, error) {
	settingsStr := "{}"
	if req.SettingsJson != nil {
		b, err := req.SettingsJson.MarshalJSON()
		if err != nil {
			return nil, fmt.Errorf("marshal settings: %w", err)
		}

		settingsStr = string(b)
	}

	questions := make([]dto.QuestionInput, len(req.Questions))

	for i, q := range req.Questions {
		logicRulesStr := "{}"
		if q.LogicRulesJson != nil {
			b, err := q.LogicRulesJson.MarshalJSON()
			if err != nil {
				return nil, fmt.Errorf("marshal logic rules for question %d: %w", i, err)
			}
			logicRulesStr = string(b)
		}

		answers := make([]dto.AnswerInput, len(q.Answers))
		for j, a := range q.Answers {
			answers[j] = dto.AnswerInput{
				ID:          a.Id,
				Text:        a.Text,
				Weight:      a.Weight,
				CategoryTag: a.CategoryTag,
			}
		}

		questions[i] = dto.QuestionInput{
			OrderNum:   q.OrderNum,
			Type:       int(q.Type),
			Text:       q.Text,
			LogicRules: logicRulesStr,
			Answers:    answers,
		}
	}

	return &dto.CreateSurveyInput{
		PsychologistID: req.PsychologistId,
		Title:          req.Title,
		Description:    req.Description,
		Settings:       settingsStr,
		Questions:      questions,
	}, nil
}

func mapListSurveysOutput(out *dto.ListSurveysOutput) *pb.ListSurveysResponse {
	surveys := make([]*pb.SurveySummary, len(out.Surveys))
	for i, survey := range out.Surveys {
		surveys[i] = &pb.SurveySummary{
			SurveyId:         survey.SurveyID,
			Title:            survey.Title,
			CompletionsCount: survey.CompletionsCount,
		}
	}

	return &pb.ListSurveysResponse{Surveys: surveys}
}
