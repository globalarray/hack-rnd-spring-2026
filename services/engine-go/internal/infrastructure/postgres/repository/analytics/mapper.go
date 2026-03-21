package analytics

import (
	repositorydto "sourcecraft.dev/benzo/testengine/internal/infrastructure/postgres/repository/analytics/dto"
	servicedto "sourcecraft.dev/benzo/testengine/internal/service/analytics/dto"
)

func mapSessionData(record *repositorydto.SessionDataRecord, responses []repositorydto.RawClientResponseRecord) *servicedto.GetSessionDataOutput {
	out := &servicedto.GetSessionDataOutput{
		SurveyID:           record.SurveyID,
		SessionID:          record.SessionID,
		ClientMetadataJSON: record.ClientMetadataJSON,
		Responses:          make([]servicedto.RawClientResponse, len(responses)),
	}

	for i, response := range responses {
		out.Responses[i] = servicedto.RawClientResponse{
			QuestionID:     response.QuestionID,
			QuestionType:   response.QuestionType,
			QuestionText:   response.QuestionText,
			SelectedWeight: response.SelectedWeight,
			CategoryTag:    response.CategoryTag,
			RawText:        response.RawText,
		}
	}

	return out
}
