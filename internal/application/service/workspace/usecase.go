// Package workspace implements workspace validation use case for ChatGPT.
package workspace

import (
	"context"
	"fmt"

	"github.com/exernia/botjanweb/pkg/constants"

	usecase "github.com/exernia/botjanweb/internal/application/service"
	"github.com/exernia/botjanweb/internal/domain"
	"github.com/exernia/botjanweb/internal/domain/entity"
)

// UseCase implements workspace validation business logic.
type UseCase struct {
	validator usecase.WorkspaceValidatorPort
}

// New creates a new workspace use case.
func New(validator usecase.WorkspaceValidatorPort) *UseCase {
	return &UseCase{
		validator: validator,
	}
}

// ValidateWorkspace validates a workspace value before QRIS generation.
// Returns WorkspaceValidation with details about the validation result.
func (uc *UseCase) ValidateWorkspace(ctx context.Context, workspace string) (*entity.WorkspaceValidation, error) {
	result := &entity.WorkspaceValidation{
		Name:     workspace,
		MaxSlots: constants.MaxWorkspaceSlots,
	}

	// Validate workspace exists in Akun ChatGPT
	exists, err := uc.validator.ValidateWorkspace(ctx, workspace)
	if err != nil {
		return nil, fmt.Errorf("gagal validasi workspace: %w", err)
	}

	if !exists {
		result.IsValid = false
		result.ErrorMessage = fmt.Sprintf("Workspace '%s' tidak ditemukan di Akun ChatGPT", workspace)
		return result, domain.ErrWorkspaceNotFound
	}

	// Count used slots
	count, err := uc.validator.CountWorkspaceSlots(ctx, workspace)
	if err != nil {
		return nil, fmt.Errorf("gagal mengecek slot workspace: %w", err)
	}
	result.UsedSlots = count

	// Check if full
	if count >= constants.MaxWorkspaceSlots {
		result.IsValid = false
		result.ErrorMessage = fmt.Sprintf("Workspace '%s' sudah penuh (%d/%d slot terpakai)", workspace, count, constants.MaxWorkspaceSlots)
		return result, domain.ErrWorkspaceFull
	}

	result.IsValid = true
	return result, nil
}

// FormatSlotStatus returns a formatted string showing slot usage.
func FormatSlotStatus(validation *entity.WorkspaceValidation) string {
	remaining := validation.MaxSlots - validation.UsedSlots
	if remaining <= 0 {
		return fmt.Sprintf("❌ %s: Penuh (%d/%d)", validation.Name, validation.UsedSlots, validation.MaxSlots)
	}
	return fmt.Sprintf("✅ %s: %d/%d terpakai, %d tersedia", validation.Name, validation.UsedSlots, validation.MaxSlots, remaining)
}
