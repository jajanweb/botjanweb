// Package http provides HTTP controllers for web endpoints.
package http

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/exernia/botjanweb/pkg/logger"

	paymentuc "github.com/exernia/botjanweb/internal/application/service/payment"
	"github.com/exernia/botjanweb/internal/domain/entity"
	"github.com/exernia/botjanweb/pkg/helper/formatter"
)

// PaymentConfirmHandler is called when a payment is matched to a pending QRIS.
type PaymentConfirmHandler func(ctx context.Context, pending *entity.PendingPayment, notification *entity.DANANotification)

// WebhookController processes incoming webhook requests.
type WebhookController struct {
	secret         string
	paymentUC      *paymentuc.UseCase
	onPaymentMatch PaymentConfirmHandler
	logger         *log.Logger
	ready          bool // Readiness status
}

// NewWebhookController creates a new webhook controller.
//
// Parameters:
//   - secret: Webhook secret for request validation
//   - paymentUC: Payment use case for processing
//   - onPaymentMatch: Callback when payment is matched
func NewWebhookController(secret string, paymentUC *paymentuc.UseCase, onPaymentMatch PaymentConfirmHandler) *WebhookController {
	return &WebhookController{
		secret:         secret,
		paymentUC:      paymentUC,
		onPaymentMatch: onPaymentMatch,
		logger:         logger.Webhook,
		ready:          true, // Ready by default
	}
}

// ServeHTTP implements http.Handler interface.
// Note: When using Router, routes are defined in routes.go instead.
func (c *WebhookController) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Route requests (legacy mode without Router)
	switch {
	case r.URL.Path == "/webhook/payment" && r.Method == http.MethodPost:
		c.handlePaymentWebhook(w, r)
	case r.URL.Path == "/health" && r.Method == http.MethodGet:
		c.handleHealth(w, r)
	default:
		http.NotFound(w, r)
	}
}

// handleHealth returns a simple health check response.
// This is for liveness probe - checks if application is running.
func (c *WebhookController) handleHealth(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":        "ok",
		"pending_count": c.paymentUC.GetPendingCount(),
		"timestamp":     time.Now().Format(time.RFC3339),
	})
}

// handleReadiness checks if the application is ready to serve traffic.
// This is for readiness probe - checks if dependencies are healthy.
func (c *WebhookController) handleReadiness(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if !c.ready {
		w.WriteHeader(http.StatusServiceUnavailable)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":    "not_ready",
			"message":   "service is shutting down",
			"timestamp": time.Now().Format(time.RFC3339),
		})
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":        "ready",
		"pending_count": c.paymentUC.GetPendingCount(),
		"timestamp":     time.Now().Format(time.RFC3339),
	})
}

// SetReady sets the readiness status.
func (c *WebhookController) SetReady(ready bool) {
	c.ready = ready
}

// handlePaymentWebhook processes incoming payment notifications.
func (c *WebhookController) handlePaymentWebhook(w http.ResponseWriter, r *http.Request) {
	// Validate secret
	secret := r.Header.Get("X-Webhook-Secret")
	if secret == "" {
		// Fallback: check Authorization header for legacy support
		secret = r.Header.Get("Authorization")
	}

	if c.secret != "" && secret != c.secret {
		c.logger.Printf("Invalid webhook secret dari %s", r.RemoteAddr)
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Parse request body
	var payload entity.WebhookPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		c.logger.Printf("Gagal parse JSON: %v", err)
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	c.logger.Printf("Webhook diterima: app=%s | title=%s", payload.App, formatter.TruncateWithEllipsis(payload.Title, 30))

	// Process the notification
	pending, notification, err := c.paymentUC.ProcessNotification(context.Background(), &payload)
	if err != nil {
		c.logger.Printf("Gagal proses notifikasi: %v", err)
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{
			"status": "ignored",
			"reason": err.Error(),
		})
		return
	}

	if pending == nil {
		c.logger.Printf("Tidak ada pending yang cocok untuk Rp%d", notification.Amount)
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "unmatched",
			"amount":  notification.Amount,
			"message": "No pending payment found for this amount",
		})
		return
	}

	c.logger.Printf("Pembayaran dicocokkan: Rp%d | MsgID: %s", notification.Amount, pending.MessageID)

	// Call the confirmation handler
	if c.onPaymentMatch != nil {
		go c.onPaymentMatch(context.Background(), pending, notification)
	}

	// Respond success
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":     "matched",
		"amount":     notification.Amount,
		"message_id": pending.MessageID,
	})
}
