package services

import (
	"context"
	"fmt"
	"log/slog"

	"tarh-script/client"
	"tarh-script/config"
	"tarh-script/models"
)

// FPService handles all FP-related API operations.
type FPService struct {
	client *client.Client
	cfg    *config.Config
	logger *slog.Logger
}

// NewFPService creates a new FPService.
func NewFPService(c *client.Client, cfg *config.Config, logger *slog.Logger) *FPService {
	return &FPService{
		client: c,
		cfg:    cfg,
		logger: logger.With(slog.String("service", "fp")),
	}
}

// Login performs FP login.
func (f *FPService) Login(ctx context.Context, state *models.WorkflowState) error {
	f.logger.Info("Step 5: FP Login")

	body := map[string]interface{}{
		"email":    f.cfg.FPEmail,
		"password": f.cfg.FPPassword,
	}

	resp, err := f.client.PostJSON(ctx, "/auth/Fplogin", body, nil)
	if err != nil {
		return fmt.Errorf("fp login: %w", err)
	}

	f.logger.Info("FP Login complete", slog.Int("status", resp.StatusCode))
	return nil
}

// VerifyOTP sends OTP for the FP user and stores the auth token.
func (f *FPService) VerifyOTP(ctx context.Context, state *models.WorkflowState) error {
	f.logger.Info("Step 6: FP OTP Verification")

	body := map[string]interface{}{
		"email": f.cfg.FPEmail,
		"otp":   f.cfg.OTP,
	}

	resp, err := f.client.PostJSON(ctx, "/auth/otp", body, nil)
	if err != nil {
		return fmt.Errorf("fp otp: %w", err)
	}

	token, err := extractHeaderToken(resp)
	if err != nil {
		return fmt.Errorf("fp otp extract token: %w", err)
	}

	f.logger.Info("FP OTP complete", slog.String("token_prefix", truncateToken(token)))

	state.FPToken = token
	return nil
}

// ApproveCompany approves the company created by the SP.
func (f *FPService) ApproveCompany(ctx context.Context, state *models.WorkflowState) error {
	f.logger.Info("Step 7: Approve Company",
		slog.String("company_id", state.CompanyID),
	)

	path := fmt.Sprintf("/company/companies/%s/approve", state.CompanyID)
	headers := client.AuthHeader(state.FPToken)

	resp, err := f.client.PostJSON(ctx, path, nil, headers)
	if err != nil {
		return fmt.Errorf("approve company: %w", err)
	}

	f.logger.Info("Company approved",
		slog.Int("status", resp.StatusCode),
	)

	return nil
}
