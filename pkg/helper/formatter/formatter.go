// Package formatter provides formatting utilities for various data types.
package formatter

import (
	"fmt"
	"regexp"
)

var digitOnlyRegex = regexp.MustCompile(`\D`)

// Phone formatting functions.

// NormalizePhone converts phone number to international format (628xxx).
// For Indonesian numbers: converts to 62xxx format
// For international numbers: returns digits-only format for consistent matching
// Examples:
//   - "08123456789" → "628123456789"
//   - "8123456789" → "628123456789"
//   - "628123456789" → "628123456789"
//   - "+628123456789" → "628123456789"
//   - "196177278545944" (international) → "196177278545944" (digits only)
func NormalizePhone(phone string) string {
	// Remove all non-digits
	digits := digitOnlyRegex.ReplaceAllString(phone, "")

	if len(digits) == 0 {
		return phone // Return original if no digits
	}

	// Check if this looks like a valid Indonesian phone number
	// Indonesian phone numbers start with 08xx, 8xx, or 628xx
	// and have 10-13 digits total (including country code)
	isIndonesianFormat := (len(digits) >= 9 && len(digits) <= 13) &&
		(digits[0] == '0' || digits[0] == '8' || (len(digits) >= 3 && digits[:2] == "62"))

	if !isIndonesianFormat {
		// International number or other format: return digits for consistent matching
		return digits
	}

	// Handle Indonesian format (0xxx)
	if digits[0] == '0' {
		return "62" + digits[1:]
	}

	// Handle without country code (8xxx)
	if digits[0] == '8' {
		return "62" + digits
	}

	// Already in international format (62xxx)
	return digits
}

// FormatPhone formats phone number to user-friendly format.
// Converts 628xxx to 08xxx for Indonesian numbers.
// Non-Indonesian numbers or LID format are returned unchanged.
// Examples:
//   - "628123456789" → "08123456789"
//   - "8123456789" → "08123456789"
//   - "08123456789" → "08123456789"
//   - "0196177278545944" → "0196177278545944" (LID, unchanged)
func FormatPhone(phone string) string {
	// Remove non-digits for validation
	digits := digitOnlyRegex.ReplaceAllString(phone, "")

	// Check if this looks like a valid Indonesian phone number
	isIndonesianFormat := (len(digits) >= 9 && len(digits) <= 13) &&
		(digits[0] == '0' || digits[0] == '8' || (len(digits) >= 3 && digits[:2] == "62"))

	if !isIndonesianFormat {
		// Not a valid Indonesian phone number (probably LID or other format)
		return phone // Return original unchanged
	}

	// Convert 628xxx → 08xxx
	if len(phone) > 2 && phone[:2] == "62" {
		return "0" + phone[2:]
	}

	// Convert 8xxx → 08xxx
	if len(phone) > 0 && phone[0] != '0' {
		return "0" + phone
	}

	return phone
}

// Currency formatting functions.

// FormatRupiah formats integer amount to Rupiah currency string.
// Examples:
//   - 10000 → "Rp10.000"
//   - 1000000 → "Rp1.000.000"
//   - 0 → "Rp0"
func FormatRupiah(amount int) string {
	str := fmt.Sprintf("%d", amount)
	n := len(str)
	if n <= 3 {
		return "Rp" + str
	}

	result := "Rp"
	start := n % 3
	if start > 0 {
		result += str[:start]
		if start < n {
			result += "."
		}
	}
	for i := start; i < n; i += 3 {
		result += str[i : i+3]
		if i+3 < n {
			result += "."
		}
	}
	return result
}

// Text formatting functions.

// TruncateWithEllipsis truncates text to maxLen characters and adds ellipsis if needed.
// Uses rune counting to properly handle multi-byte characters (e.g., emoji, Indonesian).
func TruncateWithEllipsis(text string, maxLen int) string {
	runes := []rune(text)
	if len(runes) <= maxLen {
		return text
	}
	return string(runes[:maxLen-3]) + "..."
}
