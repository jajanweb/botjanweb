# ============================================================================
# BotJanWeb Makefile
# ============================================================================
# Streamlined development workflow commands for BotJanWeb project.
#
# Usage:
#   make help      - Show all available commands
#   make air       - Start development server with hot-reload
#   make ngrok     - Expose webhook endpoint via ngrok
#   make dev       - Start air + ngrok in parallel (requires tmux)
#
# Requirements:
#   - Go 1.25+
#   - air (go install github.com/air-verse/air@latest)
#   - ngrok (https://ngrok.com/download)
#   - tmux (optional, for `make dev`)
# ============================================================================

# Project configuration
APP_NAME := botjanweb
MAIN_PKG := ./cmd/botjanweb
BUILD_DIR := ./tmp
BINARY := $(BUILD_DIR)/main
ASSETS_DIR := ./assets

# Webhook configuration
WEBHOOK_PORT := 9090

# Go commands
GO := go
GOFMT := gofmt
GOTEST := $(GO) test
GOBUILD := $(GO) build
GOMOD := $(GO) mod

# Air binary (cosmtrek/air) - check PATH first, fallback to ~/go/bin
AIR := $(shell command -v air 2>/dev/null || echo "$(HOME)/go/bin/air")

# Colors for output
COLOR_RESET := \033[0m
COLOR_GREEN := \033[32m
COLOR_YELLOW := \033[33m
COLOR_BLUE := \033[34m
COLOR_CYAN := \033[36m

# ============================================================================
# DEVELOPMENT COMMANDS
# ============================================================================

.PHONY: air
air: ## Start development server with hot-reload (air)
	@echo "$(COLOR_CYAN)🚀 Starting Air development server...$(COLOR_RESET)"
	@if [ -x "$(AIR)" ]; then \
		$(AIR); \
	else \
		echo "$(COLOR_YELLOW)⚠️  Air not found. Installing...$(COLOR_RESET)"; \
		go install github.com/air-verse/air@latest; \
		$(HOME)/go/bin/air; \
	fi

.PHONY: ngrok
ngrok: ## Expose webhook endpoint via ngrok (port 9090)
	@echo "$(COLOR_CYAN)🌐 Starting ngrok tunnel on port $(WEBHOOK_PORT)...$(COLOR_RESET)"
	@echo "$(COLOR_YELLOW)📋 Copy the HTTPS URL to your webhook notification app$(COLOR_RESET)"
	@ngrok http $(WEBHOOK_PORT)

.PHONY: dev
dev: ## Start air + ngrok in parallel using tmux
	@if command -v tmux >/dev/null 2>&1; then \
		echo "$(COLOR_CYAN)🔧 Starting development environment (air + ngrok)...$(COLOR_RESET)"; \
		tmux new-session -d -s botjanweb-dev 'make air' \; \
			split-window -h 'sleep 2 && make ngrok' \; \
			attach; \
	else \
		echo "$(COLOR_YELLOW)⚠️  tmux not found. Install tmux or run 'make air' and 'make ngrok' in separate terminals.$(COLOR_RESET)"; \
		exit 1; \
	fi

.PHONY: dev-stop
dev-stop: ## Stop tmux development session
	@tmux kill-session -t botjanweb-dev 2>/dev/null || echo "No active session"

# ============================================================================
# BUILD COMMANDS
# ============================================================================

.PHONY: build
build: ## Build the application binary
	@echo "$(COLOR_GREEN)🔨 Building $(APP_NAME)...$(COLOR_RESET)"
	@mkdir -p $(BUILD_DIR)
	@$(GOBUILD) -o $(BINARY) $(MAIN_PKG)
	@echo "$(COLOR_GREEN)✅ Binary built: $(BINARY)$(COLOR_RESET)"

.PHONY: build-prod
build-prod: ## Build optimized production binary
	@echo "$(COLOR_GREEN)🔨 Building production binary...$(COLOR_RESET)"
	@mkdir -p $(BUILD_DIR)
	@CGO_ENABLED=1 $(GOBUILD) -ldflags="-s -w" -o $(BINARY) $(MAIN_PKG)
	@echo "$(COLOR_GREEN)✅ Production binary built: $(BINARY)$(COLOR_RESET)"

.PHONY: run
run: build ## Build and run the application
	@echo "$(COLOR_GREEN)🚀 Running $(APP_NAME)...$(COLOR_RESET)"
	@$(BINARY)

.PHONY: clean
clean: ## Clean build artifacts
	@echo "$(COLOR_YELLOW)🧹 Cleaning build artifacts...$(COLOR_RESET)"
	@rm -rf $(BUILD_DIR)
	@rm -f $(APP_NAME)
	@echo "$(COLOR_GREEN)✅ Cleaned$(COLOR_RESET)"

# ============================================================================
# TEST COMMANDS
# ============================================================================

.PHONY: test
test: ## Run all tests
	@echo "$(COLOR_BLUE)🧪 Running tests...$(COLOR_RESET)"
	@$(GOTEST) -v ./...

.PHONY: test-short
test-short: ## Run tests in short mode
	@echo "$(COLOR_BLUE)🧪 Running short tests...$(COLOR_RESET)"
	@$(GOTEST) -v -short ./...

.PHONY: test-cover
test-cover: ## Run tests with coverage report
	@echo "$(COLOR_BLUE)🧪 Running tests with coverage...$(COLOR_RESET)"
	@$(GOTEST) -v -coverprofile=coverage.out ./...
	@$(GO) tool cover -html=coverage.out -o coverage.html
	@echo "$(COLOR_GREEN)✅ Coverage report: coverage.html$(COLOR_RESET)"

.PHONY: test-race
test-race: ## Run tests with race detector
	@echo "$(COLOR_BLUE)🧪 Running tests with race detector...$(COLOR_RESET)"
	@$(GOTEST) -v -race ./...

# ============================================================================
# CODE QUALITY COMMANDS
# ============================================================================

.PHONY: fmt
fmt: ## Format all Go source files
	@echo "$(COLOR_BLUE)📝 Formatting code...$(COLOR_RESET)"
	@$(GOFMT) -s -w .
	@echo "$(COLOR_GREEN)✅ Code formatted$(COLOR_RESET)"

.PHONY: lint
lint: ## Run golangci-lint (if installed)
	@if command -v golangci-lint >/dev/null 2>&1; then \
		echo "$(COLOR_BLUE)🔍 Running linter...$(COLOR_RESET)"; \
		golangci-lint run ./...; \
	else \
		echo "$(COLOR_YELLOW)⚠️  golangci-lint not found. Install: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest$(COLOR_RESET)"; \
	fi

.PHONY: vet
vet: ## Run go vet
	@echo "$(COLOR_BLUE)🔍 Running go vet...$(COLOR_RESET)"
	@$(GO) vet ./...

.PHONY: check
check: fmt vet test-short ## Run fmt, vet, and short tests

# ============================================================================
# DEPENDENCY COMMANDS
# ============================================================================

.PHONY: deps
deps: ## Download and tidy dependencies
	@echo "$(COLOR_BLUE)📦 Downloading dependencies...$(COLOR_RESET)"
	@$(GOMOD) download
	@$(GOMOD) tidy
	@echo "$(COLOR_GREEN)✅ Dependencies updated$(COLOR_RESET)"

.PHONY: deps-upgrade
deps-upgrade: ## Upgrade all dependencies
	@echo "$(COLOR_BLUE)📦 Upgrading dependencies...$(COLOR_RESET)"
	@$(GO) get -u ./...
	@$(GOMOD) tidy
	@echo "$(COLOR_GREEN)✅ Dependencies upgraded$(COLOR_RESET)"

# ============================================================================
# TOOL INSTALLATION
# ============================================================================

.PHONY: install-tools
install-tools: ## Install development tools (air, golangci-lint)
	@echo "$(COLOR_BLUE)🔧 Installing development tools...$(COLOR_RESET)"
	@go install github.com/air-verse/air@latest
	@go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	@echo "$(COLOR_GREEN)✅ Tools installed$(COLOR_RESET)"
	@echo ""
	@echo "$(COLOR_YELLOW)📋 Note: Install ngrok manually from https://ngrok.com/download$(COLOR_RESET)"
	@echo "$(COLOR_YELLOW)📋 Make sure ~/go/bin is in your PATH$(COLOR_RESET)"

.PHONY: install-air
install-air: ## Install air hot-reload tool
	@echo "$(COLOR_BLUE)🔧 Installing air (air-verse/air)...$(COLOR_RESET)"
	@go install github.com/air-verse/air@latest
	@echo "$(COLOR_GREEN)✅ Air installed to ~/go/bin/air$(COLOR_RESET)"
	@echo "$(COLOR_YELLOW)📋 Add to PATH: export PATH=\$$PATH:\$$HOME/go/bin$(COLOR_RESET)"

# ============================================================================
# UTILITY COMMANDS
# ============================================================================

.PHONY: qris-preview
qris-preview: ## Run QRIS preview tool
	@echo "$(COLOR_CYAN)🖼️  Running QRIS preview...$(COLOR_RESET)"
	@$(GO) run ./cmd/qrispreview

.PHONY: qris-extract
qris-extract: ## Run QRIS extraction tool
	@echo "$(COLOR_CYAN)📷 Running QRIS extraction...$(COLOR_RESET)"
	@$(GO) run ./cmd/qrisextract

.PHONY: env-check
env-check: ## Check required environment variables
	@echo "$(COLOR_BLUE)🔍 Checking environment variables...$(COLOR_RESET)"
	@if [ -f .env ]; then \
		echo "$(COLOR_GREEN)✅ .env file exists$(COLOR_RESET)"; \
		echo ""; \
		echo "Required variables:"; \
		grep -E "^(GROUP_JID|QRIS_STATIC_PAYLOAD|ALLOWED_SENDERS)" .env 2>/dev/null | sed 's/=.*/=***/' || echo "  (not found)"; \
		echo ""; \
		echo "Optional variables:"; \
		grep -E "^(SHEETS_ENABLED|WEBHOOK_ENABLED)" .env 2>/dev/null || echo "  (not set)"; \
	else \
		echo "$(COLOR_YELLOW)⚠️  .env file not found. Copy from .env.example$(COLOR_RESET)"; \
	fi

.PHONY: setup
setup: deps install-tools ## Initial project setup
	@echo "$(COLOR_GREEN)✅ Project setup complete!$(COLOR_RESET)"
	@echo ""
	@echo "Next steps:"
	@echo "  1. Copy .env.example to .env and configure"
	@echo "  2. Run 'make air' to start development server"
	@echo "  3. Run 'make ngrok' in another terminal for webhook"

# ============================================================================
# HELP
# ============================================================================

.PHONY: help
help: ## Show this help message
	@echo ""
	@echo "$(COLOR_CYAN)╔════════════════════════════════════════════════════════════════╗$(COLOR_RESET)"
	@echo "$(COLOR_CYAN)║              BotJanWeb Development Commands                    ║$(COLOR_RESET)"
	@echo "$(COLOR_CYAN)╚════════════════════════════════════════════════════════════════╝$(COLOR_RESET)"
	@echo ""
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  $(COLOR_GREEN)%-15s$(COLOR_RESET) %s\n", $$1, $$2}'
	@echo ""
	@echo "$(COLOR_YELLOW)Quick Start:$(COLOR_RESET)"
	@echo "  make air      Start development server with hot-reload"
	@echo "  make ngrok    Expose webhook on port $(WEBHOOK_PORT)"
	@echo "  make dev      Start both (requires tmux)"
	@echo ""

# Default target
.DEFAULT_GOAL := help
