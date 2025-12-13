// Package sheets implements Google Sheets repository for data logging.
package sheets

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/exernia/botjanweb/pkg/logger"

	"github.com/exernia/botjanweb/internal/domain/entity"
	"google.golang.org/api/option"
	"google.golang.org/api/sheets/v4"
)

// Repository implements TransactionLogPort for Google Sheets.
type Repository struct {
	service       *sheets.Service
	spreadsheetID string
	logger        *log.Logger

	// Sheet names
	qrisSheet        string
	ordersSheet      string
	akunGoogleSheet  string
	akunChatGPTSheet string
}

// NewRepository creates a new Sheets repository.
// Accepts either credentialsPath (for local dev) or credentialsJSON (for cloud deployment like Heroku).
// If credentialsJSON is provided, it takes precedence over credentialsPath.
func NewRepository(spreadsheetID, credentialsPath, credentialsJSON, qrisSheet, ordersSheet, akunGoogleSheet, akunChatGPTSheet string) (*Repository, error) {
	ctx := context.Background()

	var srv *sheets.Service
	var err error

	// Prefer JSON credentials (for cloud deployment) over file path
	if credentialsJSON != "" {
		srv, err = sheets.NewService(ctx, option.WithCredentialsJSON([]byte(credentialsJSON)))
		if err != nil {
			return nil, fmt.Errorf("failed to create sheets service from JSON: %w", err)
		}
	} else {
		srv, err = sheets.NewService(ctx, option.WithCredentialsFile(credentialsPath))
		if err != nil {
			return nil, fmt.Errorf("failed to create sheets service from file: %w", err)
		}
	}

	return &Repository{
		service:          srv,
		spreadsheetID:    spreadsheetID,
		logger:           logger.QRIS,
		qrisSheet:        qrisSheet,
		ordersSheet:      ordersSheet,
		akunGoogleSheet:  akunGoogleSheet,
		akunChatGPTSheet: akunChatGPTSheet,
	}, nil
}

// LogOrder logs an order to the appropriate sheet based on Produk field.
// Uses batchUpdate to insert row within the table and populate data.
//
// Spreadsheet column mapping:
//
//	A = No (auto, skip)
//	B = Nama
//	C = Email
//	D = Family (ChatGPT/Gemini) OR Kode Redeem (Perplexity/YouTube)
//	E = Tanggal Pesanan
//	F = Tanggal Berakhir (skip)
//	G = Amount/Nominal
//	H = Kanal
//	I = Akun/Nomor/Username
func (r *Repository) LogOrder(ctx context.Context, order *entity.Order) error {
	// Determine target sheet from Produk field
	targetSheet := r.resolveSheetName(order.Produk)

	// Get sheet ID for batchUpdate
	sheetID, err := r.getSheetID(targetSheet)
	if err != nil {
		return fmt.Errorf("failed to get sheet ID for '%s': %w", targetSheet, err)
	}

	// Check if this is a redeem-based product (Perplexity/YouTube)
	isRedeemProduct := order.Produk == "Perplexity" || order.Produk == "YouTube"

	// Find insertion position based on Family field or redeem-based product
	var lastRow int64
	if isRedeemProduct || order.Family == "" {
		// Redeem-based products or no family: insert after last filled row
		lastRow, err = r.findLastFilledRow(targetSheet)
		if err != nil {
			return fmt.Errorf("failed to find last filled row: %w", err)
		}
	} else {
		// Has family: find available slot with matching Family value
		lastRow, err = r.findAvailableSlotForFamily(targetSheet, order.Family, order.Produk)
		if err != nil {
			return fmt.Errorf("failed to find available slot for family '%s': %w", order.Family, err)
		}
	}

	wib := time.FixedZone("WIB", 7*60*60)
	tanggal := order.TanggalPesanan.In(wib).Format("2006-01-02")

	// Calculate expiry date: 1 month from order date
	tanggalBerakhir := order.TanggalPesanan.In(wib).AddDate(0, 1, 0).Format("2006-01-02")

	// Determine column D value based on product type
	var columnDValue string
	if isRedeemProduct {
		columnDValue = order.KodeRedeem // Empty initially, filled by admin later
	} else {
		columnDValue = order.Family
	}

	// Instead of inserting a new row, update the existing empty row at lastRow.
	// This prevents pushing down existing data and avoids breaking array formulas.
	//
	// For rows WITH pre-filled Family (slots 108-151 in ChatGPT/Gemini):
	//   - Only update: B (Nama), C (Email), E (Tanggal), G (Amount), H (Kanal), I (Akun)
	//   - Skip D (Family) as it's already filled
	//
	// For rows WITHOUT Family or redeem-based products:
	//   - Update all columns including D (Family or Kode Redeem)
	//
	// Strategy: Update columns separately to avoid offset issues
	requests := []*sheets.Request{
		// Update B (Nama) and C (Email)
		{
			UpdateCells: &sheets.UpdateCellsRequest{
				Start: &sheets.GridCoordinate{
					SheetId:     sheetID,
					RowIndex:    lastRow,
					ColumnIndex: 1, // Column B (0-indexed)
				},
				Rows: []*sheets.RowData{
					{
						Values: []*sheets.CellData{
							// B: Nama
							{UserEnteredValue: &sheets.ExtendedValue{StringValue: &order.Nama}},
							// C: Email (Person Smart Chip)
							{
								UserEnteredValue: &sheets.ExtendedValue{StringValue: ptr("@")},
								ChipRuns: []*sheets.ChipRun{
									{
										StartIndex: 0,
										Chip: &sheets.Chip{
											PersonProperties: &sheets.PersonProperties{
												Email:         order.Email,
												DisplayFormat: "EMAIL",
											},
										},
									},
								},
							},
						},
					},
				},
				Fields: "userEnteredValue,chipRuns",
			},
		},
		// Update D (Family or Kode Redeem) - only for redeem products or non-family orders
		{
			UpdateCells: &sheets.UpdateCellsRequest{
				Start: &sheets.GridCoordinate{
					SheetId:     sheetID,
					RowIndex:    lastRow,
					ColumnIndex: 3, // Column D (0-indexed)
				},
				Rows: []*sheets.RowData{
					{
						Values: []*sheets.CellData{
							{UserEnteredValue: &sheets.ExtendedValue{StringValue: &columnDValue}},
						},
					},
				},
				Fields: "userEnteredValue",
			},
		},
		// Update E (Tanggal Pesanan) and F (Tanggal Berakhir)
		{
			UpdateCells: &sheets.UpdateCellsRequest{
				Start: &sheets.GridCoordinate{
					SheetId:     sheetID,
					RowIndex:    lastRow,
					ColumnIndex: 4, // Column E (0-indexed)
				},
				Rows: []*sheets.RowData{
					{
						Values: []*sheets.CellData{
							// E: Tanggal Pesanan
							{UserEnteredValue: &sheets.ExtendedValue{StringValue: &tanggal}},
							// F: Tanggal Berakhir (1 month from order date)
							{UserEnteredValue: &sheets.ExtendedValue{StringValue: &tanggalBerakhir}},
						},
					},
				},
				Fields: "userEnteredValue",
			},
		},
		// Update G (Amount), H (Kanal), I (Akun)
		{
			UpdateCells: &sheets.UpdateCellsRequest{
				Start: &sheets.GridCoordinate{
					SheetId:     sheetID,
					RowIndex:    lastRow,
					ColumnIndex: 6, // Column G (0-indexed)
				},
				Rows: []*sheets.RowData{
					{
						Values: []*sheets.CellData{
							// G: Amount
							{UserEnteredValue: &sheets.ExtendedValue{NumberValue: ptr64(float64(order.Amount))}},
							// H: Kanal
							{UserEnteredValue: &sheets.ExtendedValue{StringValue: &order.Kanal}},
							// I: Akun
							{UserEnteredValue: &sheets.ExtendedValue{StringValue: &order.Akun}},
						},
					},
				},
				Fields: "userEnteredValue",
			},
		},
	}

	_, err = r.service.Spreadsheets.BatchUpdate(r.spreadsheetID, &sheets.BatchUpdateSpreadsheetRequest{
		Requests: requests,
	}).Do()

	if err != nil {
		return fmt.Errorf("failed to log order to sheet '%s': %w", targetSheet, err)
	}

	r.logger.Printf("📊 Logged order to '%s' at row %d: %s (%s)", targetSheet, lastRow+1, order.Nama, order.Email)
	return nil
}

// getSheetID returns the numeric sheet ID for a given sheet name.
func (r *Repository) getSheetID(sheetName string) (int64, error) {
	resp, err := r.service.Spreadsheets.Get(r.spreadsheetID).Do()
	if err != nil {
		return 0, err
	}

	for _, sheet := range resp.Sheets {
		if sheet.Properties.Title == sheetName {
			return sheet.Properties.SheetId, nil
		}
	}

	return 0, fmt.Errorf("sheet '%s' not found", sheetName)
}

// findLastFilledRow finds the last row with data in column B (Nama).
// Returns 0-indexed row number where new data should be inserted.
// This is used for redeem-based products (Perplexity/YouTube) or orders without Family.
func (r *Repository) findLastFilledRow(sheetName string) (int64, error) {
	// Read column B (Nama) to find last non-empty cell
	readRange := fmt.Sprintf("'%s'!B:B", sheetName)
	resp, err := r.service.Spreadsheets.Values.Get(r.spreadsheetID, readRange).Do()
	if err != nil {
		return 0, err
	}

	// Find last non-empty row in column B
	// Row 1 is title, Row 2 is header, data starts from row 3 (index 2)
	lastFilledRow := int64(1) // Default to row 2 if no data
	for i, row := range resp.Values {
		if len(row) > 0 && row[0] != nil && fmt.Sprintf("%v", row[0]) != "" {
			lastFilledRow = int64(i)
		}
	}

	// Return the row after last filled row (0-indexed)
	return lastFilledRow + 1, nil
}

// findLastFamilyRow finds the last row with data in column D (Family).
// Returns 0-indexed row number where new data should be inserted.
// This is used for orders without Family field - they should be appended
// after all rows that have Family values (including reserved slots).
func (r *Repository) findLastFamilyRow(sheetName string) (int64, error) {
	// Read column D to find last non-empty cell
	readRange := fmt.Sprintf("'%s'!D:D", sheetName)
	resp, err := r.service.Spreadsheets.Values.Get(r.spreadsheetID, readRange).Do()
	if err != nil {
		return 0, err
	}

	// Find last non-empty row in column D
	// Row 1 is title, Row 2 is header, data starts from row 3 (index 2)
	lastFamilyRow := int64(1) // Default to row 2 if no data
	for i, row := range resp.Values {
		if len(row) > 0 && row[0] != nil && fmt.Sprintf("%v", row[0]) != "" {
			lastFamilyRow = int64(i)
		}
	}

	// Return the row after last Family row (0-indexed)
	// If lastFamilyRow is 151 (row 152 in 1-indexed), we insert at 151 (0-indexed)
	return lastFamilyRow + 1, nil
}

// findAvailableSlotForFamily finds an available slot with matching Family value.
// Reserved slot area is rows 108-151 (0-indexed: 107-150).
// Returns 0-indexed row number where new data should be inserted.
// Validates slot count against product-specific limits:
// - Gemini: max 5 slots per family
// - ChatGPT: max 4 slots per workspace
// - YouTube/Perplexity: no limit specified yet (default 10)
func (r *Repository) findAvailableSlotForFamily(sheetName, family, produk string) (int64, error) {
	// Define slot limits per product
	slotLimits := map[string]int{
		"Gemini":     5,
		"ChatGPT":    4,
		"YouTube":    10, // Default until specified
		"Perplexity": 10, // Default until specified
	}

	maxSlots := slotLimits[produk]
	if maxSlots == 0 {
		maxSlots = 10 // Fallback default
	}

	// Read columns B (Nama) and D (Family) from reserved area (rows 108-151)
	// In 0-indexed: 107-150, but in A1 notation: rows 108-151
	readRange := fmt.Sprintf("'%s'!B108:D151", sheetName)
	resp, err := r.service.Spreadsheets.Values.Get(r.spreadsheetID, readRange).Do()
	if err != nil {
		return 0, fmt.Errorf("failed to read reserved slot area: %w", err)
	}

	if len(resp.Values) == 0 {
		return 0, fmt.Errorf("no reserved slots found in rows 108-151")
	}

	// Scan for matching family and count slots
	var availableSlot int64 = -1
	filledSlots := 0

	for i, row := range resp.Values {
		// Column B (Nama) is index 0, Column D (Family) is index 2
		var nama, familyVal string
		if len(row) > 0 && row[0] != nil {
			nama = strings.TrimSpace(fmt.Sprintf("%v", row[0]))
		}
		if len(row) > 2 && row[2] != nil {
			familyVal = strings.TrimSpace(fmt.Sprintf("%v", row[2]))
		}

		// Check if this row belongs to the target family
		if familyVal == family {
			if nama == "" {
				// Found empty slot with matching family
				if availableSlot == -1 {
					availableSlot = int64(107 + i) // 0-indexed: 107 = row 108
				}
			} else {
				// Slot is filled, increment counter
				filledSlots++
			}
		}
	}

	// Check if family was found at all
	if availableSlot == -1 && filledSlots == 0 {
		return 0, fmt.Errorf("family '%s' not found in reserved slot area", family)
	}

	// Check slot limit
	if filledSlots >= maxSlots {
		return 0, fmt.Errorf("slot limit reached for family '%s' (max: %d slots)", family, maxSlots)
	}

	// Check if we found an available slot
	if availableSlot == -1 {
		return 0, fmt.Errorf("no available slots for family '%s' (all slots filled)", family)
	}

	return availableSlot, nil
}

// ptr returns a pointer to the given string.
func ptr(s string) *string {
	return &s
}

// ptr64 returns a pointer to the given float64.
func ptr64(f float64) *float64 {
	return &f
}

// resolveSheetName maps Produk value to actual sheet name using Product enum.
func (r *Repository) resolveSheetName(produk string) string {
	// Parse product and get its sheet name
	product, err := entity.ParseProduct(produk)
	if err != nil {
		// Fallback to produk value if not a valid product
		return produk
	}
	return product.SheetName()
}

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

// ValidateWorkspace checks if a workspace name exists in Akun ChatGPT sheet (column C).
// Returns true if found, false otherwise.
func (r *Repository) ValidateWorkspace(ctx context.Context, workspaceName string) (bool, error) {
	// Read column C (workspace) from Akun ChatGPT sheet
	readRange := "'Akun ChatGPT'!C:C"
	resp, err := r.service.Spreadsheets.Values.Get(r.spreadsheetID, readRange).Do()
	if err != nil {
		return false, fmt.Errorf("failed to read Akun ChatGPT sheet: %w", err)
	}

	// Search for the workspace name (case-insensitive)
	workspaceLower := strings.ToLower(strings.TrimSpace(workspaceName))
	for _, row := range resp.Values {
		if len(row) > 0 {
			cellValue := strings.ToLower(strings.TrimSpace(fmt.Sprintf("%v", row[0])))
			if cellValue == workspaceLower {
				return true, nil
			}
		}
	}

	return false, nil
}

// CountWorkspaceSlots counts how many slots are used for a workspace in ChatGPT sheet (column D).
// Returns the count of non-empty rows with matching workspace value.
func (r *Repository) CountWorkspaceSlots(ctx context.Context, workspace string) (int, error) {
	// Read columns B and D from ChatGPT sheet
	// B = Nama (to check if slot is used), D = WorkSpace
	readRange := "'ChatGPT'!B:D"
	resp, err := r.service.Spreadsheets.Values.Get(r.spreadsheetID, readRange).Do()
	if err != nil {
		return 0, fmt.Errorf("failed to read ChatGPT sheet: %w", err)
	}

	count := 0
	workspaceLower := strings.ToLower(strings.TrimSpace(workspace))

	// Skip header rows (index 0 and 1 - title and header)
	for i, row := range resp.Values {
		if i < 2 {
			continue // Skip title and header rows
		}
		if len(row) < 3 {
			continue // Need at least columns B, C, D
		}

		// Column B = Nama (index 0 in this range)
		// Column D = WorkSpace (index 2 in this range)
		nama := strings.TrimSpace(fmt.Sprintf("%v", row[0]))
		workspaceCell := strings.ToLower(strings.TrimSpace(fmt.Sprintf("%v", row[2])))

		// Count if workspace matches AND slot is filled (nama not empty)
		if workspaceCell == workspaceLower && nama != "" {
			count++
		}
	}

	return count, nil
}

// AddAkunGoogle adds a new Google account to Akun Google sheet using InsertDimension.
// Columns: A=Email, B=Sandi, C=Tanggal Aktivasi, D=Tanggal Berakhir, E=Status Dibuat, F=YT Premium
func (r *Repository) AddAkunGoogle(ctx context.Context, akun *entity.AkunGoogle) error {
	sheetName := r.akunGoogleSheet

	// Get sheet ID
	sheetID, err := r.getSheetID(sheetName)
	if err != nil {
		return fmt.Errorf("failed to get sheet ID for '%s': %w", sheetName, err)
	}

	// Find last row in table (detect by checking column A for data)
	lastRow, err := r.findLastTableRowByColumn(sheetName, "A")
	if err != nil {
		return fmt.Errorf("failed to find last table row: %w", err)
	}

	wib := time.FixedZone("WIB", 7*60*60)
	tanggalAktivasi := akun.TanggalAktivasi.In(wib).Format("2006-01-02")
	tanggalBerakhir := akun.TanggalBerakhir // Already formatted as YYYY-MM-DD from usecase

	// Insert row and update cells
	requests := []*sheets.Request{
		// Insert 1 row at lastRow position
		{
			InsertDimension: &sheets.InsertDimensionRequest{
				Range: &sheets.DimensionRange{
					SheetId:    sheetID,
					Dimension:  "ROWS",
					StartIndex: lastRow,
					EndIndex:   lastRow + 1,
				},
				InheritFromBefore: true,
			},
		},
		// Update cells A through F
		{
			UpdateCells: &sheets.UpdateCellsRequest{
				Start: &sheets.GridCoordinate{
					SheetId:     sheetID,
					RowIndex:    lastRow,
					ColumnIndex: 0, // Column A
				},
				Rows: []*sheets.RowData{
					{
						Values: []*sheets.CellData{
							// A: Email (Person Smart Chip)
							{
								UserEnteredValue: &sheets.ExtendedValue{StringValue: ptr("@")},
								ChipRuns: []*sheets.ChipRun{
									{
										StartIndex: 0,
										Chip: &sheets.Chip{
											PersonProperties: &sheets.PersonProperties{
												Email:         akun.Email,
												DisplayFormat: "EMAIL",
											},
										},
									},
								},
							},
							// B: Sandi
							{UserEnteredValue: &sheets.ExtendedValue{StringValue: &akun.Sandi}},
							// C: Tanggal Aktivasi
							{UserEnteredValue: &sheets.ExtendedValue{StringValue: &tanggalAktivasi}},
							// D: Tanggal Berakhir (1 year from activation)
							{UserEnteredValue: &sheets.ExtendedValue{StringValue: &tanggalBerakhir}},
							// E: Status Dibuat (empty)
							{UserEnteredValue: &sheets.ExtendedValue{StringValue: ptr("")}},
							// F: YT Premium? (empty)
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
		return fmt.Errorf("failed to add Akun Google: %w", err)
	}

	r.logger.Printf("📊 Added Akun Google at row %d: %s", lastRow+1, akun.Email)
	return nil
}

// AddAkunChatGPT adds a new ChatGPT account to Akun ChatGPT sheet.
// Note: Akun ChatGPT doesn't have a table, so we use Append.
// Columns: A=Email, B=Sandi, C=WorkSpace, D=Status, E=Tanggal Aktivasi, F=Tanggal kena ban
func (r *Repository) AddAkunChatGPT(ctx context.Context, akun *entity.AkunChatGPT) error {
	wib := time.FixedZone("WIB", 7*60*60)
	tanggal := akun.TanggalAktivasi.In(wib).Format("2006-01-02")

	values := [][]interface{}{
		{
			akun.Email,          // A: Email
			akun.Sandi,          // B: Sandi
			akun.Workspace,      // C: WorkSpace
			akun.Status,         // D: Status (empty)
			tanggal,             // E: Tanggal Aktivasi
			akun.TanggalKenaBan, // F: Tanggal kena ban (empty)
		},
	}

	valueRange := &sheets.ValueRange{Values: values}
	appendRange := "'Akun ChatGPT'!A:F"

	_, err := r.service.Spreadsheets.Values.Append(
		r.spreadsheetID,
		appendRange,
		valueRange,
	).ValueInputOption("USER_ENTERED").Do()

	if err != nil {
		return fmt.Errorf("failed to add Akun ChatGPT: %w", err)
	}

	r.logger.Printf("📊 Added Akun ChatGPT: %s (%s)", akun.Email, akun.Workspace)
	return nil
}

// findLastTableRowByColumn finds the last row with data in the specified column.
// Returns 0-indexed row number where new data should be inserted.
func (r *Repository) findLastTableRowByColumn(sheetName string, column string) (int64, error) {
	readRange := fmt.Sprintf("'%s'!%s:%s", sheetName, column, column)
	resp, err := r.service.Spreadsheets.Values.Get(r.spreadsheetID, readRange).Do()
	if err != nil {
		return 0, err
	}

	// Find last non-empty row
	lastDataRow := int64(0)
	for i, row := range resp.Values {
		if len(row) > 0 && row[0] != "" {
			lastDataRow = int64(i)
		}
	}

	return lastDataRow + 1, nil
}

// GetAkunGoogleList fetches all Google accounts from Akun Google sheet.
// Columns: A=Email, B=Sandi, C=Tanggal Aktivasi, D=Tanggal Berakhir, E=Status Dibuat, F=YT Premium?, G=Keterangan
func (r *Repository) GetAkunGoogleList(ctx context.Context) ([]entity.AkunGoogle, error) {
	readRange := "'Akun Google'!A:G"
	resp, err := r.service.Spreadsheets.Values.Get(r.spreadsheetID, readRange).Do()
	if err != nil {
		return nil, fmt.Errorf("failed to read Akun Google sheet: %w", err)
	}

	var accounts []entity.AkunGoogle

	// Skip header row (index 0)
	for i, row := range resp.Values {
		if i == 0 {
			continue
		}
		if len(row) < 1 || row[0] == "" {
			continue // Skip empty rows
		}

		akun := entity.AkunGoogle{
			Email: strings.TrimSpace(fmt.Sprintf("%v", row[0])),
		}

		if len(row) > 1 {
			akun.Sandi = strings.TrimSpace(fmt.Sprintf("%v", row[1]))
		}
		if len(row) > 2 {
			// Parse Tanggal Aktivasi
			if tgl, err := time.Parse("2006-01-02", fmt.Sprintf("%v", row[2])); err == nil {
				akun.TanggalAktivasi = tgl
			}
		}
		if len(row) > 3 {
			akun.TanggalBerakhir = strings.TrimSpace(fmt.Sprintf("%v", row[3]))
		}
		if len(row) > 4 {
			akun.StatusDibuat = strings.TrimSpace(fmt.Sprintf("%v", row[4]))
		}
		if len(row) > 5 {
			akun.YTPremium = strings.TrimSpace(fmt.Sprintf("%v", row[5]))
		}
		if len(row) > 6 {
			akun.Keterangan = strings.TrimSpace(fmt.Sprintf("%v", row[6]))
		}

		accounts = append(accounts, akun)
	}

	return accounts, nil
}

// GetAkunChatGPTList fetches all ChatGPT accounts from Akun ChatGPT sheet.
// Columns: A=Email, B=Sandi, C=WorkSpace, D=Status, E=Tanggal Aktivasi, F=Tanggal kena ban
func (r *Repository) GetAkunChatGPTList(ctx context.Context) ([]entity.AkunChatGPT, error) {
	readRange := "'Akun ChatGPT'!A:F"
	resp, err := r.service.Spreadsheets.Values.Get(r.spreadsheetID, readRange).Do()
	if err != nil {
		return nil, fmt.Errorf("failed to read Akun ChatGPT sheet: %w", err)
	}

	var accounts []entity.AkunChatGPT

	// Skip header row (index 0)
	for i, row := range resp.Values {
		if i == 0 {
			continue
		}
		if len(row) < 1 || row[0] == "" {
			continue // Skip empty rows
		}

		akun := entity.AkunChatGPT{
			Email: strings.TrimSpace(fmt.Sprintf("%v", row[0])),
		}

		if len(row) > 1 {
			akun.Sandi = strings.TrimSpace(fmt.Sprintf("%v", row[1]))
		}
		if len(row) > 2 {
			akun.Workspace = strings.TrimSpace(fmt.Sprintf("%v", row[2]))
		}
		if len(row) > 3 {
			akun.Status = strings.TrimSpace(fmt.Sprintf("%v", row[3]))
		}
		if len(row) > 4 {
			// Parse Tanggal Aktivasi
			if tgl, err := time.Parse("2006-01-02", fmt.Sprintf("%v", row[4])); err == nil {
				akun.TanggalAktivasi = tgl
			}
		}
		if len(row) > 5 {
			akun.TanggalKenaBan = strings.TrimSpace(fmt.Sprintf("%v", row[5]))
		}

		accounts = append(accounts, akun)
	}

	return accounts, nil
}

// GetAccountListResult fetches all accounts and returns a summary with availability counts.
func (r *Repository) GetAccountListResult(ctx context.Context) (*entity.AccountListResult, error) {
	googleAccounts, err := r.GetAkunGoogleList(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get Google accounts: %w", err)
	}

	chatgptAccounts, err := r.GetAkunChatGPTList(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get ChatGPT accounts: %w", err)
	}

	result := &entity.AccountListResult{
		GoogleAccounts:  googleAccounts,
		ChatGPTAccounts: chatgptAccounts,
		TotalGoogle:     len(googleAccounts),
		TotalChatGPT:    len(chatgptAccounts),
	}

	// Count available accounts
	for i := range googleAccounts {
		if googleAccounts[i].IsAvailable() {
			result.AvailableGoogle++
		}
	}
	for i := range chatgptAccounts {
		if chatgptAccounts[i].IsAvailable() {
			result.AvailableChatGPT++
		}
	}

	return result, nil
}

// GetSlotAvailability returns slot availability for families/workspaces.
// Reads from reserved area rows 108-151 in the target sheet.
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

	// Read from reserved area rows 108-151
	// Column B = Nama, Column D = Family/Workspace
	readRange := fmt.Sprintf("'%s'!B108:D151", product)
	resp, err := r.service.Spreadsheets.Values.Get(r.spreadsheetID, readRange).Do()
	if err != nil {
		return nil, fmt.Errorf("failed to read slot area: %w", err)
	}

	// Group by family/workspace name
	slotCounts := make(map[string]*entity.SlotInfo)

	for _, row := range resp.Values {
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

		// Increment total slots (each row is a slot)
		// If nama is not empty, it's used
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

		// Filter based on availableOnly
		if availableOnly && info.AvailableSlot == 0 {
			continue
		}

		result.Slots = append(result.Slots, *info)
		result.TotalEntries++
	}

	return result, nil
}

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
