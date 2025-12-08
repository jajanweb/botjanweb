// Package parser provides command and primitive parsing utilities.
package parser

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/exernia/botjanweb/pkg/helper/formatter"
)

// ParseRupiah parses Rupiah string to integer amount.
// Handles formats: "Rp 10.000", "10000", "10.000", etc.
// Examples:
//   - "Rp 10.000" → 10000
//   - "10.000" → 10000
//   - "10000" → 10000
func ParseRupiah(str string) (int, error) {
	// Remove "Rp" prefix and whitespace
	str = strings.TrimSpace(str)
	str = strings.TrimPrefix(str, "Rp")
	str = strings.TrimSpace(str)

	// Remove thousand separators
	str = strings.ReplaceAll(str, ".", "")
	str = strings.ReplaceAll(str, ",", "")

	// Parse to integer
	amount, err := strconv.Atoi(str)
	if err != nil {
		return 0, fmt.Errorf("nominal tidak valid: %s", str)
	}

	if amount <= 0 {
		return 0, fmt.Errorf("nominal harus lebih dari 0")
	}

	return amount, nil
}

// ParsePhoneList parses and normalizes comma-separated phone numbers.
// Returns a slice of normalized phone numbers (empty items filtered out).
func ParsePhoneList(phoneListStr string) []string {
	if phoneListStr == "" {
		return nil
	}

	var phones []string
	for _, phone := range strings.Split(phoneListStr, ",") {
		phone = strings.TrimSpace(phone)
		if phone != "" {
			// Normalize at parse time for efficiency
			phones = append(phones, formatter.NormalizePhone(phone))
		}
	}
	return phones
}
