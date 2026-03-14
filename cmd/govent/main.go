// Package main is the entry point for the govent event distribution system.
package main

import (
	"flag"
	"log"
	"os"

	"github.com/jarlex/govent/config"
	"github.com/jarlex/govent/server"
	"github.com/jarlex/govent/trigger"
)

var (
	configPath = flag.String("config", "./configs/triggers.yaml", "Path to trigger configuration file")
	httpPort   = flag.String("http-port", "8080", "HTTP server port")
	grpcPort   = flag.String("grpc-port", "50051", "gRPC server port")
)

func main() {
	flag.Parse()

	// 5.9: Wire everything in main.go - initialize config, start servers on port 8080

	// Load configuration
	cfg, err := config.LoadAndValidate(*configPath)
	if err != nil {
		// Check if config file exists
		if _, statErr := os.Stat(*configPath); os.IsNotExist(statErr) {
			log.Printf("Warning: Config file not found at %s, running with no triggers", *configPath)
			cfg = &config.Config{Triggers: nil}
		} else {
			log.Fatalf("Failed to load config: %v", err)
		}
	}

	log.Printf("Loaded %d trigger(s) from configuration", len(cfg.Triggers))

	// Create trigger engine
	engine := trigger.NewEngine(cfg.Triggers)

	// Create server engine
	serverEngine := server.NewEngine(engine)

	// Create HTTP handler
	httpHandler := server.NewHTTPHandler(serverEngine)

	// Start HTTP server in a goroutine
	go func() {
		log.Printf("Starting HTTP server on port %s", *httpPort)
		if err := server.StartHTTPServer(*httpPort, httpHandler); err != nil {
			log.Printf("HTTP server error: %v", err)
		}
	}()

	// Start gRPC server in a goroutine (optional, can be disabled)
	grpcServer := server.NewSimpleGrpcServer(*grpcPort, serverEngine)
	go func() {
		log.Printf("Starting gRPC server on port %s", *grpcPort)
		if err := grpcServer.StartSimple(); err != nil {
			log.Printf("gRPC server error: %v", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown
	log.Println("Server started successfully. Press Ctrl+C to stop.")
	select {}
}
