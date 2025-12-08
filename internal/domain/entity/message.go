// Package entity defines core business entities used across all layers.
// These are pure data structures with no external dependencies.
package entity

import "time"

// Message represents an incoming WhatsApp message.
type Message struct {
	ID             string
	ChatID         string // Chat JID as string (e.g., "123456789@g.us")
	SenderID       string // Sender JID as string
	SenderPhone    string // Phone number without +
	Text           string
	Timestamp      time.Time
	IsSelfMessage  bool   // True if message is from bot itself
	IsPrivateChat  bool   // True if message is in private chat (not group)
	RecipientPhone string // Phone number of recipient (for private chats)
}
