// Package family implements family validation use case.
package family

import (
	"context"
	"fmt"

	"github.com/exernia/botjanweb/pkg/constants"

	"github.com/exernia/botjanweb/internal/application/service"
	"github.com/exernia/botjanweb/internal/domain"
	"github.com/exernia/botjanweb/internal/domain/entity"
)

// UseCase implements family validation business logic.
type UseCase struct {
	validator usecase.FamilyValidatorPort
}

// New creates a new family use case.
func New(validator usecase.FamilyValidatorPort) *UseCase {
	return &UseCase{
		validator: validator,
	}
}

// ValidateFamily validates a family value before QRIS generation.
// Returns FamilyValidation with details about the validation result.
func (uc *UseCase) ValidateFamily(ctx context.Context, family string) (*entity.FamilyValidation, error) {
	result := &entity.FamilyValidation{
		Email:    family,
		MaxSlots: constants.MaxFamilySlots,
	}

	// Check if it's a special family (skip Akun Google validation)
	if entity.IsSpecialFamily(family) {
		result.IsValid = true
		result.IsSpecial = true
		// For special families, count slots from Gemini sheet
		count, err := uc.validator.CountFamilySlots(ctx, family)
		if err != nil {
			return nil, fmt.Errorf("gagal mengecek slot family: %w", err)
		}
		result.UsedSlots = count

		// Check if full
		if count >= constants.MaxFamilySlots {
			result.IsValid = false
			result.ErrorMessage = fmt.Sprintf("Family '%s' sudah penuh (%d/%d slot terpakai)", family, count, constants.MaxFamilySlots)
			return result, domain.ErrFamilyFull
		}

		return result, nil
	}

	// Regular family: validate email exists in Akun Google
	exists, err := uc.validator.ValidateFamily(ctx, family)
	if err != nil {
		return nil, fmt.Errorf("gagal validasi family: %w", err)
	}

	if !exists {
		result.IsValid = false
		result.ErrorMessage = fmt.Sprintf("Family '%s' tidak ditemukan di Akun Google", family)
		return result, domain.ErrFamilyNotFound
	}

	// Count used slots
	count, err := uc.validator.CountFamilySlots(ctx, family)
	if err != nil {
		return nil, fmt.Errorf("gagal mengecek slot family: %w", err)
	}
	result.UsedSlots = count

	// Check if full
	if count >= constants.MaxFamilySlots {
		result.IsValid = false
		result.ErrorMessage = fmt.Sprintf("Family '%s' sudah penuh (%d/%d slot terpakai)", family, count, constants.MaxFamilySlots)
		return result, domain.ErrFamilyFull
	}

	result.IsValid = true
	return result, nil
}

// FormatSlotStatus returns a formatted string showing slot usage.
func FormatSlotStatus(validation *entity.FamilyValidation) string {
	remaining := validation.MaxSlots - validation.UsedSlots
	if remaining <= 0 {
		return fmt.Sprintf("❌ %s: Penuh (%d/%d)", validation.Email, validation.UsedSlots, validation.MaxSlots)
	}
	return fmt.Sprintf("✅ %s: %d/%d terpakai, %d tersedia", validation.Email, validation.UsedSlots, validation.MaxSlots, remaining)
}
