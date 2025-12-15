// Package entity defines core business entities used across all layers.
package entity

// SpecialFamilies lists special family accounts that don't require validation against Google Accounts.
// These are internal family accounts that are pre-configured.
var SpecialFamilies = map[string]bool{
	"Rumah Premium": true,
}

// IsSpecialFamily mengecek apakah family termasuk family khusus.
func IsSpecialFamily(family string) bool {
	return SpecialFamilies[family]
}

// FamilyValidation hasil validasi family (Gemini).
type FamilyValidation struct {
	IsValid      bool   // Apakah family valid
	IsSpecial    bool   // Apakah family khusus (skip validasi Akun Google)
	Email        string // Email family (jika bukan special)
	UsedSlots    int    // Jumlah slot terpakai
	MaxSlots     int    // Maksimal slot (5 for Gemini)
	ErrorMessage string // Error message if validation fails
}

// WorkspaceValidation hasil validasi workspace (ChatGPT).
type WorkspaceValidation struct {
	IsValid      bool   // Apakah workspace valid
	OwnerEmail   string // Email pemilik workspace (input dari user)
	UsedSlots    int    // Jumlah slot terpakai
	MaxSlots     int    // Maksimal slot (4 for ChatGPT)
	ErrorMessage string // Error message if validation fails
}

// constants.MaxFamilySlots adalah batas maksimal anggota per family (Gemini).
// constants.MaxWorkspaceSlots adalah batas maksimal anggota per workspace (ChatGPT).
