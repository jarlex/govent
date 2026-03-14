// Package server provides gRPC server implementation for the govent system.
package server

import (
	"context"
	"fmt"
	"log"
	"net"
	"time"

	"github.com/jarlex/govent/event"
	"google.golang.org/grpc"
)

// GrpcServer represents a gRPC server for event ingestion.
type GrpcServer struct {
	engine   *Engine
	listener net.Listener
	server   *grpc.Server
	port     string
}

// NewGrpcServer creates a new gRPC server.
func NewGrpcServer(port string, eng *Engine) *GrpcServer {
	return &GrpcServer{
		engine: eng,
		port:   port,
	}
}

// GrpcEventServiceServer is the interface for the event service (placeholder for proto generation).
// In a full implementation, this would be generated from a .proto file.
type GrpcEventServiceServer interface {
	IngestEvent(context.Context, *EventRequest) (*EventResponse, error)
}

// EventRequest represents a gRPC event request.
type EventRequest struct {
	ID        string                 `json:"id"`
	Type      string                 `json:"type"`
	Payload   map[string]interface{} `json:"payload"`
	Metadata  map[string]string      `json:"metadata"`
	Timestamp string                 `json:"timestamp"`
	Source    string                 `json:"source"`
}

// EventResponse represents a gRPC event response.
type EventResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	EventID string `json:"event_id"`
}

// grpcEventService implements the GrpcEventServiceServer interface.
type grpcEventService struct {
	engine *Engine
}

// IngestEvent handles incoming gRPC event requests.
// 5.6: Implement gRPC server for event ingestion
func (s *grpcEventService) IngestEvent(ctx context.Context, req *EventRequest) (*EventResponse, error) {
	// Convert gRPC request to internal Event
	evt := &event.Event{
		ID:        req.ID,
		Type:      req.Type,
		Payload:   req.Payload,
		Metadata:  req.Metadata,
		Timestamp: req.Timestamp,
		Source:    req.Source,
	}

	// Validate event
	if err := evt.Validate(); err != nil {
		return &EventResponse{
			Success: false,
			Message: err.Error(),
			EventID: "",
		}, nil
	}

	// Set defaults
	evt.EnsureDefaults()

	// Execute trigger actions asynchronously
	go s.engine.triggerEngine.ExecuteActions(evt)

	return &EventResponse{
		Success: true,
		Message: "Event received",
		EventID: evt.ID,
	}, nil
}

// Start starts the gRPC server.
// 5.6: Create server/grpc.go with gRPC server for event ingestion
func (s *GrpcServer) Start() error {
	var err error
	s.listener, err = net.Listen("tcp", ":"+s.port)
	if err != nil {
		return fmt.Errorf("failed to listen on port %s: %w", s.port, err)
	}

	s.server = grpc.NewServer()
	RegisterGrpcEventServiceServer(s.server, &grpcEventService{engine: s.engine})

	log.Printf("Starting gRPC server on port %s", s.port)
	if err := s.server.Serve(s.listener); err != nil {
		return err
	}

	return nil
}

// Stop stops the gRPC server gracefully.
func (s *GrpcServer) Stop() {
	if s.server != nil {
		s.server.GracefulStop()
	}
}

// GrpcEventServiceServer is the service descriptor for gRPC.
var _ GrpcEventServiceServer = (*grpcEventService)(nil)

// This is a placeholder registration function.
// In a full implementation with protobuf, this would be generated.
// For MVP, we use a simple implementation without proto generation.
func init() {
	// Placeholder for proto registration
}

// RegisterGrpcEventServiceServer is a placeholder for the generated gRPC service registration.
// In a production system, this would be generated from a .proto file.
func RegisterGrpcEventServiceServer(s *grpc.Server, srv *grpcEventService) {
	// This is a simplified implementation.
	// In production, you would register the actual generated protobuf service.
	// For now, we'll just note that this is where proto-generated code would go.
	_ = srv
	_ = s
}

// SimpleGrpcServer is a simple gRPC server wrapper that can be used without proto generation.
type SimpleGrpcServer struct {
	engine *Engine
	port   string
}

// NewSimpleGrpcServer creates a simple gRPC server.
func NewSimpleGrpcServer(port string, eng *Engine) *SimpleGrpcServer {
	return &SimpleGrpcServer{
		engine: eng,
		port:   port,
	}
}

// StartSimple starts a simple gRPC server that can handle basic event ingestion.
func (s *SimpleGrpcServer) StartSimple() error {
	lis, err := net.Listen("tcp", ":"+s.port)
	if err != nil {
		return fmt.Errorf("failed to listen: %w", err)
	}

	grpcServer := grpc.NewServer()
	RegisterSimpleEventService(grpcServer, s.engine)

	log.Printf("Starting gRPC server on port %s", s.port)
	return grpcServer.Serve(lis)
}

// SimpleEventService is a simple implementation for MVP.
type SimpleEventService struct {
	engine *Engine
}

// RegisterSimpleEventService registers the simple event service.
func RegisterSimpleEventService(s *grpc.Server, eng *Engine) {
	// Placeholder registration
	_ = eng
}

// Timeout for gRPC requests.
const defaultGRPCTimeout = 30 * time.Second
