package domain

import (
	"encoding/json"
	"fmt"
	"net/mail"
	"strings"
)

type ClientMetadata struct {
	values map[string]any
}

type SessionAnalytics struct {
	SurveyID       string
	SessionID      string
	ClientMetadata ClientMetadata
	Responses      []AnalyticsResponse
}

type SurveySessionSummary struct {
	SurveyID       string
	SessionID      string
	ClientMetadata ClientMetadata
	Status         string
	ResponsesCount int32
	StartedAt      string
	FinishedAt     string
}

type AnalyticsResponse struct {
	QuestionID     string
	QuestionType   QuestionType
	QuestionText   string
	SelectedWeight float64
	CategoryTag    string
	RawText        string
}

type ReportFormat string

const (
	ReportFormatClientDocx ReportFormat = "client_docx"
	ReportFormatClientHTML ReportFormat = "client_html"
	ReportFormatPsychoDocx ReportFormat = "psycho_docx"
	ReportFormatPsychoHTML ReportFormat = "psycho_html"
)

type GeneratedReport struct {
	FileName    string
	ContentType string
	Content     []byte
}

type ReportDeliveryStatus string

const (
	ReportDeliverySent   ReportDeliveryStatus = "sent"
	ReportDeliveryFailed ReportDeliveryStatus = "failed"
)

type ReportDelivery struct {
	Status       ReportDeliveryStatus
	Email        string
	FileName     string
	ContentType  string
	ErrorMessage string
}

type SubmitAnswerResult struct {
	NextQuestionID string
	IsFinished     bool
	NextQuestion   *Question
	ReportDelivery *ReportDelivery
}

func NewClientMetadata(values map[string]any) ClientMetadata {
	if values == nil {
		values = map[string]any{}
	}

	cloned := make(map[string]any, len(values))
	for key, value := range values {
		cloned[key] = value
	}

	return ClientMetadata{values: cloned}
}

func ParseClientMetadataJSON(raw string) (ClientMetadata, error) {
	if strings.TrimSpace(raw) == "" {
		return NewClientMetadata(nil), nil
	}

	values := map[string]any{}
	if err := json.Unmarshal([]byte(raw), &values); err != nil {
		return ClientMetadata{}, fmt.Errorf("%w: invalid client metadata json", ErrInvalidInput)
	}

	return NewClientMetadata(values), nil
}

func (m ClientMetadata) Values() map[string]any {
	cloned := make(map[string]any, len(m.values))
	for key, value := range m.values {
		cloned[key] = value
	}
	return cloned
}

func (m ClientMetadata) JSON() (string, error) {
	payload, err := json.Marshal(m.values)
	if err != nil {
		return "", fmt.Errorf("%w: unable to serialize client metadata", ErrInvalidInput)
	}
	return string(payload), nil
}

func (m ClientMetadata) Email() (string, error) {
	raw, ok := m.values["email"]
	if !ok {
		return "", fmt.Errorf("%w: field email is required in clientMetadata", ErrEmailRequired)
	}

	email, ok := raw.(string)
	if !ok {
		return "", fmt.Errorf("%w: field email must be a string", ErrEmailRequired)
	}

	email = strings.TrimSpace(email)
	if email == "" {
		return "", fmt.Errorf("%w: field email must not be empty", ErrEmailRequired)
	}

	if _, err := mail.ParseAddress(email); err != nil {
		return "", fmt.Errorf("%w: invalid email format", ErrEmailRequired)
	}

	return email, nil
}

func ParseReportFormat(value string) ReportFormat {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "", "client_docx", "docx":
		return ReportFormatClientDocx
	case "client_html", "html":
		return ReportFormatClientHTML
	case "psycho_docx":
		return ReportFormatPsychoDocx
	case "psycho_html":
		return ReportFormatPsychoHTML
	default:
		return ReportFormatClientDocx
	}
}

func ValidateReportFormat(value string) (ReportFormat, error) {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "", "client_docx", "docx":
		return ReportFormatClientDocx, nil
	case "client_html", "html":
		return ReportFormatClientHTML, nil
	case "psycho_docx":
		return ReportFormatPsychoDocx, nil
	case "psycho_html":
		return ReportFormatPsychoHTML, nil
	default:
		return "", fmt.Errorf("%w: unsupported reportFormat %q", ErrInvalidInput, value)
	}
}
