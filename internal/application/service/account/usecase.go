// Package account implements account management use case.
package account

import (
	"context"
	"fmt"
	"time"

	"github.com/exernia/botjanweb/internal/application/service"
	"github.com/exernia/botjanweb/internal/domain"
	"github.com/exernia/botjanweb/internal/domain/entity"
)

// UseCase implements account management business logic.
type UseCase struct {
	repo usecase.AccountRepositoryPort
}

// New creates a new account use case.
func New(repo usecase.AccountRepositoryPort) *UseCase {
	return &UseCase{
		repo: repo,
	}
}

// AddAccount adds a new account based on the command type.
func (uc *UseCase) AddAccount(ctx context.Context, cmd *entity.AddAkunCommand) error {
	switch cmd.Tipe {
	case entity.AccountTypeGoogle:
		return uc.addAkunGoogle(ctx, cmd)
	case entity.AccountTypeChatGPT:
		return uc.addAkunChatGPT(ctx, cmd)
	default:
		return domain.ErrInvalidAccType
	}
}

// addAkunGoogle adds a new Google account.
func (uc *UseCase) addAkunGoogle(ctx context.Context, cmd *entity.AddAkunCommand) error {
	// Create account entity
	now := time.Now()
	expiry := now.AddDate(1, 0, 0) // Add 1 year

	akun := &entity.AkunGoogle{
		Email:           cmd.Email,
		Sandi:           cmd.Sandi,
		TanggalAktivasi: now,
		TanggalBerakhir: expiry.Format("2006-01-02"), // Format as YYYY-MM-DD
		StatusDibuat:    "",
		YTPremium:       "",
	}

	// Add to sheet
	if err := uc.repo.AddAkunGoogle(ctx, akun); err != nil {
		return fmt.Errorf("gagal menambah akun: %w", err)
	}

	return nil
}

// addAkunChatGPT adds a new ChatGPT account.
func (uc *UseCase) addAkunChatGPT(ctx context.Context, cmd *entity.AddAkunCommand) error {
	// Create account entity
	akun := &entity.AkunChatGPT{
		Email:           cmd.Email,
		Sandi:           cmd.Sandi,
		Workspace:       cmd.Workspace,
		Status:          "",
		TanggalAktivasi: time.Now(),
		TanggalKenaBan:  "",
	}

	// Add to sheet
	if err := uc.repo.AddAkunChatGPT(ctx, akun); err != nil {
		return fmt.Errorf("gagal menambah akun: %w", err)
	}

	return nil
}

// ListAccounts fetches all accounts based on the command filter.
func (uc *UseCase) ListAccounts(ctx context.Context, cmd *entity.ListAkunCommand) (*entity.AccountListResult, error) {
	result, err := uc.repo.GetAccountListResult(ctx)
	if err != nil {
		return nil, fmt.Errorf("gagal mengambil daftar akun: %w", err)
	}

	// Filter if specific type is requested
	switch cmd.Tipe {
	case entity.AccountTypeGoogle:
		result.ChatGPTAccounts = nil
		result.TotalChatGPT = 0
		result.AvailableChatGPT = 0
	case entity.AccountTypeChatGPT:
		result.GoogleAccounts = nil
		result.TotalGoogle = 0
		result.AvailableGoogle = 0
	}

	return result, nil
}
