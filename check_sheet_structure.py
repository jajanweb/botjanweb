#!/usr/bin/env python3
"""
Check Google Sheets structure to verify column mapping
"""
import json
from google.oauth2 import service_account
from googleapiclient.discovery import build

# Load credentials
CREDENTIALS_FILE = './credentials.json'
SPREADSHEET_ID = '1Fsy2ayfT8KXgjxxUEanU-N9OnU-xlxp28UkAKrHxUUk'

# Authenticate
credentials = service_account.Credentials.from_service_account_file(
    CREDENTIALS_FILE,
    scopes=['https://www.googleapis.com/auth/spreadsheets.readonly']
)

service = build('sheets', 'v4', credentials=credentials)

# Check structure of each product sheet
sheets_to_check = ['Gemini', 'ChatGPT', 'YouTube', 'Perplexity']

print("=" * 80)
print("STRUKTUR HEADER GOOGLE SHEETS - AUDIT")
print("=" * 80)
print()

for sheet_name in sheets_to_check:
    try:
        # Read first TWO rows (row 1 might be title, row 2 is header)
        range_name = f"'{sheet_name}'!A1:Z2"
        result = service.spreadsheets().values().get(
            spreadsheetId=SPREADSHEET_ID,
            range=range_name
        ).execute()
        
        rows = result.get('values', [])
        
        print(f"📋 Sheet: {sheet_name}")
        print("-" * 80)
        
        # Show both rows to identify which is header
        if len(rows) >= 1:
            print("ROW 1 (Title/Label?):")
            for i, val in enumerate(rows[0] if len(rows[0]) > 0 else []):
                col_letter = chr(65 + i)
                print(f"  {col_letter} = {val}")
        
        if len(rows) >= 2:
            print("\nROW 2 (Header?):")
            for i, val in enumerate(rows[1] if len(rows[1]) > 0 else []):
                col_letter = chr(65 + i)
                print(f"  {col_letter} = {val}")
        print()
        
    except Exception as e:
        print(f"❌ Error reading {sheet_name}: {e}")
        print()

print("=" * 80)
print("CHECKING DATA SAMPLES (First 3 rows)")
print("=" * 80)
print()

for sheet_name in sheets_to_check:
    try:
        # Read first 4 rows (1 header + 3 data rows)
        range_name = f"'{sheet_name}'!A1:J4"
        result = service.spreadsheets().values().get(
            spreadsheetId=SPREADSHEET_ID,
            range=range_name
        ).execute()
        
        rows = result.get('values', [])
        
        print(f"📊 Sheet: {sheet_name} - Data Sample")
        print("-" * 80)
        if rows:
            print(f"Row 1 (Header): {rows[0]}")
            for idx, row in enumerate(rows[1:4], start=2):
                print(f"Row {idx} (Data):   {row}")
        else:
            print("  (No data)")
        print()
        
    except Exception as e:
        print(f"❌ Error reading {sheet_name}: {e}")
        print()

print("✅ Done!")
