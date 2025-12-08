// Package adapters contains infrastructure adapters that implement domain ports.
// These adapters follow the Hexagonal Architecture pattern (Ports & Adapters),
// allowing the domain/use-case layer to remain independent of infrastructure details.
package adapters

import (
	"context"

	"github.com/exernia/botjanweb/internal/application/service/payment"
	"github.com/exernia/botjanweb/internal/domain/entity"
	infra "github.com/exernia/botjanweb/internal/infrastructure/messaging/whatsapp"
)

// WhatsAppNotificationAdapter implements the NotificationPort interface using WhatsApp.
// This adapter translates use-case notification requests into WhatsApp-specific operations.
type WhatsAppNotificationAdapter struct {
	waClient *infra.Client
	groupJID string // Group JID for sending group notifications
}

// NewWhatsAppNotificationAdapter creates a new WhatsApp notification adapter.
func NewWhatsAppNotificationAdapter(waClient *infra.Client, groupJID string) payment.NotificationPort {
	return &WhatsAppNotificationAdapter{
		waClient: waClient,
		groupJID: groupJID,
	}
}

// SendPaymentConfirmation sends payment confirmation to customer via WhatsApp.
func (a *WhatsAppNotificationAdapter) SendPaymentConfirmation(ctx context.Context, pending *entity.PendingPayment, message string) error {
	// Guard against nil client
	if a == nil || a.waClient == nil {
		return nil // Silently skip if WhatsApp client not available
	}

	return a.waClient.SendTextReply(
		ctx,
		pending.ChatID,
		message,
		pending.OriginalMessageID,
		pending.SenderJID,
	)
}

// RevokeQRISImage revokes (deletes) the QRIS image message from WhatsApp.
func (a *WhatsAppNotificationAdapter) RevokeQRISImage(ctx context.Context, chatID, messageID string) error {
	// Guard against nil client
	if a == nil || a.waClient == nil {
		return nil // Silently skip if WhatsApp client not available
	}

	return a.waClient.RevokeMessage(ctx, chatID, messageID)
}

// SendGroupNotification sends notification to the configured group via WhatsApp.
func (a *WhatsAppNotificationAdapter) SendGroupNotification(ctx context.Context, message string, replyToMessageID string) error {
	// Guard against nil client
	if a == nil || a.waClient == nil {
		return nil // Silently skip if WhatsApp client not available
	}

	// If we have a message to reply to, use reply
	if replyToMessageID != "" {
		return a.waClient.SendTextReplyToGroup(ctx, message, replyToMessageID)
	}

	// Otherwise, send as new message
	_, err := a.waClient.SendTextToGroup(ctx, message)
	return err
}
