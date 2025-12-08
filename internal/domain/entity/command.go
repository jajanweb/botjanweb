// Package entity defines core business entities used across all layers.
package entity

// Command prefixes for bot commands.
const (
	CmdQris    = "#qris"
	CmdAddAkun = "#addakun"
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
