package engine

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/structpb"
	"sourcecraft.dev/benzo/bff/internal/application/ports"
	enginepb "sourcecraft.dev/benzo/bff/internal/clients/grpc/testenginepb"
	"sourcecraft.dev/benzo/bff/internal/domain"
)

const (
	createSurveyMethod       = "/testengine.v1.SurveyAdminService/CreateSurvey"
	listSurveysMethod        = "/testengine.v1.SurveyAdminService/ListSurveys"
	startSessionMethod       = "/testengine.v1.SessionClientService/StartSession"
	currentQuestionMethod    = "/testengine.v1.SessionClientService/GetCurrentQuestion"
	submitAnswerMethod       = "/testengine.v1.SessionClientService/SubmitAnswer"
	sessionAnalyticsMethod   = "/testengine.v1.AnalyticsService/GetSessionDataForAnalytics"
	listSurveySessionsMethod = "/testengine.v1.AnalyticsService/ListSurveySessionsForAnalytics"
)

type Client struct {
	conn grpc.ClientConnInterface
}

func NewClient(conn grpc.ClientConnInterface) *Client {
	return &Client{conn: conn}
}

func (c *Client) CreateSurvey(ctx context.Context, draft domain.SurveyDraft) (string, error) {
	settings, err := structFromMap(draft.Settings)
	if err != nil {
		return "", err
	}

	questions := make([]*enginepb.QuestionPayload, 0, len(draft.Questions))
	for _, question := range draft.Questions {
		logicRules, err := structFromMap(question.LogicRules)
		if err != nil {
			return "", err
		}

		answers := make([]*enginepb.AnswerPayload, 0, len(question.Answers))
		for _, answer := range question.Answers {
			answers = append(answers, &enginepb.AnswerPayload{
				Id:          answer.ID,
				Text:        answer.Text,
				Weight:      answer.Weight,
				CategoryTag: answer.CategoryTag,
			})
		}

		questions = append(questions, &enginepb.QuestionPayload{
			OrderNum:       question.OrderNum,
			Type:           mapQuestionTypeToProto(question.Type),
			Text:           question.Text,
			LogicRulesJson: logicRules,
			Answers:        answers,
		})
	}

	req := &enginepb.CreateSurveyRequest{
		PsychologistId: draft.PsychologistID,
		Title:          draft.Title,
		Description:    draft.Description,
		SettingsJson:   settings,
		Questions:      questions,
	}

	resp := &enginepb.CreateSurveyResponse{}
	if err := c.conn.Invoke(ctx, createSurveyMethod, req, resp); err != nil {
		return "", err
	}

	surveyID := strings.TrimSpace(resp.GetSurveyId())
	if _, err := uuid.Parse(surveyID); err != nil {
		return "", fmt.Errorf("%w: engine create survey response has invalid surveyId", domain.ErrUpstreamResponse)
	}

	return surveyID, nil
}

func (c *Client) ListSurveys(ctx context.Context, psychologistID string) ([]domain.SurveySummary, error) {
	req := &enginepb.ListSurveysRequest{PsychologistId: psychologistID}
	resp := &enginepb.ListSurveysResponse{}
	if err := c.conn.Invoke(ctx, listSurveysMethod, req, resp); err != nil {
		return nil, err
	}

	surveys := make([]domain.SurveySummary, 0, len(resp.GetSurveys()))
	for _, survey := range resp.GetSurveys() {
		surveyID := strings.TrimSpace(survey.GetSurveyId())
		if _, err := uuid.Parse(surveyID); err != nil {
			return nil, fmt.Errorf("%w: engine list surveys response has invalid surveyId", domain.ErrUpstreamResponse)
		}

		surveys = append(surveys, domain.SurveySummary{
			SurveyID:         surveyID,
			Title:            survey.GetTitle(),
			CompletionsCount: survey.GetCompletionsCount(),
		})
	}

	return surveys, nil
}

func (c *Client) StartSession(ctx context.Context, surveyID, clientMetadataJSON string) (string, *domain.Question, error) {
	req := &enginepb.StartSessionRequest{
		SurveyId:           surveyID,
		ClientMetadataJson: clientMetadataJSON,
	}
	resp := &enginepb.StartSessionResponse{}
	if err := c.conn.Invoke(ctx, startSessionMethod, req, resp); err != nil {
		return "", nil, err
	}

	sessionID := strings.TrimSpace(resp.GetSessionId())
	if _, err := uuid.Parse(sessionID); err != nil {
		return "", nil, fmt.Errorf("%w: engine start session response has invalid sessionId", domain.ErrUpstreamResponse)
	}

	question, err := mapQuestionFromProto(resp.GetFirstQuestion())
	if err != nil {
		return "", nil, err
	}

	if question == nil {
		return "", nil, fmt.Errorf("%w: engine start session response is missing firstQuestion", domain.ErrUpstreamResponse)
	}

	return sessionID, question, nil
}

func (c *Client) GetCurrentQuestion(ctx context.Context, sessionID string) (*domain.Question, error) {
	req := &enginepb.GetSessionCurrentQuestionRequest{SessionId: sessionID}
	resp := &enginepb.QuestionClientView{}
	if err := c.conn.Invoke(ctx, currentQuestionMethod, req, resp); err != nil {
		return nil, err
	}

	return mapQuestionFromProto(resp)
}

func (c *Client) SubmitAnswer(ctx context.Context, input ports.SubmitAnswerInput) (string, bool, error) {
	req := &enginepb.SubmitAnswerRequest{
		SessionId:  input.SessionID,
		QuestionId: input.QuestionID,
	}

	switch {
	case input.AnswerID != "":
		req.Payload = &enginepb.SubmitAnswerRequest_AnswerId{AnswerId: input.AnswerID}
	case input.RawText != "":
		req.Payload = &enginepb.SubmitAnswerRequest_RawText{RawText: input.RawText}
	case len(input.AnswerIDs) > 0:
		req.Payload = &enginepb.SubmitAnswerRequest_MultipleChoice{
			MultipleChoice: &enginepb.SelectedAnswers{AnswerIds: input.AnswerIDs},
		}
	default:
		return "", false, fmt.Errorf("%w: missing answer payload", domain.ErrInvalidInput)
	}

	resp := &enginepb.SubmitAnswerResponse{}
	if err := c.conn.Invoke(ctx, submitAnswerMethod, req, resp); err != nil {
		return "", false, err
	}

	return resp.GetNextQuestionId(), resp.GetIsFinished(), nil
}

func (c *Client) GetSessionAnalytics(ctx context.Context, sessionID string) (*domain.SessionAnalytics, error) {
	req := &enginepb.GetSessionDataRequest{SessionId: sessionID}
	resp := &enginepb.SessionAnalyticsResponse{}
	if err := c.conn.Invoke(ctx, sessionAnalyticsMethod, req, resp); err != nil {
		return nil, err
	}

	clientMetadata, err := domain.ParseClientMetadataJSON(resp.GetClientMetadataJson())
	if err != nil {
		return nil, fmt.Errorf("%w: engine analytics response contains invalid clientMetadataJson", domain.ErrUpstreamResponse)
	}

	responses := make([]domain.AnalyticsResponse, 0, len(resp.GetResponses()))
	for _, response := range resp.GetResponses() {
		questionID := strings.TrimSpace(response.GetQuestionId())
		if _, err := uuid.Parse(questionID); err != nil {
			return nil, fmt.Errorf("%w: engine analytics response contains invalid questionId", domain.ErrUpstreamResponse)
		}

		responses = append(responses, domain.AnalyticsResponse{
			QuestionID:     questionID,
			QuestionType:   mapQuestionTypeFromProto(response.GetQuestionType()),
			QuestionText:   response.GetQuestionText(),
			SelectedWeight: response.GetSelectedWeight(),
			CategoryTag:    response.GetCategoryTag(),
			RawText:        response.GetRawText(),
		})
	}

	surveyID := strings.TrimSpace(resp.GetSurveyId())
	if _, err := uuid.Parse(surveyID); err != nil {
		return nil, fmt.Errorf("%w: engine analytics response has invalid surveyId", domain.ErrUpstreamResponse)
	}

	respSessionID := strings.TrimSpace(resp.GetSessionId())
	if _, err := uuid.Parse(respSessionID); err != nil {
		return nil, fmt.Errorf("%w: engine analytics response has invalid sessionId", domain.ErrUpstreamResponse)
	}

	return &domain.SessionAnalytics{
		SurveyID:       surveyID,
		SessionID:      respSessionID,
		ClientMetadata: clientMetadata,
		Responses:      responses,
	}, nil
}

func (c *Client) ListSurveySessions(ctx context.Context, surveyID string) ([]domain.SurveySessionSummary, error) {
	req := &enginepb.ListSurveySessionsRequest{SurveyId: surveyID}
	resp := &enginepb.ListSurveySessionsResponse{}
	if err := c.conn.Invoke(ctx, listSurveySessionsMethod, req, resp); err != nil {
		return nil, err
	}

	sessions := make([]domain.SurveySessionSummary, 0, len(resp.GetSessions()))
	for _, session := range resp.GetSessions() {
		respSurveyID := strings.TrimSpace(session.GetSurveyId())
		if _, err := uuid.Parse(respSurveyID); err != nil {
			return nil, fmt.Errorf("%w: engine survey sessions response has invalid surveyId", domain.ErrUpstreamResponse)
		}

		respSessionID := strings.TrimSpace(session.GetSessionId())
		if _, err := uuid.Parse(respSessionID); err != nil {
			return nil, fmt.Errorf("%w: engine survey sessions response has invalid sessionId", domain.ErrUpstreamResponse)
		}

		clientMetadata, err := domain.ParseClientMetadataJSON(session.GetClientMetadataJson())
		if err != nil {
			return nil, fmt.Errorf("%w: engine survey sessions response contains invalid clientMetadataJson", domain.ErrUpstreamResponse)
		}

		sessions = append(sessions, domain.SurveySessionSummary{
			SurveyID:       respSurveyID,
			SessionID:      respSessionID,
			ClientMetadata: clientMetadata,
			Status:         strings.ToLower(strings.TrimSpace(session.GetStatus())),
			ResponsesCount: session.GetResponsesCount(),
			StartedAt:      session.GetStartedAt(),
			FinishedAt:     session.GetCompletedAt(),
		})
	}

	return sessions, nil
}

func structFromMap(values map[string]any) (*structpb.Struct, error) {
	if len(values) == 0 {
		return nil, nil
	}

	payload, err := structpb.NewStruct(values)
	if err != nil {
		return nil, fmt.Errorf("%w: invalid JSON object for protobuf struct", domain.ErrInvalidInput)
	}

	return payload, nil
}

func mapQuestionFromProto(question *enginepb.QuestionClientView) (*domain.Question, error) {
	if question == nil {
		return nil, nil
	}

	questionID := strings.TrimSpace(question.GetQuestionId())
	if _, err := uuid.Parse(questionID); err != nil {
		return nil, fmt.Errorf("%w: engine question response has invalid questionId", domain.ErrUpstreamResponse)
	}

	answers := make([]domain.AnswerOption, 0, len(question.GetAnswers()))
	for _, answer := range question.GetAnswers() {
		answerID := strings.TrimSpace(answer.GetAnswerId())
		if _, err := uuid.Parse(answerID); err != nil {
			return nil, fmt.Errorf("%w: engine question response has invalid answerId", domain.ErrUpstreamResponse)
		}

		answers = append(answers, domain.AnswerOption{
			ID:   answerID,
			Text: answer.GetText(),
		})
	}

	return &domain.Question{
		ID:      questionID,
		Type:    mapQuestionTypeFromProto(question.GetType()),
		Text:    question.GetText(),
		Answers: answers,
	}, nil
}

func mapQuestionTypeToProto(questionType domain.QuestionType) enginepb.QuestionType {
	switch questionType {
	case domain.QuestionTypeSingleChoice:
		return enginepb.QuestionType_QUESTION_TYPE_SINGLE_CHOICE
	case domain.QuestionTypeMultipleChoice:
		return enginepb.QuestionType_QUESTION_TYPE_MULTIPLE_CHOICE
	case domain.QuestionTypeScale:
		return enginepb.QuestionType_QUESTION_TYPE_SCALE
	case domain.QuestionTypeText:
		return enginepb.QuestionType_QUESTION_TYPE_TEXT
	default:
		return enginepb.QuestionType_QUESTION_TYPE_UNSPECIFIED
	}
}

func mapQuestionTypeFromProto(questionType enginepb.QuestionType) domain.QuestionType {
	switch questionType {
	case enginepb.QuestionType_QUESTION_TYPE_SINGLE_CHOICE:
		return domain.QuestionTypeSingleChoice
	case enginepb.QuestionType_QUESTION_TYPE_MULTIPLE_CHOICE:
		return domain.QuestionTypeMultipleChoice
	case enginepb.QuestionType_QUESTION_TYPE_SCALE:
		return domain.QuestionTypeScale
	case enginepb.QuestionType_QUESTION_TYPE_TEXT:
		return domain.QuestionTypeText
	default:
		return domain.QuestionTypeText
	}
}
