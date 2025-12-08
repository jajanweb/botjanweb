// Package config handles loading and parsing configuration from environment variables.
package config

import (
	"os"
	"strconv"
)

// Default configuration values.
const (
	DefaultWebhookPort = 8080 // Default port for webhook server
)

// getEnv reads an environment variable with a fallback default.
func getEnv(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}

// getEnvInt reads an environment variable as integer with a fallback default.
func getEnvInt(key string, defaultVal int) int {
	val := os.Getenv(key)
	if val == "" {
		return defaultVal
	}
	intVal, err := strconv.Atoi(val)
	if err != nil {
		return defaultVal
	}
	return intVal
}

// getEnvBool reads an environment variable as boolean.
func getEnvBool(key string, defaultVal bool) bool {
	val := os.Getenv(key)
	if val == "" {
		return defaultVal
	}
	return val == "true" || val == "1" || val == "yes"
}

// getWebhookPort reads webhook port from environment.
// Prioritizes Heroku's $PORT, then WEBHOOK_PORT, then default 8080.
func getWebhookPort() int {
	// Heroku sets $PORT environment variable
	if port := os.Getenv("PORT"); port != "" {
		if intVal, err := strconv.Atoi(port); err == nil {
			return intVal
		}
	}

	// Fallback to WEBHOOK_PORT or default 8080
	return getEnvInt("WEBHOOK_PORT", DefaultWebhookPort)
}
