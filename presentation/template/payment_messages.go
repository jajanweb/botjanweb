// Package template provides all message templates for BotJanWeb.
// This file contains payment-related message templates.
package template

import (
	"fmt"
	"strings"
	"time"

	"github.com/exernia/botjanweb/internal/domain/entity"
	"github.com/exernia/botjanweb/pkg/constants"
	"github.com/exernia/botjanweb/pkg/helper/formatter"
)

// ============================================================================
// PAYMENT NOTIFICATION TEMPLATES
// ============================================================================

// BuildPaymentConfirmation builds payment confirmation message for customer.
func BuildPaymentConfirmation(pending *entity.PendingPayment, notif *entity.DANANotification) string {
	wib := time.FixedZone("WIB", constants.WIBOffset)
	var b strings.Builder

	b.WriteString("✅ *PEMBAYARAN BERHASIL*\n\n")
	b.WriteString("━━━━━━━━━━━━━━━━━━━━\n")
	b.WriteString("📋 *Detail Transaksi:*\n")
	b.WriteString(fmt.Sprintf("• Produk: %s\n", pending.Produk))
	b.WriteString(fmt.Sprintf("• Nama: %s\n", pending.Nama))
	b.WriteString(fmt.Sprintf("• Email: %s\n", pending.Email))
	if pending.Family != "" {
		b.WriteString(fmt.Sprintf("• Family: %s\n", pending.Family))
	}
	b.WriteString(fmt.Sprintf("• Nominal: %s\n", formatter.FormatRupiah(notif.Amount)))
	b.WriteString(fmt.Sprintf("• Waktu: %s\n", notif.Timestamp.In(wib).Format(constants.DateTimeWIBFormat)))
	b.WriteString("\n🙏 Terima kasih!")

	return b.String()
}

// BuildSelfQrisPaymentNotification builds payment notification for group (self-QRIS).
func BuildSelfQrisPaymentNotification(pending *entity.PendingPayment, amount int) string {
	var b strings.Builder

	b.WriteString("💰 *PEMBAYARAN DITERIMA*\n\n")
	b.WriteString("━━━━━━━━━━━━━━━━━━━━\n")
	b.WriteString(fmt.Sprintf("*%s* telah membayar %s\n", pending.Nama, formatter.FormatRupiah(amount)))
	b.WriteString(fmt.Sprintf("📧 Email: %s\n", pending.Email))
	if pending.Family != "" {
		b.WriteString(fmt.Sprintf("👨‍👩‍👧‍👦 Family: %s\n", pending.Family))
	}
	b.WriteString(fmt.Sprintf("📱 WA: %s", formatter.FormatPhone(pending.SenderPhone)))

	return b.String()
}

// BuildOrderSavedNotification builds message after order is saved to sheet.
func BuildOrderSavedNotification(pending *entity.PendingPayment) string {
	var b strings.Builder

	b.WriteString("\n📊 *Data pesanan telah dicatat:*\n")
	b.WriteString(fmt.Sprintf("• Sheet: %s\n", pending.Produk))
	b.WriteString(fmt.Sprintf("• Nama: %s\n", pending.Nama))
	b.WriteString(fmt.Sprintf("• Email: %s\n", pending.Email))

	return b.String()
}

// BuildSheetErrorNotification builds error notification for group when sheet logging fails.
func BuildSheetErrorNotification(pending *entity.PendingPayment, errorMsg string) string {
	var b strings.Builder

	b.WriteString("\n⚠️ *GAGAL CATAT KE SPREADSHEET*\n\n")
	b.WriteString("━━━━━━━━━━━━━━━━━━━━\n")
	b.WriteString("📋 *Data Transaksi:*\n")
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
	b.WriteString("\n❌ *Error:*\n")
	// Format error message to be user-friendly
	friendlyError := formatter.FormatUserFriendlyError(errorMsg)
	b.WriteString(fmt.Sprintf("%s\n", friendlyError))
	b.WriteString("\n⚠️ *Tindakan:* Catat manual ke spreadsheet\n")

	return b.String()
}
