// Package constants contains all application-wide constants.
package constants

// Payment provider constants.
const (
	DANAPackage = "id.dana"
)

// Product defaults.
const (
	DefaultKanal = "Threads"
)

// Family plan constants.
const (
	MaxFamilySlots    = 5 // Gemini family (max 5 members)
	MaxWorkspaceSlots = 4 // ChatGPT workspace (max 4 members)
)

// Time constants.
const (
	TimezoneWIB = "Asia/Jakarta"
	WIBOffset   = 7 * 60 * 60 // 7 hours in seconds
)

// Format constants.
const (
	DateTimeWIBFormat = "02-01-2006 15:04:05 WIB"
)

// Log prefixes.
const (
	LogPrefixMain         = "[MAIN] "
	LogPrefixApp          = "[APP] "
	LogPrefixConfig       = "[CONFIG] "
	LogPrefixWebhook      = "[WEBHOOK] "
	LogPrefixWA           = "[WA] "
	LogPrefixPayment      = "[PAYMENT] "
	LogPrefixConfirmation = "[CONFIRMATION] "
	LogPrefixQRIS         = "[QRIS] "
	LogPrefixBot          = "[BOT] "
)
