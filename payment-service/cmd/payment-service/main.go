package main

import (
	"log"

	"payment-service/internal/app"
)

func main() {
	// Load configuration
	cfg := app.LoadConfig()

	// Create application with manual dependency injection (Composition Root)
	application, err := app.NewApp(cfg)
	if err != nil {
		log.Fatalf("Failed to initialize application: %v", err)
	}
	defer application.Close()

	// Run the server
	if err := application.Run(); err != nil {
		log.Fatalf("Failed to run server: %v", err)
	}
}
