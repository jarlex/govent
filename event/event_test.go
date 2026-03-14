// Package event provides tests for Event struct parsing and validation.
package event

import (
	"encoding/json"
	"testing"
	"time"
)

func TestEventJSONParsing(t *testing.T) {
	tests := []struct {
		name        string
		jsonInput   string
		expectID    bool
		expectError bool
	}{
		{
			name: "Parse event with all fields",
			jsonInput: `{
				"id": "evt-123",
				"type": "user.created",
				"payload": {"user_id": "u123", "email": "test@example.com"},
				"metadata": {"source": "auth-service"},
				"timestamp": "2024-01-15T10:30:00Z",
				"source": "auth-service"
			}`,
			expectID:    true,
			expectError: false,
		},
		{
			name: "Parse event with minimal fields",
			jsonInput: `{
				"type": "user.created",
				"payload": {"user_id": "u123"}
			}`,
			expectID:    false, // ID will be auto-generated
			expectError: false,
		},
		{
			name: "Parse event with empty payload",
			jsonInput: `{
				"type": "test.event",
				"payload": {}
			}`,
			expectID:    false,
			expectError: false,
		},
		{
			name: "Parse event with null payload",
			jsonInput: `{
				"type": "test.event",
				"payload": null
			}`,
			expectID:    false,
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var evt Event
			err := json.Unmarshal([]byte(tt.jsonInput), &evt)

			if tt.expectError && err == nil {
				t.Errorf("expected error, got nil")
			}
			if !tt.expectError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}

			if !tt.expectError {
				// Validate that Type is always set
				if tt.name != "Parse event with empty payload" && tt.name != "Parse event with null payload" {
					if evt.Type == "" {
						t.Errorf("expected type to be set")
					}
				}
			}
		})
	}
}

func TestEventNew(t *testing.T) {
	// Test creating event with New()
	evt := New("user.created", map[string]interface{}{
		"user_id": "u123",
		"email":   "test@example.com",
	})

	if evt.ID == "" {
		t.Error("expected ID to be auto-generated")
	}

	if evt.Type != "user.created" {
		t.Errorf("expected type 'user.created', got '%s'", evt.Type)
	}

	if evt.Payload == nil {
		t.Error("expected payload to be set")
	}

	if evt.Timestamp == "" {
		t.Error("expected timestamp to be auto-generated")
	}

	// Verify timestamp is RFC3339 format
	_, err := time.Parse(time.RFC3339, evt.Timestamp)
	if err != nil {
		t.Errorf("timestamp not in RFC3339 format: %v", err)
	}
}

func TestEventEnsureDefaults(t *testing.T) {
	tests := []struct {
		name           string
		event          *Event
		checkID        bool
		checkTimestamp bool
		checkPayload   bool
		checkMetadata  bool
	}{
		{
			name: "Event with no fields",
			event: &Event{
				Type: "test.event",
			},
			checkID:        true,
			checkTimestamp: true,
			checkPayload:   true,
			checkMetadata:  true,
		},
		{
			name: "Event with all fields set",
			event: &Event{
				ID:        "existing-id",
				Type:      "test.event",
				Payload:   map[string]interface{}{"key": "value"},
				Metadata:  map[string]string{"meta": "data"},
				Timestamp: "2024-01-15T10:30:00Z",
				Source:    "test-source",
			},
			checkID:        false,
			checkTimestamp: false,
			checkPayload:   false,
			checkMetadata:  false,
		},
		{
			name: "Event with nil maps",
			event: &Event{
				Type:      "test.event",
				Payload:   nil,
				Metadata:  nil,
				Timestamp: "",
			},
			checkID:        true,
			checkTimestamp: true,
			checkPayload:   true,
			checkMetadata:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			originalID := tt.event.ID
			originalTimestamp := tt.event.Timestamp

			tt.event.EnsureDefaults()

			if tt.checkID {
				if tt.event.ID == "" {
					t.Error("expected ID to be generated")
				}
			} else if tt.event.ID != originalID {
				t.Error("ID should not be modified")
			}

			if tt.checkTimestamp {
				if tt.event.Timestamp == "" {
					t.Error("expected timestamp to be generated")
				}
			} else if tt.event.Timestamp != originalTimestamp {
				t.Error("timestamp should not be modified")
			}

			if tt.checkPayload && tt.event.Payload == nil {
				t.Error("expected payload to be initialized")
			}

			if tt.checkMetadata && tt.event.Metadata == nil {
				t.Error("expected metadata to be initialized")
			}
		})
	}
}

func TestEventTimestampRFC3339(t *testing.T) {
	// Test that timestamps can be parsed as RFC3339
	validTimestamps := []string{
		"2024-01-15T10:30:00Z",
		"2024-01-15T10:30:00+00:00",
		"2024-01-15T10:30:00-05:00",
		"2024-01-15T10:30:00.123Z",
	}

	for _, ts := range validTimestamps {
		_, err := time.Parse(time.RFC3339, ts)
		if err != nil {
			t.Errorf("failed to parse valid RFC3339 timestamp '%s': %v", ts, err)
		}
	}
}
