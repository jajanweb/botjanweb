//go:build ignore

package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"google.golang.org/api/option"
	"google.golang.org/api/sheets/v4"
)

func main() {
	ctx := context.Background()
	spreadsheetID := "1Fsy2ayfT8KXgjxxUEanU-N9OnU-xlxp28UkAKrHxUUk"
	credentialsPath := "./credentials.json"

	srv, err := sheets.NewService(ctx, option.WithCredentialsFile(credentialsPath))
	if err != nil {
		log.Fatalf("Unable to create sheets service: %v", err)
	}

	resp, err := srv.Spreadsheets.Get(spreadsheetID).Do()
	if err != nil {
		log.Fatalf("Unable to retrieve spreadsheet: %v", err)
	}

	fmt.Println("📊 Available sheets:")
	for i, sheet := range resp.Sheets {
		fmt.Printf("%d. %s (ID: %d)\n", i+1, sheet.Properties.Title, sheet.Properties.SheetId)
	}

	sheetName := "Kode Perplexity"
	if len(os.Args) > 1 {
		sheetName = os.Args[1]
	}

	fmt.Printf("\n📋 Fetching '%s'...\n\n", sheetName)

	headerRange := fmt.Sprintf("'%s'!A1:J2", sheetName)
	headerResp, err := srv.Spreadsheets.Values.Get(spreadsheetID, headerRange).Do()
	if err != nil {
		log.Fatalf("Unable to read header: %v", err)
	}

	fmt.Println("📌 HEADER (Rows 1-2):")
	for i, row := range headerResp.Values {
		fmt.Printf("Row %d: ", i+1)
		for j, cell := range row {
			if cell != nil && fmt.Sprintf("%v", cell) != "" {
				fmt.Printf("[%s]='%v' ", string(rune('A'+j)), cell)
			}
		}
		fmt.Println()
	}

	dataRange := fmt.Sprintf("'%s'!A3:J15", sheetName)
	dataResp, err := srv.Spreadsheets.Values.Get(spreadsheetID, dataRange).Do()
	if err != nil {
		log.Fatalf("Unable to read data: %v", err)
	}

	fmt.Println("\n📊 SAMPLE DATA (Rows 3-15):")
	for i, row := range dataResp.Values {
		fmt.Printf("Row %d: ", i+3)
		for j, cell := range row {
			if cell != nil && fmt.Sprintf("%v", cell) != "" {
				fmt.Printf("[%s]='%v' ", string(rune('A'+j)), cell)
			}
		}
		fmt.Println()
	}

	fmt.Println("\n✅ Fetch complete!")
}
