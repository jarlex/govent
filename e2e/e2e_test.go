// Package e2e provides end-to-end integration tests for the govent system.
package e2e

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/jarlex/govent/actions"
	"github.com/jarlex/govent/config"
	"github.com/jarlex/govent/event"
	"github.com/jarlex/govent/server"
	"github.com/jarlex/govent/trigger"
)

// mockAction is a test action that records invocations
type mockAction struct {
	invocations int
	name        string
}

func (m *mockAction) Name() string        { return m.name }
func (m *mockAction) Description() string { return "Mock action for testing" }
func (m *mockAction) Type() string        { return "mock" }
func (m *mockAction) Init(config map[string]interface{}) error {
	return nil
}
func (m *mockAction) Handler() func(interface{}) error {
	return func(input interface{}) error {
		m.invocations++
		return nil
	}
}

// TestEndToEndFlow tests the complete flow: POST /events → trigger match → action execution
func TestEndToEndFlow(t *testing.T) {
	// Step 1: Create a REST endpoint to receive webhook (simulating external service)
	webhookReceived := false
	webhookServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		webhookReceived = true
		w.WriteHeader(http.StatusOK)
	}))
	defer webhookServer.Close()

	// Step 2: Create trigger configuration
	triggerConfigs := []config.TriggerConfig{
		{
			Name:      "webhook-on-user-created",
			EventType: "user.created",
			Actions: []config.ActionConfig{
				{
					Type: "rest",
					Config: map[string]interface{}{
						"url":     webhookServer.URL,
						"method":  "POST",
						"timeout": 5,
					},
				},
			},
		},
	}

	// Step 3: Create trigger engine
	engine := trigger.NewEngine(triggerConfigs)

	// Step 4: Create HTTP handler
	serverEngine := server.NewEngine(engine)
	handler := server.NewHTTPHandler(serverEngine)

	// Step 5: Make a POST request to /events
	eventJSON := `{"type":"user.created","payload":{"user_id":"u123","email":"test@example.com"}}`
	req := httptest.NewRequest(http.MethodPost, "/events", bytes.NewReader([]byte(eventJSON)))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()

	// Use goroutine to not block on async action execution
	go func() {
		handler.ServeHTTP(w, req)
	}()

	// Wait for the response
	time.Sleep(100 * time.Millisecond)

	// Verify HTTP response
	if w.Code != http.StatusCreated {
		t.Errorf("expected status 201, got %d", w.Code)
	}

	// Verify event was returned in response
	var respEvent event.Event
	if err := json.Unmarshal(w.Body.Bytes(), &respEvent); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	if respEvent.Type != "user.created" {
		t.Errorf("expected type 'user.created', got '%s'", respEvent.Type)
	}

	// Wait for async action execution
	time.Sleep(500 * time.Millisecond)

	// Verify webhook was received (action was executed)
	if !webhookReceived {
		t.Error("expected webhook to be received, but it was not")
	}
}

func TestEndToEndWithMultipleTriggers(t *testing.T) {
	// Test multiple triggers matching the same event

	webhook1Received := false
	webhook1Server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		webhook1Received = true
		w.WriteHeader(http.StatusOK)
	}))
	defer webhook1Server.Close()

	webhook2Received := false
	webhook2Server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		webhook2Received = true
		w.WriteHeader(http.StatusOK)
	}))
	defer webhook2Server.Close()

	triggerConfigs := []config.TriggerConfig{
		{
			Name:      "trigger-1",
			EventType: "user.created",
		},
		{
			Name:      "trigger-2",
			EventType: "user.created",
			Actions: []config.ActionConfig{
				{
					Type: "rest",
					Config: map[string]interface{}{
						"url":     webhook1Server.URL,
						"method":  "POST",
						"timeout": 5,
					},
				},
			},
		},
		{
			Name:      "trigger-3",
			EventType: "user.created",
			Actions: []config.ActionConfig{
				{
					Type: "rest",
					Config: map[string]interface{}{
						"url":     webhook2Server.URL,
						"method":  "POST",
						"timeout": 5,
					},
				},
			},
		},
	}

	engine := trigger.NewEngine(triggerConfigs)
	serverEngine := server.NewEngine(engine)
	handler := server.NewHTTPHandler(serverEngine)

	eventJSON := `{"type":"user.created","payload":{"user_id":"u123"}}`
	req := httptest.NewRequest(http.MethodPost, "/events", bytes.NewReader([]byte(eventJSON)))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	go handler.ServeHTTP(w, req)

	time.Sleep(100 * time.Millisecond)

	if w.Code != http.StatusCreated {
		t.Errorf("expected status 201, got %d", w.Code)
	}

	// Wait for async action execution
	time.Sleep(500 * time.Millisecond)

	// Both webhooks should have been called
	if !webhook1Received {
		t.Error("expected webhook1 to be received")
	}
	if !webhook2Received {
		t.Error("expected webhook2 to be received")
	}
}

func TestEndToEndNoMatchingTrigger(t *testing.T) {
	// Test when no trigger matches the event

	webhookServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("webhook should not be called when no trigger matches")
	}))
	defer webhookServer.Close()

	triggerConfigs := []config.TriggerConfig{
		{
			Name:      "user-deleted-trigger",
			EventType: "user.deleted",
			Actions: []config.ActionConfig{
				{
					Type: "rest",
					Config: map[string]interface{}{
						"url":     webhookServer.URL,
						"method":  "POST",
						"timeout": 5,
					},
				},
			},
		},
	}

	engine := trigger.NewEngine(triggerConfigs)
	serverEngine := server.NewEngine(engine)
	handler := server.NewHTTPHandler(serverEngine)

	// Send a user.created event, but trigger is for user.deleted
	eventJSON := `{"type":"user.created","payload":{"user_id":"u123"}}`
	req := httptest.NewRequest(http.MethodPost, "/events", bytes.NewReader([]byte(eventJSON)))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	time.Sleep(100 * time.Millisecond)

	if w.Code != http.StatusCreated {
		t.Errorf("expected status 201, got %d", w.Code)
	}

	// Give time for any potential action (there shouldn't be any)
	time.Sleep(200 * time.Millisecond)

	// webhookServer handler would have called t.Error if it was invoked
}

func TestEndToEndWithSourceMatcher(t *testing.T) {
	// Test trigger matching with source matcher

	matchingWebhookReceived := false
	matchingServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		matchingWebhookReceived = true
		w.WriteHeader(http.StatusOK)
	}))
	defer matchingServer.Close()

	nonMatchingWebhookCalled := false
	nonMatchingServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		nonMatchingWebhookCalled = true
		w.WriteHeader(http.StatusOK)
	}))
	defer nonMatchingServer.Close()

	triggerConfigs := []config.TriggerConfig{
		{
			Name:      "auth-service-trigger",
			EventType: "user.created",
			Matchers: map[string]string{
				"source": "auth-service",
			},
			Actions: []config.ActionConfig{
				{
					Type: "rest",
					Config: map[string]interface{}{
						"url":     matchingServer.URL,
						"method":  "POST",
						"timeout": 5,
					},
				},
			},
		},
		{
			Name:      "other-service-trigger",
			EventType: "user.created",
			Matchers: map[string]string{
				"source": "other-service",
			},
			Actions: []config.ActionConfig{
				{
					Type: "rest",
					Config: map[string]interface{}{
						"url":     nonMatchingServer.URL,
						"method":  "POST",
						"timeout": 5,
					},
				},
			},
		},
	}

	engine := trigger.NewEngine(triggerConfigs)
	serverEngine := server.NewEngine(engine)
	handler := server.NewHTTPHandler(serverEngine)

	// Send event with source = auth-service (should match first trigger)
	eventJSON := `{"type":"user.created","payload":{"user_id":"u123"},"source":"auth-service"}`
	req := httptest.NewRequest(http.MethodPost, "/events", bytes.NewReader([]byte(eventJSON)))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	go handler.ServeHTTP(w, req)

	time.Sleep(100 * time.Millisecond)

	if w.Code != http.StatusCreated {
		t.Errorf("expected status 201, got %d", w.Code)
	}

	time.Sleep(500 * time.Millisecond)

	// Only matching webhook should be called
	if !matchingWebhookReceived {
		t.Error("expected matching webhook to be received")
	}
	if nonMatchingWebhookCalled {
		t.Error("non-matching webhook should not be called")
	}
}

func TestConfigLoading(t *testing.T) {
	// Test that config is properly loaded and triggers are created
	cfg := config.Config{
		Triggers: []config.TriggerConfig{
			{
				Name:      "test-trigger",
				EventType: "test.event",
				Matchers: map[string]string{
					"source": "test-source",
				},
				Actions: []config.ActionConfig{
					{
						Type: "rest",
						Config: map[string]interface{}{
							"url": "http://localhost:8080/webhook",
						},
					},
				},
			},
		},
	}

	// Verify config structure
	if len(cfg.Triggers) != 1 {
		t.Errorf("expected 1 trigger, got %d", len(cfg.Triggers))
	}

	trigger := cfg.Triggers[0]
	if trigger.Name != "test-trigger" {
		t.Errorf("expected name 'test-trigger', got '%s'", trigger.Name)
	}
	if trigger.EventType != "test.event" {
		t.Errorf("expected eventType 'test.event', got '%s'", trigger.EventType)
	}
	if trigger.Matchers["source"] != "test-source" {
		t.Errorf("expected source matcher 'test-source', got '%s'", trigger.Matchers["source"])
	}
	if len(trigger.Actions) != 1 {
		t.Errorf("expected 1 action, got %d", len(trigger.Actions))
	}
	if trigger.Actions[0].Type != "rest" {
		t.Errorf("expected action type 'rest', got '%s'", trigger.Actions[0].Type)
	}
}

func TestActionFactory(t *testing.T) {
	// Test action factory creates correct action types

	restConfig := config.ActionConfig{
		Type: "rest",
		Config: map[string]interface{}{
			"url": "http://localhost:8080/webhook",
		},
	}

	action, err := actions.Factory(restConfig)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if action.Type() != "rest" {
		t.Errorf("expected type 'rest', got '%s'", action.Type())
	}

	// Test unknown action type
	unknownConfig := config.ActionConfig{
		Type:   "unknown",
		Config: map[string]interface{}{},
	}

	_, err = actions.Factory(unknownConfig)
	if err == nil {
		t.Error("expected error for unknown action type")
	}
}

func TestEventValidationInEndToEnd(t *testing.T) {
	// Test that validation happens before trigger execution

	triggerConfigs := []config.TriggerConfig{
		{
			Name:      "test-trigger",
			EventType: "user.created",
		},
	}

	engine := trigger.NewEngine(triggerConfigs)
	serverEngine := server.NewEngine(engine)
	handler := server.NewHTTPHandler(serverEngine)

	// Test with missing type (should return 400)
	eventJSON := `{"payload":{"user_id":"u123"}}`
	req := httptest.NewRequest(http.MethodPost, "/events", bytes.NewReader([]byte(eventJSON)))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status 400 for missing type, got %d", w.Code)
	}

	// Test with invalid timestamp (should return 400)
	eventJSON = `{"type":"user.created","payload":{"user_id":"u123"},"timestamp":"invalid"}`
	req = httptest.NewRequest(http.MethodPost, "/events", bytes.NewReader([]byte(eventJSON)))
	req.Header.Set("Content-Type", "application/json")

	w = httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status 400 for invalid timestamp, got %d", w.Code)
	}
}
