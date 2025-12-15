package sheets

import (
	"context"
	"fmt"
	"time"

	"github.com/exernia/botjanweb/internal/domain/entity"
	"google.golang.org/api/sheets/v4"
)

// LogOrder logs an order to the appropriate sheet based on Produk field.
// Always inserts at the last row of the table (below existing data).
//
// Spreadsheet column mapping (ChatGPT):
//
//	A = No (auto, skip)
//	B = Nama
//	C = Email
//	D = WorkSpace name (ChatGPT/Gemini) OR Kode Redeem (Perplexity/YouTube)
//	E = Paket (duration: "20 Hari" or "30 Hari" for ChatGPT)
//	F = Tanggal Pesanan
//	G = Tanggal Berakhir (calculated: +1 month)
//	H = Nominal
//	I = Kanal
//	J = Akun/Bukti Transaksi
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

	// Always insert at the last row of the table (column D determines table boundary)
	lastRow, err := r.findLastFamilyRow(targetSheet)
	if err != nil {
		return fmt.Errorf("failed to find last table row: %w", err)
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

	// Update the row at lastRow position.
	// Mapping: B=Nama, C=Email, D=WorkSpace/Redeem, E=Paket, F=TglPesanan, G=TglBerakhir, H=Nominal, I=Kanal, J=Akun
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
		// Update D (WorkSpace/Family or Kode Redeem)
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
		// Update E (Paket), F (Tanggal Pesanan), G (Tanggal Berakhir)
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
							// E: Paket (duration)
							{UserEnteredValue: &sheets.ExtendedValue{StringValue: &order.Paket}},
							// F: Tanggal Pesanan
							{UserEnteredValue: &sheets.ExtendedValue{StringValue: &tanggal}},
							// G: Tanggal Berakhir (1 month from order date)
							{UserEnteredValue: &sheets.ExtendedValue{StringValue: &tanggalBerakhir}},
						},
					},
				},
				Fields: "userEnteredValue",
			},
		},
		// Update H (Nominal), I (Kanal), J (Akun/Bukti)
		{
			UpdateCells: &sheets.UpdateCellsRequest{
				Start: &sheets.GridCoordinate{
					SheetId:     sheetID,
					RowIndex:    lastRow,
					ColumnIndex: 7, // Column H (0-indexed)
				},
				Rows: []*sheets.RowData{
					{
						Values: []*sheets.CellData{
							// H: Nominal (Amount)
							{UserEnteredValue: &sheets.ExtendedValue{NumberValue: ptr64(float64(order.Amount))}},
							// I: Kanal
							{UserEnteredValue: &sheets.ExtendedValue{StringValue: &order.Kanal}},
							// J: Akun/Bukti Transaksi
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

// findLastFamilyRow finds the last row with data in column D (Family/Workspace).
// Returns 0-indexed row number where new data should be inserted.
// All new orders are appended at the last row of the table.
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
	return lastFamilyRow + 1, nil
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
