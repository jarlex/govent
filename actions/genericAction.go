// Package actions provides action implementations and a factory for creating
// actions from configuration. It supports REST and gRPC actions, as well as
// plugin-based custom actions.
package actions

import (
	"plugin"
)

// GenericAction defines the interface that all action implementations must satisfy.
// Actions are used to handle events when triggers match.
type GenericAction interface {
	// Name returns the action's display name.
	Name() string
	// Description returns a human-readable description of what the action does.
	Description() string
	// Type returns the type identifier for this action (e.g., "rest", "grpc").
	Type() string
	// Init initializes the action with the provided configuration map.
	Init(config map[string]interface{}) error
	// Handler returns a function that executes the action with the given input.
	// The function returns an error if the action fails.
	Handler() func(interface{}) error
}

// New loads a custom action from a Go plugin (.so) file.
// The plugin must export a symbol named "Plugin" that implements GenericAction.
// Returns an error if the plugin cannot be opened or does not implement GenericAction.
func New(path string) (GenericAction, error) {
	so, err := plugin.Open(path)
	if err != nil {
		return nil, err
	}

	act, err := so.Lookup("Plugin")
	if err != nil {
		return nil, err
	}
	return act.(GenericAction), nil
}

// Execute runs the provided action with the given input.
// This is a convenience function that calls action.Handler() and executes the result.
func Execute(ac GenericAction, input interface{}) error {
	// Use the generic Handler() which returns func(interface{}) error
	return ac.Handler()(input)
}
