// Package entity defines core business entities used across all layers.
package entity

import (
	"errors"
	"strings"
	"time"
)

// AccountType represents the type of account (Google or ChatGPT).
type AccountType string

const (
	AccountTypeGoogle  AccountType = "Google"
	AccountTypeChatGPT AccountType = "ChatGPT"
)

// ParseAccountType parses a string to AccountType.
func ParseAccountType(s string) (AccountType, bool) {
	switch strings.ToLower(s) {
	case "google":
		return AccountTypeGoogle, true
	case "chatgpt":
		return AccountTypeChatGPT, true
	default:
		return "", false
	}
}

// AddAkunCommand represents a parsed #addakun command.
type AddAkunCommand struct {
	Tipe      AccountType // Google or ChatGPT (resolved type)
	Email     string      // Email address
	Sandi     string      // Password
	Workspace string      // Workspace name (ChatGPT only)

	IsHelpMode  bool   // True if #addakun sent without parameters
	AccountType string // Raw parameter from command ("google", "chatgpt")
}

// ListAkunCommand represents a parsed #listakun command.
type ListAkunCommand struct {
	Tipe        AccountType // Google or ChatGPT (optional filter)
	AccountType string      // Raw parameter from command
	IsHelpMode  bool        // True if #listakun sent without parameters
}

// AkunGoogle represents a Google account entity for the Akun Google sheet.
type AkunGoogle struct {
	Email           string    // A: Email
	Sandi           string    // B: Sandi
	TanggalAktivasi time.Time // C: Tanggal Aktivasi
	TanggalBerakhir string    // D: Tanggal Berakhir
	StatusDibuat    string    // E: Creation status (Family name)
	YTPremium       string    // F: YT Premium?
	Keterangan      string    // G: Notes/remarks
}

// IsAvailable checks if Google account is still available.
func (a *AkunGoogle) IsAvailable() bool {
	// Check if notes field contains unavailable keywords
	ket := strings.ToLower(a.Keterangan)
	unavailableKeywords := []string{"kekunci", "terkunci", "locked", "banned", "suspend", "disabled"}
	for _, kw := range unavailableKeywords {
		if strings.Contains(ket, kw) {
			return false
		}
	}

	// Check if Tanggal Berakhir has passed (simple check)
	if a.TanggalBerakhir != "" {
		// Parse common date formats
		expiry, err := parseFlexibleDate(a.TanggalBerakhir)
		if err == nil && time.Now().After(expiry) {
			return false
		}
	}

	return true
}

// AkunChatGPT represents a ChatGPT account entity for the Akun ChatGPT sheet.
type AkunChatGPT struct {
	Email           string    // A: Email
	Sandi           string    // B: Password
	Workspace       string    // C: Workspace
	Status          string    // D: Status (Safe / Banned)
	TanggalAktivasi time.Time // E: Tanggal Aktivasi
	TanggalKenaBan  string    // F: Ban date
}

// IsAvailable checks if ChatGPT account is still available.
func (a *AkunChatGPT) IsAvailable() bool {
	status := strings.ToLower(a.Status)
	// Account is unavailable if status contains "ban" or TanggalKenaBan is set
	if strings.Contains(status, "ban") || a.TanggalKenaBan != "" {
		return false
	}
	return true
}

// parseFlexibleDate parses date in various formats.
func parseFlexibleDate(s string) (time.Time, error) {
	formats := []string{
		"2 January 2006",
		"02 January 2006",
		"2 Jan 2006",
		"02 Jan 2006",
		"02/01/2006",
		"2/1/2006",
		"2006-01-02",
	}
	for _, f := range formats {
		if t, err := time.Parse(f, s); err == nil {
			return t, nil
		}
	}
	return time.Time{}, errors.New("unable to parse date")
}

// AccountListResult contains the result of listing accounts.
type AccountListResult struct {
	GoogleAccounts   []AkunGoogle
	ChatGPTAccounts  []AkunChatGPT
	TotalGoogle      int
	TotalChatGPT     int
	AvailableGoogle  int
	AvailableChatGPT int
}
