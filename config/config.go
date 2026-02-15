// Package config handles environment variable loading and application configuration.
package config

import (
	"fmt"
	"os"
	"strings"
)

// Config holds all application configuration values.
type Config struct {
	// BaseURL is the API base URL (no trailing slash).
	BaseURL string

	// FPEmail is the fixed fund-provider login email.
	FPEmail string

	// FPPassword is the fixed fund-provider login password.
	FPPassword string

	// SPPassword is the password used for SP registration and login.
	SPPassword string

	// OTP is the fixed OTP code used in all verification steps.
	OTP string
}

// Load reads configuration from environment variables with sensible defaults.
func Load() (*Config, error) {
	cfg := &Config{
		BaseURL:    envOrDefault("BASE_URL", "https://api-dev2.tarh.com.sa/api"),
		FPEmail:    envOrDefault("FP_EMAIL", "fp_user_0@sa.com"),
		FPPassword: envOrDefault("FP_PASSWORD", "secret@123"),
		SPPassword: envOrDefault("SP_PASSWORD", "Secret@1234"),
		OTP:        envOrDefault("OTP", "123456"),
	}

	// Strip trailing slash from base URL.
	cfg.BaseURL = strings.TrimRight(cfg.BaseURL, "/")

	if cfg.BaseURL == "" {
		return nil, fmt.Errorf("config: BASE_URL must not be empty")
	}

	return cfg, nil
}

// envOrDefault returns the value of the environment variable named by key,
// or fallback if the variable is not set or empty.
func envOrDefault(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
