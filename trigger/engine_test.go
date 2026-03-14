// Package trigger provides tests for trigger matching logic.
package trigger

import (
	"testing"

	"github.com/jarlex/govent/config"
	"github.com/jarlex/govent/event"
)

func TestTriggerMatching(t *testing.T) {
	tests := []struct {
		name        string
		triggerCfg  *config.TriggerConfig
		event       *event.Event
		expectMatch bool
	}{
		{
			name: "Match event by type",
			triggerCfg: &config.TriggerConfig{
				EventType: "user.created",
			},
			event: &event.Event{
				Type: "user.created",
			},
			expectMatch: true,
		},
		{
			name: "No match - different event type",
			triggerCfg: &config.TriggerConfig{
				EventType: "user.created",
			},
			event: &event.Event{
				Type: "user.deleted",
			},
			expectMatch: false,
		},
		{
			name: "Match event by source",
			triggerCfg: &config.TriggerConfig{
				EventType: "user.created",
				Matchers: map[string]string{
					"source": "auth-service",
				},
			},
			event: &event.Event{
				Type:   "user.created",
				Source: "auth-service",
			},
			expectMatch: true,
		},
		{
			name: "No match - different source",
			triggerCfg: &config.TriggerConfig{
				EventType: "user.created",
				Matchers: map[string]string{
					"source": "auth-service",
				},
			},
			event: &event.Event{
				Type:   "user.created",
				Source: "other-service",
			},
			expectMatch: false,
		},
		{
			name: "Match event with multiple matchers - all match",
			triggerCfg: &config.TriggerConfig{
				EventType: "user.created",
				Matchers: map[string]string{
					"source": "auth-service",
				},
			},
			event: &event.Event{
				Type:   "user.created",
				Source: "auth-service",
			},
			expectMatch: true,
		},
		{
			name: "No match - one matcher fails",
			triggerCfg: &config.TriggerConfig{
				EventType: "user.created",
				Matchers: map[string]string{
					"source": "auth-service",
				},
			},
			event: &event.Event{
				Type:   "user.created",
				Source: "other-service",
			},
			expectMatch: false,
		},
		{
			name: "Match with empty eventType but matching source",
			triggerCfg: &config.TriggerConfig{
				EventType: "",
				Matchers: map[string]string{
					"source": "auth-service",
				},
			},
			event: &event.Event{
				Type:   "any-type",
				Source: "auth-service",
			},
			expectMatch: true,
		},
		{
			name: "Match with empty eventType and no matchers",
			triggerCfg: &config.TriggerConfig{
				EventType: "",
				Matchers:  nil,
			},
			event: &event.Event{
				Type: "any-type",
			},
			expectMatch: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := matchesTrigger(tt.triggerCfg, tt.event)
			if result != tt.expectMatch {
				t.Errorf("expected match = %v, got %v", tt.expectMatch, result)
			}
		})
	}
}

func TestEngineMatchEvent(t *testing.T) {
	// Create trigger configs (no wildcards - exact match only)
	triggerConfigs := []config.TriggerConfig{
		{
			Name:      "user-created-trigger",
			EventType: "user.created",
			Matchers: map[string]string{
				"source": "auth-service",
			},
		},
		{
			Name:      "user-any-trigger",
			EventType: "user.anything",
		},
		{
			Name:      "order-created",
			EventType: "order.created",
		},
	}

	engine := NewEngine(triggerConfigs)

	tests := []struct {
		name          string
		event         *event.Event
		expectMatches int
	}{
		{
			name: "Event matches by type only (no matchers)",
			event: &event.Event{
				Type:   "user.created",
				Source: "other-service",
			},
			expectMatches: 0, // No match because source doesn't match "auth-service"
		},
		{
			name: "Event matches by type with matching source",
			event: &event.Event{
				Type:   "user.created",
				Source: "auth-service",
			},
			expectMatches: 1, // Matches "user-created-trigger"
		},
		{
			name: "Event matches exact type",
			event: &event.Event{
				Type: "user.anything",
			},
			expectMatches: 1, // Matches "user-any-trigger"
		},
		{
			name: "Event matches no triggers",
			event: &event.Event{
				Type: "payment.processed",
			},
			expectMatches: 0,
		},
		{
			name: "Event matches by eventType only",
			event: &event.Event{
				Type: "order.created",
			},
			expectMatches: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			matches := engine.MatchEvent(tt.event)
			if len(matches) != tt.expectMatches {
				t.Errorf("expected %d matches, got %d", tt.expectMatches, len(matches))
			}
		})
	}
}

func TestEngineMatchEventWithEmptyTriggers(t *testing.T) {
	// Test engine with no triggers
	engine := NewEngine([]config.TriggerConfig{})

	event := &event.Event{
		Type: "user.created",
	}

	matches := engine.MatchEvent(event)
	if len(matches) != 0 {
		t.Errorf("expected 0 matches, got %d", len(matches))
	}
}
