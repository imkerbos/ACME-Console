package errors

import (
	"errors"
	"fmt"
)

// Sentinel errors
var (
	ErrNotFound       = errors.New("resource not found")
	ErrInvalidInput   = errors.New("invalid input")
	ErrUnauthorized   = errors.New("unauthorized")
	ErrForbidden      = errors.New("forbidden")
	ErrConflict       = errors.New("resource conflict")
	ErrInternalServer = errors.New("internal server error")
)

// AppError represents an application-level error with context
type AppError struct {
	Code    string
	Message string
	Err     error
}

func (e *AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Err)
	}
	return e.Message
}

func (e *AppError) Unwrap() error {
	return e.Err
}

// Error constructors
func NewNotFoundError(resource string) *AppError {
	return &AppError{
		Code:    "NOT_FOUND",
		Message: fmt.Sprintf("%s not found", resource),
		Err:     ErrNotFound,
	}
}

func NewInvalidInputError(field string, reason string) *AppError {
	return &AppError{
		Code:    "INVALID_INPUT",
		Message: fmt.Sprintf("invalid %s: %s", field, reason),
		Err:     ErrInvalidInput,
	}
}

func NewConflictError(message string) *AppError {
	return &AppError{
		Code:    "CONFLICT",
		Message: message,
		Err:     ErrConflict,
	}
}

func NewInternalError(err error) *AppError {
	return &AppError{
		Code:    "INTERNAL_ERROR",
		Message: "internal server error",
		Err:     err,
	}
}

// Is checks if the error matches a target error
func Is(err, target error) bool {
	return errors.Is(err, target)
}

// As attempts to find the first error in err's chain that matches target
func As(err error, target any) bool {
	return errors.As(err, target)
}
