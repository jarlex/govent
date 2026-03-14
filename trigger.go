package govent

import (
	"github.com/jarlex/govent/actions"
	"github.com/jarlex/govent/config"
)

type Protocol int

const (
	GRPC Protocol = iota
	JSON
	TEXT
)

// Trigger represents a trigger with matching logic and actions.
type Trigger struct {
	EventType string                  // Event type to match
	Matchers  map[string]string       // Additional matchers (source, etc.)
	Actions   []actions.GenericAction // Actions to execute when triggered
	Name      string                  // Trigger name for logging
}

// NewTrigger creates a new Trigger from a TriggerConfig.
func NewTrigger(cfg config.TriggerConfig, actionInstances []actions.GenericAction) *Trigger {
	return &Trigger{
		EventType: cfg.EventType,
		Matchers:  cfg.Matchers,
		Actions:   actionInstances,
		Name:      cfg.Name,
	}
}

// Matches checks if the given event matches this trigger's criteria.
func (t *Trigger) Matches(eventType, source string) bool {
	// Check eventType match
	if t.EventType != "" && eventType != t.EventType {
		return false
	}

	// Check all matchers (AND logic - all must match)
	for key, value := range t.Matchers {
		switch key {
		case "source":
			if source != value {
				return false
			}
			// Add more matcher types as needed
		}
	}

	return true
}
