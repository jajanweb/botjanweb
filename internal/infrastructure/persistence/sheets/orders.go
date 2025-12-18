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
// Spreadsheet column mappings (verified from actual sheets):
//
// Gemini:      A=No, B=Nama, C=Email, D=Family, E=TglPesanan, F=TglBerakhir, G=Nominal, H=Kanal, I=Akun/Nomor
// ChatGPT:     A=No, B=Nama, C=Email, D=WorkSpace, E=Paket, F=TglPesanan, G=TglBerakhir, H=Nominal, I=Kanal, J=Bukti
// YouTube:     A=No, B=Nama, C=Email, D=Email Head, E=TglPesan, F=TglBerakhir, G=Status, H=Nominal, I=Kanal
// Perplexity:  A=No, B=Nama, C=Email, D=Kode Redeem, E=TglPesanan, F=TglBerakhir, G=Nominal, H=Kanal, I=Nomor/Username
func (r *Repository) LogOrder(ctx context.Context, order *entity.Order) error {
	// Determine target sheet from Produk field
	targetSheet := r.resolveSheetName(order.Produk)

	// Get sheet ID for batchUpdate
	sheetID, err := r.getSheetID(targetSheet)
	if err != nil {
		return fmt.Errorf("failed to get sheet ID for '%s': %w", targetSheet, err)
	}

	// Always insert at the last row of the table
	// Use column B (Nama) which always has data
	lastRow, err := r.findLastTableRowByColumn(targetSheet, "B")
	if err != nil {
		return fmt.Errorf("failed to find last table row: %w", err)
	}

	wib := time.FixedZone("WIB", 7*60*60)
	tanggal := order.TanggalPesanan.In(wib).Format("2006-01-02")

	// Calculate expiry date: 1 month from order date
	tanggalBerakhir := order.TanggalPesanan.In(wib).AddDate(0, 1, 0).Format("2006-01-02")

	// Build requests based on product type (different column structures)
	var requests []*sheets.Request

	// INSERT 1 ROW at lastRow position (common for all products)
	requests = append(requests, &sheets.Request{
		InsertDimension: &sheets.InsertDimensionRequest{
			Range: &sheets.DimensionRange{
				SheetId:    sheetID,
				Dimension:  "ROWS",
				StartIndex: lastRow,
				EndIndex:   lastRow + 1,
			},
			InheritFromBefore: true,
		},
	})

	// Update B (Nama) and C (Email) - common for all products
	requests = append(requests, &sheets.Request{
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
						// C: Email
						{UserEnteredValue: &sheets.ExtendedValue{StringValue: &order.Email}},
					},
				},
			},
			Fields: "userEnteredValue",
		},
	})

	// Product-specific column mappings
	switch order.Produk {
	case "Gemini":
		// D=Family, E=TglPesanan, F=TglBerakhir, G=Nominal, H=Kanal, I=Akun/Nomor
		requests = append(requests,
			// Update D (Family)
			&sheets.Request{
				UpdateCells: &sheets.UpdateCellsRequest{
					Start: &sheets.GridCoordinate{
						SheetId:     sheetID,
						RowIndex:    lastRow,
						ColumnIndex: 3, // Column D
					},
					Rows: []*sheets.RowData{
						{
							Values: []*sheets.CellData{
								{UserEnteredValue: &sheets.ExtendedValue{StringValue: &order.Family}},
							},
						},
					},
					Fields: "userEnteredValue",
				},
			},
			// Update E (Tanggal Pesanan), F (Tanggal Berakhir)
			&sheets.Request{
				UpdateCells: &sheets.UpdateCellsRequest{
					Start: &sheets.GridCoordinate{
						SheetId:     sheetID,
						RowIndex:    lastRow,
						ColumnIndex: 4, // Column E
					},
					Rows: []*sheets.RowData{
						{
							Values: []*sheets.CellData{
								{UserEnteredValue: &sheets.ExtendedValue{StringValue: &tanggal}},
								{UserEnteredValue: &sheets.ExtendedValue{StringValue: &tanggalBerakhir}},
							},
						},
					},
					Fields: "userEnteredValue",
				},
			},
			// Update G (Nominal), H (Kanal), I (Akun/Nomor)
			&sheets.Request{
				UpdateCells: &sheets.UpdateCellsRequest{
					Start: &sheets.GridCoordinate{
						SheetId:     sheetID,
						RowIndex:    lastRow,
						ColumnIndex: 6, // Column G
					},
					Rows: []*sheets.RowData{
						{
							Values: []*sheets.CellData{
								{UserEnteredValue: &sheets.ExtendedValue{NumberValue: ptr64(float64(order.Amount))}},
								{UserEnteredValue: &sheets.ExtendedValue{StringValue: &order.Kanal}},
								{UserEnteredValue: &sheets.ExtendedValue{StringValue: &order.Akun}},
							},
						},
					},
					Fields: "userEnteredValue",
				},
			},
		)

	case "ChatGPT":
		// D=WorkSpace, E=Paket, F=TglPesanan, G=TglBerakhir, H=Nominal, I=Kanal, J=Bukti
		requests = append(requests,
			// Update D (WorkSpace)
			&sheets.Request{
				UpdateCells: &sheets.UpdateCellsRequest{
					Start: &sheets.GridCoordinate{
						SheetId:     sheetID,
						RowIndex:    lastRow,
						ColumnIndex: 3, // Column D
					},
					Rows: []*sheets.RowData{
						{
							Values: []*sheets.CellData{
								{UserEnteredValue: &sheets.ExtendedValue{StringValue: &order.Family}},
							},
						},
					},
					Fields: "userEnteredValue",
				},
			},
			// Update E (Paket), F (Tanggal Pesanan), G (Tanggal Berakhir)
			&sheets.Request{
				UpdateCells: &sheets.UpdateCellsRequest{
					Start: &sheets.GridCoordinate{
						SheetId:     sheetID,
						RowIndex:    lastRow,
						ColumnIndex: 4, // Column E
					},
					Rows: []*sheets.RowData{
						{
							Values: []*sheets.CellData{
								{UserEnteredValue: &sheets.ExtendedValue{StringValue: &order.Paket}},
								{UserEnteredValue: &sheets.ExtendedValue{StringValue: &tanggal}},
								{UserEnteredValue: &sheets.ExtendedValue{StringValue: &tanggalBerakhir}},
							},
						},
					},
					Fields: "userEnteredValue",
				},
			},
			// Update H (Nominal), I (Kanal), J (Bukti Transaksi)
			&sheets.Request{
				UpdateCells: &sheets.UpdateCellsRequest{
					Start: &sheets.GridCoordinate{
						SheetId:     sheetID,
						RowIndex:    lastRow,
						ColumnIndex: 7, // Column H
					},
					Rows: []*sheets.RowData{
						{
							Values: []*sheets.CellData{
								{UserEnteredValue: &sheets.ExtendedValue{NumberValue: ptr64(float64(order.Amount))}},
								{UserEnteredValue: &sheets.ExtendedValue{StringValue: &order.Kanal}},
								{UserEnteredValue: &sheets.ExtendedValue{StringValue: &order.Akun}},
							},
						},
					},
					Fields: "userEnteredValue",
				},
			},
		)

	case "YouTube":
		// D=Email Head, E=TglPesan, F=TglBerakhir, G=Status, H=Nominal, I=Kanal
		// Note: No "Paket" column for YouTube
		emailHead := order.Family // Use Family field as Email Head for YouTube
		if emailHead == "" {
			emailHead = order.Email // Fallback to customer email
		}
		status := "Aktif" // Default status

		requests = append(requests,
			// Update D (Email Head)
			&sheets.Request{
				UpdateCells: &sheets.UpdateCellsRequest{
					Start: &sheets.GridCoordinate{
						SheetId:     sheetID,
						RowIndex:    lastRow,
						ColumnIndex: 3, // Column D
					},
					Rows: []*sheets.RowData{
						{
							Values: []*sheets.CellData{
								{UserEnteredValue: &sheets.ExtendedValue{StringValue: &emailHead}},
							},
						},
					},
					Fields: "userEnteredValue",
				},
			},
			// Update E (Tanggal Pesan), F (Tanggal Berakhir), G (Status)
			&sheets.Request{
				UpdateCells: &sheets.UpdateCellsRequest{
					Start: &sheets.GridCoordinate{
						SheetId:     sheetID,
						RowIndex:    lastRow,
						ColumnIndex: 4, // Column E
					},
					Rows: []*sheets.RowData{
						{
							Values: []*sheets.CellData{
								{UserEnteredValue: &sheets.ExtendedValue{StringValue: &tanggal}},
								{UserEnteredValue: &sheets.ExtendedValue{StringValue: &tanggalBerakhir}},
								{UserEnteredValue: &sheets.ExtendedValue{StringValue: &status}},
							},
						},
					},
					Fields: "userEnteredValue",
				},
			},
			// Update H (Nominal), I (Kanal)
			&sheets.Request{
				UpdateCells: &sheets.UpdateCellsRequest{
					Start: &sheets.GridCoordinate{
						SheetId:     sheetID,
						RowIndex:    lastRow,
						ColumnIndex: 7, // Column H
					},
					Rows: []*sheets.RowData{
						{
							Values: []*sheets.CellData{
								{UserEnteredValue: &sheets.ExtendedValue{NumberValue: ptr64(float64(order.Amount))}},
								{UserEnteredValue: &sheets.ExtendedValue{StringValue: &order.Kanal}},
							},
						},
					},
					Fields: "userEnteredValue",
				},
			},
		)

	case "Perplexity":
		// D=Kode Redeem, E=TglPesanan, F=TglBerakhir, G=Nominal, H=Kanal, I=Nomor/Username
		kodeRedeem := order.KodeRedeem // Empty initially, filled by admin later

		requests = append(requests,
			// Update D (Kode Redeem)
			&sheets.Request{
				UpdateCells: &sheets.UpdateCellsRequest{
					Start: &sheets.GridCoordinate{
						SheetId:     sheetID,
						RowIndex:    lastRow,
						ColumnIndex: 3, // Column D
					},
					Rows: []*sheets.RowData{
						{
							Values: []*sheets.CellData{
								{UserEnteredValue: &sheets.ExtendedValue{StringValue: &kodeRedeem}},
							},
						},
					},
					Fields: "userEnteredValue",
				},
			},
			// Update E (Tanggal Pesanan), F (Tanggal Berakhir)
			&sheets.Request{
				UpdateCells: &sheets.UpdateCellsRequest{
					Start: &sheets.GridCoordinate{
						SheetId:     sheetID,
						RowIndex:    lastRow,
						ColumnIndex: 4, // Column E
					},
					Rows: []*sheets.RowData{
						{
							Values: []*sheets.CellData{
								{UserEnteredValue: &sheets.ExtendedValue{StringValue: &tanggal}},
								{UserEnteredValue: &sheets.ExtendedValue{StringValue: &tanggalBerakhir}},
							},
						},
					},
					Fields: "userEnteredValue",
				},
			},
			// Update G (Nominal), H (Kanal), I (Nomor/Username)
			&sheets.Request{
				UpdateCells: &sheets.UpdateCellsRequest{
					Start: &sheets.GridCoordinate{
						SheetId:     sheetID,
						RowIndex:    lastRow,
						ColumnIndex: 6, // Column G
					},
					Rows: []*sheets.RowData{
						{
							Values: []*sheets.CellData{
								{UserEnteredValue: &sheets.ExtendedValue{NumberValue: ptr64(float64(order.Amount))}},
								{UserEnteredValue: &sheets.ExtendedValue{StringValue: &order.Kanal}},
								{UserEnteredValue: &sheets.ExtendedValue{StringValue: &order.Akun}},
							},
						},
					},
					Fields: "userEnteredValue",
				},
			},
		)

	default:
		return fmt.Errorf("unsupported product: %s", order.Produk)
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
