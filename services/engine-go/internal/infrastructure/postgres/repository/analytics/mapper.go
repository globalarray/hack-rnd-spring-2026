package analytics

import (
	"time"

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

func mapSurveySessions(records []repositorydto.SurveySessionRecord) *servicedto.ListSurveySessionsOutput {
	sessions := make([]servicedto.SurveySessionSummary, 0, len(records))

	for _, record := range records {
		completedAt := ""
		if record.CompletedAt.Valid {
			completedAt = record.CompletedAt.Time.Format(time.RFC3339)
		}

		sessions = append(sessions, servicedto.SurveySessionSummary{
			SurveyID:           record.SurveyID,
			SessionID:          record.SessionID,
			ClientMetadataJSON: record.ClientMetadataJSON,
			Status:             record.Status,
			ResponsesCount:     record.ResponsesCount,
			StartedAt:          record.StartedAt.Format(time.RFC3339),
			CompletedAt:        completedAt,
		})
	}

	return &servicedto.ListSurveySessionsOutput{Sessions: sessions}
}
