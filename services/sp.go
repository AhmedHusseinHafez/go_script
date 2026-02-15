// Package services implements each step of the Tarh workflow lifecycle.
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

// SPService handles all SP-related API operations.
type SPService struct {
	client    *client.Client
	cfg       *config.Config
	logger    *slog.Logger
	assetsDir string
}

// NewSPService creates a new SPService.
func NewSPService(c *client.Client, cfg *config.Config, logger *slog.Logger, assetsDir string) *SPService {
	return &SPService{
		client:    c,
		cfg:       cfg,
		logger:    logger.With(slog.String("service", "sp")),
		assetsDir: assetsDir,
	}
}

// Register performs SP registration and returns the national_id used.
func (s *SPService) Register(ctx context.Context, state *models.WorkflowState) error {
	s.logger.Info("Step 1: SP Register")

	nationalID := utils.RandomNationalID()
	email := utils.RandomEmail("sptest.com")
	password := s.cfg.SPPassword

	body := map[string]interface{}{
		"user_type":             2,
		"name":                  utils.RandomName(),
		"email":                 email,
		"national_id":           nationalID,
		"phone_number":          utils.RandomPhone(),
		"address":               utils.RandomAddress(),
		"password":              password,
		"password_confirmation": password,
	}

	resp, err := s.client.PostJSON(ctx, "/auth/sp/register", body, nil)
	if err != nil {
		return fmt.Errorf("sp register: %w", err)
	}

	s.logger.Info("SP Register complete",
		slog.Int("status", resp.StatusCode),
		slog.String("national_id", nationalID),
	)

	state.SPNationalID = nationalID
	return nil
}

// Login performs SP login and extracts the email from the response.
func (s *SPService) Login(ctx context.Context, state *models.WorkflowState) error {
	s.logger.Info("Step 2: SP Login")

	body := map[string]interface{}{
		"national_id": state.SPNationalID,
		"user_type":   2,
	}

	resp, err := s.client.PostJSON(ctx, "/auth/login", body, nil)
	if err != nil {
		return fmt.Errorf("sp login: %w", err)
	}

	// Extract email from response: { "user": { "email": "..." } }
	email, err := extractNestedString(resp.Body, "user", "email")
	if err != nil {
		return fmt.Errorf("sp login extract email: %w", err)
	}

	s.logger.Info("SP Login complete",
		slog.String("email", email),
	)

	state.SPEmail = email
	return nil
}

// VerifyOTP sends the OTP for the SP user and stores the auth token.
func (s *SPService) VerifyOTP(ctx context.Context, state *models.WorkflowState) error {
	s.logger.Info("Step 3: SP OTP Verification")

	body := map[string]interface{}{
		"email": state.SPEmail,
		"otp":   s.cfg.OTP,
	}

	resp, err := s.client.PostJSON(ctx, "/auth/otp", body, nil)
	if err != nil {
		return fmt.Errorf("sp otp: %w", err)
	}

	token, err := extractHeaderToken(resp)
	if err != nil {
		return fmt.Errorf("sp otp extract token: %w", err)
	}

	s.logger.Info("SP OTP complete", slog.String("token_prefix", truncateToken(token)))

	state.SPToken = token
	return nil
}

// CreateCompany creates a company under the SP and stores the company ID.
func (s *SPService) CreateCompany(ctx context.Context, state *models.WorkflowState) error {
	s.logger.Info("Step 4: Create Company (SP)")

	imagePath := filepath.Join(s.assetsDir, "test_image.png")

	fields := map[string]string{
		"email":                 utils.RandomEmail("company.com"),
		"name":                  utils.RandomCompanyName(),
		"trade_name_en":         utils.RandomCompanyName(),
		"trade_name_ar":         utils.RandomCompanyNameAr(),
		"short_name":            utils.RandomFirstName() + "Co",
		"headquarters":          utils.RandomAddress(),
		"description":           utils.RandomDescription(),
		"main_activity":         "Real Estate Development",
		"website":               utils.RandomURL("www.company"),
		"uni_number":            utils.RandomUniNumber(),
		"cr_number":             utils.RandomCRNumber(),
		"phone_number":          utils.RandomPhone(),
		"other_sector":          "Technology",
		"cr_establishment_date": utils.CREstablishmentDate(),
		"cr_expiry_date":        utils.CRExpiryDate(),
		"employee_count":        strconv.Itoa(utils.RandomInt(10, 500)),
		"industry_id":           "1",
		"sector_id":             "1",
		"linkedin_url":          utils.RandomURL("linkedin"),
		"primary_contact_email": utils.RandomEmail("contact.com"),
		"primary_contact_name":  utils.RandomName(),
		"secondary_email":       utils.RandomEmail("secondary.com"),
		"twitter_url":           utils.RandomURL("twitter"),
	}

	files := []client.FileField{
		{FieldName: "logo", FilePath: imagePath},
		{FieldName: "cr_document", FilePath: imagePath},
	}

	headers := client.AuthHeader(state.SPToken)

	resp, err := s.client.PostMultipart(ctx, "/auth/sp/company", fields, files, headers)
	if err != nil {
		return fmt.Errorf("create company: %w", err)
	}

	companyID, err := extractCompanyID(resp.Body)
	if err != nil {
		return fmt.Errorf("create company extract id: %w", err)
	}

	s.logger.Info("Company created",
		slog.String("company_id", companyID),
	)

	state.CompanyID = companyID
	return nil
}

// ---------- Helpers ----------

// extractNestedString extracts a string from a nested map: body[key1][key2].
func extractNestedString(body map[string]interface{}, keys ...string) (string, error) {
	var current interface{} = body
	for _, k := range keys {
		m, ok := current.(map[string]interface{})
		if !ok {
			return "", fmt.Errorf("expected object at key %q, got %T", k, current)
		}
		current = m[k]
	}
	s, ok := current.(string)
	if !ok {
		return "", fmt.Errorf("expected string, got %T (%v)", current, current)
	}
	return s, nil
}

// extractHeaderToken extracts the "auth-token" from the response body.
func extractHeaderToken(resp *client.Response) (string, error) {
	if resp.Body == nil {
		return "", fmt.Errorf("empty response body")
	}
	token, ok := resp.Body["auth-token"]
	if !ok {
		return "", fmt.Errorf("auth-token not found in response: %v", resp.Body)
	}
	s, ok := token.(string)
	if !ok {
		return "", fmt.Errorf("auth-token is not a string: %T", token)
	}
	return s, nil
}

// extractCompanyID tries multiple common response shapes to find the company ID.
func extractCompanyID(body map[string]interface{}) (string, error) {
	// Try: { "data": { "id": 123 } }
	if data, ok := body["data"].(map[string]interface{}); ok {
		if id, exists := data["id"]; exists {
			return toStringID(id)
		}
	}
	// Try: { "company": { "id": 123 } }
	if company, ok := body["company"].(map[string]interface{}); ok {
		if id, exists := company["id"]; exists {
			return toStringID(id)
		}
	}
	// Try: { "id": 123 }
	if id, ok := body["id"]; ok {
		return toStringID(id)
	}
	return "", fmt.Errorf("company id not found in response: %v", body)
}

// toStringID converts a numeric or string ID to a string.
func toStringID(v interface{}) (string, error) {
	switch val := v.(type) {
	case string:
		return val, nil
	case float64:
		return strconv.FormatInt(int64(val), 10), nil
	case int:
		return strconv.Itoa(val), nil
	case int64:
		return strconv.FormatInt(val, 10), nil
	default:
		return "", fmt.Errorf("unexpected id type: %T (%v)", v, v)
	}
}

// truncateToken returns the first 12 chars of a token for safe logging.
func truncateToken(token string) string {
	if len(token) > 12 {
		return token[:12] + "..."
	}
	return token
}
