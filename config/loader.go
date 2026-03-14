// Package config provides YAML configuration loading and validation.
package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// KnownActionTypes contains the list of valid action types.
var KnownActionTypes = map[string]bool{
	"rest": true,
	"grpc": true,
}

// Load reads and parses a YAML configuration file.
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse YAML: %w", err)
	}

	return &cfg, nil
}

// Validate checks the configuration for errors.
func Validate(cfg *Config) error {
	// Check for duplicate trigger names
	seenNames := make(map[string]bool)
	for _, trigger := range cfg.Triggers {
		if seenNames[trigger.Name] {
			return fmt.Errorf("duplicate trigger name: %s", trigger.Name)
		}
		seenNames[trigger.Name] = true

		// Validate required fields
		if trigger.Name == "" {
			return fmt.Errorf("trigger missing required field: name")
		}
		if trigger.EventType == "" {
			return fmt.Errorf("trigger %s missing required field: eventType", trigger.Name)
		}

		// Validate actions
		for i, action := range trigger.Actions {
			if action.Type == "" {
				return fmt.Errorf("trigger %s action %d missing required field: type", trigger.Name, i)
			}
			if !KnownActionTypes[action.Type] {
				return fmt.Errorf("trigger %s action %d: unknown action type: %s", trigger.Name, i, action.Type)
			}
			if action.Config == nil {
				return fmt.Errorf("trigger %s action %d missing required field: config", trigger.Name, i)
			}
		}
	}

	return nil
}

// LoadAndValidate reads, parses, and validates a YAML configuration file.
func LoadAndValidate(path string) (*Config, error) {
	cfg, err := Load(path)
	if err != nil {
		return nil, err
	}

	if err := Validate(cfg); err != nil {
		return nil, err
	}

	return cfg, nil
}
