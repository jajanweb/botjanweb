// Package logger provides centralized logging utilities.
package logger

import (
	"log"
	"os"

	"github.com/exernia/botjanweb/pkg/constants"
)

// Standard loggers for each component.
var (
	Main         = log.New(os.Stdout, constants.LogPrefixMain, log.LstdFlags)
	App          = log.New(os.Stdout, constants.LogPrefixApp, log.LstdFlags)
	Config       = log.New(os.Stdout, constants.LogPrefixConfig, log.LstdFlags)
	Webhook      = log.New(os.Stdout, constants.LogPrefixWebhook, log.LstdFlags)
	WhatsApp     = log.New(os.Stdout, constants.LogPrefixWA, log.LstdFlags)
	Payment      = log.New(os.Stdout, constants.LogPrefixPayment, log.LstdFlags)
	Confirmation = log.New(os.Stdout, constants.LogPrefixConfirmation, log.LstdFlags)
	QRIS         = log.New(os.Stdout, constants.LogPrefixQRIS, log.LstdFlags)
	Bot          = log.New(os.Stdout, constants.LogPrefixBot, log.LstdFlags)
)

// New creates a new logger with the given prefix.
func New(prefix string) *log.Logger {
	return log.New(os.Stdout, prefix, log.LstdFlags)
}
