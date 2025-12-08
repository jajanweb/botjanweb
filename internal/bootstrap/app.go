package bootstrap

import (
	"context"
	"fmt"
	"log"

	appservice "github.com/exernia/botjanweb/internal/application/service"
	accountuc "github.com/exernia/botjanweb/internal/application/service/account"
	familyuc "github.com/exernia/botjanweb/internal/application/service/family"
	paymentuc "github.com/exernia/botjanweb/internal/application/service/payment"
	qrisuc "github.com/exernia/botjanweb/internal/application/service/qris"
	"github.com/exernia/botjanweb/internal/bootstrap/adapters"
	"github.com/exernia/botjanweb/internal/config"
	"github.com/exernia/botjanweb/internal/domain/entity"
	infraqris "github.com/exernia/botjanweb/internal/infrastructure/external/qris"
	infrawebhook "github.com/exernia/botjanweb/internal/infrastructure/messaging/webhook"
	infrawa "github.com/exernia/botjanweb/internal/infrastructure/messaging/whatsapp"
	reposheets "github.com/exernia/botjanweb/internal/infrastructure/persistence/sheets"
	"github.com/exernia/botjanweb/pkg/logger"
	botctrl "github.com/exernia/botjanweb/presentation/handler/bot"
	httpctrl "github.com/exernia/botjanweb/presentation/handler/http"
)

// App holds all application dependencies.
type App struct {
	Config *config.Config
	Logger *log.Logger

	// Infrastructure
	WAClient      *infrawa.Client
	QrisGenerator *infraqris.Generator
	WebhookServer *infrawebhook.Server

	// Repository
	PendingStore appservice.PendingStorePort
	SheetsRepo   *reposheets.Repository

	// Use Cases
	QrisUC    *qrisuc.UseCase
	PaymentUC *paymentuc.UseCase
	FamilyUC  *familyuc.UseCase
	AccountUC *accountuc.UseCase

	// Domain Services
	ConfirmationService *paymentuc.ConfirmationService

	// Controllers
	BotHandler          *botctrl.Handler
	WebhookController   *httpctrl.WebhookController
	QRPairingController *httpctrl.QRPairingController
}

// New initializes all application components with dependency injection.
func New(ctx context.Context, assetsPath string) (*App, error) {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	app := &App{
		Config: cfg,
		Logger: logger.App,
	}

	// Initialize infrastructure layer
	if err := app.initInfrastructure(assetsPath); err != nil {
		return nil, fmt.Errorf("failed to init infrastructure: %w", err)
	}

	// Initialize repository layer
	if err := app.initRepositories(ctx); err != nil {
		return nil, fmt.Errorf("failed to init repositories: %w", err)
	}

	// Initialize use cases
	app.initUseCases()

	// Initialize controllers
	app.initControllers()

	return app, nil
}

// initUseCases sets up use case components.
func (app *App) initUseCases() {
	// QRIS use case
	app.QrisUC = qrisuc.New(
		adapters.NewQrisGeneratorAdapter(app.QrisGenerator),
		app.Config.QRISStaticPayload,
	)

	// Payment use case
	app.PaymentUC = paymentuc.New(app.PendingStore)

	// Payment confirmation service - will be initialized after WAClient is ready
	// For now, use nil adapter (will be replaced in initPaymentConfirmationService)
	sheetsPort := adapters.NewSheetsAdapter(app.SheetsRepo)
	app.ConfirmationService = paymentuc.NewConfirmationService(nil, sheetsPort)

	// Family validation use case
	if app.SheetsRepo != nil {
		app.FamilyUC = familyuc.New(app.SheetsRepo)
	}

	// Account management use case
	if app.SheetsRepo != nil {
		app.AccountUC = accountuc.New(app.SheetsRepo)
	}
}

// initControllers sets up controller components.
func (app *App) initControllers() {
	// Bot message handler with all allowed senders
	app.BotHandler = botctrl.NewHandler(
		app.QrisUC,
		app.PaymentUC,
		app.FamilyUC,
		app.AccountUC,
		app.Config.AllowedSenders,
		app.Config.SheetAkunGoogle,
		app.Config.SheetAkunChatGPT,
		app.Config.DefaultKanal,
	)

	// Webhook controller with payment confirmation service
	// Wrap the service method to match the expected signature (no error return)
	confirmHandler := func(ctx context.Context, pending *entity.PendingPayment, notif *entity.DANANotification) {
		_ = app.ConfirmationService.ConfirmPayment(ctx, pending, notif)
	}
	app.WebhookController = httpctrl.NewWebhookController(
		app.Config.WebhookSecret,
		app.PaymentUC,
		confirmHandler,
	)

	// Note: QR Pairing controller is initialized later in Run() after WAClient is created
}

// initQRPairingController initializes the QR pairing controller.
// Called from Run() after WAClient is available.
func (app *App) initQRPairingController() {
	if app.WAClient != nil {
		controller, err := httpctrl.NewQRPairingController(
			app.WAClient,
			app.Config.WebhookSecret,
		)
		if err != nil {
			log.Fatalf("Failed to initialize QR pairing controller: %v", err)
		}
		app.QRPairingController = controller
	}
}

// initPaymentConfirmationService initializes payment confirmation service with WhatsApp adapter.
// Called from Run() after WAClient is available.
func (app *App) initPaymentConfirmationService() {
	if app.WAClient != nil && app.ConfirmationService != nil {
		notificationPort := adapters.NewWhatsAppNotificationAdapter(app.WAClient, app.Config.GroupJID)
		sheetsPort := adapters.NewSheetsAdapter(app.SheetsRepo)
		app.ConfirmationService = paymentuc.NewConfirmationService(notificationPort, sheetsPort)
	}
}

// buildHTTPRouter builds the HTTP router with all registered routes.
func (app *App) buildHTTPRouter() *httpctrl.Router {
	router := httpctrl.NewRouter()
	httpctrl.RegisterRoutes(router, app.WebhookController, app.QRPairingController)
	return router
}
