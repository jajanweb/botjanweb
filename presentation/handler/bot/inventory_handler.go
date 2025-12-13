// Package bot provides WhatsApp bot message parsing and handling.
package bot

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/exernia/botjanweb/internal/domain/entity"
)

// handleCekSlotCommand handles the #cekslot command.
// Format: #cekslot <product>
// Example: #cekslot gemini, #cekslot chatgpt
func (h *Handler) handleCekSlotCommand(ctx context.Context, msg *entity.Message, text string) {
	cmd := h.parseCekSlotCommand(text)

	if cmd.IsHelpMode {
		h.sendCekSlotHelp(ctx, msg)
		return
	}

	// Get slot availability
	result, err := h.inventoryRepo.GetSlotAvailability(ctx, cmd.Product, cmd.AvailableOnly)
	if err != nil {
		h.sendErrorReply(ctx, msg, fmt.Sprintf("❌ Gagal mengecek slot: %v", err))
		return
	}

	// Format response
	h.sendSlotAvailabilityResult(ctx, msg, result)
}

// parseCekSlotCommand parses the #cekslot command.
func (h *Handler) parseCekSlotCommand(text string) *entity.CekSlotCommand {
	cmd := &entity.CekSlotCommand{
		AvailableOnly: true, // Default to showing only available
	}

	// Remove prefix
	text = strings.TrimPrefix(strings.ToLower(text), "#cekslot")
	text = strings.TrimSpace(text)

	if text == "" {
		cmd.IsHelpMode = true
		return cmd
	}

	// Parse product
	switch {
	case strings.Contains(text, "chatgpt") || strings.Contains(text, "gpt"):
		cmd.Product = "ChatGPT"
	case strings.Contains(text, "gemini") || strings.Contains(text, "google"):
		cmd.Product = "Gemini"
	case strings.Contains(text, "all") || strings.Contains(text, "semua"):
		// Show both - we'll handle this by not filtering
		cmd.Product = ""
	default:
		cmd.IsHelpMode = true
	}

	// Check for "all" flag to show all (including full)
	if strings.Contains(text, "all") || strings.Contains(text, "semua") {
		cmd.AvailableOnly = false
	}

	return cmd
}

// sendCekSlotHelp sends help message for #cekslot command.
func (h *Handler) sendCekSlotHelp(ctx context.Context, msg *entity.Message) {
	help := `📊 *Command #cekslot*

Cek ketersediaan slot Family/Workspace.

*Format:*
#cekslot <produk>

*Contoh:*
#cekslot gemini
#cekslot chatgpt
#cekslot gemini all (tampilkan semua termasuk yang penuh)

*Produk:*
• gemini - Cek slot Family Gemini (max 5)
• chatgpt - Cek slot Workspace ChatGPT (max 4)`

	h.sendErrorReply(ctx, msg, help)
}

// sendSlotAvailabilityResult formats and sends slot availability result.
func (h *Handler) sendSlotAvailabilityResult(ctx context.Context, msg *entity.Message, result *entity.SlotAvailabilityResult) {
	if result.TotalEntries == 0 {
		h.sendErrorReply(ctx, msg, fmt.Sprintf("📊 Tidak ada slot %s yang ditemukan di area reserved.", result.Product))
		return
	}

	// Sort by available slots (descending)
	sort.Slice(result.Slots, func(i, j int) bool {
		return result.Slots[i].AvailableSlot > result.Slots[j].AvailableSlot
	})

	// Build response
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("📊 *Ketersediaan Slot %s*\n\n", result.Product))

	for _, slot := range result.Slots {
		emoji := "✅"
		if slot.AvailableSlot == 0 {
			emoji = "❌"
		} else if slot.AvailableSlot == 1 {
			emoji = "⚠️"
		}

		sb.WriteString(fmt.Sprintf("%s %s: %d/%d slot tersedia\n",
			emoji, slot.Name, slot.AvailableSlot, slot.TotalSlots))
	}

	// Summary
	totalAvailable := 0
	for _, slot := range result.Slots {
		totalAvailable += slot.AvailableSlot
	}
	sb.WriteString(fmt.Sprintf("\n📈 *Total:* %d entries, %d slot tersedia", result.TotalEntries, totalAvailable))

	h.sendErrorReply(ctx, msg, sb.String())
}

// handleCekKodeCommand handles the #cekkode command.
// Format: #cekkode [all]
func (h *Handler) handleCekKodeCommand(ctx context.Context, msg *entity.Message, text string) {
	// Parse command
	text = strings.TrimPrefix(strings.ToLower(text), "#cekkode")
	text = strings.TrimSpace(text)

	availableOnly := true
	if strings.Contains(text, "all") || strings.Contains(text, "semua") {
		availableOnly = false
	}

	// Get redeem code availability
	result, err := h.inventoryRepo.GetRedeemCodeAvailability(ctx, availableOnly)
	if err != nil {
		h.sendErrorReply(ctx, msg, fmt.Sprintf("❌ Gagal mengecek kode: %v", err))
		return
	}

	// Format response
	h.sendRedeemCodeResult(ctx, msg, result)
}

// sendRedeemCodeResult formats and sends redeem code availability result.
func (h *Handler) sendRedeemCodeResult(ctx context.Context, msg *entity.Message, result *entity.RedeemCodeResult) {
	var sb strings.Builder
	sb.WriteString("🎫 *Ketersediaan Kode Redeem Perplexity*\n\n")

	if len(result.Codes) == 0 {
		if result.AvailableOnly {
			sb.WriteString("❌ Tidak ada kode yang tersedia.\n")
		} else {
			sb.WriteString("📭 Belum ada kode yang diinput.\n")
		}
	} else {
		for _, code := range result.Codes {
			emoji := "✅"
			status := "Tersedia"
			if code.TanggalAktivasi != "" {
				emoji = "❌"
				status = fmt.Sprintf("Dipakai %s", code.TanggalAktivasi)
			}

			sb.WriteString(fmt.Sprintf("%s #%d: %s\n   Email: %s\n   Status: %s\n\n",
				emoji, code.No, code.KodeRedeem, code.Email, status))
		}
	}

	// Summary
	sb.WriteString(fmt.Sprintf("📈 *Ringkasan:* %d total, %d tersedia",
		result.TotalCodes, result.AvailableCodes))

	h.sendErrorReply(ctx, msg, sb.String())
}

// handleInputKodeCommand handles the #inputkode command.
// Format: #inputkode <email> <kode>
func (h *Handler) handleInputKodeCommand(ctx context.Context, msg *entity.Message, text string) {
	cmd := h.parseInputKodeCommand(text)

	if cmd.IsHelpMode {
		h.sendInputKodeHelp(ctx, msg)
		return
	}

	// Add redeem code
	err := h.inventoryRepo.AddRedeemCode(ctx, cmd.Email, cmd.KodeRedeem)
	if err != nil {
		h.sendErrorReply(ctx, msg, fmt.Sprintf("❌ Gagal menambah kode: %v", err))
		return
	}

	// Success response
	response := fmt.Sprintf("✅ *Kode Redeem Berhasil Ditambahkan*\n\n"+
		"📧 Email: %s\n"+
		"🎫 Kode: %s\n\n"+
		"_Kode siap digunakan untuk customer._",
		cmd.Email, cmd.KodeRedeem)

	h.sendErrorReply(ctx, msg, response)
}

// parseInputKodeCommand parses the #inputkode command.
func (h *Handler) parseInputKodeCommand(text string) *entity.InputKodeCommand {
	cmd := &entity.InputKodeCommand{}

	// Remove prefix
	text = strings.TrimPrefix(strings.ToLower(text), "#inputkode")
	text = strings.TrimSpace(text)

	if text == "" {
		cmd.IsHelpMode = true
		return cmd
	}

	// Split by whitespace
	parts := strings.Fields(text)
	if len(parts) < 2 {
		cmd.IsHelpMode = true
		return cmd
	}

	cmd.Email = parts[0]
	cmd.KodeRedeem = parts[1]

	// Validate email format (basic check)
	if !strings.Contains(cmd.Email, "@") {
		cmd.IsHelpMode = true
		return cmd
	}

	return cmd
}

// sendInputKodeHelp sends help message for #inputkode command.
func (h *Handler) sendInputKodeHelp(ctx context.Context, msg *entity.Message) {
	help := `🎫 *Command #inputkode*

Menambahkan kode redeem Perplexity baru.

*Format:*
#inputkode <email> <kode_redeem>

*Contoh:*
#inputkode user@gmail.com ABC123XYZ

*Keterangan:*
• Email: akun Perplexity pemilik kode
• Kode: kode redeem dari Perplexity`

	h.sendErrorReply(ctx, msg, help)
}
