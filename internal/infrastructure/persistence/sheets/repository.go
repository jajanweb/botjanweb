// Package sheets implements Google Sheets repository for data logging.
package sheets

import (
	"context"
	"fmt"
	"log"

	"github.com/exernia/botjanweb/pkg/logger"
	"google.golang.org/api/option"
	"google.golang.org/api/sheets/v4"
)

// Repository implements TransactionLogPort for Google Sheets.
type Repository struct {
	service       *sheets.Service
	spreadsheetID string
	logger        *log.Logger

	// Sheet names
	ordersSheet      string
	akunGoogleSheet  string
	akunChatGPTSheet string
}

// NewRepository creates a new Sheets repository.
// Accepts either credentialsPath (for local dev) or credentialsJSON (for cloud deployment like Heroku).
// If credentialsJSON is provided, it takes precedence over credentialsPath.
func NewRepository(spreadsheetID, credentialsPath, credentialsJSON, ordersSheet, akunGoogleSheet, akunChatGPTSheet string) (*Repository, error) {
	ctx := context.Background()

	var srv *sheets.Service
	var err error

	// Prefer JSON credentials (for cloud deployment) over file path
	if credentialsJSON != "" {
		srv, err = sheets.NewService(ctx, option.WithCredentialsJSON([]byte(credentialsJSON)))
		if err != nil {
			return nil, fmt.Errorf("failed to create sheets service from JSON: %w", err)
		}
	} else {
		srv, err = sheets.NewService(ctx, option.WithCredentialsFile(credentialsPath))
		if err != nil {
			return nil, fmt.Errorf("failed to create sheets service from file: %w", err)
		}
	}

	return &Repository{
		service:          srv,
		spreadsheetID:    spreadsheetID,
		logger:           logger.QRIS,
		ordersSheet:      ordersSheet,
		akunGoogleSheet:  akunGoogleSheet,
		akunChatGPTSheet: akunChatGPTSheet,
	}, nil
}
