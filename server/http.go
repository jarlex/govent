// Package server provides HTTP and gRPC server implementations for the govent system.
package server

import (
	"encoding/json"
	"errors"
	"log"
	"net"
	"net/http"
	"strings"

	"github.com/jarlex/govent/event"
	"github.com/jarlex/govent/trigger"
)

// ErrPortInUse indicates that the requested port is already in use.
var ErrPortInUse = errors.New("port is already in use")

// Engine wraps the trigger engine for event processing.
type Engine struct {
	triggerEngine *trigger.Engine
}

// NewEngine creates a new server engine with the given trigger engine.
func NewEngine(te *trigger.Engine) *Engine {
	return &Engine{
		triggerEngine: te,
	}
}

// HTTPHandler handles HTTP requests for the govent server.
type HTTPHandler struct {
	engine *Engine
}

// NewHTTPHandler creates a new HTTP handler.
func NewHTTPHandler(engine *Engine) *HTTPHandler {
	return &HTTPHandler{engine: engine}
}

// ServeHTTP handles all HTTP requests.
func (h *HTTPHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.URL.Path {
	case "/events":
		if r.Method == http.MethodPost {
			h.handleEvent(w, r)
			return
		}
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	case "/health":
		if r.Method == http.MethodGet {
			h.handleHealth(w, r)
			return
		}
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	default:
		// 5.8: Implement 404 handler for unknown routes
		h.handleNotFound(w, r)
		return
	}
}

// handleEvent handles POST /events requests.
// 5.2: Implement POST /events endpoint with JSON body parsing
// 5.3: Implement Content-Type validation (return 415 if not application/json)
// 5.4: Implement event validation (return 400 if invalid)
// 5.5: Return 201 Created with event in response body
func (h *HTTPHandler) handleEvent(w http.ResponseWriter, r *http.Request) {
	// 5.3: Content-Type validation - return 415 if not application/json
	contentType := r.Header.Get("Content-Type")
	if contentType != "application/json" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnsupportedMediaType)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "Content-Type must be application/json",
		})
		return
	}

	// 5.2: Parse JSON body into Event struct
	var evt event.Event
	if err := json.NewDecoder(r.Body).Decode(&evt); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "Invalid JSON: " + err.Error(),
		})
		return
	}

	// 5.4: Event validation - return 400 if invalid
	if err := evt.Validate(); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{
			"error": err.Error(),
		})
		return
	}

	// Set defaults for missing fields (ID, timestamp)
	evt.EnsureDefaults()

	// Execute trigger actions asynchronously
	go h.engine.triggerEngine.ExecuteActions(&evt)

	// 5.5: Return 201 Created with event in response body
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(evt)
}

// handleHealth handles GET /health requests.
// 5.7: Implement GET /health endpoint returning HTTP 200 with {"status": "ok"}
func (h *HTTPHandler) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"status": "ok",
	})
}

// handleNotFound handles unknown routes.
// 5.8: Implement 404 handler for unknown routes
func (h *HTTPHandler) handleNotFound(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusNotFound)
	json.NewEncoder(w).Encode(map[string]string{
		"error": "Not Found",
	})
}

// StartHTTPServer starts the HTTP server on the specified port.
// It returns an error if the port is already in use or another error occurs.
// The handler is used to serve all HTTP requests.
func StartHTTPServer(port string, handler *HTTPHandler) error {
	addr := ":" + port

	// First, check if the port is available
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		// Check if it's a "port already in use" error
		if isAddrInUse(err) {
			log.Printf("Error: Port %s is already in use", port)
			return ErrPortInUse
		}
		log.Printf("Failed to start HTTP server on port %s: %v", port, err)
		return err
	}
	ln.Close() // Close the listener; http.ListenAndServe will create its own

	log.Printf("Starting HTTP server on port %s", port)
	if err := http.ListenAndServe(addr, handler); err != nil {
		// Check if it's a "port already in use" error
		if isAddrInUse(err) {
			log.Printf("Error: Port %s is already in use", port)
			return ErrPortInUse
		}
		return err
	}
	return nil
}

// isAddrInUse checks if the error indicates the address is already in use.
func isAddrInUse(err error) bool {
	if err == nil {
		return false
	}
	// Check for common "address already in use" error messages
	errStr := err.Error()
	return strings.Contains(errStr, "address already in use") ||
		strings.Contains(errStr, "port is already in use") ||
		strings.Contains(errStr, "bind: address already in use")
}
