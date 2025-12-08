#!/bin/bash

# ============================================================
# Web Security Testing - Quick Start
# ============================================================
# Script untuk membantu memulai testing keamanan via browser
# Usage: ./scripts/web_security_test.sh

set -e

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color
BOLD='\033[1m'

# Banner
echo -e "${BLUE}${BOLD}"
echo "╔════════════════════════════════════════════════════════════╗"
echo "║       🌐 BotjanWeb - Web Security Testing Suite           ║"
echo "║          Interactive Browser-Based Testing                 ║"
echo "╚════════════════════════════════════════════════════════════╝"
echo -e "${NC}"

# Check if bot is running
echo -e "${YELLOW}📡 Checking if bot is running...${NC}"
if curl -s http://localhost:8080/health > /dev/null 2>&1; then
    echo -e "${GREEN}✅ Bot is running on http://localhost:8080${NC}"
else
    echo -e "${RED}❌ Bot is NOT running!${NC}"
    echo ""
    echo -e "${YELLOW}Please start the bot first:${NC}"
    echo "  cd /home/exernia/Proyek/go/botjanweb"
    echo "  make dev"
    echo ""
    exit 1
fi

# Check .env configuration
echo ""
echo -e "${YELLOW}⚙️  Checking .env configuration...${NC}"
if [ -f .env ]; then
    PAIRING_TOKEN=$(grep PAIRING_TOKEN .env | cut -d '=' -f2)
    WEBHOOK_SECRET=$(grep WEBHOOK_SECRET .env | cut -d '=' -f2)
    WEBHOOK_PORT=$(grep WEBHOOK_PORT .env | cut -d '=' -f2)
    
    echo -e "${GREEN}✅ .env file found${NC}"
    echo "   PAIRING_TOKEN: ${PAIRING_TOKEN:0:20}... (${#PAIRING_TOKEN} chars)"
    echo "   WEBHOOK_SECRET: ${WEBHOOK_SECRET:0:20}... (${#WEBHOOK_SECRET} chars)"
    echo "   WEBHOOK_PORT: ${WEBHOOK_PORT}"
else
    echo -e "${RED}❌ .env file not found!${NC}"
    echo ""
    echo -e "${YELLOW}Please create .env file:${NC}"
    echo "  cp .env.security-testing .env"
    echo ""
    exit 1
fi

# Main menu
echo ""
echo -e "${BLUE}${BOLD}What would you like to do?${NC}"
echo ""
echo "1) 🚀 Open Interactive Testing Dashboard (HTML)"
echo "2) 📱 Open Pairing Page (requires ModHeader extension)"
echo "3) 🧪 Run Automated Tests (curl-based)"
echo "4) 📖 Open Testing Guide (Markdown)"
echo "5) ⚙️  Show Current Configuration"
echo "6) 🔧 Install ModHeader Extension (instructions)"
echo "7) 📊 Show Available Endpoints"
echo "8) ❌ Exit"
echo ""
read -p "Select option [1-8]: " choice

case $choice in
    1)
        echo ""
        echo -e "${GREEN}🚀 Opening Interactive Testing Dashboard...${NC}"
        echo ""
        
        # Check if xdg-open is available
        if command -v xdg-open > /dev/null; then
            xdg-open "docs/security-test-dashboard.html"
            echo -e "${GREEN}✅ Dashboard opened in browser${NC}"
        elif command -v firefox > /dev/null; then
            firefox "docs/security-test-dashboard.html" &
            echo -e "${GREEN}✅ Dashboard opened in Firefox${NC}"
        elif command -v google-chrome > /dev/null; then
            google-chrome "docs/security-test-dashboard.html" &
            echo -e "${GREEN}✅ Dashboard opened in Chrome${NC}"
        else
            echo -e "${YELLOW}⚠️  No browser command found${NC}"
            echo "Please open manually:"
            echo "  file://$(pwd)/docs/security-test-dashboard.html"
        fi
        
        echo ""
        echo -e "${BLUE}${BOLD}Dashboard Features:${NC}"
        echo "  • Interactive test execution"
        echo "  • Real-time test results"
        echo "  • Test statistics tracking"
        echo "  • Configuration management"
        echo ""
        echo -e "${YELLOW}💡 Tip: Configure your pairing token in the dashboard${NC}"
        ;;
        
    2)
        echo ""
        echo -e "${YELLOW}📱 Opening Pairing Page...${NC}"
        echo ""
        echo -e "${RED}⚠️  WARNING: This will open without authentication header!${NC}"
        echo -e "${YELLOW}Expected result: 401 Unauthorized${NC}"
        echo ""
        read -p "Continue? [y/N]: " confirm
        
        if [[ $confirm == [yY] ]]; then
            if command -v xdg-open > /dev/null; then
                xdg-open "http://localhost:8080/pairing"
            elif command -v firefox > /dev/null; then
                firefox "http://localhost:8080/pairing" &
            elif command -v google-chrome > /dev/null; then
                google-chrome "http://localhost:8080/pairing" &
            fi
            
            echo ""
            echo -e "${BLUE}${BOLD}To access with authentication:${NC}"
            echo "1. Install ModHeader extension (option 6)"
            echo "2. Configure header:"
            echo "   Name: X-Pairing-Token"
            echo "   Value: $PAIRING_TOKEN"
            echo "3. Refresh the page"
        fi
        ;;
        
    3)
        echo ""
        echo -e "${GREEN}🧪 Running Automated Tests...${NC}"
        echo ""
        
        if [ -f scripts/test_security.sh ]; then
            chmod +x scripts/test_security.sh
            ./scripts/test_security.sh
        else
            echo -e "${RED}❌ test_security.sh not found${NC}"
        fi
        ;;
        
    4)
        echo ""
        echo -e "${GREEN}📖 Opening Testing Guide...${NC}"
        echo ""
        
        if command -v code > /dev/null; then
            code docs/WEB_SECURITY_TESTING.md
            echo -e "${GREEN}✅ Guide opened in VS Code${NC}"
        elif command -v nano > /dev/null; then
            nano docs/WEB_SECURITY_TESTING.md
        elif command -v less > /dev/null; then
            less docs/WEB_SECURITY_TESTING.md
        else
            cat docs/WEB_SECURITY_TESTING.md
        fi
        ;;
        
    5)
        echo ""
        echo -e "${BLUE}${BOLD}⚙️  Current Configuration${NC}"
        echo ""
        echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
        echo -e "${YELLOW}Environment Variables:${NC}"
        echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
        grep -E "PAIRING_TOKEN|WEBHOOK_SECRET|WEBHOOK_PORT|WEBHOOK_ENABLED" .env | while read line; do
            key=$(echo "$line" | cut -d '=' -f1)
            value=$(echo "$line" | cut -d '=' -f2)
            
            # Mask secrets
            if [[ $key == *"TOKEN"* ]] || [[ $key == *"SECRET"* ]]; then
                masked="${value:0:20}... (${#value} chars)"
                echo "  $key = $masked"
            else
                echo "  $key = $value"
            fi
        done
        echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
        echo ""
        echo -e "${YELLOW}Endpoints:${NC}"
        echo "  • Health Check: http://localhost:$WEBHOOK_PORT/health"
        echo "  • Pairing Page: http://localhost:$WEBHOOK_PORT/pairing"
        echo "  • QR Endpoint:  http://localhost:$WEBHOOK_PORT/pairing/qr"
        echo "  • Webhook:      http://localhost:$WEBHOOK_PORT/webhook/payment"
        echo ""
        ;;
        
    6)
        echo ""
        echo -e "${BLUE}${BOLD}🔧 Installing ModHeader Extension${NC}"
        echo ""
        echo -e "${YELLOW}ModHeader allows you to modify HTTP headers${NC}"
        echo ""
        echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
        echo -e "${BOLD}For Chrome/Brave/Edge:${NC}"
        echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
        echo "1. Open Chrome Web Store:"
        echo "   https://chrome.google.com/webstore/detail/modheader/idgpnmonknjnojddfkpgkljpfnnfcklj"
        echo "2. Click 'Add to Chrome'"
        echo "3. Pin the extension to toolbar"
        echo ""
        echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
        echo -e "${BOLD}For Firefox:${NC}"
        echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
        echo "1. Open Firefox Add-ons:"
        echo "   https://addons.mozilla.org/en-US/firefox/addon/modify-header-value/"
        echo "2. Click 'Add to Firefox'"
        echo ""
        echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
        echo -e "${BOLD}Configuration After Install:${NC}"
        echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
        echo "1. Click ModHeader icon in toolbar"
        echo "2. Add new header:"
        echo -e "   Name: ${GREEN}X-Pairing-Token${NC}"
        echo -e "   Value: ${GREEN}$PAIRING_TOKEN${NC}"
        echo "3. Make sure ModHeader is enabled (icon colored)"
        echo "4. Visit http://localhost:8080/pairing"
        echo ""
        ;;
        
    7)
        echo ""
        echo -e "${BLUE}${BOLD}📊 Available Endpoints${NC}"
        echo ""
        echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
        echo -e "${GREEN}Public Endpoints (No Auth):${NC}"
        echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
        echo "  GET  /health        - Health check"
        echo "  GET  /ready         - Readiness check"
        echo "  GET  /healthz       - Health check alias"
        echo "  GET  /readyz        - Readiness check alias"
        echo ""
        echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
        echo -e "${YELLOW}Protected Endpoints (Requires Auth):${NC}"
        echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
        echo "  GET  /pairing       - QR pairing page"
        echo "                        Header: X-Pairing-Token"
        echo ""
        echo "  GET  /pairing/qr    - QR code API (AJAX)"
        echo "                        Header: X-Pairing-Token"
        echo ""
        echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
        echo -e "${RED}Webhook Endpoints (Requires Signature):${NC}"
        echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
        echo "  POST /webhook/payment - Payment webhook"
        echo "                          Header: X-Hub-Signature"
        echo ""
        echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
        echo -e "${BLUE}Testing Commands:${NC}"
        echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
        echo ""
        echo "# Health check (no auth)"
        echo "curl http://localhost:8080/health"
        echo ""
        echo "# Pairing page (with auth)"
        echo "curl -H \"X-Pairing-Token: $PAIRING_TOKEN\" \\"
        echo "  http://localhost:8080/pairing"
        echo ""
        echo "# QR endpoint (with auth)"
        echo "curl -H \"X-Pairing-Token: $PAIRING_TOKEN\" \\"
        echo "  http://localhost:8080/pairing/qr"
        echo ""
        ;;
        
    8)
        echo ""
        echo -e "${GREEN}👋 Goodbye!${NC}"
        echo ""
        exit 0
        ;;
        
    *)
        echo ""
        echo -e "${RED}❌ Invalid option${NC}"
        echo ""
        exit 1
        ;;
esac

echo ""
echo -e "${GREEN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo -e "${BLUE}${BOLD}Need help?${NC}"
echo "  • Read guide: docs/WEB_SECURITY_TESTING.md"
echo "  • Run automated tests: ./scripts/test_security.sh"
echo "  • Open dashboard: docs/security-test-dashboard.html"
echo -e "${GREEN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo ""
