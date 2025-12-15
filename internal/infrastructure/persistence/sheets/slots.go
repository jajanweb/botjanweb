package sheets

import (
	"context"
	"fmt"
	"strings"

	"github.com/exernia/botjanweb/internal/domain/entity"
)

// ValidateFamily checks if a family email exists in Akun Google sheet (column A).
// Returns true if found, false otherwise.
func (r *Repository) ValidateFamily(ctx context.Context, familyEmail string) (bool, error) {
	// Read column A (email) from Akun Google sheet
	readRange := "'Akun Google'!A:A"
	resp, err := r.service.Spreadsheets.Values.Get(r.spreadsheetID, readRange).Do()
	if err != nil {
		return false, fmt.Errorf("failed to read Akun Google sheet: %w", err)
	}

	// Search for the email (case-insensitive)
	familyLower := strings.ToLower(strings.TrimSpace(familyEmail))
	for _, row := range resp.Values {
		if len(row) > 0 {
			cellValue := strings.ToLower(strings.TrimSpace(fmt.Sprintf("%v", row[0])))
			if cellValue == familyLower {
				return true, nil
			}
		}
	}

	return false, nil
}

// CountFamilySlots counts how many slots are used for a family in Gemini sheet (column D).
// Returns the count of non-empty rows with matching family value.
func (r *Repository) CountFamilySlots(ctx context.Context, family string) (int, error) {
	// Read columns B and D from Gemini sheet
	// B = Nama (to check if slot is used), D = Family
	readRange := "'Gemini'!B:D"
	resp, err := r.service.Spreadsheets.Values.Get(r.spreadsheetID, readRange).Do()
	if err != nil {
		return 0, fmt.Errorf("failed to read Gemini sheet: %w", err)
	}

	count := 0
	familyLower := strings.ToLower(strings.TrimSpace(family))

	// Skip header row (index 0)
	for i, row := range resp.Values {
		if i == 0 {
			continue // Skip header
		}
		if len(row) < 3 {
			continue // Need at least columns B, C, D
		}

		// Column B = Nama (index 0 in this range)
		// Column D = Family (index 2 in this range)
		nama := strings.TrimSpace(fmt.Sprintf("%v", row[0]))
		familyCell := strings.ToLower(strings.TrimSpace(fmt.Sprintf("%v", row[2])))

		// Count if family matches AND slot is filled (nama not empty)
		if familyCell == familyLower && nama != "" {
			count++
		}
	}

	return count, nil
}

// ValidateWorkspaceEmail checks if workspace owner email exists in Akun ChatGPT sheet (column A).
// Returns true if found, false otherwise.
func (r *Repository) ValidateWorkspaceEmail(ctx context.Context, ownerEmail string) (bool, error) {
	// Read column A (Email) from Akun ChatGPT sheet
	readRange := "'Akun ChatGPT'!A:A"
	resp, err := r.service.Spreadsheets.Values.Get(r.spreadsheetID, readRange).Do()
	if err != nil {
		return false, fmt.Errorf("failed to read Akun ChatGPT sheet: %w", err)
	}

	// Search for the owner email (case-insensitive)
	emailLower := strings.ToLower(strings.TrimSpace(ownerEmail))
	for i, row := range resp.Values {
		if i == 0 {
			continue // Skip header row
		}
		if len(row) < 1 {
			continue
		}

		emailCell := strings.ToLower(strings.TrimSpace(fmt.Sprintf("%v", row[0])))
		if emailCell == emailLower {
			return true, nil
		}
	}

	return false, nil // Email not found
}

// CountWorkspaceSlots counts how many slots are used for owner email in ChatGPT sheet (column D).
// Column D now contains owner email (not workspace name).
// Returns the count of non-empty rows with matching owner email.
func (r *Repository) CountWorkspaceSlots(ctx context.Context, ownerEmail string) (int, error) {
	// Read columns B and D from ChatGPT sheet
	// B = Nama (to check if slot is used), D = Owner Email
	readRange := "'ChatGPT'!B:D"
	resp, err := r.service.Spreadsheets.Values.Get(r.spreadsheetID, readRange).Do()
	if err != nil {
		return 0, fmt.Errorf("failed to read ChatGPT sheet: %w", err)
	}

	count := 0
	emailLower := strings.ToLower(strings.TrimSpace(ownerEmail))

	// Skip header rows (index 0 and 1 - title and header)
	for i, row := range resp.Values {
		if i < 2 {
			continue // Skip title and header rows
		}
		if len(row) < 3 {
			continue // Need at least columns B, C, D
		}

		// Column B = Nama (index 0 in this range)
		// Column D = Owner Email (index 2 in this range)
		nama := strings.TrimSpace(fmt.Sprintf("%v", row[0]))
		ownerEmailCell := strings.ToLower(strings.TrimSpace(fmt.Sprintf("%v", row[2])))

		// Count if owner email matches AND slot is filled (nama not empty)
		if ownerEmailCell == emailLower && nama != "" {
			count++
		}
	}

	return count, nil
}

// GetSlotAvailability returns slot availability for families/workspaces.
// Reads ALL rows from column D (except header) in the target sheet.
// product: "ChatGPT" or "Gemini"
// availableOnly: if true, only return items with available slots
func (r *Repository) GetSlotAvailability(ctx context.Context, product string, availableOnly bool) (*entity.SlotAvailabilityResult, error) {
	// Validate product
	if product != "ChatGPT" && product != "Gemini" {
		return nil, fmt.Errorf("product must be 'ChatGPT' or 'Gemini', got: %s", product)
	}

	// Define slot limits per product
	maxSlots := 5 // Gemini default
	if product == "ChatGPT" {
		maxSlots = 4
	}

	// Read ALL rows from column B and D (except header)
	// Column B = Nama, Column D = Family/Workspace
	readRange := fmt.Sprintf("'%s'!B:D", product)
	resp, err := r.service.Spreadsheets.Values.Get(r.spreadsheetID, readRange).Do()
	if err != nil {
		return nil, fmt.Errorf("failed to read slot area: %w", err)
	}

	// Group by family/workspace name
	slotCounts := make(map[string]*entity.SlotInfo)

	// Skip header rows (first 2 rows: title and column headers)
	for i, row := range resp.Values {
		if i < 2 {
			continue
		}

		// Column B (Nama) is index 0, Column D (Family) is index 2
		var nama, familyVal string
		if len(row) > 0 && row[0] != nil {
			nama = strings.TrimSpace(fmt.Sprintf("%v", row[0]))
		}
		if len(row) > 2 && row[2] != nil {
			familyVal = strings.TrimSpace(fmt.Sprintf("%v", row[2]))
		}

		// Skip empty family values
		if familyVal == "" {
			continue
		}

		// Initialize slot info if not exists
		if _, exists := slotCounts[familyVal]; !exists {
			slotCounts[familyVal] = &entity.SlotInfo{
				Name:       familyVal,
				Product:    product,
				TotalSlots: maxSlots,
			}
		}

		// Increment used slots if nama is filled
		if nama != "" {
			slotCounts[familyVal].UsedSlots++
		}
	}

	// Calculate available slots and build result
	result := &entity.SlotAvailabilityResult{
		Product:       product,
		AvailableOnly: availableOnly,
	}

	for _, info := range slotCounts {
		info.AvailableSlot = info.TotalSlots - info.UsedSlots
		if info.AvailableSlot < 0 {
			info.AvailableSlot = 0
		}

		// Filter based on availableOnly - skip ONLY if requested AND no available slots
		if availableOnly && info.AvailableSlot == 0 {
			continue
		}

		result.Slots = append(result.Slots, *info)
		result.TotalEntries++
	}

	return result, nil
}
