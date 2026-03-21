package grpc

import (
	"sourcecraft.dev/benzo/testengine/internal/domain/models/answer"
	"sourcecraft.dev/benzo/testengine/internal/domain/models/question"
	"sourcecraft.dev/benzo/testengine/internal/gen/pb"
	"sourcecraft.dev/benzo/testengine/internal/service/session/dto"
)

func mapStartSessionRequestToStartSessionInput(req *pb.StartSessionRequest) *dto.StartSessionInput {
	return &dto.StartSessionInput{
		SurveyID:          req.SurveyId,
		ClientMetadataRaw: req.ClientMetadataJson,
	}
}

func mapAnswerToAnswerClientView(a answer.Answer) *pb.AnswerClientView {
	return &pb.AnswerClientView{
		AnswerId: a.ID,
		Text:     a.Text,
	}
}

func mapQuestionToQuestionClientView(q question.Question) *pb.QuestionClientView {
	var answers = make([]*pb.AnswerClientView, len(q.Answers))

	for i, a := range q.Answers {
		answers[i] = mapAnswerToAnswerClientView(a)
	}

	return &pb.QuestionClientView{
		QuestionId: q.ID,
		Type:       pb.QuestionType(q.Type),
		Text:       q.Title,
		Answers:    answers,
	}
}

func mapDomainToStartSessionResponse(q question.Question) *pb.StartSessionResponse {

	return &pb.StartSessionResponse{
		SessionId:     q.ID,
		FirstQuestion: mapQuestionToQuestionClientView(q),
	}
}
