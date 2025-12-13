// Package usecase defines business logic interfaces (ports).
package usecase

import (
	"context"

	"github.com/exernia/botjanweb/internal/domain/entity"
)

// --- Input Ports (Use Cases) ---

// QrisUseCase handles QRIS generation business logic.
type QrisUseCase interface {
	// GenerateQRIS creates a dynamic QRIS and returns the result.
	GenerateQRIS(ctx context.Context, cmd *entity.QrisCommand, msg *entity.Message) (*entity.QrisResult, error)
}

// PaymentUseCase handles payment matching business logic.
type PaymentUseCase interface {
	// RegisterPending adds a pending payment to the store.
	RegisterPending(pending *entity.PendingPayment)
	// GetPendingCount returns total pending payments.
	GetPendingCount() int
}

// --- Output Ports (Driven Adapters) ---

// MessagingPort defines messaging operations.
type MessagingPort interface {
	// SendText sends a text message to the default chat.
	SendText(ctx context.Context, text string) error
	// SendTextTo sends a text message to a specific chat.
	SendTextTo(ctx context.Context, chatID string, text string) error
	// SendTextReply sends a reply to a specific message.
	SendTextReply(ctx context.Context, chatID, text, quotedMsgID, quotedSenderID string) error
	// SendImage sends an image with caption, returns message ID.
	SendImage(ctx context.Context, imageData []byte, caption string) (string, error)
	// SendImageTo sends an image to a specific chat.
	SendImageTo(ctx context.Context, chatID string, imageData []byte, caption string) (string, error)
	// SendTextToGroup sends a text message to the configured group, returns message ID.
	SendTextToGroup(ctx context.Context, text string) (string, error)
	// SendTextReplyToGroup sends a reply to a specific message in the configured group.
	SendTextReplyToGroup(ctx context.Context, text, quotedMsgID string) error
	// SendImageToGroup sends an image to the configured group.
	SendImageToGroup(ctx context.Context, imageData []byte, caption string) (string, error)
	// GetOwnID returns the bot's own ID.
	GetOwnID() string
	// GetGroupJID returns the configured group JID.
	GetGroupJID() string
	// GetContactName returns the display name for a contact phone number.
	GetContactName(ctx context.Context, phone string) string
}

// QrisGeneratorPort defines QRIS generation operations.
type QrisGeneratorPort interface {
	// GenerateDynamicQRIS creates a dynamic QRIS with image.
	GenerateDynamicQRIS(baseQR string, amount int, deskripsi string) (qrisString string, imageData []byte, err error)
}

// PendingStorePort defines pending payment storage operations.
type PendingStorePort interface {
	// Add registers a new pending payment.
	Add(p *entity.PendingPayment)
	// Match finds and removes a pending payment by amount (FIFO).
	Match(amount int) *entity.PendingPayment
	// Count returns total pending payments.
	Count() int
	// StartCleanup starts background cleanup routine.
	StartCleanup()
	// StopCleanup stops the cleanup routine.
	StopCleanup()
	// Close closes the store and releases resources (for PostgreSQL).
	Close() error
}

// TransactionLogPort defines transaction logging operations.
type TransactionLogPort interface {
	// LogOrder logs an order record.
	LogOrder(ctx context.Context, order *entity.Order) error
}

// FamilyValidatorPort defines family validation operations.
type FamilyValidatorPort interface {
	// ValidateFamily checks if a family email exists in Akun Google sheet.
	// Returns true if found, false otherwise.
	ValidateFamily(ctx context.Context, familyEmail string) (bool, error)
	// CountFamilySlots counts how many slots are used for a family in Gemini sheet.
	CountFamilySlots(ctx context.Context, family string) (int, error)
}

// AccountRepositoryPort defines account management operations.
type AccountRepositoryPort interface {
	// AddAkunGoogle adds a new Google account to Akun Google sheet.
	AddAkunGoogle(ctx context.Context, akun *entity.AkunGoogle) error
	// AddAkunChatGPT adds a new ChatGPT account to Akun ChatGPT sheet.
	AddAkunChatGPT(ctx context.Context, akun *entity.AkunChatGPT) error
	// GetAccountListResult fetches all accounts and returns a summary with availability counts.
	GetAccountListResult(ctx context.Context) (*entity.AccountListResult, error)
}

// InventoryPort defines inventory checking operations.
type InventoryPort interface {
	// GetSlotAvailability returns slot availability for families/workspaces.
	// product: "ChatGPT" or "Gemini"
	// availableOnly: if true, only return items with available slots
	GetSlotAvailability(ctx context.Context, product string, availableOnly bool) (*entity.SlotAvailabilityResult, error)
	// GetRedeemCodeAvailability returns available Perplexity redeem codes.
	// availableOnly: if true, only return codes not yet activated
	GetRedeemCodeAvailability(ctx context.Context, availableOnly bool) (*entity.RedeemCodeResult, error)
	// AddRedeemCode adds a new redeem code to Kode Perplexity sheet.
	AddRedeemCode(ctx context.Context, email, kodeRedeem string) error
}
