// Package bot provides WhatsApp bot message parsing and handling.
package bot

import (
	"context"
	"log"
	"strings"

	"github.com/exernia/botjanweb/pkg/logger"

	service "github.com/exernia/botjanweb/internal/application/service"
	accountuc "github.com/exernia/botjanweb/internal/application/service/account"
	familyuc "github.com/exernia/botjanweb/internal/application/service/family"
	paymentuc "github.com/exernia/botjanweb/internal/application/service/payment"
	qrisuc "github.com/exernia/botjanweb/internal/application/service/qris"
	workspaceuc "github.com/exernia/botjanweb/internal/application/service/workspace"
	"github.com/exernia/botjanweb/internal/domain/entity"
	"github.com/exernia/botjanweb/pkg/helper/formatter"
)

// Handler processes incoming WhatsApp messages and routes them to appropriate use cases.
type Handler struct {
	qrisUC           *qrisuc.UseCase
	paymentUC        *paymentuc.UseCase
	familyUC         *familyuc.UseCase
	workspaceUC      *workspaceuc.UseCase
	accountUC        *accountuc.UseCase
	inventoryRepo    service.InventoryPort
	messaging        service.MessagingPort
	allowedSenders   []string
	logger           *log.Logger
	sheetAkunGoogle  string
	sheetAkunChatGPT string
	defaultKanal     string
}

// NewHandler creates a new message handler controller.
func NewHandler(
	qrisUC *qrisuc.UseCase,
	paymentUC *paymentuc.UseCase,
	familyUC *familyuc.UseCase,
	workspaceUC *workspaceuc.UseCase,
	accountUC *accountuc.UseCase,
	inventoryRepo service.InventoryPort,
	allowedSenders []string,
	sheetAkunGoogle string,
	sheetAkunChatGPT string,
	defaultKanal string,
) *Handler {
	return &Handler{
		qrisUC:           qrisUC,
		paymentUC:        paymentUC,
		familyUC:         familyUC,
		workspaceUC:      workspaceUC,
		accountUC:        accountUC,
		inventoryRepo:    inventoryRepo,
		allowedSenders:   allowedSenders,
		logger:           logger.Bot,
		sheetAkunGoogle:  sheetAkunGoogle,
		sheetAkunChatGPT: sheetAkunChatGPT,
		defaultKanal:     defaultKanal,
	}
}

// SetMessaging sets the messaging service (must be called before HandleMessage).
func (h *Handler) SetMessaging(m service.MessagingPort) {
	h.messaging = m
}

// HandleMessage processes an incoming message and dispatches to appropriate handler.
func (h *Handler) HandleMessage(ctx context.Context, msg *entity.Message) {
	// For self-messages (bot sending to customer), always allow
	// For other messages, check if sender is allowed
	if !msg.IsSelfMessage && !h.isFromAllowedSender(msg.SenderPhone) {
		h.logger.Printf("🚫 [DEBUG] Sender not allowed: %s (normalized: %s)",
			msg.SenderPhone, formatter.NormalizePhone(msg.SenderPhone))
		return
	}

	text := strings.TrimSpace(msg.Text)
	if text == "" {
		return
	}

	// Log command detection
	lowerText := strings.ToLower(text)
	if strings.HasPrefix(lowerText, "#") {
		h.logger.Printf("🔍 [DEBUG] Command detected: %q from %s", text, msg.SenderPhone)
	}

	// Route based on command prefix
	switch {
	case strings.HasPrefix(lowerText, "#qris"):
		h.handleQrisCommand(ctx, msg, text)
	case strings.HasPrefix(lowerText, "#addakun"):
		h.handleAddAkunCommand(ctx, msg, text)
	case strings.HasPrefix(lowerText, "#listakun"):
		h.handleListAkunCommand(ctx, msg, text)
	case strings.HasPrefix(lowerText, "#cekslot"):
		h.handleCekSlotCommand(ctx, msg, text)
	case strings.HasPrefix(lowerText, "#cekkode"):
		h.handleCekKodeCommand(ctx, msg, text)
	case strings.HasPrefix(lowerText, "#inputkode"):
		h.handleInputKodeCommand(ctx, msg, text)
	}
}

// isFromAllowedSender checks if the sender is allowed to use the bot.
func (h *Handler) isFromAllowedSender(phone string) bool {
	// No restriction if empty or wildcard "*"
	if len(h.allowedSenders) == 0 {
		return true
	}
	if len(h.allowedSenders) == 1 && h.allowedSenders[0] == "*" {
		h.logger.Printf("✅ [DEBUG] Wildcard mode: allowing all senders")
		return true
	}

	senderNormalized := formatter.NormalizePhone(phone)
	for _, allowed := range h.allowedSenders {
		if formatter.NormalizePhone(allowed) == senderNormalized {
			return true
		}
	}
	return false
}
