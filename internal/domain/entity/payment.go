// Package entity defines core business entities used across all layers.
package entity

import "time"

// PendingPayment represents a QRIS payment awaiting confirmation.
type PendingPayment struct {
	// QRIS message tracking
	MessageID         string    // ID of the QRIS image message sent by bot
	OriginalMessageID string    // ID of user's original #qris command (for reply)
	ChatID            string    // Chat JID
	SenderJID         string    // Sender JID
	SenderPhone       string    // Phone number without +
	Amount            int       // Payment amount
	CreatedAt         time.Time // When the QRIS was created
	IsSelfQris        bool      // True if created via self-message to customer
	GroupNotifMsgID   string    // ID of "QRIS TERKIRIM" notification in group (for reply threading)

	// Order data (from #qris form)
	Produk    string // Product name (determines target sheet)
	Nama      string // Customer name
	Email     string // Customer email
	Family    string // Family plan name
	Deskripsi string // Description/notes
	Kanal     string // Sales channel (default: WhatsApp)
	Akun      string // Account identifier (default: sender phone)
}

// DANANotification represents a parsed DANA payment notification.
type DANANotification struct {
	Amount     int       // Received amount
	RawMessage string    // Original notification message
	Timestamp  time.Time // When the notification was received
}

// Order represents an order to be logged to spreadsheet.
type Order struct {
	Produk         string    // Determines target sheet
	Nama           string    // B: Nama
	Email          string    // C: Email
	Family         string    // D: Family (for ChatGPT/Gemini) or empty
	KodeRedeem     string    // D: Kode Redeem (for Perplexity/YouTube)
	TanggalPesanan time.Time // E: Tanggal Pesanan
	Amount         int       // G: Amount/Nominal
	Kanal          string    // H: Kanal
	Akun           string    // I: Akun/Nomor/Username
}

// NewOrderFromPending creates an Order entity from a confirmed PendingPayment.
func NewOrderFromPending(pending *PendingPayment) *Order {
	return &Order{
		Produk:         pending.Produk,
		Nama:           pending.Nama,
		Email:          pending.Email,
		Family:         pending.Family,
		KodeRedeem:     "", // Will be filled manually by admin for redeem-based products
		TanggalPesanan: time.Now(),
		Amount:         pending.Amount,
		Kanal:          pending.Kanal,
		Akun:           pending.Akun,
	}
}

// WebhookPayload represents incoming webhook notification from Android app.
type WebhookPayload struct {
	App       string `json:"app"`       // App package name (e.g., "id.dana")
	Title     string `json:"title"`     // Notification title
	Message   string `json:"message"`   // Notification message body
	Timestamp string `json:"timestamp"` // Unix timestamp in milliseconds
}
