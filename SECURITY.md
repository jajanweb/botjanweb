# 🔒 Security Guide - BotJanWeb

## Overview

BotJanWeb adalah personal bot project dengan beberapa endpoint yang perlu diamankan. Dokumen ini menjelaskan implementasi keamanan yang ada dan best practices untuk deployment.

---

## 🛡️ Security Features Implemented

### 1. **Webhook Authentication**

**File**: `presentation/handler/http/webhook.go`

Webhook payment notifications diamankan dengan secret key validation:

```go
// Webhook menerima secret via header (2 options)
secret := r.Header.Get("X-Webhook-Secret")
if secret == "" {
    secret = r.Header.Get("Authorization")  // Fallback
}

if c.secret != "" && secret != c.secret {
    http.Error(w, "Unauthorized", http.StatusUnauthorized)
    return
}
```

**Configuration**:
```bash
# Set webhook secret (minimum 8 characters)
WEBHOOK_SECRET=your-random-secret-key-here-min-8-chars
```

**How to call webhook**:
```bash
curl -X POST https://your-app.herokuapp.com/webhook/payment \
  -H "X-Webhook-Secret: your-secret-key" \
  -H "Content-Type: application/json" \
  -d '{"app":"DANA","title":"Transfer masuk","body":"Rp50.000"}'
```

---

### 2. **QR Pairing Endpoint Protection**

**File**: `presentation/handler/http/qr_pairing.go`

QR pairing untuk production deployment diamankan dengan:

#### ✅ **Header-based Authentication** (bukan URL query)
- Token dikirim via `X-Pairing-Token` header
- Tidak muncul di server logs atau browser history
- Tidak tercuri via referrer header

```javascript
// Frontend automatically sends token via header
fetch('/pairing/qr', {
    headers: {
        'X-Pairing-Token': 'your-pairing-token'
    }
})
```

#### ✅ **Rate Limiting** (simple, personal project)
- Max 5 failed attempts per IP
- Block duration: 15 minutes
- Auto-reset after 5 minutes of no activity
- Melindungi dari brute force attacks

#### ✅ **Audit Logging**
- Log semua unauthorized attempts dengan IP address
- Log successful pairing access
- Log saat IP di-block

**Configuration**:
```bash
# Set strong pairing token (recommend: random 32+ chars)
PAIRING_TOKEN=$(openssl rand -hex 32)
```

**How to access pairing page**:
```bash
# Via curl (untuk testing)
curl -H "X-Pairing-Token: your-token" https://your-app.herokuapp.com/pairing

# Via browser (requires extension like ModHeader)
# Install ModHeader extension, add header:
#   Name: X-Pairing-Token
#   Value: your-token
# Then open: https://your-app.herokuapp.com/pairing
```

---

### 3. **Configuration Validation**

**File**: `internal/config/types.go`

Semua konfigurasi divalidasi saat startup:

```go
// Webhook secret must be strong
if len(c.WebhookSecret) < 8 {
    return fmt.Errorf("WEBHOOK_SECRET too weak (minimum 8 characters)")
}

// Phone numbers format validation
for _, phone := range c.AllowedSenders {
    if len(normalized) < 10 || len(normalized) > 15 {
        return fmt.Errorf("invalid phone number: %s", phone)
    }
}
```

**Prevents**:
- Weak secrets
- Invalid phone numbers
- Missing required fields
- Malformed configuration

---

## 🚨 Security Checklist for Deployment

### Before Deploy to Production:

- [ ] **Set strong secrets** (minimum 32 characters random):
  ```bash
  # Generate secure secrets
  WEBHOOK_SECRET=$(openssl rand -hex 32)
  PAIRING_TOKEN=$(openssl rand -hex 32)
  
  # Set in Heroku
  heroku config:set WEBHOOK_SECRET="$WEBHOOK_SECRET"
  heroku config:set PAIRING_TOKEN="$PAIRING_TOKEN"
  ```

- [ ] **Enable HTTPS only** (Heroku does this by default)
  - All communication must be encrypted
  - Check: `https://` prefix on all URLs
  
- [ ] **Restrict allowed senders**:
  ```bash
  # Only your phone numbers
  heroku config:set ALLOWED_SENDERS="+6282116086024,+6281234567890"
  ```

- [ ] **Monitor logs for security events**:
  ```bash
  # Check for [SECURITY] events
  heroku logs --tail | grep SECURITY
  ```

- [ ] **Never commit secrets to Git**:
  - Use `.env` for local development (gitignored)
  - Use Heroku config vars for production
  - Check `.env.example` for reference only

- [ ] **Backup WhatsApp session database**:
  ```bash
  # Download session from Heroku
  heroku run 'cat whatsmeow.db' > whatsmeow_backup.db
  ```

---

## 🔐 Best Practices

### 1. **Strong Secrets Generation**

**❌ WEAK** (don't use):
```bash
WEBHOOK_SECRET=secret123
PAIRING_TOKEN=password
```

**✅ STRONG** (use this):
```bash
# OpenSSL (Linux/Mac)
openssl rand -hex 32

# Python
python3 -c "import secrets; print(secrets.token_hex(32))"

# Node.js
node -e "console.log(require('crypto').randomBytes(32).toString('hex'))"
```

### 2. **Environment Variables Security**

**Local Development** (`.env` file):
```bash
# .env (never commit this!)
WEBHOOK_SECRET=local-dev-secret-at-least-8-chars
PAIRING_TOKEN=local-dev-pairing-token-min-8-chars
```

**Production** (Heroku config):
```bash
heroku config:set WEBHOOK_SECRET="$(openssl rand -hex 32)"
heroku config:set PAIRING_TOKEN="$(openssl rand -hex 32)"
```

### 3. **Webhook Security**

**Configure payment gateway** to send notifications with secret:
```bash
# Example: Configure DANA/OVO webhook URL
Webhook URL: https://your-app.herokuapp.com/webhook/payment
Header Name: X-Webhook-Secret
Header Value: <your-webhook-secret-from-heroku-config>
```

### 4. **QR Pairing Access Control**

**Browser Access** (requires ModHeader extension):
1. Install [ModHeader](https://chrome.google.com/webstore/detail/modheader/idgpnmonknjnojddfkpgkljpfnnfcklj)
2. Add request header:
   - Name: `X-Pairing-Token`
   - Value: `<your-pairing-token>`
3. Open: `https://your-app.herokuapp.com/pairing`

**Alternative** (curl for quick pairing):
```bash
curl -H "X-Pairing-Token: your-token" \
     https://your-app.herokuapp.com/pairing
```

### 5. **Monitor Security Logs**

```bash
# Watch for unauthorized attempts
heroku logs --tail | grep "SECURITY"

# Examples of log messages:
# [SECURITY] Unauthorized pairing attempt from 1.2.3.4
# [SECURITY] IP 1.2.3.4 has been blocked after 5 failed attempts
# [SECURITY] Authorized pairing page access from 1.2.3.4
```

---

## 🚧 Known Limitations (Personal Project)

**⚠️ This is a personal bot project, NOT a SaaS platform.**

### What's NOT implemented (intentionally):

1. **Advanced Rate Limiting**: Simple IP-based blocking, no distributed rate limiter
2. **IP Whitelisting**: No config option to restrict by IP (add if needed)
3. **Token Expiration**: Pairing token valid forever (rotate manually if leaked)
4. **Session Management**: No force logout mechanism (restart app to clear)
5. **2FA/OTP**: No additional authentication layers
6. **DDoS Protection**: Relies on Heroku's infrastructure
7. **Security Headers**: No CSP, HSTS, X-Frame-Options headers
8. **Input Sanitization**: Basic validation only

### If you need enterprise-grade security:
- Consider using API Gateway (e.g., Kong, Cloudflare)
- Add WAF (Web Application Firewall)
- Implement proper session management
- Add security headers middleware
- Use vault service for secrets (e.g., HashiCorp Vault)

---

## 🐛 Security Issues Reporting

**This is a personal project.** If you find security issues:

1. **Don't create public GitHub issues**
2. **Contact owner directly** (via WhatsApp or email)
3. **Provide details**: Attack vector, reproduction steps, impact
4. **Suggest fixes** (if you have ideas)

---

## 📚 References

- [OWASP Top 10](https://owasp.org/www-project-top-ten/)
- [WhatsApp Web Security](https://faq.whatsapp.com/general/security-and-privacy/)
- [Heroku Security Best Practices](https://devcenter.heroku.com/categories/security)
- [Go Security Best Practices](https://golang.org/doc/security/)

---

## 📝 Security Updates Log

| Date | Change | Reason |
|------|--------|--------|
| 2024-01-XX | Move token from URL to header | Prevent exposure in logs |
| 2024-01-XX | Add rate limiting (5 attempts) | Protect from brute force |
| 2024-01-XX | Add audit logging | Security monitoring |

---

**Last Updated**: 2024-01-XX  
**Security Level**: ⭐⭐⭐☆☆ (Good for personal project)
