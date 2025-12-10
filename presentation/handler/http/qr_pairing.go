// Package http provides HTTP controllers.
package http

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/exernia/botjanweb/internal/infrastructure/messaging/whatsapp"
	"github.com/exernia/botjanweb/pkg/logger"
	"github.com/exernia/botjanweb/presentation/view"
	"github.com/skip2/go-qrcode"
)

// QRPairingController handles WhatsApp QR code pairing via web interface.
type QRPairingController struct {
	waClient     *whatsapp.Client
	view         *view.PairingView
	logger       *log.Logger
	mu           sync.RWMutex
	currentQR    string
	lastUpdate   time.Time
	pairingToken string // Simple auth token

	// Simple rate limiting for security (personal project, not production-grade)
	rateLimitMu     sync.Mutex
	failedAttempts  map[string]int       // IP -> failed count
	lastAttemptTime map[string]time.Time // IP -> last attempt time
}

// NewQRPairingController creates a new QR pairing controller.
func NewQRPairingController(waClient *whatsapp.Client, pairingToken string) (*QRPairingController, error) {
	pairingView, err := view.NewPairingView()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize pairing view: %w", err)
	}

	return &QRPairingController{
		waClient:        waClient,
		view:            pairingView,
		logger:          logger.New("[PAIRING] "),
		pairingToken:    pairingToken,
		failedAttempts:  make(map[string]int),
		lastAttemptTime: make(map[string]time.Time),
	}, nil
}

// checkRateLimit checks if an IP is rate limited (simple protection for personal project).
// Returns true if request should be blocked.
func (c *QRPairingController) checkRateLimit(ip string) bool {
	c.rateLimitMu.Lock()
	defer c.rateLimitMu.Unlock()

	const (
		maxFailedAttempts = 5                // Block after 5 failed attempts
		blockDuration     = 15 * time.Minute // Block for 15 minutes
		resetAfter        = 5 * time.Minute  // Reset counter after 5 minutes of no attempts
	)

	now := time.Now()

	// Check if IP is currently blocked
	if lastAttempt, exists := c.lastAttemptTime[ip]; exists {
		failedCount := c.failedAttempts[ip]

		// If blocked and block period not expired yet
		if failedCount >= maxFailedAttempts && now.Sub(lastAttempt) < blockDuration {
			return true // Still blocked
		}

		// Reset counter if last attempt was long ago
		if now.Sub(lastAttempt) > resetAfter {
			delete(c.failedAttempts, ip)
			delete(c.lastAttemptTime, ip)
		}
	}

	return false // Not blocked
}

// recordFailedAttempt records a failed authentication attempt.
func (c *QRPairingController) recordFailedAttempt(ip string) {
	c.rateLimitMu.Lock()
	defer c.rateLimitMu.Unlock()

	c.failedAttempts[ip]++
	c.lastAttemptTime[ip] = time.Now()

	if c.failedAttempts[ip] >= 5 {
		c.logger.Printf("IP %s has been blocked after %d failed attempts", ip, c.failedAttempts[ip])
	}
}

// resetFailedAttempts resets failed attempts for an IP (called on successful auth).
func (c *QRPairingController) resetFailedAttempts(ip string) {
	c.rateLimitMu.Lock()
	defer c.rateLimitMu.Unlock()

	delete(c.failedAttempts, ip)
	delete(c.lastAttemptTime, ip)
}

// HandlePairingPage serves the QR code pairing HTML page.
func (c *QRPairingController) HandlePairingPage(w http.ResponseWriter, r *http.Request) {
	ip := r.RemoteAddr

	// Check rate limiting first (protect from brute force)
	if c.checkRateLimit(ip) {
		c.logger.Printf("Rate limit exceeded for %s - request blocked", ip)
		http.Error(w, "Too many failed attempts - please try again later", http.StatusTooManyRequests)
		return
	}

	// Security: Authenticate via header (not URL to prevent exposure in logs)
	// Supports X-Pairing-Token (recommended) or Authorization header
	token := r.Header.Get("X-Pairing-Token")
	if token == "" {
		token = r.Header.Get("Authorization")
	}

	if token != c.pairingToken {
		c.logger.Printf("Unauthorized pairing attempt from %s", ip)
		c.recordFailedAttempt(ip)
		http.Error(w, "Unauthorized - Valid X-Pairing-Token header required", http.StatusUnauthorized)
		return
	}

	// Successful auth - reset failed attempts
	c.resetFailedAttempts(ip)
	c.logger.Printf("Authorized pairing page access from %s", ip)

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

	// Start pairing process if not already started
	c.mu.Lock()
	pairingActive := c.currentQR != "" || !c.lastUpdate.IsZero()
	c.mu.Unlock()

	if !pairingActive {
		c.logger.Println("Starting QR pairing on-demand...")
		// Use background context - pairing must survive beyond this HTTP request
		go func() {
			if err := c.StartPairing(context.Background()); err != nil {
				c.logger.Printf("Failed to start pairing: %v", err)
			}
		}()
		// Give pairing a moment to start before rendering page
		time.Sleep(100 * time.Millisecond)
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
	ip := r.RemoteAddr

	// Check rate limiting (lighter than pairing page - this is polling)
	if c.checkRateLimit(ip) {
		http.Error(w, "Too many requests", http.StatusTooManyRequests)
		return
	}

	// Security: Authenticate via header (not URL to prevent exposure in logs)
	// Supports X-Pairing-Token (recommended) or Authorization header
	token := r.Header.Get("X-Pairing-Token")
	if token == "" {
		token = r.Header.Get("Authorization")
	}

	if token != c.pairingToken {
		// No detailed log spam for polling requests - just record and reject silently
		c.recordFailedAttempt(ip)
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

	c.logger.Println("QR channel obtained, starting event handler goroutine")

	// Start goroutine to handle QR events
	go func() {
		c.logger.Println("QR event handler goroutine started")
		for evt := range qrChan {
			c.logger.Printf("Received QR event: %s", evt.Event)
			switch evt.Event {
			case "code":
				// Generate QR code as base64
				qrBase64 := generateQRDataURL(evt.Code)

				c.mu.Lock()
				c.currentQR = qrBase64
				c.lastUpdate = time.Now()
				c.mu.Unlock()

				c.logger.Println("✅ QR code generated and stored")

			case "success":
				c.logger.Println("✅ Pairing successful")
				c.mu.Lock()
				c.currentQR = ""
				c.mu.Unlock()
				return

			case "timeout":
				log.Println("[PAIRING] ⏱️ QR code expired/timeout")
				c.mu.Lock()
				c.currentQR = ""
				c.mu.Unlock()
				return
			}
		}
		c.logger.Println("QR channel closed")
	}()

	c.logger.Println("Connecting WhatsApp websocket...")
	// Now connect the websocket (event handler should already be set)
	if err := c.waClient.ConnectWithEventHandler(); err != nil {
		return fmt.Errorf("failed to connect WhatsApp: %w", err)
	}

	c.logger.Println("✅ WhatsApp websocket connected")
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
