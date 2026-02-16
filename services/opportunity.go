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

// OpportunityService handles opportunity creation from a funding request.
type OpportunityService struct {
	client    *client.Client
	cfg       *config.Config
	logger    *slog.Logger
	assetsDir string
}

// NewOpportunityService creates a new OpportunityService.
func NewOpportunityService(c *client.Client, cfg *config.Config, logger *slog.Logger, assetsDir string) *OpportunityService {
	return &OpportunityService{
		client:    c,
		cfg:       cfg,
		logger:    logger.With(slog.String("service", "opportunity")),
		assetsDir: assetsDir,
	}
}

// CreateOpportunity converts a funding request into an opportunity.
func (o *OpportunityService) CreateOpportunity(ctx context.Context, state *models.WorkflowState) error {
	o.logger.Info("Step 9: Create Opportunity (FP)",
		slog.String("funding_request_id", state.FundingRequestID),
	)

	imagePath := filepath.Join(o.assetsDir, "test_image.png")
	pdfPath := filepath.Join(o.assetsDir, "test_document.pdf")

	fundingGoal := utils.RandomInt(1000000, 50000000)
	totalShares := utils.RandomInt(1000, 100000)
	pricePerShare := fundingGoal / totalShares

	fields := map[string]string{
		"status":                    "2",
		"project_name_ar":           utils.RandomProjectNameAr(),
		"project_name_en":           utils.RandomProjectName(),
		"project_location_ar":       "الرياض، المملكة العربية السعودية",
		"project_location_en":       utils.RandomAddress(),
		"project_description_ar":    utils.RandomDescriptionAr(),
		"project_description_en":    utils.RandomDescription(),
		"funding_goal":              strconv.Itoa(fundingGoal),
		"minimum_investment":        strconv.Itoa(utils.RandomInt(100, 2000)),   // Adjusted for 10,000 credit limit
		"maximum_investment":        strconv.Itoa(utils.RandomInt(2000, 10000)), // Adjusted for 10,000 credit limit
		"offering_start_at":         utils.OfferingStartAt(),
		"offering_end_at":           utils.OfferingEndAt(),
		"expected_apr_percent":      utils.RandomFloat(5.0, 15.0),
		"expected_roi_percent":      utils.RandomFloat(10.0, 30.0),
		"expected_irr_percent":      utils.RandomFloat(8.0, 25.0),
		"funding_duration":          strconv.Itoa(utils.RandomInt(6, 36)),
		"total_shares_offered":      strconv.Itoa(totalShares),
		"price_per_share":           strconv.Itoa(pricePerShare),
		"distribution":              strconv.Itoa(utils.RandomInt(1, 5)), // 1:Monthly 2:Quarterly 3:Half Year 4:Yearly 5:At Exit
		"risk_level":                strconv.Itoa(utils.RandomInt(1, 3)),
		"platform_coverage":         utils.RandomFloat(1.0, 5.0),
		"max_investment_per_retail": strconv.Itoa(utils.RandomInt(200, 10000)), // Adjusted for 10,000 credit limit
		"max_number_of_retail":      strconv.Itoa(utils.RandomInt(100, 1000)),
		"fund_return_period":        strconv.Itoa(utils.RandomInt(6, 24)),
		"success_threshold":         strconv.Itoa(utils.RandomInt(60, 90)),
		"offering_type":             "3",
		"configuration_type":        "3",
		"funding_type[name_ar]":     "تمويل بالمشاركة",
		"funding_type[name_en]":     "Partnership Funding",
		"exit_strategy[name_ar]":    "بيع الحصص",
		"exit_strategy[name_en]":    "Equity Sale",
		"project_video_url":         utils.RandomRealEstateVideoURL(),
	}

	// Build file list: gallery image, hero image, and 6 documents.
	files := []client.FileField{
		{FieldName: "gallery_images[0]", FilePath: imagePath},
		{FieldName: "main_hero_image", FilePath: imagePath},
	}

	// Add 6 document entries (indices 0-5).
	for i := 0; i < 6; i++ {
		prefix := fmt.Sprintf("documents[%d]", i)
		fields[prefix+"[type]"] = "1"
		labelAr, labelEn := utils.RandomLabel()
		fields[prefix+"[label_ar]"] = labelAr
		fields[prefix+"[label_en]"] = labelEn
		files = append(files, client.FileField{
			FieldName: prefix + "[file]",
			FilePath:  pdfPath,
		})
	}

	path := fmt.Sprintf("/v1/funding/%s/to-opportunity", state.FundingRequestID)
	headers := client.AuthHeader(state.FPToken)

	resp, err := o.client.PostMultipart(ctx, path, fields, files, headers)
	if err != nil {
		return fmt.Errorf("create opportunity: %w", err)
	}

	opportunityID, err := extractOpportunityID(resp.Body)
	if err != nil {
		// Non-fatal: log warning but continue (the opportunity was created).
		o.logger.Warn("Could not extract opportunity ID, but request succeeded",
			slog.String("raw_body", string(resp.RawBody)),
		)
		state.OpportunityID = "unknown"
	} else {
		state.OpportunityID = opportunityID
	}

	o.logger.Info("Opportunity created",
		slog.String("opportunity_id", state.OpportunityID),
		slog.Int("status", resp.StatusCode),
	)

	return nil
}

// extractOpportunityID tries multiple response shapes.
func extractOpportunityID(body map[string]interface{}) (string, error) {
	if data, ok := body["data"].(map[string]interface{}); ok {
		if id, exists := data["id"]; exists {
			return toStringID(id)
		}
	}
	if opp, ok := body["opportunity"].(map[string]interface{}); ok {
		if id, exists := opp["id"]; exists {
			return toStringID(id)
		}
	}
	if id, ok := body["id"]; ok {
		return toStringID(id)
	}
	return "", fmt.Errorf("opportunity id not found in response: %v", body)
}
