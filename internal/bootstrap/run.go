// Package bootstrap handles application initialization and dependency injection.
package bootstrap

import (
	"context"
	"fmt"

	"github.com/exernia/botjanweb/internal/domain/entity"
	infrawebhook "github.com/exernia/botjanweb/internal/infrastructure/messaging/webhook"
	infrawa "github.com/exernia/botjanweb/internal/infrastructure/messaging/whatsapp"
)

// Run starts all application services.
func (app *App) Run(ctx context.Context) error {
	// Create WhatsApp client with message handler
	waClient, err := infrawa.NewClient(
		ctx,
		app.Config.WhatsAppDBPath,
		app.Config.GroupJID,
	)
	if err != nil {
		return fmt.Errorf("failed to create WhatsApp client: %w", err)
	}
	app.WAClient = waClient

	// Set messaging service to bot handler
	app.BotHandler.SetMessaging(app.WAClient)

	// Set message handler
	app.WAClient.SetMessageHandler(app.createMessageHandler())

	// Re-initialize payment confirmation service with WhatsApp adapter (must be after WAClient is created)
	app.initPaymentConfirmationService()

	// Initialize QR pairing controller (must be after WAClient is created)
	if app.Config.WebhookEnabled && app.WAClient != nil {
		app.initQRPairingController()
	}

	// Start webhook server if enabled
	if app.Config.WebhookEnabled {
		// Build HTTP router with all routes
		router := app.buildHTTPRouter()

		app.WebhookServer = infrawebhook.NewServer(
			app.Config.WebhookPort,
			router,
		)
		if err := app.WebhookServer.Start(); err != nil {
			return fmt.Errorf("failed to start webhook server: %w", err)
		}
		app.Logger.Printf("✅ Webhook server started on port %d", app.Config.WebhookPort)
	}

	// CRITICAL: Start QR pairing BEFORE WAClient.Run() to avoid race condition
	if !app.WAClient.IsLoggedIn() && app.QRPairingController != nil {
		app.Logger.Println("📱 Device not paired. Starting QR pairing...")
		if err := app.QRPairingController.StartPairing(ctx); err != nil {
			app.Logger.Printf("⚠️ Failed to start QR pairing: %v", err)
			// Don't return error, still run the client
		} else {
			// Get hostname for Heroku or use localhost for dev
			hostname := "localhost"
			if app.Config.HerokuAppName != "" {
				hostname = app.Config.HerokuAppName + ".herokuapp.com"
			} else {
				hostname = fmt.Sprintf("localhost:%d", app.Config.WebhookPort)
			}
			app.Logger.Printf("🌐 QR Pairing available at: http://%s/pairing?token=%s",
				hostname, app.Config.WebhookSecret)
		}
	}

	// Run WhatsApp client (blocking)
	// If device not logged in, this will just wait for pairing via web interface
	return app.WAClient.Run(ctx)
}

// Shutdown gracefully stops all services.
func (app *App) Shutdown(ctx context.Context) error {
	app.Logger.Println("🛑 Initiating graceful shutdown...")

	// Mark as not ready (stop accepting new requests)
	if app.WebhookController != nil {
		app.WebhookController.SetReady(false)
		app.Logger.Println("   → Webhook marked as not ready")
	}

	// Stop webhook server
	if app.WebhookServer != nil {
		app.Logger.Println("   → Stopping webhook server...")
		if err := app.WebhookServer.Stop(ctx); err != nil {
			app.Logger.Printf("   ⚠️ Error stopping webhook server: %v", err)
		} else {
			app.Logger.Println("   ✅ Webhook server stopped")
		}
	}

	// Disconnect WhatsApp
	if app.WAClient != nil {
		app.Logger.Println("   → Disconnecting from WhatsApp...")
		app.WAClient.Disconnect()
		app.Logger.Println("   ✅ WhatsApp disconnected")
	}

	// Stop cleanup goroutine
	if app.PendingStore != nil {
		app.Logger.Println("   → Stopping pending payment cleanup...")
		app.PendingStore.StopCleanup()
		app.Logger.Println("   ✅ Cleanup stopped")

		// Close database connection (PostgreSQL)
		app.Logger.Println("   → Closing pending store...")
		if err := app.PendingStore.Close(); err != nil {
			app.Logger.Printf("   ⚠️ Error closing pending store: %v", err)
		} else {
			app.Logger.Println("   ✅ Pending store closed")
		}
	}

	app.Logger.Println("✅ Graceful shutdown complete")
	return nil
}

// createMessageHandler creates the WhatsApp message handler function.
func (app *App) createMessageHandler() infrawa.MessageHandler {
	return func(ctx context.Context, msg *entity.Message) {
		app.BotHandler.HandleMessage(ctx, msg)
	}
}
