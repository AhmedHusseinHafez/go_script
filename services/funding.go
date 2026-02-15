package services

import (
	"context"
	"fmt"
	"log/slog"
	"path/filepath"
	"strconv"

	"tarh-script/client"
	"tarh-script/config"
	"tarh-script/models"
	"tarh-script/utils"
)

// FundingService handles funding request creation.
type FundingService struct {
	client    *client.Client
	cfg       *config.Config
	logger    *slog.Logger
	assetsDir string
}

// NewFundingService creates a new FundingService.
func NewFundingService(c *client.Client, cfg *config.Config, logger *slog.Logger, assetsDir string) *FundingService {
	return &FundingService{
		client:    c,
		cfg:       cfg,
		logger:    logger.With(slog.String("service", "funding")),
		assetsDir: assetsDir,
	}
}

// CreateFundingRequest creates a funding request under the SP's company.
func (f *FundingService) CreateFundingRequest(ctx context.Context, state *models.WorkflowState) error {
	f.logger.Info("Step 8: Create Funding Request (SP)",
		slog.String("company_id", state.CompanyID),
	)

	imagePath := filepath.Join(f.assetsDir, "test_image.png")

	projectValue := utils.RandomInt(1000000, 50000000)
	fundingAmount := utils.RandomInt(500000, projectValue)

	fields := map[string]string{
		"project_name":             utils.RandomProjectName(),
		"sector_id":               "1",
		"industry_id":             "1",
		"project_value":           strconv.Itoa(projectValue),
		"requested_funding_amount": strconv.Itoa(fundingAmount),
		"funding_duration":         strconv.Itoa(utils.RandomInt(6, 36)),
		"project_description":      utils.RandomDescription(),
		"project_location":         utils.RandomAddress(),
		"latitude":                 utils.RandomLatitude(),
		"longitude":                utils.RandomLongitude(),
		"objectives":               utils.RandomDescription(),
		"approvables[0][type]":     "file",
		"approvables[0][status]":   "1",
	}

	files := []client.FileField{
		{FieldName: "approvables[0][key]", FilePath: imagePath},
	}

	headers := client.MergeHeaders(
		client.AuthHeader(state.SPToken),
		map[string]string{"Company-Id": state.CompanyID},
	)

	resp, err := f.client.PostMultipart(ctx, "/v1/funding/sp/funding-request", fields, files, headers)
	if err != nil {
		return fmt.Errorf("create funding request: %w", err)
	}

	fundingID, err := extractFundingRequestID(resp.Body)
	if err != nil {
		return fmt.Errorf("extract funding request id: %w", err)
	}

	f.logger.Info("Funding request created",
		slog.String("funding_request_id", fundingID),
	)

	state.FundingRequestID = fundingID
	return nil
}

// extractFundingRequestID tries multiple response shapes.
func extractFundingRequestID(body map[string]interface{}) (string, error) {
	// Try: { "data": { "id": 123 } }
	if data, ok := body["data"].(map[string]interface{}); ok {
		if id, exists := data["id"]; exists {
			return toStringID(id)
		}
	}
	// Try: { "funding_request": { "id": 123 } }
	if fr, ok := body["funding_request"].(map[string]interface{}); ok {
		if id, exists := fr["id"]; exists {
			return toStringID(id)
		}
	}
	// Try top level
	if id, ok := body["id"]; ok {
		return toStringID(id)
	}
	return "", fmt.Errorf("funding request id not found in response: %v", body)
}
