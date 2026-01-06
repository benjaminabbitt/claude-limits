package api

import (
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"os"
	"runtime"
	"time"

	apierrors "github.com/benjaminabbitt/claude-limits/internal/errors"
	"github.com/benjaminabbitt/claude-limits/internal/models"
	"github.com/benjaminabbitt/claude-limits/internal/version"
)

// DefaultBaseURL is the default Anthropic API endpoint
const DefaultBaseURL = "https://api.anthropic.com"

// Retry configuration
const (
	maxRetries     = 3
	initialBackoff = 500 * time.Millisecond
	maxBackoff     = 5 * time.Second
)

// userAgent returns a User-Agent string matching Claude Code format
func userAgent() string {
	return fmt.Sprintf("claude-code/%s (%s; %s) Go/%s",
		version.Version, runtime.GOOS, runtime.GOARCH, runtime.Version()[2:])
}

// Client is the Anthropic OAuth API client
type Client struct {
	accessToken string
	baseURL     string
	httpClient  *http.Client
}

// ClientOption configures a Client
type ClientOption func(*Client)

// WithBaseURL sets a custom base URL for the API client
func WithBaseURL(baseURL string) ClientOption {
	return func(c *Client) {
		c.baseURL = baseURL
	}
}

// WithHTTPClient sets a custom HTTP client
func WithHTTPClient(httpClient *http.Client) ClientOption {
	return func(c *Client) {
		c.httpClient = httpClient
	}
}

// NewClient creates a new API client with the given OAuth access token.
// The base URL can be overridden via CLAUDE_API_BASE_URL environment variable
// or WithBaseURL option.
func NewClient(accessToken string, opts ...ClientOption) *Client {
	c := &Client{
		accessToken: accessToken,
		baseURL:     DefaultBaseURL,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}

	// Check environment variable for base URL override
	if envURL := os.Getenv("CLAUDE_API_BASE_URL"); envURL != "" {
		c.baseURL = envURL
	}

	// Apply options (can override env var)
	for _, opt := range opts {
		opt(c)
	}

	return c
}

// isRetriable returns true if the status code indicates a retriable error
func isRetriable(statusCode int) bool {
	switch statusCode {
	case http.StatusTooManyRequests,
		http.StatusServiceUnavailable,
		http.StatusGatewayTimeout,
		http.StatusBadGateway:
		return true
	default:
		return statusCode >= 500
	}
}

// backoffDuration calculates exponential backoff with jitter
func backoffDuration(attempt int) time.Duration {
	backoff := float64(initialBackoff) * math.Pow(2, float64(attempt))
	if backoff > float64(maxBackoff) {
		backoff = float64(maxBackoff)
	}
	return time.Duration(backoff)
}

// GetUsage fetches the current usage from Anthropic API with automatic retry
func (c *Client) GetUsage() (*models.Usage, error) {
	reqURL := fmt.Sprintf("%s/api/oauth/usage", c.baseURL)

	var lastErr error
	for attempt := 0; attempt <= maxRetries; attempt++ {
		if attempt > 0 {
			time.Sleep(backoffDuration(attempt - 1))
		}

		usage, err, retry := c.doRequest(reqURL)
		if err == nil {
			return usage, nil
		}
		lastErr = err
		if !retry {
			return nil, err
		}
	}

	return nil, fmt.Errorf("request failed after %d retries: %w", maxRetries, lastErr)
}

// doRequest performs a single HTTP request and returns whether it should be retried
func (c *Client) doRequest(reqURL string) (*models.Usage, error, bool) {
	req, err := http.NewRequest("GET", reqURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err), false
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", userAgent())
	req.Header.Set("Authorization", "Bearer "+c.accessToken)
	req.Header.Set("anthropic-beta", "oauth-2025-04-20")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		// Network errors are retriable
		return nil, fmt.Errorf("failed to make request: %w", err), true
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		retriable := isRetriable(resp.StatusCode)
		msg := http.StatusText(resp.StatusCode)
		if len(body) > 0 {
			var errResp struct {
				Error string `json:"error"`
			}
			if json.Unmarshal(body, &errResp) == nil && errResp.Error != "" {
				msg = errResp.Error
			}
		}
		return nil, apierrors.NewAPIError(resp.StatusCode, msg, retriable), retriable
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err), true
	}

	var usage models.Usage
	if err := json.Unmarshal(body, &usage); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err), false
	}

	return &usage, nil, false
}
