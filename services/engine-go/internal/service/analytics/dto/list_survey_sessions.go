package dto

type ListSurveySessionsInput struct {
	SurveyID string
}

type SurveySessionSummary struct {
	SurveyID           string
	SessionID          string
	ClientMetadataJSON string
	Status             string
	ResponsesCount     int32
	StartedAt          string
	CompletedAt        string
}

type ListSurveySessionsOutput struct {
	Sessions []SurveySessionSummary
}
