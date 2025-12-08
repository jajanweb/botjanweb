// Package payment contains payment-related use cases and domain services.
package payment

import (
	"context"

	"github.com/exernia/botjanweb/internal/domain/entity"
)

// NotificationPort defines the interface for sending notifications.
// This port abstracts the notification infrastructure (WhatsApp, email, etc.)
// allowing the use case to remain independent of specific notification implementations.
type NotificationPort interface {
	// SendPaymentConfirmation sends payment confirmation to customer
	SendPaymentConfirmation(ctx context.Context, pending *entity.PendingPayment, message string) error

	// RevokeQRISImage revokes (deletes) the QRIS image message
	RevokeQRISImage(ctx context.Context, chatID, messageID string) error

	// SendGroupNotification sends notification to group
	SendGroupNotification(ctx context.Context, message string, replyToMessageID string) error
}

// SheetsPort defines the interface for Google Sheets operations.
type SheetsPort interface {
	// LogOrder saves order to Google Sheets
	LogOrder(ctx context.Context, order *entity.Order) error
}
