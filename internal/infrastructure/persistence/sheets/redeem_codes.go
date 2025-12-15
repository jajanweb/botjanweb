package sheets

import (
	"context"
	"fmt"
	"strings"

	"github.com/exernia/botjanweb/internal/domain/entity"
	"google.golang.org/api/sheets/v4"
)

// GetRedeemCodeAvailability returns available Perplexity redeem codes.
// Reads from "Kode Perplexity" sheet.
// Columns: A=No, B=Email, C=Kode redeem, D=Tanggal aktivasi, E=Tanggal berakhir
// availableOnly: if true, only return codes where D (Tanggal aktivasi) is empty
func (r *Repository) GetRedeemCodeAvailability(ctx context.Context, availableOnly bool) (*entity.RedeemCodeResult, error) {
	// Read from Kode Perplexity sheet (skip header row 1)
	readRange := "'Kode Perplexity'!A2:E"
	resp, err := r.service.Spreadsheets.Values.Get(r.spreadsheetID, readRange).Do()
	if err != nil {
		return nil, fmt.Errorf("failed to read Kode Perplexity sheet: %w", err)
	}

	result := &entity.RedeemCodeResult{
		AvailableOnly: availableOnly,
	}

	for i, row := range resp.Values {
		// Skip header row (already skipped in range)
		if len(row) < 1 {
			continue
		}

		code := entity.RedeemCodeInfo{
			No: i + 2, // Row number (1-indexed, plus header)
		}

		// Parse columns
		if len(row) > 0 && row[0] != nil {
			// Column A: No (usually just row number, but use index)
			code.No = i + 2
		}
		if len(row) > 1 && row[1] != nil {
			code.Email = strings.TrimSpace(fmt.Sprintf("%v", row[1]))
		}
		if len(row) > 2 && row[2] != nil {
			code.KodeRedeem = strings.TrimSpace(fmt.Sprintf("%v", row[2]))
		}
		if len(row) > 3 && row[3] != nil {
			code.TanggalAktivasi = strings.TrimSpace(fmt.Sprintf("%v", row[3]))
		}
		if len(row) > 4 && row[4] != nil {
			code.TanggalBerakhir = strings.TrimSpace(fmt.Sprintf("%v", row[4]))
		}

		// Skip empty rows (no email or kode)
		if code.Email == "" && code.KodeRedeem == "" {
			continue
		}

		result.TotalCodes++

		// Check if available (Tanggal aktivasi is empty)
		isAvailable := code.TanggalAktivasi == ""
		if isAvailable {
			result.AvailableCodes++
		}

		// Filter based on availableOnly
		if availableOnly && !isAvailable {
			continue
		}

		result.Codes = append(result.Codes, code)
	}

	return result, nil
}

// AddRedeemCode adds a new redeem code to Kode Perplexity sheet.
// Inserts a new row with Email and Kode redeem.
// Columns: A=No (auto), B=Email, C=Kode redeem, D=Tanggal aktivasi (empty), E=Tanggal berakhir (empty)
func (r *Repository) AddRedeemCode(ctx context.Context, email, kodeRedeem string) error {
	sheetName := "Kode Perplexity"

	// Get sheet ID
	sheetID, err := r.getSheetID(sheetName)
	if err != nil {
		return fmt.Errorf("failed to get sheet ID for '%s': %w", sheetName, err)
	}

	// Find last row with data in column B (Email)
	readRange := fmt.Sprintf("'%s'!B:B", sheetName)
	resp, err := r.service.Spreadsheets.Values.Get(r.spreadsheetID, readRange).Do()
	if err != nil {
		return fmt.Errorf("failed to read %s: %w", sheetName, err)
	}

	// Find last non-empty row (0-indexed)
	lastRow := int64(0)
	for i, row := range resp.Values {
		if len(row) > 0 && row[0] != nil && fmt.Sprintf("%v", row[0]) != "" {
			lastRow = int64(i)
		}
	}
	// Insert after last row
	insertRow := lastRow + 1

	// Calculate the No value (row number - 1 for header)
	noValue := float64(insertRow) // No is just the row number minus header

	// Insert row and update cells
	requests := []*sheets.Request{
		// Insert 1 row at insertRow position
		{
			InsertDimension: &sheets.InsertDimensionRequest{
				Range: &sheets.DimensionRange{
					SheetId:    sheetID,
					Dimension:  "ROWS",
					StartIndex: insertRow,
					EndIndex:   insertRow + 1,
				},
				InheritFromBefore: true,
			},
		},
		// Update cells A through E
		{
			UpdateCells: &sheets.UpdateCellsRequest{
				Start: &sheets.GridCoordinate{
					SheetId:     sheetID,
					RowIndex:    insertRow,
					ColumnIndex: 0, // Column A
				},
				Rows: []*sheets.RowData{
					{
						Values: []*sheets.CellData{
							// A: No
							{UserEnteredValue: &sheets.ExtendedValue{NumberValue: &noValue}},
							// B: Email (Person Smart Chip)
							{
								UserEnteredValue: &sheets.ExtendedValue{StringValue: ptr("@")},
								ChipRuns: []*sheets.ChipRun{
									{
										StartIndex: 0,
										Chip: &sheets.Chip{
											PersonProperties: &sheets.PersonProperties{
												Email:         email,
												DisplayFormat: "EMAIL",
											},
										},
									},
								},
							},
							// C: Kode redeem
							{UserEnteredValue: &sheets.ExtendedValue{StringValue: &kodeRedeem}},
							// D: Tanggal aktivasi (empty)
							{UserEnteredValue: &sheets.ExtendedValue{StringValue: ptr("")}},
							// E: Tanggal berakhir (empty)
							{UserEnteredValue: &sheets.ExtendedValue{StringValue: ptr("")}},
						},
					},
				},
				Fields: "userEnteredValue,chipRuns",
			},
		},
	}

	_, err = r.service.Spreadsheets.BatchUpdate(r.spreadsheetID, &sheets.BatchUpdateSpreadsheetRequest{
		Requests: requests,
	}).Do()

	if err != nil {
		return fmt.Errorf("failed to add redeem code: %w", err)
	}

	r.logger.Printf("📊 Added redeem code at row %d: %s (%s)", insertRow+1, email, kodeRedeem)
	return nil
}
