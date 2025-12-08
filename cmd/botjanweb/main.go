// Package main is the entry point for BotJanWeb.
package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/exernia/botjanweb/pkg/logger"

	"github.com/exernia/botjanweb/internal/bootstrap"
)

func main() {
	log := logger.Main
	log.Println("Starting BotJanWeb...")

	ctx := context.Background()

	// Initialize application with all dependencies
	log.Println("Initializing application...")
	app, err := bootstrap.New(ctx, "assets")
	if err != nil {
		log.Fatalf("Initialization error: %v", err)
	}

	log.Printf("Config loaded. Group: %s", app.Config.GroupJID)
	log.Println("============================================")
	log.Println("Scan QR code with WhatsApp if first run")
	log.Println("============================================")

	// Setup graceful shutdown channel
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)

	// Run application in goroutine
	errChan := make(chan error, 1)
	go func() {
		if err := app.Run(ctx); err != nil {
			errChan <- err
		}
	}()

	// Wait for either termination signal or application error
	select {
	case sig := <-sigChan:
		log.Printf("Received signal: %v. Initiating graceful shutdown...", sig)
	case err := <-errChan:
		log.Printf("Application error: %v. Shutting down...", err)
	}

	// Graceful shutdown with 30 second timeout
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := app.Shutdown(shutdownCtx); err != nil {
		log.Printf("Shutdown error: %v", err)
		os.Exit(1)
	}

	log.Println("BotJanWeb stopped gracefully.")
}
