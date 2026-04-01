package api

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"
)

// Common errors

var (
	ErrMissingAPIKey    = errors.New("ANTHROPIC_API_KEY is required")
	ErrInvalidRequest   = errors.New("invalid request")
	ErrRateLimited      = errors.New("rate limited")
	ErrAuthentication   = errors.New("authentication failed")
	ErrForbidden        = errors.New("access forbidden")
	ErrNotFound         = errors.New("resource not found")
	ErrInternalError    = errors.New("internal server error")
	ErrOverloaded       = errors.New("service overloaded")
	ErrContextTooLong   = errors.New("conversation too long")
	ErrInvalidModel     = errors.New("invalid model")
)

// ErrorConverter converts API errors to user-friendly messages
func ErrorConverter(err error) error {
	var apiErr *APIError
	if errors.As(err, &apiErr) {
		switch apiErr.Type {
		case "authentication_error":
			return fmt.Errorf("%w: please check your API key", ErrAuthentication)
		case "authorization_error":
			return fmt.Errorf("%w: you don't have access to this resource", ErrForbidden)
		case "not_found_error":
			return fmt.Errorf("%w: the requested resource was not found", ErrNotFound)
		case "rate_limit_error":
			return fmt.Errorf("%w: please wait before making more requests", ErrRateLimited)
		case "overloaded_error":
			return fmt.Errorf("%w: please try again later", ErrOverloaded)
		case "invalid_request_error":
			if strings.Contains(apiErr.Message, "messages too long") {
				return fmt.Errorf("%w: conversation exceeds model's context window", ErrContextTooLong)
			}
			return fmt.Errorf("%w: %s", ErrInvalidRequest, apiErr.Message)
		case "internal_server_error":
			return fmt.Errorf("%w: an unexpected error occurred", ErrInternalError)
		default:
			return apiErr
		}
	}
	return err
}

// GetErrorMessage returns a user-friendly error message
func GetErrorMessage(err error) string {
	err = ErrorConverter(err)
	if err == nil {
		return "An unknown error occurred"
	}
	return err.Error()
}

// IsRetryable returns true if the error is retryable
func IsRetryable(err error) bool {
	var apiErr *APIError
	if errors.As(err, &apiErr) {
		switch apiErr.Type {
		case "rate_limit_error", "overloaded_error", "internal_server_error":
			return true
		}
	}
	return false
}

// ShouldRetry returns true if the request should be retried
func ShouldRetry(err error, attempt, maxAttempts int) bool {
	if attempt >= maxAttempts {
		return false
	}

	var apiErr *APIError
	if errors.As(err, &apiErr) {
		switch apiErr.Type {
		case "rate_limit_error", "overloaded_error", "internal_server_error":
			return true
		}
	}

	// Network errors are retryable
	if strings.Contains(err.Error(), "connection") ||
		strings.Contains(err.Error(), "timeout") ||
		strings.Contains(err.Error(), "network") {
		return true
	}

	return false
}

// GetRetryAfter returns the duration to wait before retrying
func GetRetryAfter(err error) time.Duration {
	var apiErr *APIError
	if errors.As(err, &apiErr) {
		// Try to parse retry-after from error
		// In production, this would come from headers
		switch apiErr.Type {
		case "rate_limit_error":
			return 5 * time.Second
		case "overloaded_error":
			return 10 * time.Second
		}
	}
	return time.Second
}

// HTTPStatusCode returns the HTTP status code for an error
func HTTPStatusCode(err error) int {
	var apiErr *APIError
	if errors.As(err, &apiErr) {
		switch apiErr.Type {
		case "authentication_error":
			return http.StatusUnauthorized
		case "authorization_error":
			return http.StatusForbidden
		case "not_found_error":
			return http.StatusNotFound
		case "rate_limit_error":
			return http.StatusTooManyRequests
		case "overloaded_error":
			return http.StatusServiceUnavailable
		case "invalid_request_error":
			return http.StatusBadRequest
		case "internal_server_error":
			return http.StatusInternalServerError
		}
	}
	return http.StatusInternalServerError
}

// APIErrorWithCode is an API error with an HTTP status code
type APIErrorWithCode struct {
	Code    int
	Message string
}

func (e *APIErrorWithCode) Error() string {
	return e.Message
}

// WrapAPIError wraps an error with HTTP status code information
func WrapAPIError(err error, statusCode int) error {
	return &APIErrorWithCode{
		Code:    statusCode,
		Message: err.Error(),
	}
}
