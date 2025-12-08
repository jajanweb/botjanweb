# Testing Webhook Payment Endpoint

## ⚠️ Masalah yang Ditemukan dari Log

### 1. Method GET Tidak Didukung (404)
```
method=GET path="/webhook/payment" status=404
```
**Solusi**: Gunakan method **POST**, bukan GET.

### 2. Invalid Webhook Secret (401)
```
[WEBHOOK] Invalid webhook secret dari 10.1.95.250:53646
status=401
```
**Solusi**: Sertakan header `X-Webhook-Secret` dengan nilai yang benar.

---

## ✅ Cara Test yang Benar

### Menggunakan cURL

```bash
# Production (Heroku)
curl -X POST https://botjanweb-e6734da46d03.herokuapp.com/webhook/payment \
  -H "Content-Type: application/json" \
  -H "X-Webhook-Secret: c6bab3ab7e4dec519e3387df5e85fcc4e9e8fadee594f855ffba070a7711a0ce" \
  -d '{
    "app": "id.dana",
    "title": "DANA",
    "message": "Kamu berhasil menerima Rp15.000 dari John Doe",
    "timestamp": "1733645400000"
  }'
```

**Expected Response:**
- Status: `200 OK`
- Body: `{"message":"Webhook received"}`

---

### Menggunakan Postman / Thunder Client

#### 1. Setup Request
- **Method**: POST
- **URL**: `https://botjanweb-e6734da46d03.herokuapp.com/webhook/payment`

#### 2. Headers
```
Content-Type: application/json
X-Webhook-Secret: c6bab3ab7e4dec519e3387df5e85fcc4e9e8fadee594f855ffba070a7711a0ce
```

#### 3. Body (JSON)
```json
{
  "app": "id.dana",
  "title": "DANA",
  "message": "Kamu berhasil menerima Rp25.000 dari Jane Smith",
  "timestamp": "1733645400000"
}
```

**⚠️ Important Notes:**
- `app` must be exactly `"id.dana"` (DANA package name)
- `message` must contain **"berhasil menerima"** and **"Rp"** to be recognized
- `timestamp` is Unix timestamp in **milliseconds** (not seconds)
- Example timestamp: `1733645400000` = 2025-12-08 09:30:00 UTC

---

## 🧪 Testing Lokal (Development)

### 1. Start Development Server
```bash
make air
```

### 2. Test Webhook
```bash
curl -X POST http://localhost:9090/webhook/payment \
  -H "Content-Type: application/json" \
  -H "X-Webhook-Secret: dev-secret-key" \
  -d '{
    "app": "id.dana",
    "title": "DANA",
    "message": "Kamu berhasil menerima Rp25.000 dari Test User",
    "timestamp": "1733645400000"
  }'
```

---

## 📋 Parameter Webhook Payment

| Field | Type | Required | Deskripsi |
|-------|------|----------|-----------|
| `app` | string | ✅ | Package name aplikasi: `"id.dana"` |
| `title` | string | ✅ | Judul notifikasi (biasanya `"DANA"`) |
| `message` | string | ✅ | Isi notifikasi. Harus mengandung:<br>- `"berhasil menerima"`<br>- `"Rp"` diikuti nominal<br>Contoh: `"Kamu berhasil menerima Rp15.000 dari John"` |
| `timestamp` | string | ✅ | Unix timestamp dalam **milliseconds**<br>Contoh: `"1733645400000"` |

### Regex Pattern untuk `message`:
- Parser mencari: `Rp\s?([\d.]+)`
- Contoh valid: 
  - `"Rp15.000"`
  - `"Rp 25000"`
  - `"Rp150.000"`

---

## 🔍 Verifikasi di Log Heroku

Setelah test webhook, cek log:

```bash
heroku logs --tail -a botjanweb | grep -E "WEBHOOK|BOT"
```

**Log yang diharapkan:**
```
[WEBHOOK] Webhook diterima: app=id.dana | title=DANA
[WEBHOOK] Pembayaran dicocokkan: Rp15000 | MsgID: 3EB0...
[BOT] 💰 Pembayaran confirmed untuk pending payment
```

---

## ❌ Common Errors

### Error 401: Invalid webhook secret
**Penyebab**: Header `X-Webhook-Secret` salah atau tidak ada.

**Solusi**:
```bash
# Cek secret yang benar
heroku config:get WEBHOOK_SECRET -a botjanweb

# Gunakan value tersebut di header X-Webhook-Secret
```

### Error 404: Not Found
**Penyebab**: 
1. Path salah (bukan `/webhook/payment`)
2. Method GET digunakan (harus POST)

**Solusi**: Gunakan path `/webhook/payment` dengan method POST.

### Error 500: Internal Server Error
**Penyebab**: Kesalahan di code (cek log Heroku).

**Solusi**:
```bash
heroku logs --tail -a botjanweb
```

---

## 🎯 Flow Webhook Payment

```
QRIS Provider (Xendit/Midtrans) 
    ↓
    POST /webhook/payment (dengan X-Webhook-Secret)
    ↓
BotJanWeb validates secret
    ↓
    ✅ Process payment
    ↓
    Update database (mark payment CONFIRMED)
    ↓
    Send WhatsApp notification to group
    ↓
    Return 200 OK
```

---

## 🔐 Security Notes

1. **Secret Rotation**: Ganti WEBHOOK_SECRET secara berkala
   ```bash
   heroku config:set WEBHOOK_SECRET=$(openssl rand -hex 32) -a botjanweb
   ```

2. **IP Whitelist** (Opsional): Batasi webhook hanya dari IP provider tertentu

3. **HTTPS Only**: Webhook hanya menerima HTTPS di production

---

## 📚 Reference

- Code: `internal/infrastructure/messaging/webhook/handler.go`
- Config: `WEBHOOK_SECRET` environment variable
- Endpoint: `POST /webhook/payment`
