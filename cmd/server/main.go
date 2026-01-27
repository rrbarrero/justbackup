package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/rrbarrero/justbackup/internal/shared/infrastructure/container"

	_ "github.com/rrbarrero/justbackup/docs"
)

// @title JustBackup API
// @version 1.0
// @description API for JustBackup application
// @BasePath /api/v1
// @securityDefinitions.basic BasicAuth
func main() {
	fmt.Println("Starting JustBackup Server...")

	// Initialize the dependency injection container
	container, err := container.InitializeContainer()
	if err != nil {
		log.Fatal("Failed to initialize application container:", err)
	}

	// Create context that listens for the interrupt signal from the OS
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Set up channel to listen for interrupt and terminate signals
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Start the application in a separate goroutine
	go func() {
		if err := container.Run(ctx); err != nil {
			log.Printf("Server error: %v", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	fmt.Println("Server is running. Press Ctrl+C to stop.")
	<-sigChan
	fmt.Println("\nShutting down server...")

	// Cancel the context to stop all background services
	cancel()

	// Close the container to release resources
	if err := container.Close(); err != nil {
		log.Printf("Error closing container: %v", err)
	}

	fmt.Println("Server stopped")
}
