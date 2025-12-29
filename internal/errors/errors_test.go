package errors

import (
	"errors"
	"testing"
)

func TestAuthError(t *testing.T) {
	err := NewAuthError("browser", ErrCookieNotFound)

	// Test Error() method
	msg := err.Error()
	if msg == "" {
		t.Error("AuthError.Error() returned empty string")
	}

	// Test Unwrap
	if !errors.Is(err, ErrCookieNotFound) {
		t.Error("AuthError should unwrap to ErrCookieNotFound")
	}

	// Test source
	if err.Source != "browser" {
		t.Errorf("AuthError.Source = %q, want %q", err.Source, "browser")
	}
}

func TestAPIError(t *testing.T) {
	err := NewAPIError(503, "Service Unavailable", true)

	if err.StatusCode != 503 {
		t.Errorf("APIError.StatusCode = %d, want 503", err.StatusCode)
	}
	if !err.Retriable {
		t.Error("APIError.Retriable should be true")
	}

	msg := err.Error()
	if msg == "" {
		t.Error("APIError.Error() returned empty string")
	}
}

func TestCacheError(t *testing.T) {
	underlying := errors.New("file not found")
	err := NewCacheError("read", "/path/to/cache", underlying)

	if err.Operation != "read" {
		t.Errorf("CacheError.Operation = %q, want %q", err.Operation, "read")
	}
	if err.Path != "/path/to/cache" {
		t.Errorf("CacheError.Path = %q, want %q", err.Path, "/path/to/cache")
	}

	// Test Unwrap
	if !errors.Is(err, underlying) {
		t.Error("CacheError should unwrap to underlying error")
	}

	msg := err.Error()
	if msg == "" {
		t.Error("CacheError.Error() returned empty string")
	}
}

func TestQueryError(t *testing.T) {
	err := NewQueryError("five_hour", ErrNoMatch)

	if err.Query != "five_hour" {
		t.Errorf("QueryError.Query = %q, want %q", err.Query, "five_hour")
	}

	// Test Unwrap
	if !errors.Is(err, ErrNoMatch) {
		t.Error("QueryError should unwrap to ErrNoMatch")
	}

	msg := err.Error()
	if msg == "" {
		t.Error("QueryError.Error() returned empty string")
	}
}

func TestSentinelErrors(t *testing.T) {
	// Verify sentinel errors are distinct
	sentinels := []error{
		ErrAuthRequired,
		ErrCookieNotFound,
		ErrOrgIDNotFound,
		ErrCacheExpired,
		ErrNoMatch,
		ErrRequestFailed,
		ErrResponseParse,
	}

	for i, a := range sentinels {
		for j, b := range sentinels {
			if i != j && errors.Is(a, b) {
				t.Errorf("Sentinel errors %v and %v should not match", a, b)
			}
		}
	}
}

func TestIsAndAs(t *testing.T) {
	err := NewAuthError("browser", ErrCookieNotFound)

	// Test Is wrapper
	if !Is(err, ErrCookieNotFound) {
		t.Error("Is should find wrapped ErrCookieNotFound")
	}

	// Test As wrapper
	var authErr *AuthError
	if !As(err, &authErr) {
		t.Error("As should find AuthError")
	}
	if authErr.Source != "browser" {
		t.Errorf("AuthError.Source = %q, want %q", authErr.Source, "browser")
	}
}
