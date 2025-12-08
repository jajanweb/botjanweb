// Package bot provides WhatsApp bot message parsing and handling.
package bot

import (
	"context"

	"github.com/exernia/botjanweb/internal/domain/entity"
	"github.com/exernia/botjanweb/pkg/helper/parser"
	"github.com/exernia/botjanweb/presentation/template"
)

// handleAddAkunCommand processes #addakun commands.
func (h *Handler) handleAddAkunCommand(ctx context.Context, msg *entity.Message, text string) {
	cmd, err := parser.ParseAddAkunCommand(text)
	if err != nil {
		h.sendErrorReply(ctx, msg, "❌ "+err.Error())
		return
	}

	if cmd.IsHelpMode {
		h.handleAddAkunHelp(ctx, msg, cmd.AccountType)
		return
	}

	h.handleAddAkunForm(ctx, msg, cmd)
}

// handleAddAkunHelp sends the form template based on account type.
func (h *Handler) handleAddAkunHelp(ctx context.Context, msg *entity.Message, accountType string) {
	h.logger.Printf("📋 Sending AddAkun form (type=%s) to %s", accountType, msg.SenderPhone)

	var formTemplate, help string

	switch accountType {
	case string(parser.ProductParamGoogle):
		formTemplate = template.AddAkunGoogleFormTemplate
		help = template.AddAkunGoogleFormHelp
	case string(parser.ProductParamChatGPT):
		formTemplate = template.AddAkunChatGPTFormTemplate
		help = template.AddAkunChatGPTFormHelp
	default:
		if err := h.messaging.SendTextReply(ctx, msg.ChatID, template.AddAkunGeneralHelp, msg.ID, msg.SenderID); err != nil {
			h.logger.Printf("Gagal kirim general help: %v", err)
		}
		return
	}

	if err := h.messaging.SendTextReply(ctx, msg.ChatID, formTemplate, msg.ID, msg.SenderID); err != nil {
		h.logger.Printf("Gagal kirim form template: %v", err)
		return
	}

	if err := h.messaging.SendTextTo(ctx, msg.ChatID, help); err != nil {
		h.logger.Printf("Gagal kirim help: %v", err)
	}
}

// handleAddAkunForm processes form-based #addakun commands.
func (h *Handler) handleAddAkunForm(ctx context.Context, msg *entity.Message, cmd *entity.AddAkunCommand) {
	h.logger.Printf("📝 AddAkun: %s | %s | Workspace: %s", cmd.Tipe, cmd.Email, cmd.Workspace)

	// Check if account use case is available
	if h.accountUC == nil {
		h.sendErrorReply(ctx, msg, "❌ Fitur tambah akun tidak tersedia. Google Sheets belum dikonfigurasi.")
		return
	}

	if err := h.accountUC.AddAccount(ctx, cmd); err != nil {
		h.logger.Printf("Gagal tambah akun: %v", err)
		h.sendErrorReply(ctx, msg, "❌ "+err.Error())
		return
	}

	successMsg := template.BuildAddAkunSuccess(cmd)
	if err := h.messaging.SendTextReply(ctx, msg.ChatID, successMsg, msg.ID, msg.SenderID); err != nil {
		h.logger.Printf("Gagal kirim konfirmasi: %v", err)
		return
	}

	h.logger.Printf("Akun %s berhasil ditambahkan: %s", cmd.Tipe, cmd.Email)
}

// handleListAkunCommand processes #listakun commands.
func (h *Handler) handleListAkunCommand(ctx context.Context, msg *entity.Message, text string) {
	cmd, err := parser.ParseListAkunCommand(text)
	if err != nil {
		h.sendErrorReply(ctx, msg, "❌ "+err.Error())
		return
	}

	if cmd.IsHelpMode {
		h.sendListAkunHelp(ctx, msg)
		return
	}

	h.sendListAkunResult(ctx, msg, cmd)
}

// sendListAkunHelp sends help message for #listakun.
func (h *Handler) sendListAkunHelp(ctx context.Context, msg *entity.Message) {
	help := template.ListAkunHelp
	if err := h.messaging.SendTextReply(ctx, msg.ChatID, help, msg.ID, msg.SenderID); err != nil {
		h.logger.Printf("Gagal kirim listakun help: %v", err)
	}
}

// sendListAkunResult fetches and sends account list.
func (h *Handler) sendListAkunResult(ctx context.Context, msg *entity.Message, cmd *entity.ListAkunCommand) {
	h.logger.Printf("📋 ListAkun: filter=%s", cmd.Tipe)

	// Check if account use case is available
	if h.accountUC == nil {
		h.sendErrorReply(ctx, msg, "❌ Fitur list akun tidak tersedia. Google Sheets belum dikonfigurasi.")
		return
	}

	result, err := h.accountUC.ListAccounts(ctx, cmd)
	if err != nil {
		h.logger.Printf("Gagal list akun: %v", err)
		h.sendErrorReply(ctx, msg, "❌ "+err.Error())
		return
	}

	response := template.BuildListAkunResult(result, cmd.Tipe)
	if err := h.messaging.SendTextReply(ctx, msg.ChatID, response, msg.ID, msg.SenderID); err != nil {
		h.logger.Printf("Gagal kirim hasil list: %v", err)
	}
}
