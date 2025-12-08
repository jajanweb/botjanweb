// Package http provides HTTP controllers and routing for web endpoints.
//
// This file defines all HTTP routes for the application.
// Similar to Laravel's routes/api.php, all endpoint definitions are centralized here.
package http

// RegisterRoutes registers all HTTP routes to the router.
// This is the central place for all endpoint definitions.
//
// Usage:
//
//	router := http.NewRouter()
//	http.RegisterRoutes(router, webhookController, qrPairingController)
//	http.ListenAndServe(":8080", router)
func RegisterRoutes(router *Router, webhook *WebhookController, qrPairing *QRPairingController) {
	// ============================================================
	// Health & Status Routes (Kubernetes/Heroku probes)
	// ============================================================
	// /health - Liveness probe: Is the app running?
	// Returns 200 if process is alive (even if dependencies are down)
	router.GET("/health", "Health check endpoint (liveness probe)", webhook.handleHealth)

	// /ready - Readiness probe: Is the app ready to serve traffic?
	// Returns 200 only if all critical dependencies are healthy
	router.GET("/ready", "Readiness check endpoint (readiness probe)", webhook.handleReadiness)

	// Aliases for compatibility
	router.GET("/healthz", "Health check alias", webhook.handleHealth)
	router.GET("/readyz", "Readiness check alias", webhook.handleReadiness)

	// ============================================================
	// WhatsApp Pairing Routes (Protected by token)
	// ============================================================
	if qrPairing != nil {
		router.GET("/pairing", "WhatsApp QR code pairing page", qrPairing.HandlePairingPage)
		router.GET("/pairing/qr", "Get current QR code (AJAX endpoint)", qrPairing.HandleQRCodeAPI)
	}

	// ============================================================
	// Webhook Routes - Payment Notifications
	// ============================================================
	router.POST("/webhook/payment", "Receive payment notifications from DANA/payment gateway", webhook.handlePaymentWebhook)

	// ============================================================
	// Future Routes (placeholder examples)
	// ============================================================
	// router.GET("/api/pending", "List all pending payments", webhook.handleListPending)
	// router.DELETE("/api/pending/:id", "Cancel a pending payment", webhook.handleCancelPending)
	// router.GET("/api/transactions", "List transaction history", webhook.handleListTransactions)
}
