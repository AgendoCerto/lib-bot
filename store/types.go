// Package store provides types for data storage and versioning.
package store

import "errors"

// Static errors for better error handling.
var (
	ErrServiceNotConfigured = errors.New("store service not properly configured")
	ErrValidationFailed     = errors.New("validation failed")
)

// Versioned represents a versioned document.
type Versioned struct {
	ID       string `json:"id"`
	Status   string `json:"status"` // development|production
	Checksum string `json:"checksum"`
	Data     []byte `json:"data"` // Normalized design JSON
}

// ValidationError represents validation errors.
type ValidationError struct {
	Issues []ValidationIssue
}

// Error implements the error interface.
func (e ValidationError) Error() string {
	return "validation failed"
}

// IsValidationError checks if an error is a validation error.
func IsValidationError(err error) bool {
	var validationErr ValidationError
	return errors.As(err, &validationErr)
}
