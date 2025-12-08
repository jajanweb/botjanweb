// Package formatter provides formatting utilities for various data types.
package formatter

import (
	"regexp"
	"strings"
)

// Error message patterns and their user-friendly replacements.
var errorPatterns = []struct {
	pattern     *regexp.Regexp
	replacement string
}{
	{
		// Pattern: "family 'email' not found in reserved slot area (rows X-Y)"
		pattern:     regexp.MustCompile(`family '[^']+' not found in reserved slot area \(rows \d+-\d+\)`),
		replacement: "Family belum terdaftar di sistem. Hubungi admin untuk registrasi family.",
	},
	{
		// Pattern: "failed to find available slot for family 'email': ..."
		// Note: This must come AFTER the specific "family not found" pattern
		pattern:     regexp.MustCompile(`failed to find available slot for family '[^']+': .+`),
		replacement: "Slot untuk family ini tidak tersedia.",
	},
	{
		// Pattern: "no available slot found for family 'email'"
		pattern:     regexp.MustCompile(`no available slot found for family '[^']+'`),
		replacement: "Tidak ada slot kosong tersedia untuk family ini.",
	},
	{
		// Pattern: database/connection errors - must match whole line
		pattern:     regexp.MustCompile(`(?i)(database|connection|network|timeout) error.*`),
		replacement: "Terjadi gangguan koneksi. Silakan coba lagi dalam beberapa saat.",
	},
	{
		// Pattern: any "failed to..." technical messages (generic fallback)
		pattern:     regexp.MustCompile(`^failed to (.+): .+$`),
		replacement: "Gagal $1. Silakan coba lagi atau hubungi admin.",
	},
}

// FormatUserFriendlyError converts technical error messages to user-friendly Indonesian messages.
// This function is used to transform raw error messages before displaying them to users in WhatsApp groups.
//
// Examples:
//   - "family 'user@example.com' not found in reserved slot area (rows 108-151)"
//     → "Family belum terdaftar di sistem. Hubungi admin untuk registrasi family."
//   - "failed to find available slot for family 'user@example.com': some reason"
//     → "Slot untuk family ini tidak tersedia."
//   - "database error: connection timeout"
//     → "Terjadi gangguan koneksi. Silakan coba lagi dalam beberapa saat."
func FormatUserFriendlyError(errorMsg string) string {
	if errorMsg == "" {
		return "Terjadi kesalahan. Silakan hubungi admin."
	}

	// Clean up the error message
	msg := strings.TrimSpace(errorMsg)

	// Apply multiple passes to handle nested errors
	maxPasses := 3
	for pass := 0; pass < maxPasses; pass++ {
		originalMsg := msg

		// Try to match against known patterns
		for _, ep := range errorPatterns {
			if ep.pattern.MatchString(msg) {
				// Replace with user-friendly message
				msg = ep.pattern.ReplaceAllString(msg, ep.replacement)
				break // Exit inner loop after first match
			}
		}

		// If no change was made, we're done
		if msg == originalMsg {
			break
		}
	}

	// If the message is still too technical or unchanged, provide generic message
	if strings.Contains(msg, "failed to") || strings.Contains(msg, "error:") || len(msg) > 150 {
		return "Terjadi kesalahan dalam proses pencatatan. Silakan catat manual atau hubungi admin."
	}

	// If message looks technical/unparsed, use generic message
	if !strings.Contains(msg, "Family") && !strings.Contains(msg, "Slot") &&
		!strings.Contains(msg, "Tidak") && !strings.Contains(msg, "Gagal") &&
		!strings.Contains(msg, "gangguan") && !strings.Contains(msg, "kesalahan") {
		return "Terjadi kesalahan dalam proses pencatatan. Silakan catat manual atau hubungi admin."
	}

	return msg
}
