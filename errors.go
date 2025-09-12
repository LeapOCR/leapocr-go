package ocr

import (
	"fmt"
	"net/http"
)

// ErrorType represents different types of SDK errors
type ErrorType string

const (
	// ErrorTypeInvalidConfig represents configuration validation errors
	ErrorTypeInvalidConfig ErrorType = "invalid_config"
	// ErrorTypeValidationError represents input validation errors
	ErrorTypeValidationError ErrorType = "validation_error"
	// ErrorTypeHTTPError represents HTTP-level errors
	ErrorTypeHTTPError ErrorType = "http_error"
	// ErrorTypeAPIError represents API-level errors
	ErrorTypeAPIError ErrorType = "api_error"
	// ErrorTypeTimeout represents timeout errors
	ErrorTypeTimeout ErrorType = "timeout"
	// ErrorTypeUploadError represents file upload errors
	ErrorTypeUploadError ErrorType = "upload_error"
	// ErrorTypeJobError represents job processing errors
	ErrorTypeJobError ErrorType = "job_error"
	// ErrorTypeUnknown represents unknown errors
	ErrorTypeUnknown ErrorType = "unknown"
)

// SDKError is the main error type for the SDK
type SDKError struct {
	Type       ErrorType
	Message    string
	StatusCode int
	Cause      error
}

// Error implements the error interface
func (e *SDKError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("%s: %s (caused by: %v)", e.Type, e.Message, e.Cause)
	}
	return fmt.Sprintf("%s: %s", e.Type, e.Message)
}

// Unwrap returns the underlying cause
func (e *SDKError) Unwrap() error {
	return e.Cause
}

// IsTimeout returns true if the error is a timeout error
func (e *SDKError) IsTimeout() bool {
	return e.Type == ErrorTypeTimeout
}

// IsHTTPError returns true if the error is an HTTP error
func (e *SDKError) IsHTTPError() bool {
	return e.Type == ErrorTypeHTTPError
}

// IsAPIError returns true if the error is an API error
func (e *SDKError) IsAPIError() bool {
	return e.Type == ErrorTypeAPIError
}

// IsValidationError returns true if the error is a validation error
func (e *SDKError) IsValidationError() bool {
	return e.Type == ErrorTypeValidationError
}

// NewSDKError creates a new SDK error
func NewSDKError(errorType ErrorType, message string, cause error) *SDKError {
	return &SDKError{
		Type:    errorType,
		Message: message,
		Cause:   cause,
	}
}

// NewHTTPError creates a new HTTP error with status code
func NewHTTPError(statusCode int, message string, cause error) *SDKError {
	return &SDKError{
		Type:       ErrorTypeHTTPError,
		Message:    message,
		StatusCode: statusCode,
		Cause:      cause,
	}
}

// WrapHTTPError wraps an HTTP response error
func WrapHTTPError(resp *http.Response, cause error) *SDKError {
	message := fmt.Sprintf("HTTP request failed with status %d", resp.StatusCode)
	if resp.Status != "" {
		message = fmt.Sprintf("HTTP request failed: %s", resp.Status)
	}

	return &SDKError{
		Type:       ErrorTypeHTTPError,
		Message:    message,
		StatusCode: resp.StatusCode,
		Cause:      cause,
	}
}

// IsRetryable returns true if the error is retryable
func (e *SDKError) IsRetryable() bool {
	switch e.Type {
	case ErrorTypeTimeout, ErrorTypeHTTPError:
		// Retry on timeout and certain HTTP errors
		if e.Type == ErrorTypeHTTPError {
			return e.StatusCode >= 500 || e.StatusCode == 408 || e.StatusCode == 429
		}
		return true
	default:
		return false
	}
}
