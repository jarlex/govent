// Package event provides the Event struct and validation for the govent system.
// It defines the core event structure used throughout the system for event
// processing, validation, and trigger matching.
package event

import (
	"time"

	"github.com/google/uuid"
)

// Event represents an event in the govent system.
// It contains all required fields for event processing, including unique
// identification, type classification, payload data, and metadata.
//
// JSON tags are provided for serialization and deserialization.
// The ID and Timestamp fields are auto-generated if not provided.
type Event struct {
	// ID is the unique identifier for the event. If empty, it will be
	// automatically generated as a UUID when EnsureDefaults is called.
	ID string `json:"id"`
	// Type is the event type classification (e.g., "user.created", "order.completed").
	// This field is required and is used for trigger matching.
	Type string `json:"type"`
	// Payload contains the event data as key-value pairs.
	// This is the main content of the event and is passed to actions.
	Payload map[string]interface{} `json:"payload"`
	// Metadata contains additional context about the event.
	// This field is optional and is not used for trigger matching.
	Metadata map[string]string `json:"metadata,omitempty"`
	// Timestamp is the event creation time in RFC 3339 format.
	// If empty, it will be automatically set to the current UTC time
	// when EnsureDefaults is called.
	Timestamp string `json:"timestamp"`
	// Source identifies the origin service that created the event.
	// This field is optional and can be used for trigger matching via matchers.
	Source string `json:"source,omitempty"`
}

// New creates a new Event with the specified type and payload.
// The event ID is generated as a new UUID and the timestamp is set to
// the current UTC time in RFC 3339 format. The Payload map is initialized
// to an empty map if nil is passed.
func New(eventType string, payload map[string]interface{}) *Event {
	now := time.Now().UTC().Format(time.RFC3339)
	return &Event{
		ID:        uuid.New().String(),
		Type:      eventType,
		Payload:   payload,
		Timestamp: now,
	}
}

// EnsureDefaults sets default values for missing fields.
// It generates a UUID for the ID if empty, sets the current UTC timestamp
// in RFC 3339 format if Timestamp is empty, and initializes Payload and
// Metadata to empty maps if they are nil. This method should be called
// before processing an event to ensure all required fields have values.
func (e *Event) EnsureDefaults() {
	if e.ID == "" {
		e.ID = uuid.New().String()
	}
	if e.Timestamp == "" {
		e.Timestamp = time.Now().UTC().Format(time.RFC3339)
	}
	if e.Payload == nil {
		e.Payload = make(map[string]interface{})
	}
	if e.Metadata == nil {
		e.Metadata = make(map[string]string)
	}
}
