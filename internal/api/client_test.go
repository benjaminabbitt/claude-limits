package api

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestNewClient(t *testing.T) {
	c := NewClient("test-token")

	if c.accessToken != "test-token" {
		t.Errorf("accessToken = %q, want %q", c.accessToken, "test-token")
	}
	if c.baseURL != DefaultBaseURL {
		t.Errorf("baseURL = %q, want %q", c.baseURL, DefaultBaseURL)
	}
}

func TestNewClientWithOptions(t *testing.T) {
	customURL := "https://custom.example.com"
	customClient := &http.Client{Timeout: 60 * time.Second}

	c := NewClient("token",
		WithBaseURL(customURL),
		WithHTTPClient(customClient),
	)

	if c.baseURL != customURL {
		t.Errorf("baseURL = %q, want %q", c.baseURL, customURL)
	}
	if c.httpClient != customClient {
		t.Error("httpClient not set correctly")
	}
}

func TestNewClientWithEnvVar(t *testing.T) {
	// Set env var
	t.Setenv("CLAUDE_API_BASE_URL", "https://env.example.com")

	c := NewClient("token")

	if c.baseURL != "https://env.example.com" {
		t.Errorf("baseURL = %q, want %q", c.baseURL, "https://env.example.com")
	}
}

func TestNewClientOptionOverridesEnv(t *testing.T) {
	t.Setenv("CLAUDE_API_BASE_URL", "https://env.example.com")

	c := NewClient("token", WithBaseURL("https://option.example.com"))

	if c.baseURL != "https://option.example.com" {
		t.Errorf("baseURL = %q, want option URL", c.baseURL)
	}
}

func TestIsRetriable(t *testing.T) {
	tests := []struct {
		statusCode int
		expected   bool
	}{
		{http.StatusOK, false},
		{http.StatusBadRequest, false},
		{http.StatusUnauthorized, false},
		{http.StatusForbidden, false},
		{http.StatusNotFound, false},
		{http.StatusTooManyRequests, true},
		{http.StatusInternalServerError, true},
		{http.StatusBadGateway, true},
		{http.StatusServiceUnavailable, true},
		{http.StatusGatewayTimeout, true},
	}

	for _, tt := range tests {
		result := isRetriable(tt.statusCode)
		if result != tt.expected {
			t.Errorf("isRetriable(%d) = %v, want %v", tt.statusCode, result, tt.expected)
		}
	}
}

func TestBackoffDuration(t *testing.T) {
	// First attempt: 500ms
	d0 := backoffDuration(0)
	if d0 != 500*time.Millisecond {
		t.Errorf("backoffDuration(0) = %v, want 500ms", d0)
	}

	// Second attempt: 1000ms
	d1 := backoffDuration(1)
	if d1 != 1000*time.Millisecond {
		t.Errorf("backoffDuration(1) = %v, want 1000ms", d1)
	}

	// Should cap at maxBackoff
	d10 := backoffDuration(10)
	if d10 > maxBackoff {
		t.Errorf("backoffDuration(10) = %v, should not exceed %v", d10, maxBackoff)
	}
}

func TestGetUsageSuccess(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify request
		if r.Method != "GET" {
			t.Errorf("Method = %s, want GET", r.Method)
		}
		if r.URL.Path != "/api/oauth/usage" {
			t.Errorf("Path = %s, want /api/oauth/usage", r.URL.Path)
		}

		// Check Authorization header
		auth := r.Header.Get("Authorization")
		if auth != "Bearer test-token" {
			t.Errorf("Authorization = %q, want 'Bearer test-token'", auth)
		}

		// Check anthropic-beta header
		beta := r.Header.Get("anthropic-beta")
		if beta != "oauth-2025-04-20" {
			t.Errorf("anthropic-beta = %q, want 'oauth-2025-04-20'", beta)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"five_hour": {"utilization": 75.5}}`))
	}))
	defer server.Close()

	c := NewClient("test-token", WithBaseURL(server.URL))
	usage, err := c.GetUsage()

	if err != nil {
		t.Fatalf("GetUsage failed: %v", err)
	}
	if usage == nil {
		t.Fatal("GetUsage returned nil usage")
	}
}

func TestGetUsageRetry(t *testing.T) {
	attempts := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		if attempts < 3 {
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"test": "data"}`))
	}))
	defer server.Close()

	c := NewClient("token", WithBaseURL(server.URL))
	usage, err := c.GetUsage()

	if err != nil {
		t.Fatalf("GetUsage failed after retries: %v", err)
	}
	if usage == nil {
		t.Fatal("GetUsage returned nil usage")
	}
	if attempts != 3 {
		t.Errorf("Expected 3 attempts, got %d", attempts)
	}
}

func TestGetUsageNonRetriableError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer server.Close()

	c := NewClient("token", WithBaseURL(server.URL))
	_, err := c.GetUsage()

	if err == nil {
		t.Error("GetUsage should fail on 401")
	}
}

func TestUserAgent(t *testing.T) {
	ua := userAgent()
	if ua == "" {
		t.Error("userAgent returned empty string")
	}
	// Should contain product name
	if len(ua) < 10 {
		t.Error("userAgent too short")
	}
}
