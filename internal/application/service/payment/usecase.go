// Package payment implements payment matching use case.
package payment

import (
	"context"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/exernia/botjanweb/pkg/constants"

	"github.com/exernia/botjanweb/internal/application/service"
	"github.com/exernia/botjanweb/internal/domain"
	"github.com/exernia/botjanweb/internal/domain/entity"
)

var danaAmountRegex = regexp.MustCompile(`Rp\s?([\d.]+)`)

// UseCase implements PaymentUseCase.
type UseCase struct {
	store usecase.PendingStorePort
}

// New creates a new payment use case.
func New(store usecase.PendingStorePort) *UseCase {
	return &UseCase{
		store: store,
	}
}

// RegisterPending adds a pending payment to the store.
func (uc *UseCase) RegisterPending(pending *entity.PendingPayment) {
	uc.store.Add(pending)
}

// GetPendingCount returns total pending payments.
func (uc *UseCase) GetPendingCount() int {
	return uc.store.Count()
}

// ProcessNotification processes a webhook payload and tries to match it with pending payment.
func (uc *UseCase) ProcessNotification(ctx context.Context, payload *entity.WebhookPayload) (*entity.PendingPayment, *entity.DANANotification, error) {
	// Check if this is a DANA payment notification
	if !IsDANAPaymentNotification(payload.App, payload.Message) {
		return nil, nil, fmt.Errorf("bukan notifikasi pembayaran DANA")
	}

	// Parse the DANA notification
	timestamp := ParseWebhookTimestamp(payload.Timestamp)
	notification, err := ParseDANANotification(payload.Message, timestamp)
	if err != nil {
		return nil, nil, err
	}

	// Try to match with pending payment
	matched := uc.store.Match(notification.Amount)
	return matched, notification, nil
}

// ParseDANANotification parses a DANA payment notification message.
func ParseDANANotification(message string, timestamp time.Time) (*entity.DANANotification, error) {
	matches := danaAmountRegex.FindStringSubmatch(message)
	if len(matches) < 2 {
		return nil, fmt.Errorf("%w: tidak ditemukan nominal Rupiah", domain.ErrInvalidNotification)
	}

	amountStr := strings.ReplaceAll(matches[1], ".", "")
	amount, err := strconv.Atoi(amountStr)
	if err != nil || amount <= 0 {
		return nil, fmt.Errorf("%w: nominal tidak valid: %s", domain.ErrInvalidNotification, matches[1])
	}

	return &entity.DANANotification{
		Amount:     amount,
		RawMessage: message,
		Timestamp:  timestamp,
	}, nil
}

// IsDANAPaymentNotification checks if notification is a DANA payment receipt.
func IsDANAPaymentNotification(appPackage, message string) bool {
	if appPackage != constants.DANAPackage {
		return false
	}
	lower := strings.ToLower(message)
	return strings.Contains(lower, "berhasil menerima") && strings.Contains(lower, "rp")
}

// ParseWebhookTimestamp converts Unix milliseconds string to time.Time.
func ParseWebhookTimestamp(ts string) time.Time {
	if millis, err := strconv.ParseInt(ts, 10, 64); err == nil {
		return time.UnixMilli(millis)
	}
	return time.Now()
}
