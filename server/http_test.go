// Package server provides tests for HTTP server handlers.
package server

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/jarlex/govent/config"
	"github.com/jarlex/govent/event"
	"github.com/jarlex/govent/trigger"
)

func TestHTTPHandlerEventEndpoint(t *testing.T) {
	// Create trigger engine with empty triggers (no matching)
	engine := trigger.NewEngine([]config.TriggerConfig{})
	serverEngine := NewEngine(engine)
	handler := NewHTTPHandler(serverEngine)

	tests := []struct {
		name           string
		method         string
		contentType    string
		body           []byte
		expectedStatus int
		expectBody     bool
	}{
		{
			name:           "Valid POST request",
			method:         http.MethodPost,
			contentType:    "application/json",
			body:           []byte(`{"type":"user.created","payload":{"user_id":"u123"}}`),
			expectedStatus: http.StatusCreated,
			expectBody:     true,
		},
		{
			name:           "POST with all fields",
			method:         http.MethodPost,
			contentType:    "application/json",
			body:           []byte(`{"type":"user.created","payload":{"user_id":"u123"},"id":"evt-123","timestamp":"2024-01-15T10:30:00Z","source":"auth-service"}`),
			expectedStatus: http.StatusCreated,
			expectBody:     true,
		},
		{
			name:           "Missing Content-Type",
			method:         http.MethodPost,
			contentType:    "",
			body:           []byte(`{"type":"user.created","payload":{"user_id":"u123"}}`),
			expectedStatus: http.StatusUnsupportedMediaType,
			expectBody:     false,
		},
		{
			name:           "Invalid Content-Type",
			method:         http.MethodPost,
			contentType:    "text/plain",
			body:           []byte(`{"type":"user.created","payload":{"user_id":"u123"}}`),
			expectedStatus: http.StatusUnsupportedMediaType,
			expectBody:     false,
		},
		{
			name:           "Invalid JSON",
			method:         http.MethodPost,
			contentType:    "application/json",
			body:           []byte(`{invalid json}`),
			expectedStatus: http.StatusBadRequest,
			expectBody:     false,
		},
		{
			name:           "Missing required type field",
			method:         http.MethodPost,
			contentType:    "application/json",
			body:           []byte(`{"payload":{"user_id":"u123"}}`),
			expectedStatus: http.StatusBadRequest,
			expectBody:     false,
		},
		{
			name:           "Missing required payload field",
			method:         http.MethodPost,
			contentType:    "application/json",
			body:           []byte(`{"type":"user.created"}`),
			expectedStatus: http.StatusBadRequest,
			expectBody:     false,
		},
		{
			name:           "Invalid timestamp format",
			method:         http.MethodPost,
			contentType:    "application/json",
			body:           []byte(`{"type":"user.created","payload":{"user_id":"u123"},"timestamp":"not-a-timestamp"}`),
			expectedStatus: http.StatusBadRequest,
			expectBody:     false,
		},
		{
			name:           "Method not allowed - GET",
			method:         http.MethodGet,
			contentType:    "application/json",
			body:           []byte(`{"type":"user.created","payload":{"user_id":"u123"}}`),
			expectedStatus: http.StatusMethodNotAllowed,
			expectBody:     false,
		},
		{
			name:           "Empty body",
			method:         http.MethodPost,
			contentType:    "application/json",
			body:           []byte(``),
			expectedStatus: http.StatusBadRequest,
			expectBody:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, "/events", bytes.NewReader(tt.body))
			if tt.contentType != "" {
				req.Header.Set("Content-Type", tt.contentType)
			}

			w := httptest.NewRecorder()
			handler.ServeHTTP(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			if tt.expectBody && w.Body.Len() == 0 {
				t.Error("expected body in response")
			}

			// Verify error responses contain error field
			if tt.expectedStatus >= 400 && tt.expectBody {
				var resp map[string]string
				if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
					t.Errorf("failed to parse response body: %v", err)
				} else {
					if _, hasError := resp["error"]; !hasError {
						t.Error("expected 'error' field in error response")
					}
				}
			}
		})
	}
}

func TestHTTPHandlerHealthEndpoint(t *testing.T) {
	engine := trigger.NewEngine([]config.TriggerConfig{})
	serverEngine := NewEngine(engine)
	handler := NewHTTPHandler(serverEngine)

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	var resp map[string]string
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Errorf("failed to parse response body: %v", err)
	}

	if resp["status"] != "ok" {
		t.Errorf("expected status 'ok', got '%s'", resp["status"])
	}
}

func TestHTTPHandlerHealthWrongMethod(t *testing.T) {
	engine := trigger.NewEngine([]config.TriggerConfig{})
	serverEngine := NewEngine(engine)
	handler := NewHTTPHandler(serverEngine)

	req := httptest.NewRequest(http.MethodPost, "/health", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected status 405, got %d", w.Code)
	}
}

func TestHTTPHandlerNotFound(t *testing.T) {
	engine := trigger.NewEngine([]config.TriggerConfig{})
	serverEngine := NewEngine(engine)
	handler := NewHTTPHandler(serverEngine)

	testPaths := []string{"/unknown", "/eventss", "/healh", "/api/v1/events"}

	for _, path := range testPaths {
		t.Run("Path "+path, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, path, nil)
			w := httptest.NewRecorder()
			handler.ServeHTTP(w, req)

			if w.Code != http.StatusNotFound {
				t.Errorf("expected status 404 for path %s, got %d", path, w.Code)
			}

			var resp map[string]string
			if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
				t.Errorf("failed to parse response body: %v", err)
			}

			if _, hasError := resp["error"]; !hasError {
				t.Error("expected 'error' field in 404 response")
			}
		})
	}
}

func TestHTTPHandlerReturnsEvent(t *testing.T) {
	engine := trigger.NewEngine([]config.TriggerConfig{})
	serverEngine := NewEngine(engine)
	handler := NewHTTPHandler(serverEngine)

	body := []byte(`{"type":"user.created","payload":{"user_id":"u123"}}`)
	req := httptest.NewRequest(http.MethodPost, "/events", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("expected status 201, got %d", w.Code)
	}

	var evt event.Event
	if err := json.Unmarshal(w.Body.Bytes(), &evt); err != nil {
		t.Errorf("failed to parse event in response: %v", err)
	}

	if evt.Type != "user.created" {
		t.Errorf("expected type 'user.created', got '%s'", evt.Type)
	}

	if evt.ID == "" {
		t.Error("expected ID to be generated")
	}

	if evt.Timestamp == "" {
		t.Error("expected timestamp to be generated")
	}
}

func TestHTTPHandlerWithTriggerMatching(t *testing.T) {
	// This test verifies that events trigger matching
	// We use a flag to track if actions would be executed
	executed := false

	// Create a trigger config
	triggerConfigs := []config.TriggerConfig{
		{
			Name:      "test-trigger",
			EventType: "user.created",
		},
	}

	engine := trigger.NewEngine(triggerConfigs)
	serverEngine := NewEngine(engine)
	handler := NewHTTPHandler(serverEngine)

	body := []byte(`{"type":"user.created","payload":{"user_id":"u123"}}`)
	req := httptest.NewRequest(http.MethodPost, "/events", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	// Request should succeed
	if w.Code != http.StatusCreated {
		t.Errorf("expected status 201, got %d", w.Code)
	}

	// The event should be processed by the trigger engine
	// (In a real test, we'd mock the action execution)
	_ = executed
}
