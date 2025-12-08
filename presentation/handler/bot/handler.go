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
	"github.com/exernia/botjanweb/internal/domain/entity"
	"github.com/exernia/botjanweb/pkg/helper/formatter"
)

// Handler processes incoming WhatsApp messages and routes them to appropriate use cases.
type Handler struct {
	qrisUC           *qrisuc.UseCase
	paymentUC        *paymentuc.UseCase
	familyUC         *familyuc.UseCase
	accountUC        *accountuc.UseCase
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
	accountUC *accountuc.UseCase,
	allowedSenders []string,
	sheetAkunGoogle string,
	sheetAkunChatGPT string,
	defaultKanal string,
) *Handler {
	return &Handler{
		qrisUC:           qrisUC,
		paymentUC:        paymentUC,
		familyUC:         familyUC,
		accountUC:        accountUC,
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
		return
	}

	text := strings.TrimSpace(msg.Text)
	if text == "" {
		return
	}

	// Route based on command prefix
	switch {
	case strings.HasPrefix(strings.ToLower(text), "#qris"):
		h.handleQrisCommand(ctx, msg, text)
	case strings.HasPrefix(strings.ToLower(text), "#addakun"):
		h.handleAddAkunCommand(ctx, msg, text)
	case strings.HasPrefix(strings.ToLower(text), "#listakun"):
		h.handleListAkunCommand(ctx, msg, text)
	}
}

// isFromAllowedSender checks if the sender is allowed to use the bot.
func (h *Handler) isFromAllowedSender(phone string) bool {
	if len(h.allowedSenders) == 0 {
		return true // No restriction
	}

	senderNormalized := formatter.NormalizePhone(phone)
	for _, allowed := range h.allowedSenders {
		if formatter.NormalizePhone(allowed) == senderNormalized {
			return true
		}
	}
	return false
}
