// Package trigger provides the trigger engine for matching events and executing actions.
package trigger

import (
	"log"
	"sync"

	"github.com/jarlex/govent/actions"
	"github.com/jarlex/govent/config"
	"github.com/jarlex/govent/event"
)

// Engine manages triggers and executes matching actions for events.
type Engine struct {
	triggers []Trigger
}

// Trigger wraps a config.TriggerConfig with instantiated actions.
type Trigger struct {
	Config  config.TriggerConfig
	Actions []actions.GenericAction
}

// NewEngine creates a new trigger engine from configurations.
func NewEngine(triggerConfigs []config.TriggerConfig) *Engine {
	engine := &Engine{
		triggers: make([]Trigger, 0, len(triggerConfigs)),
	}

	for _, cfg := range triggerConfigs {
		// Create action instances from config
		actionInstances := make([]actions.GenericAction, 0, len(cfg.Actions))
		for _, actionCfg := range cfg.Actions {
			action, err := actions.Factory(actionCfg)
			if err != nil {
				log.Printf("Failed to create action %s for trigger %s: %v", actionCfg.Type, cfg.Name, err)
				continue
			}
			actionInstances = append(actionInstances, action)
		}

		engine.triggers = append(engine.triggers, Trigger{
			Config:  cfg,
			Actions: actionInstances,
		})
	}

	return engine
}

// MatchEvent checks which triggers match the given event.
func (e *Engine) MatchEvent(evt *event.Event) []Trigger {
	matched := make([]Trigger, 0)

	for _, t := range e.triggers {
		if matchesTrigger(&t.Config, evt) {
			matched = append(matched, t)
		}
	}

	return matched
}

// matchesTrigger checks if a trigger configuration matches the event.
func matchesTrigger(cfg *config.TriggerConfig, evt *event.Event) bool {
	// Check eventType match
	if cfg.EventType != "" && evt.Type != cfg.EventType {
		return false
	}

	// Check all matchers (AND logic - all must match)
	for key, value := range cfg.Matchers {
		switch key {
		case "source":
			if evt.Source != value {
				return false
			}
			// Add more matcher types as needed
		}
	}

	return true
}

// ExecuteActions executes all actions for matched triggers asynchronously.
// Action failures do not prevent other actions from running.
func (e *Engine) ExecuteActions(evt *event.Event) {
	matchedTriggers := e.MatchEvent(evt)

	if len(matchedTriggers) == 0 {
		log.Printf("No triggers matched for event %s of type %s", evt.ID, evt.Type)
		return
	}

	log.Printf("Event %s matched %d trigger(s)", evt.ID, len(matchedTriggers))

	// Execute all actions asynchronously
	var wg sync.WaitGroup

	for _, t := range matchedTriggers {
		for _, action := range t.Actions {
			wg.Add(1)
			go executeAction(action, evt, &wg)
		}
	}

	wg.Wait()
}

// executeAction runs a single action asynchronously.
func executeAction(action actions.GenericAction, evt *event.Event, wg *sync.WaitGroup) {
	defer wg.Done()

	log.Printf("Executing action %s for event %s", action.Name(), evt.ID)

	err := action.Handler()(evt)
	if err != nil {
		log.Printf("Action %s failed for event %s: %v", action.Name(), evt.ID, err)
		return
	}

	log.Printf("Action %s completed successfully for event %s", action.Name(), evt.ID)
}
