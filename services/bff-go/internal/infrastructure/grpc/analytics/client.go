package analytics

import (
	"context"
	"fmt"
	"strings"

	"google.golang.org/grpc"
	analyticspb "sourcecraft.dev/benzo/bff/internal/clients/grpc/analyticspb"
	"sourcecraft.dev/benzo/bff/internal/domain"
)

const generateReportMethod = "/analytics.v1.AnalyticsService/GenerateReport"

type Client struct {
	conn grpc.ClientConnInterface
}

func NewClient(conn grpc.ClientConnInterface) *Client {
	return &Client{conn: conn}
}

func (c *Client) GenerateReport(ctx context.Context, analytics domain.SessionAnalytics, format domain.ReportFormat) (*domain.GeneratedReport, error) {
	clientMetadataJSON, err := analytics.ClientMetadata.JSON()
	if err != nil {
		return nil, err
	}

	responses := make([]*analyticspb.RawClientResponse, 0, len(analytics.Responses))
	for _, response := range analytics.Responses {
		responses = append(responses, &analyticspb.RawClientResponse{
			QuestionId:     response.QuestionID,
			QuestionType:   mapQuestionType(response.QuestionType),
			QuestionText:   response.QuestionText,
			SelectedWeight: response.SelectedWeight,
			CategoryTag:    response.CategoryTag,
			RawText:        response.RawText,
		})
	}

	req := &analyticspb.GenerateReportRequest{
		SessionId:          analytics.SessionID,
		Format:             mapReportFormat(format),
		ClientMetadataJson: clientMetadataJSON,
		Responses:          responses,
	}

	resp := &analyticspb.GenerateReportResponse{}
	if err := c.conn.Invoke(ctx, generateReportMethod, req, resp); err != nil {
		return nil, err
	}

	fileName := strings.TrimSpace(resp.GetSuggestedFilename())
	if fileName == "" {
		return nil, fmt.Errorf("%w: analytics response is missing filename", domain.ErrUpstreamResponse)
	}

	contentType := strings.TrimSpace(resp.GetContentType())
	if contentType == "" {
		return nil, fmt.Errorf("%w: analytics response is missing content type", domain.ErrUpstreamResponse)
	}

	if len(resp.GetFileContent()) == 0 {
		return nil, fmt.Errorf("%w: analytics response is missing file content", domain.ErrUpstreamResponse)
	}

	return &domain.GeneratedReport{
		FileName:    fileName,
		ContentType: contentType,
		Content:     resp.GetFileContent(),
	}, nil
}

func mapQuestionType(questionType domain.QuestionType) analyticspb.QuestionType {
	switch questionType {
	case domain.QuestionTypeText:
		return analyticspb.QuestionType_TEXT
	default:
		return analyticspb.QuestionType_CHOICE
	}
}

func mapReportFormat(format domain.ReportFormat) analyticspb.ReportFormat {
	switch format {
	case domain.ReportFormatClientHTML:
		return analyticspb.ReportFormat_REPORT_FORMAT_HTML
	case domain.ReportFormatPsychoDocx:
		return analyticspb.ReportFormat_REPORT_FORMAT_PSYCHO_DOCX
	case domain.ReportFormatPsychoHTML:
		return analyticspb.ReportFormat_REPORT_FORMAT_PSYCHO_HTML
	default:
		return analyticspb.ReportFormat_REPORT_FORMAT_DOCX
	}
}
