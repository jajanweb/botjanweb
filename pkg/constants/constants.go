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

// QRIS rendering constants.
const (
	DefaultQRSize = 256
	QRMargin      = 10

	// Image dimensions
	QRWidth  = 1080
	QRHeight = 1920

	// Colors
	ColorWhite = "#FFFFFF"
	ColorBlack = "#000000"
)

// Family plan constants.
const (
	MaxFamilySlots = 5
)

// Time constants.
const (
	TimezoneWIB = "Asia/Jakarta"
	WIBOffset   = 7 * 60 * 60 // 7 hours in seconds
)

// Format constants.
const (
	DateFormat          = "02-01-2006"
	DateTimeFormat      = "02-01-2006 15:04:05"
	DateTimeWIBFormat   = "02-01-2006 15:04:05 WIB"
	ShortDateFormat     = "02/01"
	ShortDateTimeFormat = "02/01 15:04"
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

// Special family groups that require special handling.
var SpecialFamilies = map[string]bool{
	"jancokbot": true,
	"jancok":    true,
}
