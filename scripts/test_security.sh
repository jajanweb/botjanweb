#!/bin/bash

# 🔒 BotJanWeb Security Testing Suite
# Testing security improvements: header auth, rate limiting, audit logging

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

BASE_URL="http://localhost:8080"
VALID_TOKEN="test-pairing-token-min-8-chars"
VALID_SECRET="test-secret-for-webhook-min-8-chars"

echo -e "${BLUE}╔════════════════════════════════════════════╗${NC}"
echo -e "${BLUE}║  🔒 BotJanWeb Security Testing Suite      ║${NC}"
echo -e "${BLUE}╚════════════════════════════════════════════╝${NC}"
echo ""

passed=0
failed=0

# Helper function to run test
run_test() {
  local test_num=$1
  local test_name=$2
  local expected=$3
  local actual=$4
  
  echo -e "${YELLOW}Test $test_num:${NC} $test_name"
  if [ "$actual" = "$expected" ]; then
    echo -e "  ${GREEN}✅ PASS${NC} - Got $actual"
    ((passed++))
  else
    echo -e "  ${RED}❌ FAIL${NC} - Got $actual, expected $expected"
    ((failed++))
  fi
  echo ""
}

# Check if bot is running
echo -e "${BLUE}🔍 Checking if bot is running on $BASE_URL...${NC}"
if ! curl -s -o /dev/null -w "%{http_code}" --connect-timeout 2 $BASE_URL/health | grep -q "200"; then
  echo -e "${RED}❌ ERROR: Bot is not running!${NC}"
  echo ""
  echo "Please start the bot first:"
  echo "  Terminal 1: make dev"
  echo "  Terminal 2: ./test_security.sh"
  echo ""
  exit 1
fi
echo -e "${GREEN}✅ Bot is running${NC}"
echo ""

# ============================================================
# Test Suite 1: QR Pairing Endpoint Authentication
# ============================================================
echo -e "${BLUE}═══ Test Suite 1: QR Pairing Authentication ═══${NC}"
echo ""

# Test 1: Pairing tanpa token
response=$(curl -s -o /dev/null -w "%{http_code}" $BASE_URL/pairing)
run_test "1.1" "Pairing tanpa token (should reject)" "401" "$response"

# Test 2: Pairing dengan token salah
response=$(curl -s -o /dev/null -w "%{http_code}" \
  $BASE_URL/pairing -H "X-Pairing-Token: wrong-token")
run_test "1.2" "Pairing dengan token salah (should reject)" "401" "$response"

# Test 3: Pairing dengan token benar
response=$(curl -s -o /dev/null -w "%{http_code}" \
  $BASE_URL/pairing -H "X-Pairing-Token: $VALID_TOKEN")
run_test "1.3" "Pairing dengan token benar (should accept)" "200" "$response"

# Test 4: Pairing dengan Authorization header (fallback)
response=$(curl -s -o /dev/null -w "%{http_code}" \
  $BASE_URL/pairing -H "Authorization: $VALID_TOKEN")
run_test "1.4" "Pairing dengan Authorization header (fallback)" "200" "$response"

# ============================================================
# Test Suite 2: Rate Limiting (Brute Force Protection)
# ============================================================
echo -e "${BLUE}═══ Test Suite 2: Rate Limiting (Brute Force) ═══${NC}"
echo ""
echo -e "${YELLOW}Test 2.1:${NC} Simulating brute force attack (10 attempts)"

# Reset dengan successful auth dulu
curl -s -o /dev/null $BASE_URL/pairing -H "X-Pairing-Token: $VALID_TOKEN"
sleep 1

pass_count=0
fail_count=0

for i in {1..10}; do
  response=$(curl -s -o /dev/null -w "%{http_code}" \
    $BASE_URL/pairing -H "X-Pairing-Token: brute-force-$i")
  
  if [ $i -le 5 ]; then
    expected="401"
    status_text="Unauthorized"
  else
    expected="429"
    status_text="Rate Limited"
  fi
  
  if [ "$response" = "$expected" ]; then
    echo -e "  Attempt $i: ${GREEN}✅ $response ($status_text)${NC}"
    ((pass_count++))
  else
    echo -e "  Attempt $i: ${RED}❌ $response (expected $expected)${NC}"
    ((fail_count++))
  fi
  sleep 0.2
done

if [ $fail_count -eq 0 ]; then
  echo -e "  ${GREEN}✅ PASS${NC} - Rate limiting works correctly"
  ((passed++))
else
  echo -e "  ${RED}❌ FAIL${NC} - Rate limiting not working as expected"
  ((failed++))
fi
echo ""

# Test 2.2: Verify IP still blocked
echo -e "${YELLOW}Test 2.2:${NC} Verify IP still blocked after 5 attempts"
response=$(curl -s -o /dev/null -w "%{http_code}" \
  $BASE_URL/pairing -H "X-Pairing-Token: another-wrong-token")
run_test "2.2" "Request after block (should be rate limited)" "429" "$response"

# Test 2.3: Successful auth resets counter
echo -e "${YELLOW}Test 2.3:${NC} Successful auth resets failed attempts counter"
# Wait for block to expire or use correct token to reset
curl -s -o /dev/null $BASE_URL/pairing -H "X-Pairing-Token: $VALID_TOKEN"
sleep 1
# Try wrong token again (should get 401, not 429)
response=$(curl -s -o /dev/null -w "%{http_code}" \
  $BASE_URL/pairing -H "X-Pairing-Token: test-after-reset")
run_test "2.3" "Wrong token after successful auth (counter reset)" "401" "$response"

# ============================================================
# Test Suite 3: QR API Endpoint (AJAX Polling)
# ============================================================
echo -e "${BLUE}═══ Test Suite 3: QR API Endpoint (Polling) ═══${NC}"
echo ""

# Reset counter
curl -s -o /dev/null $BASE_URL/pairing -H "X-Pairing-Token: $VALID_TOKEN"
sleep 1

# Test 3.1: API tanpa token
response=$(curl -s -o /dev/null -w "%{http_code}" $BASE_URL/pairing/qr)
run_test "3.1" "QR API tanpa token (should reject)" "401" "$response"

# Test 3.2: API dengan token benar
response=$(curl -s -o /dev/null -w "%{http_code}" \
  $BASE_URL/pairing/qr -H "X-Pairing-Token: $VALID_TOKEN")
run_test "3.2" "QR API dengan token benar (should accept)" "200" "$response"

# ============================================================
# Test Suite 4: Webhook Secret Validation
# ============================================================
echo -e "${BLUE}═══ Test Suite 4: Webhook Secret Validation ═══${NC}"
echo ""

# Test 4.1: Webhook tanpa secret
response=$(curl -s -o /dev/null -w "%{http_code}" \
  -X POST $BASE_URL/webhook/payment \
  -H "Content-Type: application/json" \
  -d '{"app":"id.dana","title":"Test","body":"Rp1000"}')
run_test "4.1" "Webhook tanpa secret (should reject)" "401" "$response"

# Test 4.2: Webhook dengan secret salah
response=$(curl -s -o /dev/null -w "%{http_code}" \
  -X POST $BASE_URL/webhook/payment \
  -H "X-Webhook-Secret: wrong-secret" \
  -H "Content-Type: application/json" \
  -d '{"app":"id.dana","title":"Test","body":"Rp1000"}')
run_test "4.2" "Webhook dengan secret salah (should reject)" "401" "$response"

# Test 4.3: Webhook dengan secret benar
response=$(curl -s -o /dev/null -w "%{http_code}" \
  -X POST $BASE_URL/webhook/payment \
  -H "X-Webhook-Secret: $VALID_SECRET" \
  -H "Content-Type: application/json" \
  -d '{"app":"id.dana","title":"Test","body":"Rp1000"}')
run_test "4.3" "Webhook dengan secret benar (should accept)" "200" "$response"

# Test 4.4: Webhook dengan Authorization header (fallback)
response=$(curl -s -o /dev/null -w "%{http_code}" \
  -X POST $BASE_URL/webhook/payment \
  -H "Authorization: $VALID_SECRET" \
  -H "Content-Type: application/json" \
  -d '{"app":"id.dana","title":"Test","body":"Rp1000"}')
run_test "4.4" "Webhook dengan Authorization header (fallback)" "200" "$response"

# ============================================================
# Test Summary
# ============================================================
echo ""
echo -e "${BLUE}╔════════════════════════════════════════════╗${NC}"
echo -e "${BLUE}║           📊 Test Summary                  ║${NC}"
echo -e "${BLUE}╚════════════════════════════════════════════╝${NC}"
echo ""

total=$((passed + failed))
pass_rate=$((passed * 100 / total))

echo -e "Total Tests:  ${BLUE}$total${NC}"
echo -e "Passed:       ${GREEN}$passed${NC}"
echo -e "Failed:       ${RED}$failed${NC}"
echo -e "Pass Rate:    ${BLUE}$pass_rate%${NC}"
echo ""

if [ $failed -eq 0 ]; then
  echo -e "${GREEN}🎉 All tests passed! Security is working correctly.${NC}"
  echo ""
  echo -e "${BLUE}📖 Next steps:${NC}"
  echo "  1. Check security logs: tail -f tmp/main.log | grep SECURITY"
  echo "  2. Review SECURITY_TESTING.md for detailed test scenarios"
  echo "  3. Deploy to production and test in prod environment"
  exit 0
else
  echo -e "${RED}⚠️  Some tests failed. Please review the implementation.${NC}"
  echo ""
  echo -e "${BLUE}🐛 Debugging tips:${NC}"
  echo "  1. Check bot logs: tail -f tmp/main.log"
  echo "  2. Verify .env configuration matches test expectations"
  echo "  3. Ensure PAIRING_TOKEN=$VALID_TOKEN in .env"
  echo "  4. Ensure WEBHOOK_SECRET=$VALID_SECRET in .env"
  echo "  5. Restart bot if config was changed"
  exit 1
fi
