// Package template provides all message templates for BotJanWeb.
// This file contains QRIS-related message templates.
package template

import (
	"fmt"
	"strings"

	"github.com/exernia/botjanweb/internal/domain/entity"
	"github.com/exernia/botjanweb/pkg/helper/formatter"
)

// ============================================================================
// QRIS TEMPLATES - Help & Forms
// ============================================================================

// QrisGeneralHelp is sent when #qris is called without product parameter.
const QrisGeneralHelp = `📋 *PANDUAN QRIS*
━━━━━━━━━━━━━━━━━━━━

Gunakan command sesuai produk:

• *#qris google* → Order Gemini (GDrive 2TB + AI Pro)
• *#qris chatgpt* → Order ChatGPT Pro

Contoh:
#qris google
#qris chatgpt`

// QrisGeminiFormTemplate is the template for Gemini/Google orders.
const QrisGeminiFormTemplate = `#qris google
───────────────────
Nama: 
Email: 
Family: 
Nominal: 
Kanal: 
Akun: 
───────────────────`

// QrisGeminiFormHelp is the help for Gemini orders.
const QrisGeminiFormHelp = `📋 *PANDUAN ORDER GEMINI*
━━━━━━━━━━━━━━━━━━━━

📦 *Produk:* Gemini (GDrive 2TB + Gemini AI Pro)

📝 *Keterangan:*
• *Nama* - Nama lengkap (wajib)
• *Email* - Alamat Gmail (wajib)
• *Family* - Nama family plan (wajib)
• *Nominal* - Jumlah pembayaran (wajib)
• *Kanal* - Channel pembelian (default: Threads)
• *Akun* - Username/akun (opsional)

📌 *Contoh:*
#qris google
───────────────────
Nama: John Doe
Email: john@example.com
Family: Rumah Premium
Nominal: 49901
Kanal: Threads
Akun: @johndoe
───────────────────`

// QrisChatGPTFormTemplate is the template for ChatGPT orders.
const QrisChatGPTFormTemplate = `#qris chatgpt
───────────────────
Nama: 
Email: 
Workspace: 
Paket: 
Nominal: 
Kanal: 
───────────────────`

// QrisChatGPTFormHelp is the help for ChatGPT orders.
const QrisChatGPTFormHelp = `📋 *PANDUAN ORDER CHATGPT*
━━━━━━━━━━━━━━━━━━━━

📦 *Produk:* ChatGPT Pro

📝 *Keterangan:*
• *Nama* - Nama lengkap (wajib)
• *Email* - Alamat email (wajib)
• *Workspace* - Nama workspace (wajib)
• *Paket* - Paket langganan (wajib)
• *Nominal* - Jumlah pembayaran (wajib)
• *Kanal* - Channel pembelian (default: Threads)

📌 *Contoh:*
#qris chatgpt
───────────────────
Nama: John Doe
Email: john@example.com
Workspace: TeamAlpha
Paket: Pro
Nominal: 75000
Kanal: Threads
───────────────────`

// ============================================================================
// QRIS IMAGE CAPTION TEMPLATES
// ============================================================================

// BuildQRISCaption builds caption for QRIS image (full pending payment info).
func BuildQRISCaption(amount int, deskripsi string) string {
	var b strings.Builder

	b.WriteString("💳 *QRIS PEMBAYARAN*\n\n")
	b.WriteString(fmt.Sprintf("💰 Nominal: %s\n", formatter.FormatRupiah(amount)))
	if deskripsi != "" {
		b.WriteString(fmt.Sprintf("📋 %s\n", deskripsi))
	}
	b.WriteString("\n📱 Scan QRIS di atas untuk bayar")

	return b.String()
}

// BuildQRISCaptionFromPending builds caption for QRIS image from pending payment.
func BuildQRISCaptionFromPending(pending *entity.PendingPayment) string {
	var b strings.Builder

	b.WriteString("💳 *QRIS PEMBAYARAN*\n\n")
	b.WriteString("━━━━━━━━━━━━━━━━━━━━\n")
	b.WriteString("📋 *Detail Pesanan:*\n")
	b.WriteString(fmt.Sprintf("• Produk: %s\n", pending.Produk))
	b.WriteString(fmt.Sprintf("• Nama: %s\n", pending.Nama))
	b.WriteString(fmt.Sprintf("• Email: %s\n", pending.Email))
	if pending.Family != "" {
		b.WriteString(fmt.Sprintf("• Family: %s\n", pending.Family))
	}
	b.WriteString(fmt.Sprintf("• Nominal: %s\n", formatter.FormatRupiah(pending.Amount)))
	b.WriteString(fmt.Sprintf("• Kanal: %s\n", pending.Kanal))
	if pending.Akun != "" {
		b.WriteString(fmt.Sprintf("• Akun: %s\n", pending.Akun))
	}
	b.WriteString("\n📱 Scan QRIS di atas untuk bayar")

	return b.String()
}

// BuildQrisFormCaption builds simple caption for form template.
func BuildQrisFormCaption(cmd *entity.QrisCommand) string {
	return fmt.Sprintf("📝 *Form Order %s*\n\nIsi form di atas dan kirim ulang.", cmd.Produk)
}

// BuildLegacyQrisCaption builds legacy QRIS caption (simplified).
func BuildLegacyQrisCaption(nama, email, family, kanal string, amount int) string {
	var b strings.Builder

	b.WriteString("💳 *QRIS PEMBAYARAN*\n\n")
	b.WriteString(fmt.Sprintf("• Nama: %s\n", nama))
	b.WriteString(fmt.Sprintf("• Email: %s\n", email))
	if family != "" {
		b.WriteString(fmt.Sprintf("• Family: %s\n", family))
	}
	b.WriteString(fmt.Sprintf("• Nominal: %s\n", formatter.FormatRupiah(amount)))
	b.WriteString(fmt.Sprintf("• Kanal: %s\n", kanal))
	b.WriteString("\n📱 Scan QRIS untuk bayar")

	return b.String()
}

// BuildSelfQrisNotification builds initial notification for self-QRIS (before payment).
func BuildSelfQrisNotification(cmd *entity.QrisCommand, recipientPhone string) string {
	var b strings.Builder

	b.WriteString("🔔 *PESANAN BARU*\n\n")
	b.WriteString("━━━━━━━━━━━━━━━━━━━━\n")
	b.WriteString(fmt.Sprintf("👤 %s sedang order:\n", cmd.Nama))
	b.WriteString(fmt.Sprintf("📧 Email: %s\n", cmd.Email))
	if cmd.Family != "" {
		b.WriteString(fmt.Sprintf("👨‍👩‍👧‍👦 Family: %s\n", cmd.Family))
	}
	b.WriteString(fmt.Sprintf("💰 Nominal: %s\n", formatter.FormatRupiah(cmd.Amount)))
	b.WriteString(fmt.Sprintf("📱 WA: %s\n", formatter.FormatPhone(recipientPhone)))
	b.WriteString("\n⏳ Menunggu pembayaran...")

	return b.String()
}
