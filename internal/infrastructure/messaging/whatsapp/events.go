// Package whatsapp wraps WhatsMeow for WhatsApp connectivity.
package whatsapp

import (
	"context"
	"strings"
	"time"

	"github.com/exernia/botjanweb/internal/domain/entity"
	"github.com/exernia/botjanweb/pkg/helper/formatter"
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

	// For private chat, extract recipient phone properly
	if entityMsg.IsPrivateChat {
		// For self-messages (IsFromMe=true), recipient is the chat partner
		// For incoming messages, sender is the other person
		var recipientJID string

		if info.IsFromMe {
			// I sent message: recipient is the chat JID
			recipientJID = info.Chat.User
		} else {
			// They sent message: sender is the other person (but we don't use this for self-QRIS)
			recipientJID = info.Sender.User
		}

		// Only normalize if it looks like a phone number (not a group ID)
		// Phone numbers: 10-15 digits, groups: usually longer
		// s.whatsapp.net = user, g.us = group, lid = hidden
		if info.Chat.Server == "s.whatsapp.net" || info.Chat.Server == "lid" {
			entityMsg.RecipientPhone = formatter.NormalizePhone(recipientJID)
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
