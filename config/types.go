// Package config provides types and loading for YAML-based trigger configuration.
// It defines the configuration structures used to configure triggers and actions
// in the govent system.
package config

// ActionConfig represents the configuration for a single action.
// Actions are executed when a trigger's conditions are met.
//
// The Type field specifies the action implementation to use (e.g., "rest" or "grpc").
// The Config field contains implementation-specific settings.
type ActionConfig struct {
	// Type specifies the action type. Supported values are "rest" for HTTP actions
	// and "grpc" for gRPC actions.
	Type string `yaml:"type"`
	// Config contains the action-specific configuration as key-value pairs.
	// For REST actions, this typically includes url, method, headers, and timeout.
	// For gRPC actions, this typically includes address, service, method, and timeout.
	Config map[string]interface{} `yaml:"config"`
}

// TriggerConfig represents the configuration for a single trigger.
// A trigger defines when its actions should be executed based on event matching.
type TriggerConfig struct {
	// Name is a unique identifier for this trigger within the configuration.
	Name string `yaml:"name"`
	// EventType specifies the event type that this trigger responds to.
	// An empty EventType matches all events (use with caution).
	EventType string `yaml:"eventType"`
	// Matchers provides additional matching criteria beyond EventType.
	// Currently supported matchers:
	//   - "source": matches the event's Source field exactly
	// All matchers must match (AND logic) for the trigger to fire.
	Matchers map[string]string `yaml:"matchers"`
	// Actions is the list of actions to execute when this trigger matches an event.
	// All actions are executed asynchronously; individual action failures
	// do not prevent other actions from running.
	Actions []ActionConfig `yaml:"actions"`
}

// Config represents the root configuration structure for the govent system.
// It contains a list of trigger configurations that define how events are processed.
type Config struct {
	// Triggers is the list of trigger configurations.
	// Each trigger defines matching criteria and actions to execute.
	Triggers []TriggerConfig `yaml:"triggers"`
}
