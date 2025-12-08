// Package validator provides input validation utilities.
package validator

import (
	"regexp"
	"strings"

	"github.com/exernia/botjanweb/pkg/helper/formatter"
)

var (
	emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)
)

// ValidateEmail checks if email format is valid.
func ValidateEmail(email string) bool {
	email = strings.TrimSpace(email)
	if email == "" {
		return false
	}
	return emailRegex.MatchString(email)
}

// ValidatePhone checks if phone number is valid Indonesian format.
// Must be 10-15 digits and start with 628 (Indonesian mobile) after normalization.
func ValidatePhone(phone string) bool {
	normalized := formatter.NormalizePhone(phone)

	// Check if normalized (empty return means no valid digits)
	if normalized == "" || normalized == phone && len(phone) > 13 {
		return false
	}

	// Must be 10-15 digits total
	if len(normalized) < 10 || len(normalized) > 15 {
		return false
	}

	// Must start with 628 (Indonesian mobile)
	if len(normalized) >= 3 && normalized[:3] != "628" {
		return false
	}

	return true
}

// ValidateNonEmpty checks if string is non-empty after trimming.
func ValidateNonEmpty(s string) bool {
	return strings.TrimSpace(s) != ""
}

// ValidateAmount checks if amount is positive.
func ValidateAmount(amount int) bool {
	return amount > 0
}
