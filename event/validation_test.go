// Package event provides tests for event validation.
package event

import (
	"strings"
	"testing"
)

func TestEventValidation(t *testing.T) {
	tests := []struct {
		name        string
		event       *Event
		expectError bool
		errorField  string
	}{
		{
			name: "Valid event with all required fields",
			event: &Event{
				Type:      "user.created",
				Payload:   map[string]interface{}{"user_id": "u123"},
				Timestamp: "2024-01-15T10:30:00Z",
			},
			expectError: false,
		},
		{
			name: "Valid event with minimal fields",
			event: &Event{
				Type:    "user.created",
				Payload: map[string]interface{}{},
			},
			expectError: false,
		},
		{
			name: "Invalid event - missing type",
			event: &Event{
				Payload: map[string]interface{}{"user_id": "u123"},
			},
			expectError: true,
			errorField:  "type",
		},
		{
			name: "Invalid event - empty type",
			event: &Event{
				Type:    "",
				Payload: map[string]interface{}{"user_id": "u123"},
			},
			expectError: true,
			errorField:  "type",
		},
		{
			name: "Invalid event - nil payload",
			event: &Event{
				Type:    "user.created",
				Payload: nil,
			},
			expectError: true,
			errorField:  "payload",
		},
		{
			name: "Invalid event - invalid timestamp format",
			event: &Event{
				Type:      "user.created",
				Payload:   map[string]interface{}{"user_id": "u123"},
				Timestamp: "not-a-valid-timestamp",
			},
			expectError: true,
			errorField:  "timestamp",
		},
		{
			name: "Valid event - empty timestamp (optional)",
			event: &Event{
				Type:      "user.created",
				Payload:   map[string]interface{}{"user_id": "u123"},
				Timestamp: "",
			},
			expectError: false,
		},
		{
			name: "Valid event - RFC3339 timestamp",
			event: &Event{
				Type:      "user.created",
				Payload:   map[string]interface{}{"user_id": "u123"},
				Timestamp: "2024-01-15T10:30:00Z",
			},
			expectError: false,
		},
		{
			name: "Valid event - RFC3339 with timezone",
			event: &Event{
				Type:      "user.created",
				Payload:   map[string]interface{}{"user_id": "u123"},
				Timestamp: "2024-01-15T10:30:00+05:30",
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.event.Validate()

			if tt.expectError && err == nil {
				t.Errorf("expected error, got nil")
			}
			if !tt.expectError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}

			if tt.expectError && err != nil {
				if !strings.Contains(err.Error(), tt.errorField) {
					t.Errorf("expected error to mention '%s', got '%s'", tt.errorField, err.Error())
				}
			}
		})
	}
}

func TestEventValidatePartial(t *testing.T) {
	tests := []struct {
		name        string
		event       *Event
		expectError bool
	}{
		{
			name: "Event with valid timestamp",
			event: &Event{
				Type:      "user.created",
				Timestamp: "2024-01-15T10:30:00Z",
			},
			expectError: false,
		},
		{
			name: "Event with invalid timestamp",
			event: &Event{
				Type:      "user.created",
				Timestamp: "invalid",
			},
			expectError: true,
		},
		{
			name: "Event with empty timestamp",
			event: &Event{
				Type:      "user.created",
				Timestamp: "",
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.event.ValidatePartial()

			if tt.expectError && err == nil {
				t.Errorf("expected error, got nil")
			}
			if !tt.expectError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestEventIsValid(t *testing.T) {
	tests := []struct {
		name     string
		event    *Event
		expected bool
	}{
		{
			name: "Valid event returns true",
			event: &Event{
				Type:    "user.created",
				Payload: map[string]interface{}{"user_id": "u123"},
			},
			expected: true,
		},
		{
			name: "Invalid event returns false",
			event: &Event{
				Type:    "",
				Payload: map[string]interface{}{"user_id": "u123"},
			},
			expected: false,
		},
		{
			name: "Invalid event with nil payload returns false",
			event: &Event{
				Type:    "user.created",
				Payload: nil,
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.event.IsValid()
			if result != tt.expected {
				t.Errorf("expected IsValid() = %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestValidationError(t *testing.T) {
	err := &ValidationError{
		Field:   "type",
		Message: "field is required",
	}

	expectedMsg := "validation error on field 'type': field is required"
	if err.Error() != expectedMsg {
		t.Errorf("expected error message '%s', got '%s'", expectedMsg, err.Error())
	}
}
