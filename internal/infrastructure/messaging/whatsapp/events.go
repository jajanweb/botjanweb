// Package whatsapp wraps WhatsMeow for WhatsApp connectivity.
package whatsapp

import (
	"context"
	"strings"
	"time"

	"github.com/exernia/botjanweb/internal/domain/entity"
	"github.com/exernia/botjanweb/pkg/helper/formatter"
	"go.mau.fi/whatsmeow/types"
	"go.mau.fi/whatsmeow/types/events"
)

// eventHandler handles WhatsMeow events and forwards messages to the bot handler.
func (c *Client) eventHandler(evt interface{}) {
	switch v := evt.(type) {
	case *events.Message:
		c.handleIncomingMessage(v)

	case *events.Connected:
		c.logger.Printf("✅ WebSocket connected to WhatsApp servers")

	case *events.PairSuccess:
		c.logger.Printf("🎉 Pairing successful! Device: %s", v.ID.String())

	case *events.LoggedOut:
		c.logger.Printf("🚪 Logged out from WhatsApp (reason: %s)", v.Reason)

	case *events.Disconnected:
		c.logger.Println("⚠️ Disconnected from WhatsApp")
	}
}

// handleIncomingMessage converts WhatsMeow message to entity.Message.
func (c *Client) handleIncomingMessage(evt *events.Message) {
	if c.handler == nil {
		c.logger.Println("⚠️ [DEBUG] Handler is nil, skipping message")
		return
	}

	// Skip unavailable messages that whatsmeow couldn't decrypt
	// These show as WARN in logs: "Unavailable message ... (type: "")"
	// Common with first messages from new contacts during E2E setup
	if evt.UnavailableRequestID != "" {
		c.logger.Printf("⚠️ [DEBUG] Skipping unavailable message %s from %s (E2E key not synced yet)",
			evt.Info.ID, evt.Info.Sender.User)
		return
	}

	msg := evt.Message
	if msg == nil {
		c.logger.Printf("⚠️ [DEBUG] Message content is nil for %s from %s",
			evt.Info.ID, evt.Info.Sender.User)
		return
	}

	// Skip broadcast messages
	if evt.Info.IsIncomingBroadcast() {
		return
	}

	text := ""
	switch {
	case msg.GetConversation() != "":
		text = msg.GetConversation()
	case msg.ExtendedTextMessage != nil && msg.ExtendedTextMessage.Text != nil:
		text = msg.ExtendedTextMessage.GetText()
	default:
		// Only handle text messages
		return
	}

	// Log incoming message for debugging
	c.logger.Printf("📨 [DEBUG] Received: from=%s server=%s chat=%s text=%q",
		evt.Info.Sender.User, evt.Info.Sender.Server, evt.Info.Chat.String(), truncateText(text, 50))

	info := evt.Info

	// Determine sender phone - handle LID (Linked ID) format
	senderPhone := info.Sender.User
	if info.Sender.Server == "lid" {
		// LID users have hidden phone numbers - try to resolve
		pn, err := c.wm.Store.LIDs.GetPNForLID(context.Background(), info.Sender)
		if err == nil && !pn.IsEmpty() {
			senderPhone = pn.User
			c.logger.Printf("📞 Resolved sender LID %s → PN %s", info.Sender.User, pn.User)
		} else {
			c.logger.Printf("⚠️ [DEBUG] Could not resolve sender LID %s to phone number", info.Sender.User)
			// Keep LID as fallback - may still match if ALLOWED_SENDERS contains LID
		}
	}

	entityMsg := &entity.Message{
		ID:            info.ID,
		ChatID:        info.Chat.String(),
		SenderID:      info.Sender.String(),
		SenderPhone:   senderPhone,
		Text:          strings.TrimSpace(text),
		Timestamp:     info.Timestamp,
		IsSelfMessage: info.IsFromMe,
		IsPrivateChat: info.Chat.Server != "g.us",
	}

	// For private chat, extract recipient phone
	if entityMsg.IsPrivateChat {
		// For outgoing messages (self-QRIS), use chat partner's JID
		if info.IsFromMe {
			// Only extract phone if it's s.whatsapp.net (regular phone number)
			switch info.Chat.Server {
			case "s.whatsapp.net":
				entityMsg.RecipientPhone = formatter.NormalizePhone(info.Chat.User)
			case "lid":
				// LID users have hidden phone numbers
				// Try to resolve using the LID mapping store
				pn, err := c.wm.Store.LIDs.GetPNForLID(context.Background(), info.Chat)
				if err == nil && !pn.IsEmpty() {
					entityMsg.RecipientPhone = formatter.NormalizePhone(pn.User)
					c.logger.Printf("📞 Resolved LID %s → PN %s", info.Chat.User, pn.User)
				} else {
					// Fallback: Try calling GetUserInfo to get updated mapping
					entityMsg.RecipientPhone = c.tryResolveLIDPhone(info.Chat)
				}
			}
		}
	}

	// Preserve timestamp if zero
	if entityMsg.Timestamp.IsZero() {
		entityMsg.Timestamp = time.Now()
	}

	// Run handler in goroutine to avoid blocking WhatsApp event loop
	// Use background context - let operations complete naturally
	go func() {
		ctx := context.Background()
		c.handler(ctx, entityMsg)
	}()
}

// tryResolveLIDPhone attempts to resolve a phone number from a LID JID.
// It first checks the local cache/store, then tries GetUserInfo as a fallback.
// Returns "Private User" if resolution fails.
func (c *Client) tryResolveLIDPhone(lid types.JID) string {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Try GetUserInfo which may trigger LID sync and populate mapping in the store
	_, err := c.wm.GetUserInfo(ctx, []types.JID{lid})
	if err != nil {
		c.logger.Printf("⚠️ GetUserInfo failed for LID %s: %v", lid.User, err)
		return "Private User"
	}

	// After GetUserInfo, check if we got a LID-PN mapping stored
	pn, err := c.wm.Store.LIDs.GetPNForLID(ctx, lid)
	if err == nil && !pn.IsEmpty() {
		c.logger.Printf("📞 Resolved LID %s → PN %s (via GetUserInfo)", lid.User, pn.User)
		return formatter.NormalizePhone(pn.User)
	}

	return "Private User"
}

// truncateText truncates text for logging purposes.
func truncateText(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}
