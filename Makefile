.PHONY: help print go-dep go-mod-upgrade gen gen-testdata run bin test clean kill version release update

# Paths
MAKEFILE_DIR := $(dir $(abspath $(lastword $(MAKEFILE_LIST))))
BIN_DIR := $(MAKEFILE_DIR).bin
DIST_DIR := $(MAKEFILE_DIR).dist
PB_CMD_DIR := $(MAKEFILE_DIR)pkg/cmd/pocketbase
MIGRATIONS_DIR := $(PB_CMD_DIR)/pb_migrations
CODEGEN_DIR := $(MAKEFILE_DIR)pkg/pb/codegen
DATA := $(PB_CMD_DIR)/pb_data
TEMPLATE := $(CODEGEN_DIR)/_templates/google_tokens.go
MODELS := $(CODEGEN_DIR)/models
BINARY := $(BIN_DIR)/wellknown-pb

# GitHub
GH_OWNER := joeblew999
GH_REPO := wellknown

.DEFAULT_GOAL := help

## help: Show this help message
help:
	@echo "════════════════════════════════════════════════════════════════"
	@echo "  Wellknown Development Toolkit"
	@echo "════════════════════════════════════════════════════════════════"
	@echo ""
	@grep -E '^##' $(MAKEFILE_LIST) | sed 's/^## //' | awk -F: '{printf "  %-20s %s\n", $$1, $$2}'
	@echo ""
	@echo "════════════════════════════════════════════════════════════════"
	@echo "💡 Quick Start:"
	@echo "   make go-dep        Install development tools"
	@echo "   make run           Start PocketBase (port 8090)"
	@echo "   make gen           Generate template and models"
	@echo "════════════════════════════════════════════════════════════════"


## print: Show all Makefile variables and paths
print:
	@echo "=== Makefile Debug Info ==="
	@echo ""
	@echo "📁 Paths:"
	@echo "  MAKEFILE_DIR    = $(MAKEFILE_DIR)"
	@echo "  BIN_DIR         = $(BIN_DIR)"
	@echo "  DIST_DIR        = $(DIST_DIR)"
	@echo "  PB_CMD_DIR      = $(PB_CMD_DIR)"
	@echo "  MIGRATIONS_DIR  = $(MIGRATIONS_DIR)"
	@echo "  CODEGEN_DIR     = $(CODEGEN_DIR)"
	@echo ""
	@echo "📄 Files:"
	@echo "  BINARY          = $(BINARY)"
	@echo "  DATA            = $(DATA)"
	@echo "  TEMPLATE        = $(TEMPLATE)"
	@echo "  MODELS          = $(MODELS)"
	@echo ""
	@echo "🐙 GitHub:"
	@echo "  GH_OWNER        = $(GH_OWNER)"
	@echo "  GH_REPO         = $(GH_REPO)"

## go-dep: Install development tools (pocketbase-gogen, gh)
go-dep:
	@echo "📦 Installing development tools..."
	go install github.com/snonky/pocketbase-gogen@v0.7.0
	go install github.com/cli/cli/v2/cmd/gh@latest
	@echo "✅ Tools installed"

## go-mod-upgrade: Upgrade Go module dependencies
go-mod-upgrade:
	go install github.com/oligot/go-mod-upgrade@latest
	go-mod-upgrade

## gen: Generate PocketBase template and type-safe models
gen:
	@echo "📝 Generating PocketBase template from schema..."
	@if [ ! -d "$(DATA)" ]; then \
		echo "❌ $(DATA) not found!"; \
		echo "   Run 'make serve' once to initialize PocketBase"; \
		exit 1; \
	fi
	@mkdir -p $(dir $(TEMPLATE))
	pocketbase-gogen template $(DATA) $(TEMPLATE) --package models
	@echo "✅ Template: $(TEMPLATE)"
	@echo ""
	@echo "🔧 Generating type-safe models from template..."
	pocketbase-gogen generate $(TEMPLATE) $(MODELS)/proxies.go --package models --utils --hooks
	@echo "✅ Models: $(MODELS)/{proxies,utils,proxy_hooks}.go"

## gen-testdata: Generate test data using Go reflection
gen-testdata:
	@echo "🔧 Generating test data..."
	@go run . gen-testdata -v

## run: Run PocketBase server (port 8090)
run:
	@echo "🚀 Starting PocketBase..."
	@echo "Admin UI: http://localhost:8090/_/"
	go run . pb

## bin: Build PocketBase server binary into BIN
bin:
	@echo "🏗️  Building PocketBase server..."
	@mkdir -p $(BIN_DIR)
	go build -o $(BINARY) .
	@echo "✅ Binary: $(BINARY)"

## test: Build and test PocketBase API endpoints
test: bin
	@echo "🧪 Testing PocketBase API endpoints..."
	@lsof -ti:8090 | xargs kill -9 2>/dev/null || true
	@sleep 1
	@$(BINARY) pb &
	@SERVER_PID=$$! && \
	sleep 4 && \
	echo "" && \
	echo "=== Testing Banking API ===" && \
	echo "" && \
	echo "1. POST /api/banking/accounts:" && \
	curl -s -X POST http://localhost:8090/api/banking/accounts \
	  -H "Content-Type: application/json" \
	  -d '{"user_id":"test123","account_number":"ACC001","account_name":"Checking","account_type":"checking","balance":1000.50,"currency":"USD","is_active":true}' \
	  -w "\nHTTP Status: %{http_code}\n" && \
	echo "" && \
	echo "2. GET /api/banking/accounts?user_id=test123:" && \
	curl -s "http://localhost:8090/api/banking/accounts?user_id=test123" -w "\nHTTP Status: %{http_code}\n" && \
	echo "" && \
	kill $$SERVER_PID 2>/dev/null || true

## clean: Clean build artifacts and generated files
clean:
	@echo "🧹 Cleaning generated files..."
	rm -rf $(MAKEFILE_DIR)tmp/ $(BIN_DIR) $(DIST_DIR) tests/e2e/generated/
	rm -f $(MODELS)/*.go
	@echo "✅ Cleaned"

## kill: Kill process on port 8090
kill:
	@echo "🔫 Killing processes on port 8090..."
	@lsof -ti:8090 | xargs kill -9 2>/dev/null || echo "   No processes found"
	@echo "✅ Port 8090 freed"

## version: Show current PocketBase binary version
version:
	@echo "📋 PocketBase binary version:"
	@if [ ! -f $(BINARY) ]; then \
		echo "❌ Binary not found: $(BINARY)"; \
		echo "   Run: make build"; \
		exit 1; \
	fi
	@strings $(BINARY) | grep "github.com/pocketbase/pocketbase/core.Version=" | head -1 | sed 's/.*Version=/   /' | tr -d '"'


## release: Build & create GitHub release (multi-platform)
release:
	@echo "🚀 Creating GitHub release..."
	@if ! command -v gh >/dev/null 2>&1; then \
		echo "⚙️  Installing GitHub CLI..."; \
		go install github.com/cli/cli/v2/cmd/gh@latest; \
	fi
	@if ! gh auth status >/dev/null 2>&1; then \
		echo "⚠️  GitHub CLI not authenticated"; \
		echo "   Run: gh auth login"; \
		exit 1; \
	fi
	@LAST_TAG=$$(git describe --tags --abbrev=0 2>/dev/null || echo "v0.0.0"); \
	LAST_NUM=$$(echo $$LAST_TAG | cut -d. -f3); \
	NEXT_NUM=$$(($$LAST_NUM + 1)); \
	NEXT_TAG=$$(echo $$LAST_TAG | sed "s/\\.$$LAST_NUM$$/.$${NEXT_NUM}/"); \
	echo "📌 Last tag: $$LAST_TAG"; \
	echo "📌 Next tag: $$NEXT_TAG"; \
	read -p "Use $$NEXT_TAG? [Y/n/custom]: " CONFIRM; \
	if [ -z "$$CONFIRM" ] || [ "$$CONFIRM" = "y" ] || [ "$$CONFIRM" = "Y" ]; then \
		VERSION=$$NEXT_TAG; \
	elif [ "$$CONFIRM" = "n" ] || [ "$$CONFIRM" = "N" ]; then \
		read -p "Enter version tag: " VERSION; \
	else \
		VERSION=$$CONFIRM; \
	fi; \
	echo "📦 Building for multiple platforms with version $$VERSION..."; \
	mkdir -p $(DIST_DIR); \
	LDFLAGS="-X github.com/pocketbase/pocketbase/core.Version=$$VERSION"; \
	cd $(PB_CMD_DIR) && \
	CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build -ldflags "$$LDFLAGS" -o $(DIST_DIR)/wellknown-pb-darwin-arm64 . & \
	CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -ldflags "$$LDFLAGS" -o $(DIST_DIR)/wellknown-pb-darwin-amd64 . & \
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags "$$LDFLAGS" -o $(DIST_DIR)/wellknown-pb-linux-amd64 . & \
	CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -ldflags "$$LDFLAGS" -o $(DIST_DIR)/wellknown-pb-linux-arm64 . & \
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -ldflags "$$LDFLAGS" -o $(DIST_DIR)/wellknown-pb-windows-amd64.exe . & \
	CGO_ENABLED=0 GOOS=windows GOARCH=arm64 go build -ldflags "$$LDFLAGS" -o $(DIST_DIR)/wellknown-pb-windows-arm64.exe . & \
	wait && \
	cd $(DIST_DIR) && \
	cp wellknown-pb-darwin-arm64 wellknown-pb && zip wellknown-pb_darwin_arm64.zip wellknown-pb && rm wellknown-pb && \
	cp wellknown-pb-darwin-amd64 wellknown-pb && zip wellknown-pb_darwin_amd64.zip wellknown-pb && rm wellknown-pb && \
	cp wellknown-pb-linux-amd64 wellknown-pb && zip wellknown-pb_linux_amd64.zip wellknown-pb && rm wellknown-pb && \
	cp wellknown-pb-linux-arm64 wellknown-pb && zip wellknown-pb_linux_arm64.zip wellknown-pb && rm wellknown-pb && \
	cp wellknown-pb-windows-amd64.exe wellknown-pb.exe && zip wellknown-pb_windows_amd64.zip wellknown-pb.exe && rm wellknown-pb.exe && \
	cp wellknown-pb-windows-arm64.exe wellknown-pb.exe && zip wellknown-pb_windows_arm64.zip wellknown-pb.exe && rm wellknown-pb.exe && \
	cd $(MAKEFILE_DIR) && \
	git tag -a "$$VERSION" -m "Release $$VERSION" && \
	git push origin "$$VERSION" && \
	gh release create "$$VERSION" $(DIST_DIR)/wellknown-pb*.zip \
		--title "PocketBase Server $$VERSION" \
		--notes "Release $$VERSION" && \
	echo "✅ Release $$VERSION created!"


## update: Update PocketBase binary from GitHub releases
update:
	@echo "⬇️  Updating from GitHub releases..."
	@if [ ! -f $(BINARY) ]; then \
		echo "❌ Binary not found: $(BINARY)"; \
		echo "   Run: make build"; \
		exit 1; \
	fi
	$(BINARY) update
	@echo "✅ Update complete!"