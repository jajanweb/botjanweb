// Package repository defines domain repository interfaces (ports).
package repository

import (
	"context"

	"github.com/exernia/botjanweb/internal/domain/entity"
)

// PendingPaymentRepository defines the contract for pending payment storage.
type PendingPaymentRepository interface {
	// Add registers a new pending payment.
	Add(payment *entity.PendingPayment)

	// Match finds and removes the oldest pending payment matching the amount (FIFO).
	Match(amount int) *entity.PendingPayment

	// GetByMessageID retrieves a pending payment by its WhatsApp message ID.
	GetByMessageID(messageID string) *entity.PendingPayment

	// RemoveByMessageID removes a pending payment by message ID.
	RemoveByMessageID(messageID string) bool

	// ListPending returns all pending payments (for admin/debugging).
	ListPending() []*entity.PendingPayment

	// StartCleanup starts the cleanup routine.
	StartCleanup(ctx context.Context)

	// StopCleanup stops the cleanup routine.
	StopCleanup()

	// Close closes the repository and releases resources.
	Close() error
}
