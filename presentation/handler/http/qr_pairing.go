// Package http provides HTTP controllers.
package http

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/exernia/botjanweb/internal/infrastructure/messaging/whatsapp"
	"github.com/exernia/botjanweb/presentation/view"
	"github.com/skip2/go-qrcode"
)

// QRPairingController handles WhatsApp QR code pairing via web interface.
type QRPairingController struct {
	waClient     *whatsapp.Client
	view         *view.PairingView
	mu           sync.RWMutex
	currentQR    string
	lastUpdate   time.Time
	pairingToken string // Simple auth token
}

// NewQRPairingController creates a new QR pairing controller.
func NewQRPairingController(waClient *whatsapp.Client, pairingToken string) (*QRPairingController, error) {
	pairingView, err := view.NewPairingView()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize pairing view: %w", err)
	}

	return &QRPairingController{
		waClient:     waClient,
		view:         pairingView,
		pairingToken: pairingToken,
	}, nil
}

// HandlePairingPage serves the QR code pairing HTML page.
func (c *QRPairingController) HandlePairingPage(w http.ResponseWriter, r *http.Request) {
	// Simple authentication check
	token := r.URL.Query().Get("token")
	if token != c.pairingToken {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Check if already paired
	if c.waClient.IsLoggedIn() {
		err := c.view.RenderSuccess(w, view.PairingSuccessData{
			DeviceID: c.waClient.GetOwnID(),
		})
		if err != nil {
			http.Error(w, "Failed to render page", http.StatusInternalServerError)
		}
		return
	}

	// Show pairing page
	err := c.view.RenderPage(w, view.PairingPageData{
		Token: c.pairingToken,
	})
	if err != nil {
		http.Error(w, "Failed to render page", http.StatusInternalServerError)
	}
}

// HandleQRCodeAPI returns the current QR code as JSON (for AJAX polling).
func (c *QRPairingController) HandleQRCodeAPI(w http.ResponseWriter, r *http.Request) {
	// Simple authentication check
	token := r.URL.Query().Get("token")
	if token != c.pairingToken {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	c.mu.RLock()
	currentQR := c.currentQR
	lastUpdate := c.lastUpdate
	c.mu.RUnlock()

	response := map[string]interface{}{
		"timestamp": lastUpdate.Unix(),
	}

	// Check if already logged in
	if c.waClient.IsLoggedIn() {
		response["status"] = "success"
		response["device_id"] = c.waClient.GetOwnID()
	} else if currentQR != "" {
		response["qr"] = currentQR
		response["status"] = "waiting"
	} else {
		// QR not ready yet or expired
		response["status"] = "loading"
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// StartPairing initiates the QR code pairing process.
// This should be called BEFORE WAClient.Run() to avoid race condition.
func (c *QRPairingController) StartPairing(ctx context.Context) error {
	if c.waClient.IsLoggedIn() {
		return nil // Already logged in
	}

	// Get QR channel BEFORE connecting
	qrChan, err := c.waClient.GetQRChannelForPairing(ctx)
	if err != nil {
		return fmt.Errorf("failed to get QR channel: %w", err)
	}

	// Start goroutine to handle QR events
	go func() {
		for evt := range qrChan {
			switch evt.Event {
			case "code":
				// Generate QR code as data URL
				qrDataURL := generateQRDataURL(evt.Code)

				c.mu.Lock()
				c.currentQR = qrDataURL
				c.lastUpdate = time.Now()
				c.mu.Unlock()

			case "success":
				c.mu.Lock()
				c.currentQR = ""
				c.mu.Unlock()
				return

			case "timeout":
				c.mu.Lock()
				c.currentQR = ""
				c.mu.Unlock()
				return
			}
		}
	}()

	// Now connect the websocket (event handler should already be set)
	if err := c.waClient.ConnectWithEventHandler(); err != nil {
		return fmt.Errorf("failed to connect for pairing: %w", err)
	}

	return nil
}

// generateQRDataURL converts QR code text to data URL for HTML display.
// This generates the QR code locally without sending sensitive data to external APIs.
func generateQRDataURL(code string) string {
	// Generate QR code as PNG with error correction level Medium
	qrCode, err := qrcode.New(code, qrcode.Medium)
	if err != nil {
		return ""
	}

	// Set size to 300x300 pixels
	qrCode.DisableBorder = false

	// Generate PNG bytes
	pngBytes, err := qrCode.PNG(300)
	if err != nil {
		return ""
	}

	// Convert to data URL (base64 encoded)
	base64Str := base64.StdEncoding.EncodeToString(pngBytes)
	return fmt.Sprintf("data:image/png;base64,%s", base64Str)
}
