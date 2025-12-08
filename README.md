# BotJanWeb

A production-ready WhatsApp group bot built in Go with modular, clean architecture. BotJanWeb integrates with WhatsApp (via WhatsMeow), dynamic QRIS generation, Google Sheets for data logging, and automatic payment notification via DANA.

## Features

### 1. `#qris` - Dynamic QRIS Generator with Order Forms

Generate dynamic QRIS payment codes for products with guided order forms. Supports multiple products with specific validation rules.

**Format:**
```
#qris <product>
```

**Supported Products:**
- `google` - Google Workspace/Gemini accounts (family slots, max 5)
- `chatgpt` - ChatGPT accounts (workspace validation)

**Examples:**
```
#qris google
#qris chatgpt
```

**Behavior:**
1. Bot sends product-specific order form template
2. Customer fills out the form and sends back
3. Bot validates the order:
   - **Google**: Checks if family still has available slots (max 5)
   - **ChatGPT**: Validates workspace availability
4. If valid, bot generates dynamic QRIS and sends to customer directly (Self-QRIS)
5. Bot notifies group with order details
6. Order is logged to product-specific Google Sheet
7. Pending payment registered for automatic confirmation

**Order Form Fields:**

*Google Products:*
```
━━━━━━━━━━━━━━━━━━━━
Nama:
Gmail:
No HP:
Nominal:
Product:
Jenis:
━━━━━━━━━━━━━━━━━━━━
```

*ChatGPT Products:*
```
━━━━━━━━━━━━━━━━━━━━
Nama:
Email:
No HP:
Nominal:
Product:
Workspace:
━━━━━━━━━━━━━━━━━━━━
```

**Validation Rules:**
- **Google Family**: Maximum 5 member slots, rejects if full
- **ChatGPT Workspace**: Checks existing workspace name, prevents duplicates
- **Phone Format**: Normalizes to format 08xxx or 8xxx
- **Email**: Basic email format validation

### 2. `#addakun` - Add Account Management

Add new Google or ChatGPT accounts to the system for tracking.

**Format:**
```
#addakun <product>
```

**Supported Products:**
- `google` - Add Google Workspace/Gemini account with family details
- `chatgpt` - Add ChatGPT account with workspace name

**Examples:**
```
#addakun google
#addakun chatgpt
```

**Behavior:**
1. Bot sends account form template for the specified product
2. Admin fills out account details
3. Bot validates and saves to account sheet (`Akun Google` or `Akun ChatGPT`)
4. Sends confirmation with account summary

**Account Form Fields:**

*Google Account:*
```
━━━━━━━━━━━━━━━━━━━━
Nama Family:
Gmail:
Password:
Slot: [angka]
━━━━━━━━━━━━━━━━━━━━
```

*ChatGPT Account:*
```
━━━━━━━━━━━━━━━━━━━━
Workspace:
Email:
Password:
━━━━━━━━━━━━━━━━━━━━
```

### 3. `#listakun` - List Available Accounts

View all registered accounts with availability status.

**Format:**
```
#listakun
```

**Behavior:**
- Lists all Google accounts with family slot usage
- Lists all ChatGPT accounts with workspace names
- Shows which accounts are full or have available slots
- Only accessible by allowed senders

**Example Output:**
```
📋 Akun Google:
• Family Budi (budi@gmail.com) - 3/5 slot terisi
• Family Sinta (sinta@gmail.com) - 5/5 slot terisi ❌

📋 Akun ChatGPT:
• Workspace Alpha (alpha@mail.com)
• Workspace Beta (beta@mail.com)
```

### 4. Self-QRIS Flow (Direct Payment to Customer)

When a customer submits a valid order form, the bot automatically:
1. Generates a dynamic QRIS with the specified amount
2. Sends the QRIS directly to the customer's WhatsApp (DM)
3. Sends a notification to the group with order details
4. Registers the payment as "pending" for automatic confirmation

**Key Features:**
- **Direct Customer Experience**: Customer receives QRIS in private chat, not in group
- **Group Notifications**: Admin group gets notified about new orders
- **Privacy**: Payment details stay between bot and customer
- **No Manual Steps**: Fully automated from order to QRIS delivery

**Flow Example:**
```
Customer → Fills order form
    ↓
Bot → Validates (family slots, workspace, etc.)
    ↓
Bot → Generates QRIS
    ↓
Bot → Sends QRIS to customer DM ✅
    ↓
Bot → Notifies group "New order from 08xxx..."
    ↓
Customer → Pays via DANA/banking app
    ↓
Webhook → Receives payment notification
    ↓
Bot → Sends confirmation to customer & group
```

### 5. Automatic Payment Confirmation (DANA)

Automatically detect DANA payments and send confirmation as a reply to the QRIS image.

**How it works:**
1. When bot generates QRIS via `#qris`, the payment is registered as "pending"
2. Android Nomad Gateway forwards DANA push notifications to webhook
3. Bot matches incoming payment with pending QRIS by amount
4. Confirmation message is sent as a **reply** to the original QRIS image in customer's DM
5. Group also receives payment confirmation notification

**Requirements:**
- [Android Nomad Gateway](https://github.com/AzharRiv662/android-nomad-gateway) app installed
- ngrok or public URL for webhook endpoint
- DANA app with payment notifications enabled

## Project Structure (Clean Architecture)

```
botjanweb/
├── cmd/
│   └── botjanweb/          # Main bot entry point
├── internal/
│   ├── application/        # Application business logic
│   │   ├── service/        # Application services (qris, payment, family, account)
│   │   ├── usecase/        # Reserved: Complex use case orchestrations
│   │   └── dto/            # Reserved: Data transfer objects
│   ├── domain/             # Core business layer
│   │   ├── entity/         # Domain entities
│   │   └── repository/     # Repository interfaces (ports)
│   ├── infrastructure/     # External services adapters
│   │   ├── persistence/    # Database implementations
│   │   │   ├── memory/     # In-memory pending store
│   │   │   ├── postgres/   # PostgreSQL implementation
│   │   │   └── sheets/     # Google Sheets integration
│   │   ├── messaging/      # Messaging adapters
│   │   │   ├── whatsapp/   # WhatsMeow client wrapper
│   │   │   └── webhook/    # HTTP server infrastructure
│   │   └── external/       # External integrations
│   │       └── qris/       # QRIS generation & rendering
│   ├── bootstrap/          # Dependency injection wiring
│   └── config/             # Configuration loading
├── presentation/           # Presentation layer
│   ├── handler/            # Request handlers
│   │   ├── bot/            # WhatsApp message handling
│   │   └── http/           # HTTP webhook handling
│   └── template/           # Message templates
├── pkg/                    # Shared utilities (framework-agnostic)
│   ├── constants/          # Application constants
│   ├── helper/
│   │   ├── formatter/      # Rupiah, phone formatters (used across all layers)
│   │   ├── parser/         # Command parsers
│   │   └── validator/      # Input validators
│   └── logger/             # Centralized logging
└── assets/                 # Static resources (templates, fonts)
```

### Architecture Layers

Following Uncle Bob's Clean Architecture:

1. **Domain (innermost)**: Core business entities with no external dependencies
2. **Application**: Application-specific business rules and orchestration
3. **Infrastructure**: External services implementations (database, messaging, QRIS)
4. **Presentation (outermost)**: Handlers, templates, and user-facing components

Dependencies flow inward: Presentation → Application → Domain

## Setup

### Prerequisites

- Go 1.21+
- A WhatsApp account dedicated for the bot
- Google Cloud project with Sheets API enabled
- Service account credentials JSON file

### 1. Clone and Install Dependencies

```bash
cd botjanweb
go mod tidy
```

### 2. Configure Environment

Copy `.env.example` to `.env` and fill in your values:

```bash
cp .env.example .env
```

**Required Configuration:**

| Variable | Description |
|----------|-------------|
| `WHATSAPP_DB_PATH` | Path to SQLite database for WhatsApp session (default: `./whatsmeow.db`) |
| `ALLOWED_SENDERS` | Comma-separated phone numbers (without +), e.g., `6282116086024,6282219931715` |
| `GROUP_JID` | Target WhatsApp group JID, e.g., `123456789-1234567890@g.us` |
| `QRIS_STATIC_PAYLOAD` | Your static QRIS string (EMV format) |

**Google Sheets Configuration (optional):**

| Variable | Description |
|----------|-------------|
| `SHEETS_ENABLED` | Set to `true` to enable Google Sheets logging |
| `GOOGLE_SPREADSHEET_ID` | ID of your Google Spreadsheet |
| `GOOGLE_CREDENTIALS_PATH` | Path to service account JSON (default: `./credentials.json`) |
| `SHEET_GEMINI` | Sheet name for Gemini/Google orders (default: `Gemini`) |
| `SHEET_CHATGPT` | Sheet name for ChatGPT orders (default: `ChatGPT`) |
| `SHEET_YOUTUBE` | Sheet name for YouTube Premium orders (default: `YouTube`) |
| `SHEET_PERPLEXITY` | Sheet name for Perplexity orders (default: `Perplexity`) |
| `SHEET_AKUN_GOOGLE` | Sheet name for Google accounts (default: `Akun Google`) |
| `SHEET_AKUN_CHATGPT` | Sheet name for ChatGPT accounts (default: `Akun ChatGPT`) |

**Webhook Configuration (optional, for payment notifications):**

| Variable | Description |
|----------|-------------|
| `WEBHOOK_ENABLED` | Set to `true` to enable webhook server |
| `WEBHOOK_PORT` | Port for webhook server (default: `8080`) |
| `WEBHOOK_SECRET` | Secret key for webhook validation |

### 3. Google Sheets Setup

1. Create a new Google Spreadsheet
2. Create the following sheets with their respective headers:

**Product Sheets (Orders):**

**Gemini** (Google Workspace/Gemini orders):
| Timestamp | Nama | Gmail | No HP | Nominal | Product | Jenis | Family | WA Message ID |
|-----------|------|-------|-------|---------|---------|-------|--------|---------------|

**ChatGPT** (ChatGPT orders):
| Timestamp | Nama | Email | No HP | Nominal | Product | Workspace | WA Message ID |
|-----------|------|-------|-------|---------|---------|-----------|---------------|

**YouTube** (YouTube Premium orders):
| Timestamp | Nama | Email | No HP | Nominal | Product | Family | WA Message ID |
|-----------|------|-------|-------|---------|---------|--------|---------------|

**Perplexity** (Perplexity AI orders):
| Timestamp | Nama | Email | No HP | Nominal | Product | WA Message ID |
|-----------|------|-------|-------|---------|---------|---------------|

**Account Sheets (Inventory):**

**Akun Google** (Google account management):
| Timestamp | Nama Family | Gmail | Password | Slot | Tersisa |
|-----------|-------------|-------|----------|------|----------|

**Akun ChatGPT** (ChatGPT account management):
| Timestamp | Workspace | Email | Password |
|-----------|-----------|-------|----------|

4. Share the spreadsheet with your service account email (found in credentials JSON)

### 4. Finding Your Group JID

To find your WhatsApp group JID:

1. Run the bot once with any group JID
2. The bot will log incoming messages with their chat JID
3. Send a test message to your target group
4. Copy the JID from the logs (format: `123456789-1234567890@g.us`)

### 5. Build and Run

```bash
# Build
go build -o botjanweb ./cmd/botjanweb

# Run
./botjanweb
```

On first run, a QR code will be displayed. Scan it with WhatsApp to pair the bot account.

## 🚀 Deployment

### **Production Deployment to Heroku**

BotJanWeb includes built-in web-based QR pairing for production deployment:

#### **Quick Deploy** (< 10 minutes):
```bash
heroku create my-bot
heroku buildpacks:set heroku/go
heroku config:set GROUP_JID="123456@g.us"
heroku config:set QRIS_STATIC_PAYLOAD="..."
heroku config:set WEBHOOK_SECRET="secret"
git push heroku main
```

#### **Pairing WhatsApp in Production**:
```
https://my-bot.herokuapp.com/pairing?token=YOUR_SECRET
```
No terminal access needed! Scan QR code directly in browser.

#### **Documentation**:
- 📖 **[QUICKSTART_HEROKU.md](./QUICKSTART_HEROKU.md)** - 10-minute quick start
- 📚 **[HEROKU_DEPLOY.md](./HEROKU_DEPLOY.md)** - Complete deployment guide
- 🔐 **[WHATSAPP_PAIRING.md](./WHATSAPP_PAIRING.md)** - Pairing in production explained

### **Key Features for Production**:
- ✅ Web-based QR pairing (`/pairing` endpoint)
- ✅ Health & readiness endpoints (`/health`, `/ready`)
- ✅ Webhook for payment notifications
- ✅ Graceful shutdown handling
- ✅ Heroku $PORT auto-detection
- ✅ Session persistence options (Postgres/S3)

---

## Architecture

The project follows Uncle Bob's Clean Architecture with clear separation of concerns:

### Layers

| Layer | Purpose | Example Packages |
|-------|---------|-----------------|
| **Domain** | Core business entities | `domain/entity` |
| **Use Cases** | Application business rules | `usecase/qris`, `usecase/payment`, `usecase/order` |
| **Controllers** | Input adapters | `controller/bot`, `controller/http` |
| **Infrastructure** | External services | `infrastructure/whatsapp`, `infrastructure/qris` |
| **Repository** | Data access | `repository/memory`, `repository/sheets` |

### Dependency Injection

All dependencies are wired in `bootstrap/app.go` following the Dependency Inversion Principle. Use cases depend on interfaces (ports), not concrete implementations.

## Adding New Commands

To add a new command:

1. Add a new entity in `internal/domain/entity/command.go`
2. Create a new use case in `internal/usecase/<command>/usecase.go`
3. Add parser function in `internal/controller/bot/handler.go`
4. Update `HandleMessage()` to route to the new command
5. Wire the new use case in `internal/bootstrap/app.go`

## 🔒 Security

BotJanWeb implements security best practices for personal bot projects:

### Implemented Security Features

- ✅ **Header-based Authentication**: Tokens sent via headers (not URL) to prevent exposure in logs
- ✅ **Rate Limiting**: Simple IP-based blocking (5 attempts, 15-minute block)
- ✅ **Audit Logging**: All security events logged with IP addresses
- ✅ **Secret Validation**: Minimum 8 characters for webhook secrets
- ✅ **Configuration Validation**: All configs validated at startup
- ✅ **Webhook Authentication**: X-Webhook-Secret header validation

### Quick Security Setup

```bash
# Generate strong secrets
WEBHOOK_SECRET=$(openssl rand -hex 32)
PAIRING_TOKEN=$(openssl rand -hex 32)

# Set in Heroku
heroku config:set WEBHOOK_SECRET="$WEBHOOK_SECRET"
heroku config:set PAIRING_TOKEN="$PAIRING_TOKEN"

# Restrict allowed senders
heroku config:set ALLOWED_SENDERS="+6282116086024"
```

### Accessing QR Pairing (Production)

The QR pairing endpoint requires authentication via header:

**Using ModHeader extension** (recommended for browsers):
1. Install [ModHeader](https://chrome.google.com/webstore/detail/modheader/idgpnmonknjnojddfkpgkljpfnnfcklj)
2. Add header: `X-Pairing-Token: <your-token>`
3. Visit: `https://your-app.herokuapp.com/pairing`

**Using curl**:
```bash
curl -H "X-Pairing-Token: your-token" \
     https://your-app.herokuapp.com/pairing
```

### Monitor Security Events

```bash
# Watch security logs
heroku logs --tail | grep SECURITY

# Examples:
# [SECURITY] Unauthorized pairing attempt from 1.2.3.4
# [SECURITY] IP 1.2.3.4 blocked after 5 failed attempts
# [SECURITY] Authorized access from 1.2.3.4
```

**📖 Full Security Documentation**: See [SECURITY.md](./SECURITY.md) for complete security guide, best practices, and deployment checklist.

## Dependencies

- [go.mau.fi/whatsmeow](https://github.com/tulir/whatsmeow) - WhatsApp Web API
- [github.com/fyvri/go-qris](https://github.com/fyvri/go-qris) - QRIS generation
- [github.com/skip2/go-qrcode](https://github.com/skip2/go-qrcode) - QR code image generation
- [google.golang.org/api/sheets/v4](https://pkg.go.dev/google.golang.org/api/sheets/v4) - Google Sheets API
- [github.com/joho/godotenv](https://github.com/joho/godotenv) - Environment file loading

## License

MIT

