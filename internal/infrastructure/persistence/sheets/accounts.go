package sheets

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/exernia/botjanweb/internal/domain/entity"
	"google.golang.org/api/sheets/v4"
)

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
							// A: Email (text biasa)
							{UserEnteredValue: &sheets.ExtendedValue{StringValue: &akun.Email}},
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
				Fields: "userEnteredValue",
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
