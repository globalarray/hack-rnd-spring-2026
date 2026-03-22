package dto

import "sourcecraft.dev/benzo/testengine/internal/domain/models/question"

type StartSessionInput struct {
	SurveyID          string
	ClientMetadataRaw string
	ShareLinkID       string
}

type StartSessionOutput struct {
	SessionID     string
	FirstQuestion *question.Question
}
