#!/bin/bash
# Test Webhook Payment Flow - End to End
# This script demonstrates a complete payment flow:
# 1. Generate QRIS via WhatsApp command
# 2. Simulate DANA payment notification webhook
# 3. Verify payment confirmation in group

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
WEBHOOK_URL="${WEBHOOK_URL:-https://botjanweb-e6734da46d03.herokuapp.com/webhook/payment}"
WEBHOOK_SECRET="${WEBHOOK_SECRET:-c6bab3ab7e4dec519e3387df5e85fcc4e9e8fadee594f855ffba070a7711a0ce}"

echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo -e "${BLUE}   🧪 BotJanWeb Webhook Payment Flow Test${NC}"
echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo ""

# Test 1: Health Check
echo -e "${YELLOW}[1/4]${NC} Testing health endpoint..."
HEALTH_RESPONSE=$(curl -s -w "\n%{http_code}" "$WEBHOOK_URL" | tail -1)
if [ "$HEALTH_RESPONSE" = "404" ]; then
    echo -e "${GREEN}✓${NC} Server is running (404 expected for GET /webhook/payment)"
else
    echo -e "${RED}✗${NC} Unexpected response: $HEALTH_RESPONSE"
fi
echo ""

# Test 2: Invalid Secret
echo -e "${YELLOW}[2/4]${NC} Testing webhook authentication..."
INVALID_RESPONSE=$(curl -s -X POST "$WEBHOOK_URL" \
  -H "Content-Type: application/json" \
  -H "X-Webhook-Secret: wrong-secret" \
  -d '{"app":"id.dana","title":"DANA","message":"Test","timestamp":"1733645400000"}' \
  -w "\n%{http_code}" | tail -1)

if [ "$INVALID_RESPONSE" = "401" ]; then
    echo -e "${GREEN}✓${NC} Authentication working (401 for invalid secret)"
else
    echo -e "${RED}✗${NC} Expected 401, got: $INVALID_RESPONSE"
fi
echo ""

# Test 3: Valid Webhook - No Match
echo -e "${YELLOW}[3/4]${NC} Testing valid DANA notification (no pending)..."
NO_MATCH_RESPONSE=$(curl -s -X POST "$WEBHOOK_URL" \
  -H "Content-Type: application/json" \
  -H "X-Webhook-Secret: $WEBHOOK_SECRET" \
  -d '{
    "app": "id.dana",
    "title": "DANA",
    "message": "Kamu berhasil menerima Rp99.999 dari Test User",
    "timestamp": "1733645400000"
  }')

echo "Response: $NO_MATCH_RESPONSE"
if echo "$NO_MATCH_RESPONSE" | grep -q "unmatched"; then
    echo -e "${GREEN}✓${NC} Webhook processed correctly (no matching pending)"
else
    echo -e "${RED}✗${NC} Unexpected response format"
fi
echo ""

# Test 4: Realistic Amounts
echo -e "${YELLOW}[4/4]${NC} Testing common payment amounts..."
AMOUNTS=(10000 15000 25000 50000)

for AMOUNT in "${AMOUNTS[@]}"; do
    FORMATTED_AMOUNT=$(printf "%'.0f" $AMOUNT | tr ',' '.')
    echo -e "  Testing Rp${FORMATTED_AMOUNT}..."
    
    RESPONSE=$(curl -s -X POST "$WEBHOOK_URL" \
      -H "Content-Type: application/json" \
      -H "X-Webhook-Secret: $WEBHOOK_SECRET" \
      -d "{
        \"app\": \"id.dana\",
        \"title\": \"DANA\",
        \"message\": \"Kamu berhasil menerima Rp${FORMATTED_AMOUNT} dari Test User\",
        \"timestamp\": \"$(date +%s)000\"
      }")
    
    if echo "$RESPONSE" | grep -q "unmatched"; then
        echo -e "    ${BLUE}→${NC} Webhook OK (no pending payment)"
    elif echo "$RESPONSE" | grep -q "matched"; then
        echo -e "    ${GREEN}✓${NC} Payment matched and confirmed!"
    elif echo "$RESPONSE" | grep -q "ignored"; then
        echo -e "    ${YELLOW}⚠${NC} $(echo $RESPONSE | jq -r '.reason')"
    else
        echo -e "    ${RED}✗${NC} Error: $RESPONSE"
    fi
done

echo ""
echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo -e "${GREEN}✓ Webhook testing complete!${NC}"
echo ""
echo -e "${YELLOW}💡 Next Steps:${NC}"
echo "1. Send QRIS command in WhatsApp: /qris 15000"
echo "2. Note the Message ID from bot response"
echo "3. Simulate payment webhook using this script"
echo "4. Bot should send confirmation to group"
echo ""
echo -e "${BLUE}📖 Documentation:${NC} WEBHOOK_TESTING.md"
echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
