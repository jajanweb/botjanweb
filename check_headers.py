#!/usr/bin/env python3
from google.oauth2 import service_account
from googleapiclient.discovery import build

credentials = service_account.Credentials.from_service_account_file(
    './credentials.json',
    scopes=['https://www.googleapis.com/auth/spreadsheets.readonly']
)
service = build('sheets', 'v4', credentials=credentials)
SPREADSHEET_ID = '1Fsy2ayfT8KXgjxxUEanU-N9OnU-xlxp28UkAKrHxUUk'

print("="*80)
print("STRUKTUR AKTUAL SEMUA PRODUCT SHEETS")
print("="*80)
print()

for sheet in ['Gemini', 'ChatGPT', 'YouTube', 'Perplexity']:
    result = service.spreadsheets().values().get(
        spreadsheetId=SPREADSHEET_ID,
        range=f'{sheet}!A2:K2'
    ).execute()
    headers = result.get('values', [[]])[0] if result.get('values') else []
    
    print(f'Sheet: {sheet}')
    print('-'*80)
    for i, h in enumerate(headers):
        print(f'  {chr(65+i)} = {h}')
    print()
