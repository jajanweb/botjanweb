// Package config handles loading and parsing configuration from environment variables.
//
// This package is organized into the following files:
//   - types.go: Config struct and validation methods
//   - loader.go: Configuration loading from environment
//   - env.go: Environment variable helper functions
//
// All configuration is loaded once at startup and should be treated as immutable.
package config

import (
	"fmt"
	"strings"
)

// Config holds all application configuration values.
// All values are loaded once at startup and should be treated as read-only.
type Config struct {
	// WhatsMeow database path for session storage
	WhatsAppDBPath string

	// List of allowed sender phone numbers (without +, e.g., "6282116086024")
	AllowedSenders []string

	// Target WhatsApp group JID (e.g., "123456789-1234567890@g.us")
	GroupJID string

	// Static QRIS payload string (EMV format) to convert to dynamic
	QRISStaticPayload string

	// Merchant/toko name to display on QRIS image and receipts
	MerchantName string

	// Google Sheets configuration
	SheetsEnabled         bool   // Toggle to enable/disable Google Sheets integration
	GoogleSpreadsheetID   string // Google Spreadsheet ID
	GoogleCredentialsPath string // Path to Google credentials JSON file (for local dev)
	GoogleCredentialsJSON string // Google credentials JSON content (for Heroku/cloud deployment)

	// Sheet names for different data types
	SheetOrders      string // Name of the orders sheet (customer orders)
	SheetAkunGoogle  string // Name of the Google accounts sheet (account management)
	SheetAkunChatGPT string // Name of the ChatGPT accounts sheet (account management)

	// Default values for orders
	DefaultKanal string // Default sales channel (e.g., "Threads")

	// Webhook configuration for payment notifications
	WebhookEnabled bool   // Toggle to enable/disable webhook server
	WebhookPort    int    // Port number for webhook server
	WebhookSecret  string // Secret key for request validation

	// Database configuration
	DatabaseURL string // PostgreSQL connection URL (optional, uses in-memory if empty)

	// Heroku configuration
	HerokuAppName string // Heroku app name for generating URLs (optional)
}

// validate checks that all required configuration values are present and valid.
func (c *Config) validate() error {
	// Required: GroupJID
	if c.GroupJID == "" {
		return fmt.Errorf("GROUP_JID is required")
	}
	if !strings.HasSuffix(c.GroupJID, "@g.us") {
		return fmt.Errorf("GROUP_JID must be a group JID (must end with @g.us), got: %s", c.GroupJID)
	}

	// Required: QRIS payload
	if c.QRISStaticPayload == "" {
		return fmt.Errorf("QRIS_STATIC_PAYLOAD is required")
	}
	if len(c.QRISStaticPayload) < 50 {
		return fmt.Errorf("QRIS_STATIC_PAYLOAD seems invalid (too short: %d characters)", len(c.QRISStaticPayload))
	}

	// Required: Merchant name
	if c.MerchantName == "" {
		return fmt.Errorf("MERCHANT_NAME is required for QRIS display")
	}

	// Required: At least one allowed sender
	if len(c.AllowedSenders) == 0 {
		return fmt.Errorf("ALLOWED_SENDERS is required (at least one phone number)")
	}

	// Validate phone numbers format (already normalized by ParsePhoneList)
	for i, phone := range c.AllowedSenders {
		// Phone is already normalized, just validate length
		if len(phone) < 10 || len(phone) > 15 {
			return fmt.Errorf("ALLOWED_SENDERS[%d] invalid phone number: %s (must be 10-15 digits)", i, phone)
		}
	}

	// Google Sheets config is only required if SHEETS_ENABLED=true
	if c.SheetsEnabled {
		if c.GoogleSpreadsheetID == "" {
			return fmt.Errorf("GOOGLE_SPREADSHEET_ID is required when SHEETS_ENABLED=true")
		}
		// Either credentials path OR JSON content must be provided
		if c.GoogleCredentialsPath == "" && c.GoogleCredentialsJSON == "" {
			return fmt.Errorf("either GOOGLE_CREDENTIALS_PATH or GOOGLE_CREDENTIALS_JSON is required when SHEETS_ENABLED=true")
		}
	}

	// Webhook config validation if enabled
	if c.WebhookEnabled {
		if c.WebhookPort <= 0 || c.WebhookPort > 65535 {
			return fmt.Errorf("WEBHOOK_PORT must be between 1-65535, got: %d", c.WebhookPort)
		}
		if c.WebhookSecret == "" {
			return fmt.Errorf("WEBHOOK_SECRET is required when WEBHOOK_ENABLED=true (for security)")
		}
		if len(c.WebhookSecret) < 8 {
			return fmt.Errorf("WEBHOOK_SECRET too weak (minimum 8 characters), got: %d characters", len(c.WebhookSecret))
		}
	}

	return nil
}
