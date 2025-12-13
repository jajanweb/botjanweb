// Package bot provides WhatsApp bot message parsing and handling.
package bot

import (
	"context"
	"fmt"
	"time"

	"github.com/exernia/botjanweb/internal/domain/entity"
	"github.com/exernia/botjanweb/pkg/helper/formatter"
	"github.com/exernia/botjanweb/pkg/helper/parser"
	"github.com/exernia/botjanweb/presentation/template"
)

// handleQrisCommand routes #qris commands for normal users or self-messages.
func (h *Handler) handleQrisCommand(ctx context.Context, msg *entity.Message, text string) {
	// Self-QRIS: bot sends to customer in private chat
	if msg.IsSelfMessage && msg.IsPrivateChat {
		h.handleQrisSelf(ctx, msg, text)
		return
	}

	cmd, err := parser.ParseQrisCommand(text, h.defaultKanal)
	if err != nil {
		h.sendErrorReply(ctx, msg, "❌ "+err.Error())
		return
	}

	if cmd.IsHelpMode {
		h.sendQrisHelp(ctx, msg, cmd.ProductType)
		return
	}

	if cmd.IsFormMode {
		h.handleQrisForm(ctx, msg, cmd)
		return
	}

	// Legacy format not allowed in group
	h.sendErrorReply(ctx, msg, "❌ Format tidak valid. Gunakan format:\n\n#qris google\natau\n#qris chatgpt\n\nuntuk melihat form yang benar.")
}

// handleQrisForm handles form-based #qris for products (Gemini, ChatGPT, etc.).
func (h *Handler) handleQrisForm(ctx context.Context, msg *entity.Message, cmd *entity.QrisCommand) {
	h.logger.Printf("💳 QRIS Form: %s | %s | %s", cmd.Produk, cmd.Nama, cmd.Email)

	// Validate Family if provided (for Gemini)
	if _, errorMsg, err := h.validateFamilyOrWorkspace(ctx, "Family", cmd.Family); err != nil {
		h.sendErrorReply(ctx, msg, errorMsg)
		// Also send to group
		if _, err := h.messaging.SendTextToGroup(ctx, errorMsg); err != nil {
			h.logger.Printf("⚠️ Gagal kirim error ke grup: %v", err)
		}
		return
	}

	// Validate Workspace if provided (for ChatGPT)
	if _, errorMsg, err := h.validateFamilyOrWorkspace(ctx, "Workspace", cmd.Workspace); err != nil {
		h.sendErrorReply(ctx, msg, errorMsg)
		// Also send to group
		if _, err := h.messaging.SendTextToGroup(ctx, errorMsg); err != nil {
			h.logger.Printf("⚠️ Gagal kirim error ke grup: %v", err)
		}
		return
	}

	result, err := h.qrisUC.GenerateQRIS(ctx, cmd, msg)
	if err != nil {
		h.logger.Printf("❌ Gagal generate QRIS: %v", err)
		h.sendErrorReply(ctx, msg, "❌ Gagal generate QRIS")
		return
	}

	// Send QRIS image without caption to group (clean)
	qrisMsgID, err := h.messaging.SendImage(ctx, result.ImageData, "")
	if err != nil {
		h.logger.Printf("❌ Gagal kirim QRIS: %v", err)
		return
	}

	// Send caption as reply to the QRIS image in group
	caption := template.BuildQrisFormCaption(cmd)
	if err := h.messaging.SendTextReplyToGroup(ctx, caption, qrisMsgID); err != nil {
		h.logger.Printf("⚠️ Gagal kirim caption: %v (QRIS tetap terkirim)", err)
		// Continue - QRIS already sent successfully
	}

	pending := &entity.PendingPayment{
		MessageID:         qrisMsgID,
		OriginalMessageID: msg.ID,
		ChatID:            msg.ChatID,
		SenderJID:         msg.SenderID,
		SenderPhone:       msg.SenderPhone,
		Amount:            result.Amount,
		CreatedAt:         time.Now(),
		Produk:            cmd.Produk,
		Nama:              cmd.Nama,
		Email:             cmd.Email,
		Family:            cmd.Family,
		Deskripsi:         cmd.Deskripsi,
		Kanal:             cmd.Kanal,
		Akun:              cmd.Akun,
	}

	h.paymentUC.RegisterPending(pending)
	h.logger.Printf("✅ QRIS terkirim (form), pending registered: MsgID=%s", qrisMsgID)
}

// handleQrisSelf processes #qris commands sent by bot itself in private chat.
// Required fields: nominal and produk. Other fields are optional.
func (h *Handler) handleQrisSelf(ctx context.Context, msg *entity.Message, text string) {
	cmd, err := parser.ParseSelfQrisCommand(text)
	if err != nil {
		h.logger.Printf("❌ Self-QRIS parse error: %v", err)
		return
	}

	h.logger.Printf("💳 Self-QRIS: Rp%d | Ke: %s", cmd.Amount, msg.RecipientPhone)

	// Validate Family if provided (for Gemini)
	if _, errorMsg, err := h.validateFamilyOrWorkspace(ctx, "Family", cmd.Family); err != nil {
		// Send error to group
		errorMsg = "❌ Self-QRIS Gagal: " + errorMsg[2:] // Remove "❌ " prefix and add Self-QRIS prefix
		if _, err := h.messaging.SendTextToGroup(ctx, errorMsg); err != nil {
			h.logger.Printf("⚠️ Gagal kirim error ke grup: %v", err)
		}
		return
	}

	// Validate Workspace if provided (for ChatGPT)
	if _, errorMsg, err := h.validateFamilyOrWorkspace(ctx, "Workspace", cmd.Workspace); err != nil {
		// Send error to group
		errorMsg = "❌ Self-QRIS Gagal: " + errorMsg[2:] // Remove "❌ " prefix and add Self-QRIS prefix
		if _, err := h.messaging.SendTextToGroup(ctx, errorMsg); err != nil {
			h.logger.Printf("⚠️ Gagal kirim error ke grup: %v", err)
		}
		return
	}

	result, err := h.qrisUC.GenerateQRIS(ctx, cmd, msg)
	if err != nil {
		h.logger.Printf("❌ Gagal generate QRIS: %v", err)
		return
	}

	// Send QRIS image without caption (clean)
	qrisMsgID, err := h.messaging.SendImageTo(ctx, msg.ChatID, result.ImageData, "")
	if err != nil {
		h.logger.Printf("❌ Gagal kirim QRIS ke customer: %v", err)
		return
	}

	// Send caption as reply to the QRIS image
	caption := template.BuildQRISCaption(result.Amount, result.Deskripsi)
	if err := h.messaging.SendTextReply(ctx, msg.ChatID, caption, qrisMsgID, msg.SenderID); err != nil {
		h.logger.Printf("⚠️ Gagal kirim caption: %v (QRIS tetap terkirim)", err)
		// Continue - QRIS already sent successfully
	}

	notif := template.BuildSelfQrisNotification(cmd, msg.RecipientPhone)
	groupNotifMsgID, err := h.messaging.SendTextToGroup(ctx, notif)
	if err != nil {
		h.logger.Printf("⚠️ Gagal kirim notifikasi ke grup: %v", err)
	}

	// Auto-fill Akun with customer phone if not provided
	akun := cmd.Akun
	if akun == "" {
		akun = msg.RecipientPhone
	}

	pending := &entity.PendingPayment{
		MessageID:         qrisMsgID,
		OriginalMessageID: msg.ID,
		ChatID:            msg.ChatID,
		SenderJID:         msg.SenderID,
		SenderPhone:       msg.RecipientPhone,
		Amount:            cmd.Amount,
		CreatedAt:         time.Now(),
		IsSelfQris:        true,
		GroupNotifMsgID:   groupNotifMsgID,
		Produk:            cmd.Produk,
		Nama:              cmd.Nama,
		Email:             cmd.Email,
		Family:            cmd.Family,
		Deskripsi:         cmd.Deskripsi,
		Kanal:             cmd.Kanal,
		Akun:              akun,
	}

	h.paymentUC.RegisterPending(pending)
	h.logger.Printf("✅ Self-QRIS terkirim ke %s, notif ke grup (ID: %s)", formatter.FormatPhone(msg.RecipientPhone), groupNotifMsgID)
}

// validateFamilyOrWorkspace validates family or workspace field.
// Returns validation result and error message if validation fails.
// For Family: validates against Akun Google (Gemini family emails)
// For Workspace: validates against Akun ChatGPT (ChatGPT workspace names)
func (h *Handler) validateFamilyOrWorkspace(ctx context.Context, fieldName, fieldValue string) (interface{}, string, error) {
	if fieldValue == "" {
		return nil, "", nil
	}

	switch fieldName {
	case "Family":
		// Family validation for Gemini products
		if h.familyUC == nil {
			return nil, "", nil
		}
		validation, err := h.familyUC.ValidateFamily(ctx, fieldValue)
		if err != nil || !validation.IsValid {
			errorMsg := fmt.Sprintf("❌ Validasi %s gagal", fieldName)
			if validation != nil && validation.ErrorMessage != "" {
				errorMsg = "❌ " + validation.ErrorMessage
			}
			h.logger.Printf("Validasi %s gagal: %v", fieldName, err)
			return validation, errorMsg, err
		}
		h.logger.Printf("Validasi %s berhasil: %s (%d/%d slots)", fieldName, fieldValue, validation.UsedSlots, validation.MaxSlots)
		return validation, "", nil

	case "Workspace":
		// Workspace validation for ChatGPT products
		if h.workspaceUC == nil {
			return nil, "", nil
		}
		validation, err := h.workspaceUC.ValidateWorkspace(ctx, fieldValue)
		if err != nil || !validation.IsValid {
			errorMsg := fmt.Sprintf("❌ Validasi %s gagal", fieldName)
			if validation != nil && validation.ErrorMessage != "" {
				errorMsg = "❌ " + validation.ErrorMessage
			}
			h.logger.Printf("Validasi %s gagal: %v", fieldName, err)
			return validation, errorMsg, err
		}
		h.logger.Printf("Validasi %s berhasil: %s (%d/%d slots)", fieldName, fieldValue, validation.UsedSlots, validation.MaxSlots)
		return validation, "", nil

	default:
		return nil, "", nil
	}
}

// sendQrisHelp sends help/template based on product type.
func (h *Handler) sendQrisHelp(ctx context.Context, msg *entity.Message, productType string) {
	switch productType {
	case string(parser.ProductParamGoogle):
		_ = h.messaging.SendTextReply(ctx, msg.ChatID, template.QrisGeminiFormTemplate, msg.ID, msg.SenderID)
		_ = h.messaging.SendTextTo(ctx, msg.ChatID, template.QrisGeminiFormHelp)
	case string(parser.ProductParamChatGPT):
		_ = h.messaging.SendTextReply(ctx, msg.ChatID, template.QrisChatGPTFormTemplate, msg.ID, msg.SenderID)
		_ = h.messaging.SendTextTo(ctx, msg.ChatID, template.QrisChatGPTFormHelp)
	default:
		_ = h.messaging.SendTextReply(ctx, msg.ChatID, template.QrisGeneralHelp, msg.ID, msg.SenderID)
	}
}

// sendErrorReply sends an error message as a reply.
func (h *Handler) sendErrorReply(ctx context.Context, msg *entity.Message, text string) {
	if err := h.messaging.SendTextReply(ctx, msg.ChatID, text, msg.ID, msg.SenderID); err != nil {
		h.logger.Printf("Gagal kirim error reply: %v", err)
	}
}
