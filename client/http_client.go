// Package client provides a reusable HTTP client with retry logic,
// structured logging, JSON helpers, and multipart form-data support.
package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"math"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// ---------- Configuration ----------

const (
	defaultTimeout    = 60 * time.Second
	defaultMaxRetries = 3
	retryBaseDelay    = 1 * time.Second
	retryMaxDelay     = 15 * time.Second
)

// ---------- Client ----------

// Client wraps http.Client with convenience methods for the Tarh API.
type Client struct {
	base       string
	httpClient *http.Client
	logger     *slog.Logger
	maxRetries int
}

// New creates a new API client.
func New(baseURL string, logger *slog.Logger) *Client {
	if logger == nil {
		logger = slog.Default()
	}
	return &Client{
		base: strings.TrimRight(baseURL, "/"),
		httpClient: &http.Client{
			Timeout: defaultTimeout,
		},
		logger:     logger,
		maxRetries: defaultMaxRetries,
	}
}

// ---------- Response wrapper ----------

// Response holds the raw HTTP response body parsed as a generic map plus status code.
type Response struct {
	StatusCode int
	Body       map[string]interface{}
	RawBody    []byte
}

// ---------- JSON POST ----------

// PostJSON sends a JSON POST request and returns the parsed response.
func (c *Client) PostJSON(ctx context.Context, path string, body map[string]interface{}, headers map[string]string) (*Response, error) {
	url := c.base + path

	payload, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("marshal request body: %w", err)
	}

	c.logger.Info("POST JSON",
		slog.String("url", url),
		slog.String("body", truncate(string(payload), 500)),
	)

	var resp *Response
	err = c.withRetry(ctx, func() error {
		req, reqErr := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(payload))
		if reqErr != nil {
			return fmt.Errorf("create request: %w", reqErr)
		}

		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Accept", "application/json")
		for k, v := range headers {
			req.Header.Set(k, v)
		}

		resp, reqErr = c.doRequest(req)
		return reqErr
	})

	return resp, err
}

// ---------- Multipart POST ----------

// FileField describes a file to include in a multipart request.
type FileField struct {
	FieldName string // form field name, e.g. "logo", "documents[0][file]"
	FilePath  string // absolute path on disk
}

// PostMultipart sends a multipart/form-data POST request.
func (c *Client) PostMultipart(ctx context.Context, path string, fields map[string]string, files []FileField, headers map[string]string) (*Response, error) {
	url := c.base + path

	c.logger.Info("POST Multipart",
		slog.String("url", url),
		slog.Int("fields", len(fields)),
		slog.Int("files", len(files)),
	)

	var resp *Response
	err := c.withRetry(ctx, func() error {
		body := &bytes.Buffer{}
		writer := multipart.NewWriter(body)

		// Write text fields.
		for k, v := range fields {
			if writeErr := writer.WriteField(k, v); writeErr != nil {
				return fmt.Errorf("write field %s: %w", k, writeErr)
			}
		}

		// Write file fields.
		for _, ff := range files {
			part, createErr := writer.CreateFormFile(ff.FieldName, filepath.Base(ff.FilePath))
			if createErr != nil {
				return fmt.Errorf("create form file %s: %w", ff.FieldName, createErr)
			}
			f, openErr := os.Open(ff.FilePath)
			if openErr != nil {
				return fmt.Errorf("open file %s: %w", ff.FilePath, openErr)
			}
			if _, copyErr := io.Copy(part, f); copyErr != nil {
				f.Close()
				return fmt.Errorf("copy file %s: %w", ff.FilePath, copyErr)
			}
			f.Close()
		}

		if closeErr := writer.Close(); closeErr != nil {
			return fmt.Errorf("close multipart writer: %w", closeErr)
		}

		req, reqErr := http.NewRequestWithContext(ctx, http.MethodPost, url, body)
		if reqErr != nil {
			return fmt.Errorf("create request: %w", reqErr)
		}

		req.Header.Set("Content-Type", writer.FormDataContentType())
		req.Header.Set("Accept", "application/json")
		for k, v := range headers {
			req.Header.Set(k, v)
		}

		resp, reqErr = c.doRequest(req)
		return reqErr
	})

	return resp, err
}

// ---------- Internal ----------

// doRequest executes the HTTP request and parses the response.
func (c *Client) doRequest(req *http.Request) (*Response, error) {
	httpResp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("http do: %w", err)
	}
	defer httpResp.Body.Close()

	rawBody, err := io.ReadAll(httpResp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response body: %w", err)
	}

	resp := &Response{
		StatusCode: httpResp.StatusCode,
		RawBody:    rawBody,
	}

	// Attempt JSON parse; non-JSON responses keep Body nil.
	var parsed map[string]interface{}
	if jsonErr := json.Unmarshal(rawBody, &parsed); jsonErr == nil {
		resp.Body = parsed
	}

	c.logger.Info("Response",
		slog.Int("status", httpResp.StatusCode),
		slog.String("body", truncate(string(rawBody), 800)),
	)

	// Treat 4xx/5xx as errors.
	if httpResp.StatusCode >= 400 {
		return resp, fmt.Errorf("HTTP %d: %s", httpResp.StatusCode, truncate(string(rawBody), 300))
	}

	return resp, nil
}

// withRetry executes fn with exponential back-off on transient failures.
func (c *Client) withRetry(ctx context.Context, fn func() error) error {
	var lastErr error
	for attempt := 0; attempt <= c.maxRetries; attempt++ {
		if attempt > 0 {
			delay := backoffDelay(attempt)
			c.logger.Warn("Retrying request",
				slog.Int("attempt", attempt),
				slog.Duration("delay", delay),
			)
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(delay):
			}
		}

		lastErr = fn()
		if lastErr == nil {
			return nil
		}

		// Only retry on server errors (5xx) or timeout-like issues.
		if !isRetryable(lastErr) {
			return lastErr
		}
	}
	return fmt.Errorf("all %d retries exhausted: %w", c.maxRetries, lastErr)
}

// isRetryable determines if an error warrants a retry.
func isRetryable(err error) bool {
	if err == nil {
		return false
	}
	msg := err.Error()
	// Retry on 5xx or connection-level errors.
	return strings.Contains(msg, "HTTP 5") ||
		strings.Contains(msg, "connection refused") ||
		strings.Contains(msg, "timeout") ||
		strings.Contains(msg, "EOF")
}

// backoffDelay returns the delay for the given retry attempt using exponential back-off.
func backoffDelay(attempt int) time.Duration {
	d := retryBaseDelay * time.Duration(math.Pow(2, float64(attempt-1)))
	if d > retryMaxDelay {
		d = retryMaxDelay
	}
	return d
}

// truncate shortens a string to maxLen characters, appending "..." if truncated.
func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}

// ---------- Header helpers ----------

// AuthHeader returns an Authorization Bearer header map.
func AuthHeader(token string) map[string]string {
	return map[string]string{
		"Authorization": "Bearer " + token,
	}
}

// MergeHeaders merges multiple header maps into one. Later values overwrite earlier ones.
func MergeHeaders(maps ...map[string]string) map[string]string {
	merged := make(map[string]string)
	for _, m := range maps {
		for k, v := range m {
			merged[k] = v
		}
	}
	return merged
}
