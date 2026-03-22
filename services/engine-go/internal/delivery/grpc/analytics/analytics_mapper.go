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

func mapListSurveySessionsOutput(out *servicedto.ListSurveySessionsOutput) *pb.ListSurveySessionsResponse {
	sessions := make([]*pb.SurveySessionSummary, 0, len(out.Sessions))

	for _, session := range out.Sessions {
		sessions = append(sessions, &pb.SurveySessionSummary{
			SurveyId:           session.SurveyID,
			SessionId:          session.SessionID,
			ClientMetadataJson: session.ClientMetadataJSON,
			Status:             session.Status,
			ResponsesCount:     session.ResponsesCount,
			StartedAt:          session.StartedAt,
			CompletedAt:        session.CompletedAt,
		})
	}

	return &pb.ListSurveySessionsResponse{Sessions: sessions}
}
