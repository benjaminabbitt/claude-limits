package errors

import (
	"errors"
	"fmt"
)

// Sentinel errors for programmatic error handling
var (
	ErrAuthRequired     = errors.New("authentication required")
	ErrCookieNotFound   = errors.New("session cookie not found")
	ErrOrgIDNotFound    = errors.New("organization ID not found")
	ErrCacheExpired     = errors.New("cache expired")
	ErrNoMatch          = errors.New("no match found")
	ErrRequestFailed    = errors.New("request failed")
	ErrResponseParse    = errors.New("failed to parse response")
)

// AuthError represents an authentication-related error
type AuthError struct {
	Source string // "flag", "env", "browser"
	Err    error
}

func (e *AuthError) Error() string {
	return fmt.Sprintf("authentication error (%s): %v", e.Source, e.Err)
}

func (e *AuthError) Unwrap() error {
	return e.Err
}

// NewAuthError creates a new AuthError
func NewAuthError(source string, err error) *AuthError {
	return &AuthError{Source: source, Err: err}
}

// APIError represents an API request error with status code
type APIError struct {
	StatusCode int
	Message    string
	Retriable  bool
}

func (e *APIError) Error() string {
	return fmt.Sprintf("API error (status %d): %s", e.StatusCode, e.Message)
}

// NewAPIError creates a new APIError
func NewAPIError(statusCode int, message string, retriable bool) *APIError {
	return &APIError{
		StatusCode: statusCode,
		Message:    message,
		Retriable:  retriable,
	}
}

// CacheError represents a cache-related error
type CacheError struct {
	Operation string // "read", "write", "parse"
	Path      string
	Err       error
}

func (e *CacheError) Error() string {
	return fmt.Sprintf("cache %s error (%s): %v", e.Operation, e.Path, e.Err)
}

func (e *CacheError) Unwrap() error {
	return e.Err
}

// NewCacheError creates a new CacheError
func NewCacheError(operation, path string, err error) *CacheError {
	return &CacheError{Operation: operation, Path: path, Err: err}
}

// QueryError represents a search/query error
type QueryError struct {
	Query string
	Err   error
}

func (e *QueryError) Error() string {
	return fmt.Sprintf("query error for %q: %v", e.Query, e.Err)
}

func (e *QueryError) Unwrap() error {
	return e.Err
}

// NewQueryError creates a new QueryError
func NewQueryError(query string, err error) *QueryError {
	return &QueryError{Query: query, Err: err}
}

// Is checks if target error matches any of our sentinel errors
func Is(err, target error) bool {
	return errors.Is(err, target)
}

// As finds the first error in err's chain that matches target
func As(err error, target interface{}) bool {
	return errors.As(err, target)
}
