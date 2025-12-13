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

📌 *Contoh:*
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
• *Sandi* - Password akun (wajib)

📌 *Contoh:*
#addakun google
───────────────────
Email: john@example.com
Sandi: password123
───────────────────`

// AddAkunChatGPTFormTemplate is the template for adding ChatGPT accounts.
const AddAkunChatGPTFormTemplate = `#addakun chatgpt
───────────────────
Email: 
Sandi: 
Workspace: 
───────────────────`

// AddAkunChatGPTFormHelp is the help for adding ChatGPT accounts.
const AddAkunChatGPTFormHelp = `📋 *PANDUAN TAMBAH AKUN CHATGPT*

━━━━━━━━━━━━━━━━━━━━
📝 *Keterangan:*
• *Email* - Alamat email (wajib)
• *Sandi* - Password akun (wajib)
• *Workspace* - Nama workspace (wajib)

📌 *Contoh:*
#addakun chatgpt
───────────────────
Email: john@example.com
Sandi: password123
Workspace: TeamAlpha
───────────────────`

// ============================================================================
// LIST AKUN TEMPLATES
// ============================================================================

// ListAkunHelp is the help message for #listakun command.
const ListAkunHelp = `📋 *PANDUAN LIST AKUN*

━━━━━━━━━━━━━━━━━━━━
Perintah ini menampilkan daftar akun yang terdaftar.

📌 *Contoh:*
#listakun`

// ============================================================================
// ADD AKUN RESPONSE TEMPLATES
// ============================================================================

// BuildAddAkunSuccess builds success message for adding account.
func BuildAddAkunSuccess(cmd *entity.AddAkunCommand) string {
	// For Google: show Family, for ChatGPT: show Workspace
	var detailField string
	var detailValue string

	if cmd.Tipe == entity.AccountTypeGoogle {
		detailField = "Family"
		detailValue = "-" // Google accounts don't track family in AddAkun
	} else {
		detailField = "Workspace"
		detailValue = cmd.Workspace
		if detailValue == "" {
			detailValue = "-"
		}
	}

	return fmt.Sprintf(`✅ *AKUN BERHASIL DITAMBAHKAN*

━━━━━━━━━━━━━━━━━━━━
📋 *Detail:*
• Tipe: %s
• Email: %s
• %s: %s

📊 Data telah tersimpan di spreadsheet`, cmd.Tipe, cmd.Email, detailField, detailValue)
} // BuildListAkunResult builds the account list message.
func BuildListAkunResult(result *entity.AccountListResult, filter entity.AccountType) string {
	var msg string

	msg += "📋 *DAFTAR AKUN*\n\n"
	msg += "━━━━━━━━━━━━━━━━━━━━\n"

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
