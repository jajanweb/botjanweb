package payment

import (
	"context"

	"github.com/exernia/botjanweb/internal/domain/entity"
	"github.com/exernia/botjanweb/pkg/logger"
	"github.com/exernia/botjanweb/presentation/template"
)

var confirmationLogger = logger.Confirmation

// ConfirmationService handles payment confirmation workflows.
// This service orchestrates the entire payment confirmation process including
// customer notification, QRIS revocation, group notification, and order logging.
type ConfirmationService struct {
	notifier NotificationPort
	sheets   SheetsPort
}

// NewConfirmationService creates a new payment confirmation service.
func NewConfirmationService(notifier NotificationPort, sheets SheetsPort) *ConfirmationService {
	return &ConfirmationService{
		notifier: notifier,
		sheets:   sheets,
	}
}

// ConfirmPayment handles the complete payment confirmation workflow.
// This is called when a payment is matched with a pending QRIS.
func (s *ConfirmationService) ConfirmPayment(ctx context.Context, pending *entity.PendingPayment, notif *entity.DANANotification) error {
	confirmationLogger.Printf("💰 Payment confirmed: Rp%d | %s | %s",
		notif.Amount, pending.Nama, pending.Email)

	// 1. Send confirmation to customer
	if err := s.sendPaymentConfirmation(ctx, pending, notif); err != nil {
		confirmationLogger.Printf("❌ Failed to send confirmation: %v", err)
		// Don't return error, continue with other steps
	}

	// 2. Revoke QRIS image (delete for everyone)
	if err := s.revokeQRISImage(ctx, pending); err != nil {
		confirmationLogger.Printf("⚠️ Failed to revoke QRIS image: %v", err)
		// Don't return error, continue with other steps
	}

	// 3. Send group notification (for self-QRIS only)
	if pending.IsSelfQris {
		if err := s.sendGroupNotification(ctx, pending, notif.Amount); err != nil {
			confirmationLogger.Printf("⚠️ Failed to send group notification: %v", err)
			// Don't return error, continue with other steps
		}
	}

	// 4. Save order to Google Sheets (if enabled and valid)
	if err := s.saveOrderToSheets(ctx, pending); err != nil {
		confirmationLogger.Printf("❌ Failed to save to Sheets: %v", err)
		// Send error notification
		s.notifySheetError(ctx, pending, err.Error())
		// Don't return error, payment is already confirmed
	} else if s.sheets != nil && pending.Produk != "" {
		// Only send success notification if sheets is enabled and product exists
		s.notifySheetSuccess(ctx, pending)
	}

	return nil
}

// sendPaymentConfirmation sends payment confirmation message to customer.
func (s *ConfirmationService) sendPaymentConfirmation(ctx context.Context, pending *entity.PendingPayment, notif *entity.DANANotification) error {
	confirmMsg := template.BuildPaymentConfirmation(pending, notif)

	if err := s.notifier.SendPaymentConfirmation(ctx, pending, confirmMsg); err != nil {
		return err
	}

	confirmationLogger.Printf("✅ Confirmation sent to %s", pending.ChatID)
	return nil
}

// revokeQRISImage revokes (deletes) the QRIS image message.
func (s *ConfirmationService) revokeQRISImage(ctx context.Context, pending *entity.PendingPayment) error {
	if pending.MessageID == "" {
		return nil
	}

	if err := s.notifier.RevokeQRISImage(ctx, pending.ChatID, pending.MessageID); err != nil {
		return err
	}

	confirmationLogger.Printf("🗑️ QRIS image revoked: %s", pending.MessageID)
	return nil
}

// sendGroupNotification sends payment notification to group (for self-QRIS).
func (s *ConfirmationService) sendGroupNotification(ctx context.Context, pending *entity.PendingPayment, amount int) error {
	groupNotif := template.BuildSelfQrisPaymentNotification(pending, amount)

	if err := s.notifier.SendGroupNotification(ctx, groupNotif, pending.GroupNotifMsgID); err != nil {
		return err
	}

	if pending.GroupNotifMsgID != "" {
		confirmationLogger.Printf("📢 Self-QRIS payment notification sent to group (reply to %s)", pending.GroupNotifMsgID)
	} else {
		confirmationLogger.Printf("📢 Self-QRIS payment notification sent to group")
	}
	return nil
}

// saveOrderToSheets saves order to Google Sheets if enabled and valid.
func (s *ConfirmationService) saveOrderToSheets(ctx context.Context, pending *entity.PendingPayment) error {
	// Skip if Sheets not enabled or no product data
	if s.sheets == nil || pending.Produk == "" {
		return nil
	}

	// Validate product first
	if _, err := entity.ParseProduct(pending.Produk); err != nil {
		confirmationLogger.Printf("⚠️ Product '%s' is not valid, skipping Sheets logging", pending.Produk)
		return nil
	}

	// Create order and save
	order := entity.NewOrderFromPending(pending)
	if err := s.sheets.LogOrder(ctx, order); err != nil {
		return err
	}

	return nil
}

// notifySheetError notifies about Google Sheets save error.
func (s *ConfirmationService) notifySheetError(ctx context.Context, pending *entity.PendingPayment, errMsg string) {
	errorNotif := template.BuildSheetErrorNotification(pending, errMsg)

	// Send to group
	var replyToMsgID string
	if pending.IsSelfQris && pending.GroupNotifMsgID != "" {
		// Self-QRIS: reply to the group notification message
		replyToMsgID = pending.GroupNotifMsgID
		confirmationLogger.Printf("📢 Error notification sent to group (reply to %s)", pending.GroupNotifMsgID)
	} else {
		// Group order or self-QRIS without GroupNotifMsgID: send as new message
		confirmationLogger.Printf("📢 Error notification sent to group")
	}

	if err := s.notifier.SendGroupNotification(ctx, errorNotif, replyToMsgID); err != nil {
		confirmationLogger.Printf("⚠️ Failed to send error notification to group: %v", err)
	}

	// For non-self-QRIS, also notify user personally
	if !pending.IsSelfQris {
		errorMsg := "⚠️ Pembayaran berhasil tapi gagal mencatat ke spreadsheet. Silakan catat manual."
		if err := s.notifier.SendPaymentConfirmation(ctx, pending, errorMsg); err != nil {
			confirmationLogger.Printf("⚠️ Failed to send error message to user: %v", err)
		}
	}
}

// notifySheetSuccess notifies about successful Google Sheets save.
func (s *ConfirmationService) notifySheetSuccess(ctx context.Context, pending *entity.PendingPayment) {
	if pending.IsSelfQris {
		confirmationLogger.Printf("📊 Self-QRIS recorded: %s → Sheet %s | Rp%d",
			pending.SenderPhone, pending.Produk, pending.Amount)
	} else {
		confirmationLogger.Printf("📊 Order saved to sheet: %s", pending.Produk)

		// Send order saved notification to user
		savedMsg := template.BuildOrderSavedNotification(pending)
		if err := s.notifier.SendPaymentConfirmation(ctx, pending, savedMsg); err != nil {
			confirmationLogger.Printf("⚠️ Failed to send order saved notification: %v", err)
		}
	}
}
