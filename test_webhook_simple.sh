#!/bin/bash
# Simple Webhook Test Script
# Usage: ./test_webhook_simple.sh <amount>

AMOUNT=${1:-15000}
WEBHOOK_URL="https://botjanweb-e6734da46d03.herokuapp.com/webhook/payment"
SECRET="c6bab3ab7e4dec519e3387df5e85fcc4e9e8fadee594f855ffba070a7711a0ce"

echo "🧪 Testing webhook for Rp${AMOUNT}..."
echo ""

RESPONSE=$(curl -s -X POST "$WEBHOOK_URL" \
  -H "Content-Type: application/json" \
  -H "X-Webhook-Secret: $SECRET" \
  -d "{
    \"app\": \"id.dana\",
    \"title\": \"DANA\",
    \"message\": \"Kamu berhasil menerima Rp${AMOUNT} dari Test User\",
    \"timestamp\": \"$(date +%s)000\"
  }" \
  -w "\nHTTP_STATUS:%{http_code}")

HTTP_CODE=$(echo "$RESPONSE" | grep "HTTP_STATUS" | cut -d: -f2)
BODY=$(echo "$RESPONSE" | grep -v "HTTP_STATUS")

echo "📥 Response:"
echo "$BODY" | jq -C . 2>/dev/null || echo "$BODY"
echo ""

if [ "$HTTP_CODE" = "200" ]; then
    if echo "$BODY" | grep -q '"status":"matched"'; then
        echo "✅ Payment MATCHED! Bot should send notification to group."
    elif echo "$BODY" | grep -q '"status":"unmatched"'; then
        echo "⚠️  No pending payment found for Rp${AMOUNT}"
        echo "💡 Create QRIS first: Send '/qris $AMOUNT' to WhatsApp group"
    elif echo "$BODY" | grep -q '"status":"ignored"'; then
        REASON=$(echo "$BODY" | jq -r '.reason' 2>/dev/null || echo "unknown")
        echo "⚠️  Webhook ignored: $REASON"
    fi
elif [ "$HTTP_CODE" = "401" ]; then
    echo "❌ Unauthorized - Check your WEBHOOK_SECRET"
else
    echo "❌ HTTP $HTTP_CODE"
fi
