package httpapi

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"google.golang.org/grpc/codes"
	grpcstatus "google.golang.org/grpc/status"
	"sourcecraft.dev/benzo/bff/internal/application/ports"
	"sourcecraft.dev/benzo/bff/internal/application/usecase"
	"sourcecraft.dev/benzo/bff/internal/domain"
)

type Handler struct {
	log      *slog.Logger
	surveys  *usecase.SurveyUseCase
	sessions *usecase.SessionUseCase
	auth     *usecase.AuthUseCase
}

func NewRouter(log *slog.Logger, surveys *usecase.SurveyUseCase, sessions *usecase.SessionUseCase, authUseCase *usecase.AuthUseCase) http.Handler {
	handler := &Handler{
		log:      log,
		surveys:  surveys,
		sessions: sessions,
		auth:     authUseCase,
	}

	router := chi.NewRouter()
	router.Use(middleware.RequestID)
	router.Use(middleware.RealIP)
	router.Use(middleware.Logger)
	router.Use(middleware.Recoverer)
	router.Use(cors.Handler(cors.Options{
		AllowedOrigins: []string{"*"},
		AllowedMethods: []string{"GET", "POST", "PATCH", "OPTIONS"},
		AllowedHeaders: []string{"Accept", "Authorization", "Content-Type"},
	}))

	router.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})

	router.Route("/api/v1", func(r chi.Router) {
		r.Post("/surveys", handler.createSurvey)
		r.Get("/surveys", handler.listSurveys)
		r.Get("/sessions/{sessionId}/analytics", handler.getSessionAnalytics)
		r.Post("/sessions/{sessionId}/report/send", handler.sendSessionReport)
		r.Get("/auth/profile", handler.getProfile)
		r.Patch("/auth/profile", handler.updateProfile)
		r.Post("/auth/invitations", handler.createInvitation)
		r.Post("/auth/users/block", handler.blockUser)
		r.Post("/auth/users/unblock", handler.unblockUser)
	})

	router.Route("/public/v1", func(r chi.Router) {
		r.Post("/sessions", handler.startSession)
		r.Post("/sessions/start", handler.startSession)
		r.Get("/sessions/{sessionId}/current-question", handler.getCurrentQuestion)
		r.Post("/sessions/{sessionId}/answers", handler.submitAnswer)
		r.Post("/auth/login", handler.login)
		r.Post("/auth/refresh", handler.refreshToken)
		r.Post("/auth/register", handler.register)
		r.Get("/profiles/{userId}", handler.getPublicProfile)
	})

	return router
}

type createSurveyRequest struct {
	PsychologistID string                 `json:"psychologistId"`
	Title          string                 `json:"title"`
	Description    string                 `json:"description"`
	Settings       map[string]any         `json:"settings"`
	Questions      []createSurveyQuestion `json:"questions"`
}

type createSurveyQuestion struct {
	OrderNum   uint32               `json:"orderNum"`
	Type       string               `json:"type"`
	Text       string               `json:"text"`
	LogicRules map[string]any       `json:"logicRules"`
	Answers    []createSurveyAnswer `json:"answers"`
}

type createSurveyAnswer struct {
	ID          string  `json:"id"`
	Text        string  `json:"text"`
	Weight      float64 `json:"weight"`
	CategoryTag string  `json:"categoryTag"`
}

type startSessionRequest struct {
	SurveyID       string         `json:"surveyId"`
	ClientMetadata map[string]any `json:"clientMetadata"`
}

type submitAnswerRequest struct {
	QuestionID string   `json:"questionId"`
	AnswerID   string   `json:"answerId"`
	RawText    string   `json:"rawText"`
	AnswerIDs  []string `json:"answerIds"`
}

type sendReportRequest struct {
	ReportFormat string `json:"reportFormat"`
}

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type refreshTokenRequest struct {
	RefreshToken string `json:"refreshToken"`
}

type registerRequest struct {
	Token    string `json:"token"`
	Password string `json:"password"`
}

type updateProfileRequest struct {
	PhotoURL string `json:"photoUrl"`
	About    string `json:"about"`
}

type createInvitationRequest struct {
	FullName    string `json:"fullName"`
	Phone       string `json:"phone"`
	Email       string `json:"email"`
	Role        string `json:"role"`
	AccessUntil string `json:"accessUntil"`
	ExpiresAt   string `json:"expiresAt"`
}

type userEmailRequest struct {
	Email string `json:"email"`
}

func (h *Handler) createSurvey(w http.ResponseWriter, r *http.Request) {
	var req createSurveyRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, err)
		return
	}

	draft := domain.SurveyDraft{
		PsychologistID: req.PsychologistID,
		Title:          req.Title,
		Description:    req.Description,
		Settings:       req.Settings,
		Questions:      make([]domain.SurveyQuestionDraft, 0, len(req.Questions)),
	}

	for _, question := range req.Questions {
		questionType, err := domain.ParseQuestionType(question.Type)
		if err != nil {
			writeError(w, err)
			return
		}

		answers := make([]domain.SurveyAnswerDraft, 0, len(question.Answers))
		for _, answer := range question.Answers {
			answers = append(answers, domain.SurveyAnswerDraft{
				ID:          answer.ID,
				Text:        answer.Text,
				Weight:      answer.Weight,
				CategoryTag: answer.CategoryTag,
			})
		}

		draft.Questions = append(draft.Questions, domain.SurveyQuestionDraft{
			OrderNum:   question.OrderNum,
			Type:       questionType,
			Text:       question.Text,
			LogicRules: question.LogicRules,
			Answers:    answers,
		})
	}

	surveyID, err := h.surveys.CreateSurvey(r.Context(), draft)
	if err != nil {
		writeError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, map[string]string{"surveyId": surveyID})
}

func (h *Handler) listSurveys(w http.ResponseWriter, r *http.Request) {
	psychologistID := r.URL.Query().Get("psychologistId")
	if psychologistID == "" {
		psychologistID = r.URL.Query().Get("psychologist_id")
	}

	surveys, err := h.surveys.ListSurveys(r.Context(), psychologistID)
	if err != nil {
		writeError(w, err)
		return
	}

	type surveyItem struct {
		SurveyID         string `json:"surveyId"`
		Title            string `json:"title"`
		CompletionsCount int32  `json:"completionsCount"`
	}

	items := make([]surveyItem, 0, len(surveys))
	for _, survey := range surveys {
		items = append(items, surveyItem{
			SurveyID:         survey.SurveyID,
			Title:            survey.Title,
			CompletionsCount: survey.CompletionsCount,
		})
	}

	writeJSON(w, http.StatusOK, map[string]any{"surveys": items})
}

func (h *Handler) startSession(w http.ResponseWriter, r *http.Request) {
	var req startSessionRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, err)
		return
	}

	metadata := domain.NewClientMetadata(req.ClientMetadata)
	sessionID, firstQuestion, err := h.sessions.StartSession(r.Context(), req.SurveyID, metadata)
	if err != nil {
		writeError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, map[string]any{
		"sessionId":     sessionID,
		"firstQuestion": mapQuestion(firstQuestion),
	})
}

func (h *Handler) getCurrentQuestion(w http.ResponseWriter, r *http.Request) {
	sessionID := chi.URLParam(r, "sessionId")
	question, err := h.sessions.GetCurrentQuestion(r.Context(), sessionID)
	if err != nil {
		writeError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, mapQuestion(question))
}

func (h *Handler) submitAnswer(w http.ResponseWriter, r *http.Request) {
	sessionID := chi.URLParam(r, "sessionId")

	var req submitAnswerRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, err)
		return
	}

	result, err := h.sessions.SubmitAnswer(r.Context(), ports.SubmitAnswerInput{
		SessionID:  sessionID,
		QuestionID: req.QuestionID,
		AnswerID:   req.AnswerID,
		RawText:    req.RawText,
		AnswerIDs:  req.AnswerIDs,
	})
	if err != nil {
		writeError(w, err)
		return
	}

	response := map[string]any{
		"nextQuestionId": result.NextQuestionID,
		"isFinished":     result.IsFinished,
	}

	if result.NextQuestion != nil {
		response["nextQuestion"] = mapQuestion(result.NextQuestion)
	}

	if result.ReportDelivery != nil {
		response["reportDelivery"] = map[string]any{
			"status":       result.ReportDelivery.Status,
			"email":        result.ReportDelivery.Email,
			"fileName":     result.ReportDelivery.FileName,
			"contentType":  result.ReportDelivery.ContentType,
			"errorMessage": result.ReportDelivery.ErrorMessage,
		}
	}

	writeJSON(w, http.StatusOK, response)
}

func (h *Handler) getSessionAnalytics(w http.ResponseWriter, r *http.Request) {
	sessionID := chi.URLParam(r, "sessionId")
	analytics, err := h.sessions.GetSessionAnalytics(r.Context(), sessionID)
	if err != nil {
		writeError(w, err)
		return
	}

	responses := make([]map[string]any, 0, len(analytics.Responses))
	for _, response := range analytics.Responses {
		responses = append(responses, map[string]any{
			"questionId":     response.QuestionID,
			"questionType":   response.QuestionType,
			"questionText":   response.QuestionText,
			"selectedWeight": response.SelectedWeight,
			"categoryTag":    response.CategoryTag,
			"rawText":        response.RawText,
		})
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"surveyId":       analytics.SurveyID,
		"sessionId":      analytics.SessionID,
		"clientMetadata": analytics.ClientMetadata.Values(),
		"responses":      responses,
	})
}

func (h *Handler) sendSessionReport(w http.ResponseWriter, r *http.Request) {
	sessionID := chi.URLParam(r, "sessionId")

	var req sendReportRequest
	if err := decodeJSON(r, &req); err != nil && !errors.Is(err, errEmptyBody) {
		writeError(w, err)
		return
	}

	delivery, err := h.sessions.SendClientReport(r.Context(), sessionID, domain.ParseReportFormat(req.ReportFormat))
	if err != nil {
		writeError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"status":      delivery.Status,
		"email":       delivery.Email,
		"fileName":    delivery.FileName,
		"contentType": delivery.ContentType,
	})
}

func (h *Handler) login(w http.ResponseWriter, r *http.Request) {
	var req loginRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, err)
		return
	}

	tokens, err := h.auth.Login(r.Context(), req.Email, req.Password)
	if err != nil {
		writeError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, mapTokens(tokens))
}

func (h *Handler) refreshToken(w http.ResponseWriter, r *http.Request) {
	var req refreshTokenRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, err)
		return
	}

	tokens, err := h.auth.RefreshToken(r.Context(), req.RefreshToken)
	if err != nil {
		writeError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, mapTokens(tokens))
}

func (h *Handler) register(w http.ResponseWriter, r *http.Request) {
	var req registerRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, err)
		return
	}

	tokens, err := h.auth.Register(r.Context(), req.Token, req.Password)
	if err != nil {
		writeError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, mapTokens(tokens))
}

func (h *Handler) getProfile(w http.ResponseWriter, r *http.Request) {
	profile, err := h.auth.GetProfile(r.Context(), r.Header.Get("Authorization"))
	if err != nil {
		writeError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, mapProfile(profile))
}

func (h *Handler) updateProfile(w http.ResponseWriter, r *http.Request) {
	var req updateProfileRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, err)
		return
	}

	profile, err := h.auth.UpdateProfile(r.Context(), r.Header.Get("Authorization"), domain.ProfileUpdate{
		PhotoURL: req.PhotoURL,
		About:    req.About,
	})
	if err != nil {
		writeError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, mapProfile(profile))
}

func (h *Handler) getPublicProfile(w http.ResponseWriter, r *http.Request) {
	userID := chi.URLParam(r, "userId")

	profile, err := h.auth.GetPublicProfile(r.Context(), userID)
	if err != nil {
		writeError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"fullName": profile.FullName,
		"photoUrl": profile.PhotoURL,
		"about":    profile.About,
	})
}

func (h *Handler) createInvitation(w http.ResponseWriter, r *http.Request) {
	var req createInvitationRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, err)
		return
	}

	accessUntil, err := parseFlexibleDate(req.AccessUntil)
	if err != nil {
		writeError(w, fmt.Errorf("%w: accessUntil must be RFC3339 or YYYY-MM-DD", domain.ErrInvalidInput))
		return
	}

	expiresAt, err := parseFlexibleDateTime(req.ExpiresAt, true)
	if err != nil {
		writeError(w, fmt.Errorf("%w: expiresAt must be RFC3339 or YYYY-MM-DD", domain.ErrInvalidInput))
		return
	}

	invitation, err := h.auth.CreateInvitation(r.Context(), r.Header.Get("Authorization"), domain.InvitationDraft{
		FullName:    req.FullName,
		Phone:       req.Phone,
		Email:       req.Email,
		Role:        req.Role,
		AccessUntil: accessUntil,
		ExpiresAt:   expiresAt,
	})
	if err != nil {
		writeError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, map[string]any{
		"invitationToken": invitation.Token,
		"invitationUrl":   invitation.URL,
	})
}

func (h *Handler) blockUser(w http.ResponseWriter, r *http.Request) {
	var req userEmailRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, err)
		return
	}

	if err := h.auth.BlockUser(r.Context(), r.Header.Get("Authorization"), req.Email); err != nil {
		writeError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{"status": "blocked", "email": req.Email})
}

func (h *Handler) unblockUser(w http.ResponseWriter, r *http.Request) {
	var req userEmailRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, err)
		return
	}

	if err := h.auth.UnblockUser(r.Context(), r.Header.Get("Authorization"), req.Email); err != nil {
		writeError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{"status": "unblocked", "email": req.Email})
}

func mapQuestion(question *domain.Question) map[string]any {
	if question == nil {
		return nil
	}

	answers := make([]map[string]any, 0, len(question.Answers))
	for _, answer := range question.Answers {
		answers = append(answers, map[string]any{
			"answerId": answer.ID,
			"text":     answer.Text,
		})
	}

	return map[string]any{
		"questionId": question.ID,
		"type":       question.Type,
		"text":       question.Text,
		"answers":    answers,
	}
}

func mapTokens(tokens *domain.AuthTokens) map[string]any {
	if tokens == nil {
		return nil
	}

	return map[string]any{
		"accessToken":  tokens.AccessToken,
		"refreshToken": tokens.RefreshToken,
		"expiresIn":    tokens.ExpiresIn,
		"role":         tokens.Role,
	}
}

func mapProfile(profile *domain.UserProfile) map[string]any {
	if profile == nil {
		return nil
	}

	payload := map[string]any{
		"id":          profile.ID,
		"email":       profile.Email,
		"fullName":    profile.FullName,
		"phone":       profile.Phone,
		"role":        profile.Role,
		"status":      profile.Status,
		"photoUrl":    profile.PhotoURL,
		"about":       profile.About,
		"accessUntil": "",
	}

	if !profile.AccessUntil.IsZero() {
		payload["accessUntil"] = profile.AccessUntil.Format(time.RFC3339)
	}

	return payload
}

func parseFlexibleDate(value string) (time.Time, error) {
	return parseFlexibleDateTime(value, false)
}

func parseFlexibleDateTime(value string, endOfDay bool) (time.Time, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return time.Time{}, fmt.Errorf("empty time")
	}

	if parsed, err := time.Parse(time.RFC3339, value); err == nil {
		return parsed, nil
	}

	if parsed, err := time.Parse("2006-01-02", value); err == nil {
		if endOfDay {
			return parsed.Add(23*time.Hour + 59*time.Minute + 59*time.Second), nil
		}
		return parsed, nil
	}

	return time.Time{}, fmt.Errorf("unsupported time format")
}

var errEmptyBody = errors.New("empty body")

func decodeJSON(r *http.Request, dest any) error {
	if r.Body == nil {
		return errEmptyBody
	}

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(dest); err != nil {
		if strings.Contains(err.Error(), "EOF") {
			return errEmptyBody
		}
		return fmtError(err)
	}

	if decoder.More() {
		return fmtError(errors.New("request body must contain a single JSON object"))
	}

	return nil
}

func writeJSON(w http.ResponseWriter, statusCode int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	_ = json.NewEncoder(w).Encode(payload)
}

func writeError(w http.ResponseWriter, err error) {
	statusCode := http.StatusInternalServerError
	code := "internal_error"
	message := err.Error()

	switch {
	case errors.Is(err, errEmptyBody):
		statusCode = http.StatusBadRequest
		code = "invalid_request"
		message = "request body is required"
	case errors.Is(err, domain.ErrInvalidInput), errors.Is(err, domain.ErrEmailRequired):
		statusCode = http.StatusBadRequest
		code = "invalid_request"
	case errors.Is(err, domain.ErrReportDeliveryDisabled):
		statusCode = http.StatusServiceUnavailable
		code = "report_delivery_disabled"
	default:
		if statusErr, ok := grpcstatus.FromError(err); ok {
			code = strings.ToLower(statusErr.Code().String())
			message = statusErr.Message()
			switch statusErr.Code() {
			case codes.InvalidArgument:
				statusCode = http.StatusBadRequest
			case codes.NotFound:
				statusCode = http.StatusNotFound
			case codes.AlreadyExists:
				statusCode = http.StatusConflict
			case codes.FailedPrecondition:
				statusCode = http.StatusPreconditionFailed
			case codes.PermissionDenied:
				statusCode = http.StatusForbidden
			case codes.Unauthenticated:
				statusCode = http.StatusUnauthorized
			case codes.Unimplemented:
				statusCode = http.StatusNotImplemented
			case codes.Unavailable:
				statusCode = http.StatusServiceUnavailable
			default:
				statusCode = http.StatusBadGateway
			}
		}
	}

	writeJSON(w, statusCode, map[string]any{
		"error": map[string]string{
			"code":    code,
			"message": message,
		},
	})
}

func fmtError(err error) error {
	return fmt.Errorf("%w: %s", domain.ErrInvalidInput, err.Error())
}
