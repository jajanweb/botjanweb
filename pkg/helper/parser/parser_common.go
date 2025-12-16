// Package parser provides command and form parsing utilities.
package parser

// ============================================================================
// COMMON CONSTANTS
// ============================================================================

// Form field names (case-insensitive).
const (
	fieldProduk    = "produk"
	fieldNama      = "nama"
	fieldEmail     = "email"
	fieldFamily    = "family"
	fieldNominal   = "nominal"
	fieldKanal     = "kanal"
	fieldAkun      = "akun"
	fieldWorkspace = "workspace"
	fieldPaket     = "paket"
	fieldTipe      = "tipe"
	fieldSandi     = "sandi"
	fieldUntuk     = "untuk" // Target phone for self-QRIS
)

// ProductParam represents the product/account type from command parameter.
// Used for both #qris and #addakun commands.
type ProductParam string

const (
	ProductParamGoogle  ProductParam = "google"
	ProductParamChatGPT ProductParam = "chatgpt"
	ProductParamUnknown ProductParam = ""
)
