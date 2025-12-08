// Package bootstrap handles application initialization and dependency injection.
package bootstrap

import (
	"context"
	"fmt"

	infraqris "github.com/exernia/botjanweb/internal/infrastructure/external/qris"
	repomemory "github.com/exernia/botjanweb/internal/infrastructure/persistence/memory"
	repopostgres "github.com/exernia/botjanweb/internal/infrastructure/persistence/postgres"
	reposheets "github.com/exernia/botjanweb/internal/infrastructure/persistence/sheets"
)

// initInfrastructure initializes infrastructure layer components.
func (app *App) initInfrastructure(assetsPath string) error {
	// QRIS Generator
	app.QrisGenerator = infraqris.NewGenerator(assetsPath)

	app.Logger.Printf("✅ Infrastructure initialized (assetsPath: %s)", assetsPath)
	return nil
}

// initRepositories initializes repository layer components.
func (app *App) initRepositories(ctx context.Context) error {
	// Pending payment store: PostgreSQL (production) or In-memory (development)
	if app.Config.DatabaseURL != "" {
		// Use PostgreSQL (production/Heroku)
		app.Logger.Println("📊 Using PostgreSQL pending store")
		pgStore, err := repopostgres.NewPendingStore(ctx, app.Config.DatabaseURL)
		if err != nil {
			return fmt.Errorf("failed to init PostgreSQL pending store: %w", err)
		}
		app.PendingStore = pgStore
		app.PendingStore.StartCleanup()
	} else {
		// Use in-memory store (local development)
		app.Logger.Println("📊 Using in-memory pending store (development mode)")
		memStore := repomemory.NewPendingStore()
		app.PendingStore = memStore
		app.PendingStore.StartCleanup()
	}

	// Google Sheets repository (optional, only if enabled)
	if app.Config.SheetsEnabled {
		repo, err := reposheets.NewRepository(
			app.Config.GoogleSpreadsheetID,
			app.Config.GoogleCredentialsPath,
			"",                          // qrisSheet (legacy, not used)
			app.Config.SheetOrders,      // ordersSheet
			app.Config.SheetAkunGoogle,  // akunGoogleSheet
			app.Config.SheetAkunChatGPT, // akunChatGPTSheet
		)
		if err != nil {
			return fmt.Errorf("failed to init sheets repository: %w", err)
		}
		app.SheetsRepo = repo
		app.Logger.Printf("✅ Google Sheets repository initialized")
	}

	app.Logger.Printf("✅ Repositories initialized")
	return nil
}
