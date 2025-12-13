// Package entity defines core business entities used across all layers.
package entity

// Command prefixes for bot commands.
const (
	CmdQris      = "#qris"
	CmdAddAkun   = "#addakun"
	CmdCekSlot   = "#cekslot"
	CmdCekKode   = "#cekkode"
	CmdInputKode = "#inputkode"
)

// QrisCommand represents a parsed #qris command.
type QrisCommand struct {
	// Common fields
	Produk string // Product name (Gemini, ChatGPT, etc.)
	Nama   string // Customer name
	Email  string // Customer email
	Amount int    // Payment amount

	// Gemini-specific
	Family string // Family plan name

	// ChatGPT-specific
	Workspace string // Workspace name
	Paket     string // Package type

	// Optional fields
	Kanal string // Sales channel (default: Threads)
	Akun  string // Account identifier

	// Self-QRIS specific
	TargetPhone string // Target phone number for self-QRIS (e.g., untuk:6281234567890)
	Deskripsi   string // Description/notes for QRIS

	// Mode flags
	IsFormMode  bool   // True if parsed from form format
	IsHelpMode  bool   // True if command sent without parameters
	ProductType string // Raw parameter ("google", "chatgpt")
}

// CekSlotCommand represents a parsed #cekslot command.
type CekSlotCommand struct {
	Product       string // "chatgpt" or "gemini"
	AvailableOnly bool   // Only show available slots
	IsHelpMode    bool   // True if command sent without parameters
}

// InputKodeCommand represents a parsed #inputkode command.
type InputKodeCommand struct {
	Email      string // Email for the redeem code
	KodeRedeem string // The redeem code
	IsHelpMode bool   // True if command sent without parameters
}
