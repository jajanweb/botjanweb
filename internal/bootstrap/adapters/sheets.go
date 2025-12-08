// Package adapters contains infrastructure adapters that implement domain ports.
package adapters

import (
	"context"

	"github.com/exernia/botjanweb/internal/application/service/payment"
	"github.com/exernia/botjanweb/internal/domain/entity"
	"github.com/exernia/botjanweb/internal/infrastructure/persistence/sheets"
)

// SheetsAdapter implements the SheetsPort interface.
// This adapter translates use-case sheet operations into repository operations.
type SheetsAdapter struct {
	repo *sheets.Repository
}

// NewSheetsAdapter creates a new Sheets adapter.
// Returns nil if repository is nil (Sheets not enabled).
func NewSheetsAdapter(repo *sheets.Repository) payment.SheetsPort {
	if repo == nil {
		return nil
	}
	return &SheetsAdapter{repo: repo}
}

// LogOrder saves order to Google Sheets.
func (a *SheetsAdapter) LogOrder(ctx context.Context, order *entity.Order) error {
	return a.repo.LogOrder(ctx, order)
}
