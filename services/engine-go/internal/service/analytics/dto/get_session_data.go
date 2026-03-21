package dto

type GetSessionDataInput struct {
	SessionID string
}

type RawClientResponse struct {
	QuestionID     string
	QuestionType   int32
	QuestionText   string
	SelectedWeight float64
	CategoryTag    string
	RawText        string
}

type GetSessionDataOutput struct {
	SurveyID           string
	SessionID          string
	ClientMetadataJSON string
	Responses          []RawClientResponse
}
