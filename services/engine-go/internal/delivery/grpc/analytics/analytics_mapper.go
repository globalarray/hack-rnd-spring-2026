package analytics

import (
	"sourcecraft.dev/benzo/testengine/internal/gen/pb"
	servicedto "sourcecraft.dev/benzo/testengine/internal/service/analytics/dto"
)

func mapGetSessionDataOutput(out *servicedto.GetSessionDataOutput) *pb.SessionAnalyticsResponse {
	responses := make([]*pb.RawClientResponse, len(out.Responses))
	for i, response := range out.Responses {
		responses[i] = &pb.RawClientResponse{
			QuestionId:     response.QuestionID,
			QuestionType:   pb.QuestionType(response.QuestionType),
			QuestionText:   response.QuestionText,
			SelectedWeight: response.SelectedWeight,
			CategoryTag:    response.CategoryTag,
			RawText:        response.RawText,
		}
	}

	return &pb.SessionAnalyticsResponse{
		SurveyId:           out.SurveyID,
		SessionId:          out.SessionID,
		ClientMetadataJson: out.ClientMetadataJSON,
		Responses:          responses,
	}
}
