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
		return
	}

	// Skip unavailable messages that whatsmeow couldn't decrypt
	// These show as WARN in logs: "Unavailable message ... (type: "")"
	// Common with first messages from new contacts during E2E setup
	if evt.UnavailableRequestID != "" {
		return
	}

	msg := evt.Message
	if msg == nil {
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

	info := evt.Info
	entityMsg := &entity.Message{
		ID:            info.ID,
		ChatID:        info.Chat.String(),
		SenderID:      info.Sender.String(),
		SenderPhone:   info.Sender.User,
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
