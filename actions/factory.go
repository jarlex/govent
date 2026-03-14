// Package actions provides action implementations and factory for creating actions from configuration.
package actions

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/jarlex/govent/config"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// Factory creates an action instance from an ActionConfig.
func Factory(cfg config.ActionConfig) (GenericAction, error) {
	switch cfg.Type {
	case "rest":
		return NewRestActionFromConfig(cfg.Config)
	case "grpc":
		return NewGrpcActionFromConfig(cfg.Config)
	default:
		return nil, fmt.Errorf("unknown action type: %s", cfg.Type)
	}
}

// restActionConfig represents the REST action configuration.
type restActionConfig struct {
	URL     string
	Method  string
	Headers map[string]string
	Timeout int
}

// NewRestActionFromConfig creates a new REST action from configuration map.
func NewRestActionFromConfig(configMap map[string]interface{}) (GenericAction, error) {
	cfg := restActionConfig{
		Method:  "POST",
		Timeout: 30,
	}

	// Validate required URL
	if url, ok := configMap["url"].(string); ok && url != "" {
		cfg.URL = url
	} else {
		return nil, fmt.Errorf("REST action requires 'url' in config")
	}

	if method, ok := configMap["method"].(string); ok {
		cfg.Method = method
	}
	if headers, ok := configMap["headers"].(map[string]interface{}); ok {
		cfg.Headers = make(map[string]string)
		for k, v := range headers {
			if strVal, ok := v.(string); ok {
				cfg.Headers[k] = strVal
			}
		}
	}
	if timeout, ok := configMap["timeout"].(int); ok {
		cfg.Timeout = timeout
	}

	// Create and initialize the REST action
	action := &restActionImpl{
		name:        "REST Action",
		description: fmt.Sprintf("HTTP %s request to %s", cfg.Method, cfg.URL),
		url:         cfg.URL,
		method:      cfg.Method,
		headers:     cfg.Headers,
		timeout:     cfg.Timeout,
	}

	return action, nil
}

// restActionImpl is the concrete implementation of a REST action.
type restActionImpl struct {
	name        string
	description string
	url         string
	method      string
	headers     map[string]string
	timeout     int
}

func (a *restActionImpl) Name() string        { return a.name }
func (a *restActionImpl) Description() string { return a.description }
func (a *restActionImpl) Type() string        { return restType }

func (a *restActionImpl) Init(config map[string]interface{}) error {
	// Configuration is already handled in NewRestActionFromConfig
	// This method exists to satisfy the GenericAction interface
	return nil
}

// Handler returns a function that executes the REST action.
// It sends an HTTP request with the event as JSON body.
func (a *restActionImpl) Handler() func(interface{}) error {
	return func(input interface{}) error {
		// Get event from input - support both event.Event and generic map
		var body []byte
		var err error

		switch v := input.(type) {
		case map[string]interface{}:
			// Serialize generic map to JSON
			body, err = json.Marshal(v)
			if err != nil {
				return fmt.Errorf("failed to marshal input: %w", err)
			}
		default:
			// Try to serialize as JSON
			body, err = json.Marshal(v)
			if err != nil {
				return fmt.Errorf("failed to marshal input: %w", err)
			}
		}

		// Create HTTP request
		req, err := http.NewRequest(a.method, a.url, bytes.NewReader(body))
		if err != nil {
			return fmt.Errorf("failed to create request: %w", err)
		}

		// Set headers
		req.Header.Set("Content-Type", "application/json")
		for key, value := range a.headers {
			req.Header.Set(key, value)
		}

		// Create HTTP client with timeout
		client := &http.Client{
			Timeout: time.Duration(a.timeout) * time.Second,
		}

		// Send request
		resp, err := client.Do(req)
		if err != nil {
			return fmt.Errorf("request failed: %w", err)
		}
		defer resp.Body.Close()

		// Handle response codes - 4.4: Handle HTTP response codes (200 = success, 500 = failure)
		if resp.StatusCode >= 200 && resp.StatusCode < 300 {
			// Success (2xx)
			return nil
		}

		// Failure (non-2xx)
		return fmt.Errorf("HTTP %d: request failed", resp.StatusCode)
	}
}

// grpcActionConfig represents the gRPC action configuration.
type grpcActionConfig struct {
	Address string
	Service string
	Method  string
	Timeout int
}

// grpcActionImpl is the concrete implementation of a gRPC action.
type grpcActionImpl struct {
	name        string
	description string
	address     string
	service     string
	method      string
	timeout     int
}

// NewGrpcActionFromConfig creates a new gRPC action from configuration map.
func NewGrpcActionFromConfig(configMap map[string]interface{}) (GenericAction, error) {
	cfg := grpcActionConfig{
		Timeout: 30,
	}

	// Validate required address
	if address, ok := configMap["address"].(string); ok && address != "" {
		cfg.Address = address
	} else {
		return nil, fmt.Errorf("gRPC action requires 'address' in config")
	}

	// Validate required service
	if service, ok := configMap["service"].(string); ok && service != "" {
		cfg.Service = service
	} else {
		return nil, fmt.Errorf("gRPC action requires 'service' in config")
	}

	if method, ok := configMap["method"].(string); ok {
		cfg.Method = method
	}
	if timeout, ok := configMap["timeout"].(int); ok {
		cfg.Timeout = timeout
	}

	// Create and initialize the gRPC action
	action := &grpcActionImpl{
		name:        "gRPC Action",
		description: fmt.Sprintf("gRPC call to %s/%s", cfg.Address, cfg.Service),
		address:     cfg.Address,
		service:     cfg.Service,
		method:      cfg.Method,
		timeout:     cfg.Timeout,
	}

	return action, nil
}

func (a *grpcActionImpl) Name() string        { return a.name }
func (a *grpcActionImpl) Description() string { return a.description }
func (a *grpcActionImpl) Type() string        { return "grpc" }

func (a *grpcActionImpl) Init(config map[string]interface{}) error {
	// Configuration is already handled in NewGrpcActionFromConfig
	// This method exists to satisfy the GenericAction interface
	return nil
}

// Handler returns a function that executes the gRPC action.
// 4.5-4.8: Implement gRPC connection, timeout, and error handling.
func (a *grpcActionImpl) Handler() func(interface{}) error {
	return func(input interface{}) error {
		// 4.6: Implement gRPC connection with address and service
		// Create gRPC connection with insecure credentials (for MVP)
		conn, err := grpc.Dial(
			a.address,
			grpc.WithTransportCredentials(insecure.NewCredentials()),
			grpc.WithBlock(),
		)
		if err != nil {
			// 4.8: Handle gRPC connection failures
			return fmt.Errorf("gRPC connection failed to %s: %w", a.address, err)
		}
		defer conn.Close()

		// 4.7: Implement timeout handling for gRPC actions
		ctx, cancel := context.WithTimeout(context.Background(), time.Duration(a.timeout)*time.Second)
		defer cancel()

		// Serialize input to JSON (as a simple payload for MVP)
		// In a full implementation, this would use protobuf serialization
		payload, err := json.Marshal(input)
		if err != nil {
			return fmt.Errorf("failed to marshal input: %w", err)
		}

		// For MVP, we'll just verify the connection is working
		// A full implementation would invoke the actual gRPC service method
		// This demonstrates the connection and timeout work correctly
		_ = payload // payload would be sent via gRPC in full implementation

		// Simple health check - verify connection is alive
		// In production, this would call the actual service method
		err = conn.Invoke(ctx, a.service+"/"+a.method, payload, payload)
		if err != nil {
			// Return error for connection failures or method invocation failures
			return fmt.Errorf("gRPC invoke failed: %w", err)
		}

		return nil
	}
}
