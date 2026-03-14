// Package event provides event validation functions for the govent system.
package event

import (
	"fmt"
	"time"
)

// ValidationError represents an event validation error.
type ValidationError struct {
	Field   string
	Message string
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("validation error on field '%s': %s", e.Field, e.Message)
}

// Validate checks if the event has all required fields and valid values.
func (e *Event) Validate() error {
	// Validate required field: Type
	if e.Type == "" {
		return &ValidationError{Field: "type", Message: "field is required"}
	}

	// Validate required field: Payload
	if e.Payload == nil {
		return &ValidationError{Field: "payload", Message: "field is required"}
	}

	// Validate timestamp format if provided
	if e.Timestamp != "" {
		if _, err := time.Parse(time.RFC3339, e.Timestamp); err != nil {
			return &ValidationError{Field: "timestamp", Message: "must be in RFC 3339 format"}
		}
	}

	return nil
}

// ValidatePartial validates fields that are present but may be invalid.
// This is useful for detecting partial validation issues.
func (e *Event) ValidatePartial() error {
	// If timestamp is present, validate its format
	if e.Timestamp != "" {
		if _, err := time.Parse(time.RFC3339, e.Timestamp); err != nil {
			return &ValidationError{Field: "timestamp", Message: "must be in RFC 3339 format"}
		}
	}

	return nil
}

// IsValid returns true if the event passes validation.
func (e *Event) IsValid() bool {
	return e.Validate() == nil
}
