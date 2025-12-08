// Package template provides all message templates for BotJanWeb.
// This file contains account management message templates.
package template

import (
	"fmt"

	"github.com/exernia/botjanweb/internal/domain/entity"
)

// ============================================================================
// ADD AKUN TEMPLATES - Help & Forms
// ============================================================================

// AddAkunGeneralHelp is sent when #addakun is called without type parameter.
const AddAkunGeneralHelp = `📋 *PANDUAN TAMBAH AKUN*
━━━━━━━━━━━━━━━━━━━━

Gunakan command sesuai tipe:

• *#addakun google* → Tambah Akun Google
• *#addakun chatgpt* → Tambah Akun ChatGPT

Contoh:
#addakun google
#addakun chatgpt`

// AddAkunGoogleFormTemplate is the template for adding Google accounts.
const AddAkunGoogleFormTemplate = `#addakun google
───────────────────
Email: 
Sandi: 
───────────────────`

// AddAkunGoogleFormHelp is the help for adding Google accounts.
const AddAkunGoogleFormHelp = `📋 *PANDUAN TAMBAH AKUN GOOGLE*
━━━━━━━━━━━━━━━━━━━━

📝 *Keterangan:*
• *Email* - Alamat Gmail (wajib)
• *Family* - Nama family plan (wajib)

📌 *Contoh:*
#addakun google
───────────────────
Email: john@example.com
Family: Rumah Premium
───────────────────`

// AddAkunChatGPTFormTemplate is the template for adding ChatGPT accounts.
const AddAkunChatGPTFormTemplate = `#addakun chatgpt
───────────────────
Email:
Sandi:
Workspace: 
Paket: 
───────────────────`

// AddAkunChatGPTFormHelp is the help for adding ChatGPT accounts.
const AddAkunChatGPTFormHelp = `📋 *PANDUAN TAMBAH AKUN CHATGPT*
━━━━━━━━━━━━━━━━━━━━

📝 *Keterangan:*
• *Email* - Alamat email (wajib)
• *Workspace* - Nama workspace (wajib)
• *Paket* - Paket langganan (wajib)

📌 *Contoh:*
#addakun chatgpt
───────────────────
Email: john@example.com
Workspace: TeamAlpha
Paket: Pro
───────────────────`

// ============================================================================
// LIST AKUN TEMPLATES
// ============================================================================

// ListAkunHelp is the help message for #listakun command.
const ListAkunHelp = `📋 *PANDUAN LIST AKUN*
━━━━━━━━━━━━━━━━━━━━

Perintah ini menampilkan daftar akun yang terdaftar.

Contoh:
#listakun`

// ============================================================================
// ADD AKUN RESPONSE TEMPLATES
// ============================================================================

// BuildAddAkunSuccess builds success message for adding account.
func BuildAddAkunSuccess(cmd *entity.AddAkunCommand) string {
	familyOrWorkspace := cmd.Workspace
	if familyOrWorkspace == "" {
		familyOrWorkspace = "-"
	}

	return fmt.Sprintf(`✅ *AKUN BERHASIL DITAMBAHKAN*

━━━━━━━━━━━━━━━━━━━━
• Tipe: %s
• Email: %s
• Workspace: %s

📊 Data telah tersimpan di spreadsheet`, cmd.Tipe, cmd.Email, familyOrWorkspace)
} // BuildListAkunResult builds the account list message.
func BuildListAkunResult(result *entity.AccountListResult, filter entity.AccountType) string {
	var msg string

	msg += "━━━━━━━━━━━━━━━━━━━━\n"
	msg += "📋 *DAFTAR AKUN*\n"
	msg += "━━━━━━━━━━━━━━━━━━━━\n\n"

	// Show Google accounts
	if filter == "" || filter == entity.AccountTypeGoogle {
		msg += "🔵 *AKUN GOOGLE*\n"
		msg += fmt.Sprintf("Total: %d | Tersedia: %d | Tidak Tersedia: %d\n",
			result.TotalGoogle, result.AvailableGoogle, result.TotalGoogle-result.AvailableGoogle)
		msg += "───────────────────\n"

		if len(result.GoogleAccounts) == 0 {
			msg += "_Belum ada akun_\n"
		} else {
			for _, acc := range result.GoogleAccounts {
				status := "✅"
				if !acc.IsAvailable() {
					status = "❌"
				}
				msg += fmt.Sprintf("%s %s\n", status, acc.Email)
				if acc.Keterangan != "" {
					msg += fmt.Sprintf("   └ %s\n", acc.Keterangan)
				}
			}
		}
		msg += "\n"
	}

	// Show ChatGPT accounts
	if filter == "" || filter == entity.AccountTypeChatGPT {
		msg += "🟢 *AKUN CHATGPT*\n"
		msg += fmt.Sprintf("Total: %d | Tersedia: %d | Tidak Tersedia: %d\n",
			result.TotalChatGPT, result.AvailableChatGPT, result.TotalChatGPT-result.AvailableChatGPT)
		msg += "───────────────────\n"

		if len(result.ChatGPTAccounts) == 0 {
			msg += "_Belum ada akun_\n"
		} else {
			for _, acc := range result.ChatGPTAccounts {
				status := "✅"
				if !acc.IsAvailable() {
					status = "❌"
				}
				msg += fmt.Sprintf("%s %s (%s)\n", status, acc.Email, acc.Workspace)
				if acc.Status != "" && !acc.IsAvailable() {
					msg += fmt.Sprintf("   └ %s\n", acc.Status)
				}
			}
		}
	}

	msg += "\n━━━━━━━━━━━━━━━━━━━━\n"
	msg += "✅ = Tersedia | ❌ = Tidak Tersedia"

	return msg
}
