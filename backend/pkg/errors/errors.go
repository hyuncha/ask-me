package errors

import (
	"fmt"
	"net/http"
)

// AppError represents an application error with HTTP status code
type AppError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Status  int    `json:"-"`
	Err     error  `json:"-"`
}

func (e *AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Err)
	}
	return e.Message
}

// Common error constructors
func New(code, message string, status int) *AppError {
	return &AppError{
		Code:    code,
		Message: message,
		Status:  status,
	}
}

func Wrap(err error, code, message string, status int) *AppError {
	return &AppError{
		Code:    code,
		Message: message,
		Status:  status,
		Err:     err,
	}
}

// Predefined errors
var (
	ErrUnauthorized = New("UNAUTHORIZED", "Unauthorized access", http.StatusUnauthorized)
	ErrForbidden    = New("FORBIDDEN", "Access forbidden", http.StatusForbidden)
	ErrNotFound     = New("NOT_FOUND", "Resource not found", http.StatusNotFound)
	ErrBadRequest   = New("BAD_REQUEST", "Invalid request", http.StatusBadRequest)
	ErrInternal     = New("INTERNAL_ERROR", "Internal server error", http.StatusInternalServerError)

	// Domain specific errors
	ErrInvalidCredentials = New("INVALID_CREDENTIALS", "Invalid credentials", http.StatusUnauthorized)
	ErrUserNotFound       = New("USER_NOT_FOUND", "User not found", http.StatusNotFound)
	ErrKnowledgeNotFound  = New("KNOWLEDGE_NOT_FOUND", "Knowledge document not found", http.StatusNotFound)
	ErrSubscriptionRequired = New("SUBSCRIPTION_REQUIRED", "Subscription required for this feature", http.StatusForbidden)
	ErrRateLimitExceeded  = New("RATE_LIMIT_EXCEEDED", "Rate limit exceeded", http.StatusTooManyRequests)
)
