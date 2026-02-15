// Tarh Workflow Orchestrator
//
// Performs the full SP → FP → Funding → Opportunity lifecycle
// with randomized data on every run.
//
// Usage:
//
//	cp .env.example .env   # edit values as needed
//	go run .
package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"time"

	"tarh-script/client"
	"tarh-script/config"
	"tarh-script/models"
	"tarh-script/services"
	"tarh-script/utils"
)

func main() {
	os.Exit(run())
}

func run() int {
	// ── Logger ──────────────────────────────────────────
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	slog.SetDefault(logger)

	// ── Load .env (best effort) ────────────────────────
	loadDotEnv()

	// ── Config ─────────────────────────────────────────
	cfg, err := config.Load()
	if err != nil {
		logger.Error("Failed to load configuration", slog.String("error", err.Error()))
		return 1
	}
	logger.Info("Configuration loaded",
		slog.String("base_url", cfg.BaseURL),
	)

	// ── Assets ─────────────────────────────────────────
	assetsDir, _ := filepath.Abs("assets")
	if err := utils.EnsureAssets(assetsDir); err != nil {
		logger.Error("Failed to ensure test assets", slog.String("error", err.Error()))
		return 1
	}
	logger.Info("Test assets ready", slog.String("dir", assetsDir))

	// ── HTTP Client ────────────────────────────────────
	httpClient := client.New(cfg.BaseURL, logger)

	// ── Workflow State ─────────────────────────────────
	state := &models.WorkflowState{}

	// ── Services ───────────────────────────────────────
	spSvc := services.NewSPService(httpClient, cfg, logger, assetsDir)
	fpSvc := services.NewFPService(httpClient, cfg, logger)
	fundingSvc := services.NewFundingService(httpClient, cfg, logger, assetsDir)
	opportunitySvc := services.NewOpportunityService(httpClient, cfg, logger, assetsDir)

	// ── Context with global timeout ────────────────────
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	// ── Orchestration Pipeline ─────────────────────────
	start := time.Now()
	logger.Info("═══════════════════════════════════════════════")
	logger.Info("  Tarh Workflow Orchestrator – Starting Run")
	logger.Info("═══════════════════════════════════════════════")

	steps := []struct {
		name string
		fn   func(context.Context, *models.WorkflowState) error
	}{
		{"SP Register", spSvc.Register},
		{"SP Login", spSvc.Login},
		{"SP OTP", spSvc.VerifyOTP},
		{"Create Company", spSvc.CreateCompany},
		{"FP Login", fpSvc.Login},
		{"FP OTP", fpSvc.VerifyOTP},
		{"Approve Company", fpSvc.ApproveCompany},
		{"Create Funding Request", fundingSvc.CreateFundingRequest},
		{"Create Opportunity", opportunitySvc.CreateOpportunity},
	}

	for i, step := range steps {
		stepNum := i + 1
		logger.Info(fmt.Sprintf("──── Step %d/%d: %s ────", stepNum, len(steps), step.name))

		stepStart := time.Now()
		if err := step.fn(ctx, state); err != nil {
			logger.Error("Step failed — aborting workflow",
				slog.Int("step", stepNum),
				slog.String("name", step.name),
				slog.String("error", err.Error()),
				slog.Duration("elapsed", time.Since(stepStart)),
			)
			return 1
		}

		logger.Info(fmt.Sprintf("✓ Step %d/%d complete", stepNum, len(steps)),
			slog.String("name", step.name),
			slog.Duration("elapsed", time.Since(stepStart)),
		)
	}

	// ── Summary ────────────────────────────────────────
	elapsed := time.Since(start)
	logger.Info("═══════════════════════════════════════════════")
	logger.Info("  Workflow Complete!")
	logger.Info("═══════════════════════════════════════════════")
	logger.Info("Summary",
		slog.String("sp_national_id", state.SPNationalID),
		slog.String("sp_email", state.SPEmail),
		slog.String("company_id", state.CompanyID),
		slog.String("funding_request_id", state.FundingRequestID),
		slog.String("opportunity_id", state.OpportunityID),
		slog.Duration("total_elapsed", elapsed),
	)

	return 0
}

// loadDotEnv is a minimal .env loader. It reads key=value pairs from a .env
// file (if present) and sets them as environment variables only when not
// already set, so real env vars always take precedence.
func loadDotEnv() {
	data, err := os.ReadFile(".env")
	if err != nil {
		return // file doesn't exist — that's fine
	}

	for _, line := range splitLines(string(data)) {
		line = trimSpace(line)
		if line == "" || line[0] == '#' {
			continue
		}
		parts := splitFirst(line, '=')
		if len(parts) != 2 {
			continue
		}
		key := trimSpace(parts[0])
		val := trimSpace(parts[1])
		// Strip surrounding quotes.
		val = stripQuotes(val)
		// Only set if not already in env.
		if os.Getenv(key) == "" {
			os.Setenv(key, val)
		}
	}
}

// ---------- tiny string helpers (avoid importing extra packages) ----------

func splitLines(s string) []string {
	var lines []string
	start := 0
	for i := 0; i < len(s); i++ {
		if s[i] == '\n' {
			lines = append(lines, s[start:i])
			start = i + 1
		}
	}
	if start < len(s) {
		lines = append(lines, s[start:])
	}
	return lines
}

func trimSpace(s string) string {
	i, j := 0, len(s)
	for i < j && (s[i] == ' ' || s[i] == '\t' || s[i] == '\r') {
		i++
	}
	for j > i && (s[j-1] == ' ' || s[j-1] == '\t' || s[j-1] == '\r') {
		j--
	}
	return s[i:j]
}

func splitFirst(s string, sep byte) []string {
	for i := 0; i < len(s); i++ {
		if s[i] == sep {
			return []string{s[:i], s[i+1:]}
		}
	}
	return []string{s}
}

func stripQuotes(s string) string {
	if len(s) >= 2 {
		if (s[0] == '"' && s[len(s)-1] == '"') || (s[0] == '\'' && s[len(s)-1] == '\'') {
			return s[1 : len(s)-1]
		}
	}
	return s
}
