// Package config handles loading and parsing configuration from environment variables.
package config

import (
	"github.com/exernia/botjanweb/pkg/constants"
	"github.com/exernia/botjanweb/pkg/helper/parser"
	"github.com/joho/godotenv"
)

// Load reads configuration from environment variables.
// It attempts to load .env file first (for local development),
// then reads from the environment.
func Load() (*Config, error) {
	// Try loading .env file, ignore error if not found (production may use real env vars)
	_ = godotenv.Load()

	cfg := &Config{
		WhatsAppDBPath:        getEnv("WHATSAPP_DB_PATH", "./whatsmeow.db"),
		GroupJID:              getEnv("GROUP_JID", ""),
		QRISStaticPayload:     getEnv("QRIS_STATIC_PAYLOAD", ""),
		MerchantName:          getEnv("MERCHANT_NAME", "JAJAN WEB"),
		SheetsEnabled:         getEnvBool("SHEETS_ENABLED", false),
		GoogleSpreadsheetID:   getEnv("GOOGLE_SPREADSHEET_ID", ""),
		GoogleCredentialsPath: getEnv("GOOGLE_CREDENTIALS_PATH", "./credentials.json"),
		SheetQRISTransactions: getEnv("SHEET_QRIS_TRANSACTIONS", "TransaksiQRIS"),
		SheetOrders:           getEnv("SHEET_ORDERS", "Pemesanan"),
		SheetAkunGoogle:       getEnv("SHEET_AKUN_GOOGLE", "Akun Google"),
		SheetAkunChatGPT:      getEnv("SHEET_AKUN_CHATGPT", "Akun ChatGPT"),
		DefaultKanal:          getEnv("DEFAULT_KANAL", constants.DefaultKanal),
		WebhookEnabled:        getEnvBool("WEBHOOK_ENABLED", false),
		WebhookPort:           getWebhookPort(),
		WebhookSecret:         getEnv("WEBHOOK_SECRET", ""),
		DatabaseURL:           getEnv("DATABASE_URL", ""),
		HerokuAppName:         getEnv("HEROKU_APP_NAME", ""),
	}

	// Parse allowed senders (comma-separated) using common utility
	cfg.AllowedSenders = parser.ParsePhoneList(getEnv("ALLOWED_SENDERS", ""))

	// Validate required fields
	if err := cfg.validate(); err != nil {
		return nil, err
	}

	return cfg, nil
}
