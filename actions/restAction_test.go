// Package actions provides tests for action implementations.
package actions

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestRestActionCreation(t *testing.T) {
	tests := []struct {
		name        string
		config      map[string]interface{}
		expectError bool
		errorMsg    string
	}{
		{
			name: "Valid REST config with URL",
			config: map[string]interface{}{
				"url": "http://localhost:8080/webhook",
			},
			expectError: false,
		},
		{
			name: "Valid REST config with all options",
			config: map[string]interface{}{
				"url":     "http://localhost:8080/webhook",
				"method":  "POST",
				"headers": map[string]interface{}{"X-Custom-Header": "value"},
				"timeout": 30,
			},
			expectError: false,
		},
		{
			name: "Missing URL",
			config: map[string]interface{}{
				"method": "POST",
			},
			expectError: true,
			errorMsg:    "REST action requires 'url' in config",
		},
		{
			name: "Empty URL",
			config: map[string]interface{}{
				"url": "",
			},
			expectError: true,
			errorMsg:    "REST action requires 'url' in config",
		},
		{
			name: "Default method is POST",
			config: map[string]interface{}{
				"url": "http://localhost:8080/webhook",
			},
			expectError: false,
		},
		{
			name: "Custom GET method",
			config: map[string]interface{}{
				"url":    "http://localhost:8080/webhook",
				"method": "GET",
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			action, err := NewRestActionFromConfig(tt.config)

			if tt.expectError {
				if err == nil {
					t.Errorf("expected error, got nil")
				}
				if err != nil && tt.errorMsg != "" && err.Error() != tt.errorMsg {
					t.Errorf("expected error message '%s', got '%s'", tt.errorMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				if action == nil {
					t.Error("expected action to be created")
				}
			}
		})
	}
}

func TestRestActionSuccessResponse(t *testing.T) {
	// Create a test server that returns 200 OK
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	// Create REST action pointing to test server
	action, err := NewRestActionFromConfig(map[string]interface{}{
		"url":     server.URL,
		"method":  "POST",
		"timeout": 5,
	})
	if err != nil {
		t.Fatalf("failed to create action: %v", err)
	}

	// Execute action
	testEvent := map[string]interface{}{
		"type": "user.created",
		"id":   "evt-123",
	}

	handler := action.Handler()
	err = handler(testEvent)

	if err != nil {
		t.Errorf("expected success, got error: %v", err)
	}
}

func TestRestActionErrorResponse(t *testing.T) {
	// Create a test server that returns 500 Internal Server Error
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	// Create REST action pointing to test server
	action, err := NewRestActionFromConfig(map[string]interface{}{
		"url":     server.URL,
		"method":  "POST",
		"timeout": 5,
	})
	if err != nil {
		t.Fatalf("failed to create action: %v", err)
	}

	// Execute action
	testEvent := map[string]interface{}{
		"type": "user.created",
		"id":   "evt-123",
	}

	handler := action.Handler()
	err = handler(testEvent)

	if err == nil {
		t.Error("expected error for 500 response, got nil")
	}
}

func TestRestActionTimeout(t *testing.T) {
	// Create a slow test server that never responds
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(10 * time.Second) // Sleep longer than timeout
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	// Create REST action with short timeout
	action, err := NewRestActionFromConfig(map[string]interface{}{
		"url":     server.URL,
		"method":  "POST",
		"timeout": 1, // 1 second timeout
	})
	if err != nil {
		t.Fatalf("failed to create action: %v", err)
	}

	// Execute action
	testEvent := map[string]interface{}{
		"type": "user.created",
		"id":   "evt-123",
	}

	handler := action.Handler()
	err = handler(testEvent)

	if err == nil {
		t.Error("expected timeout error, got nil")
	}
}

func TestRestActionCustomHeaders(t *testing.T) {
	var receivedHeaders http.Header

	// Create a test server that captures headers
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedHeaders = r.Header.Clone()
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	// Create REST action with custom headers
	action, err := NewRestActionFromConfig(map[string]interface{}{
		"url":    server.URL,
		"method": "POST",
		"headers": map[string]interface{}{
			"X-Custom-Header": "custom-value",
			"X-Another":       "another-value",
		},
		"timeout": 5,
	})
	if err != nil {
		t.Fatalf("failed to create action: %v", err)
	}

	// Execute action
	testEvent := map[string]interface{}{
		"type": "user.created",
		"id":   "evt-123",
	}

	handler := action.Handler()
	err = handler(testEvent)

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// Verify custom headers were sent
	if receivedHeaders.Get("X-Custom-Header") != "custom-value" {
		t.Errorf("expected X-Custom-Header 'custom-value', got '%s'", receivedHeaders.Get("X-Custom-Header"))
	}
	if receivedHeaders.Get("X-Another") != "another-value" {
		t.Errorf("expected X-Another 'another-value', got '%s'", receivedHeaders.Get("X-Another"))
	}
}

func TestRestActionHTTPMethods(t *testing.T) {
	testMethods := []string{"GET", "POST", "PUT", "DELETE"}

	for _, method := range testMethods {
		t.Run("Method "+method, func(t *testing.T) {
			var receivedMethod string

			// Create a test server that captures the method
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				receivedMethod = r.Method
				w.WriteHeader(http.StatusOK)
			}))
			defer server.Close()

			// Create REST action with specific method
			action, err := NewRestActionFromConfig(map[string]interface{}{
				"url":     server.URL,
				"method":  method,
				"timeout": 5,
			})
			if err != nil {
				t.Fatalf("failed to create action: %v", err)
			}

			// Execute action
			handler := action.Handler()
			err = handler(map[string]interface{}{"type": "test"})

			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}

			if receivedMethod != method {
				t.Errorf("expected method '%s', got '%s'", method, receivedMethod)
			}
		})
	}
}

func TestRestActionDescription(t *testing.T) {
	action, err := NewRestActionFromConfig(map[string]interface{}{
		"url":    "http://example.com/webhook",
		"method": "POST",
	})
	if err != nil {
		t.Fatalf("failed to create action: %v", err)
	}

	if action.Name() == "" {
		t.Error("expected non-empty name")
	}

	if action.Description() == "" {
		t.Error("expected non-empty description")
	}

	if action.Type() != "rest" {
		t.Errorf("expected type 'rest', got '%s'", action.Type())
	}
}
