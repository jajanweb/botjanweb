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

// ValidateWorkspace validates a workspace owner email before QRIS generation.
// Input: ownerEmail (e.g., "gptadmin03@jajanweb.id")
// Returns WorkspaceValidation with slot details.
func (uc *UseCase) ValidateWorkspace(ctx context.Context, ownerEmail string) (*entity.WorkspaceValidation, error) {
	result := &entity.WorkspaceValidation{
		OwnerEmail: ownerEmail,
		MaxSlots:   constants.MaxWorkspaceSlots,
	}

	// Validate email exists in Akun ChatGPT sheet
	exists, err := uc.validator.ValidateWorkspaceEmail(ctx, ownerEmail)
	if err != nil {
		return nil, fmt.Errorf("gagal validasi workspace: %w", err)
	}

	if !exists {
		result.IsValid = false
		result.ErrorMessage = fmt.Sprintf("Email workspace '%s' tidak ditemukan di Akun ChatGPT", ownerEmail)
		return result, domain.ErrWorkspaceNotFound
	}

	// Count used slots for this owner email in ChatGPT sheet (column D)
	count, err := uc.validator.CountWorkspaceSlots(ctx, ownerEmail)
	if err != nil {
		return nil, fmt.Errorf("gagal mengecek slot workspace: %w", err)
	}
	result.UsedSlots = count

	// Check if full
	if count >= constants.MaxWorkspaceSlots {
		result.IsValid = false
		result.ErrorMessage = fmt.Sprintf("Workspace '%s' sudah penuh (%d/%d slot terpakai)", ownerEmail, count, constants.MaxWorkspaceSlots)
		return result, domain.ErrWorkspaceFull
	}

	result.IsValid = true
	return result, nil
}

// FormatSlotStatus returns a formatted string showing slot usage.
func FormatSlotStatus(validation *entity.WorkspaceValidation) string {
	remaining := validation.MaxSlots - validation.UsedSlots
	if remaining <= 0 {
		return fmt.Sprintf("❌ %s: Penuh (%d/%d)", validation.OwnerEmail, validation.UsedSlots, validation.MaxSlots)
	}
	return fmt.Sprintf("✅ %s: %d/%d terpakai, %d tersedia", validation.OwnerEmail, validation.UsedSlots, validation.MaxSlots, remaining)
}
