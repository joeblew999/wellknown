.PHONY: help print go-dep go-mod-upgrade gen gen-testdata run bin test health clean kill version env-list env-validate env-example env-generate-example env-sync env-sync-dockerfile env-sync-flytoml env-generate-local env-generate-production env-sync-secrets env-sync-secrets-production release update fly-auth fly-launch fly-volume fly-secrets fly-deploy fly-status fly-logs fly-ssh fly-destroy certs-install certs-init certs-generate certs-clean certs-status

# Paths
MAKEFILE_DIR := $(dir $(abspath $(lastword $(MAKEFILE_LIST))))
BIN_DIR := $(MAKEFILE_DIR).bin
DIST_DIR := $(MAKEFILE_DIR).dist

# Multi-service data directory structure
# All runtime data is organized under .data/ (gitignored)
DATA_DIR := $(MAKEFILE_DIR).data
# PocketBase runtime: databases, storage/, backups/
PB_DATA_DIR := $(DATA_DIR)/pb
# Future: NATS JetStream state for HA
NATS_DATA_DIR := $(DATA_DIR)/nats
# Development HTTPS certificates (mkcert-generated)
CERTS_DIR := $(DATA_DIR)/certs

# PocketBase source directories (version controlled)
# Database migrations (Go) - Note: pb_hooks and *.go files moved to main.go
MIGRATIONS_DIR := $(MAKEFILE_DIR)pkg/cmd/pocketbase/pb_migrations
CODEGEN_DIR := $(MAKEFILE_DIR)pkg/pb/codegen

# Generated files
# Template contains ALL collections (users, google_tokens, accounts, transactions, etc.)
TEMPLATE := $(CODEGEN_DIR)/_templates/schema.go
MODELS := $(CODEGEN_DIR)/models
BINARY := $(BIN_DIR)/wellknown-pb

# GitHub
GH_OWNER := joeblew999
GH_REPO := wellknown

# Fly.io
FLY_APP_NAME := $(shell grep "^app = " fly.toml 2>/dev/null | cut -d'"' -f2)
FLY_REGION := $(shell grep "^primary_region = " fly.toml 2>/dev/null | cut -d'"' -f2)
FLY := $(shell command -v flyctl 2>/dev/null || command -v fly 2>/dev/null)

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
	@echo "   make run           Start unified server (API + Demo UI on port 8090)"
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
	@echo "  DATA_DIR        = $(DATA_DIR)"
	@echo "  PB_DATA_DIR     = $(PB_DATA_DIR)"
	@echo "  NATS_DATA_DIR   = $(NATS_DATA_DIR)"
	@echo "  MIGRATIONS_DIR  = $(MIGRATIONS_DIR)"
	@echo "  CODEGEN_DIR     = $(CODEGEN_DIR)"
	@echo ""
	@echo "📄 Files:"
	@echo "  BINARY          = $(BINARY)"
	@echo "  TEMPLATE        = $(TEMPLATE)"
	@echo "  MODELS          = $(MODELS)"
	@echo ""
	@echo "🐙 GitHub:"
	@echo "  GH_OWNER        = $(GH_OWNER)"
	@echo "  GH_REPO         = $(GH_REPO)"
	@echo ""
	@echo "✈️  Fly.io:"
	@echo "  FLY_APP_NAME    = $(FLY_APP_NAME)"
	@echo "  FLY_REGION      = $(FLY_REGION)"
	@echo "  FLY             = $(FLY)"

## go-dep: Install development tools (pocketbase-gogen, gh, flyctl, mkcert)
go-dep:
	@echo "📦 Installing development tools..."
	go install github.com/snonky/pocketbase-gogen@v0.7.0
	go install github.com/cli/cli/v2/cmd/gh@latest
	go install github.com/superfly/flyctl@latest
	go install filippo.io/mkcert@latest
	@echo "✅ Tools installed"
	@echo ""
	@echo "Installed:"
	@echo "  - pocketbase-gogen (PocketBase code generation)"
	@echo "  - gh (GitHub CLI)"
	@echo "  - flyctl (Fly.io CLI)"
	@echo "  - mkcert (Local HTTPS certificates for development)"
	@echo ""
	@echo "💡 Next steps for HTTPS:"
	@echo "   make certs-init       Initialize local CA (one-time)"
	@echo "   make certs-generate   Generate certificates"
	@echo "   make run-https        Start server with HTTPS"

## go-mod-tidy: Tidy Go module dependencies
go-mod-tidy:
	go mod tidy

## go-mod-upgrade: Upgrade Go module dependencies
go-mod-upgrade:
	go install github.com/oligot/go-mod-upgrade@latest
	go-mod-upgrade

## run: Run PocketBase server with HTTPS (uses .env.local)
run: gen
	@if [ ! -f ".env.local" ]; then \
		echo "❌ .env.local not found!"; \
		echo "   Copy .env.local and configure with your localhost OAuth credentials"; \
		exit 1; \
	fi
	@if [ ! -f "$(CERTS_DIR)/cert.pem" ] || [ ! -f "$(CERTS_DIR)/key.pem" ]; then \
		echo "❌ Certificates not found! Run: make certs-generate"; \
		exit 1; \
	fi
	@mkdir -p $(PB_DATA_DIR) $(NATS_DATA_DIR)
	@echo "📋 Loading .env.local..."
	@set -a && . ./.env.local && set +a && \
		HTTPS_ENABLED=true \
		CERT_FILE=$(CERTS_DIR)/cert.pem \
		KEY_FILE=$(CERTS_DIR)/key.pem \
		go run . serve --https=0.0.0.0:8443 $(ARGS)

## mcp: Run MCP server for Claude Desktop integration (stdio)
mcp:
	@echo "🤖 Starting MCP server..."
	@echo "📡 MCP server will communicate via stdio"
	@echo "💡 Configure in Claude Desktop: ~/Library/Application Support/Claude/claude_desktop_config.json"
	@echo ""
	@echo "📋 Add this to your Claude Desktop config:"
	@echo '  {'
	@echo '    "mcpServers": {'
	@echo '      "pocketbase": {'
	@echo '        "command": "$(shell pwd)/$(BINARY)",'
	@echo '        "args": ["mcp"],'
	@echo '        "env": {'
	@echo '          "PB_DATA": "$(PB_DATA_DIR)"'
	@echo '        }'
	@echo '      }'
	@echo '    }'
	@echo '  }'
	@echo ""
	@mkdir -p $(PB_DATA_DIR) $(NATS_DATA_DIR)
	go run . mcp

## gen: Generate PocketBase template and type-safe models off the Pocketbase database itself
gen:
	@echo "📝 Generating PocketBase template from schema..."
	@if [ ! -d "$(PB_DATA_DIR)" ]; then \
		echo "❌ $(PB_DATA_DIR) not found!"; \
		echo "   Run 'make run' once to initialize PocketBase"; \
		exit 1; \
	fi
	@mkdir -p $(dir $(TEMPLATE))
	pocketbase-gogen template $(PB_DATA_DIR) $(TEMPLATE) --package models
	@echo "✅ Template: $(TEMPLATE)"
	@echo ""
	@echo "🔧 Generating type-safe models from template..."
	pocketbase-gogen generate $(TEMPLATE) $(MODELS)/proxies.go --package models --utils --hooks
	@echo "✅ Models: $(MODELS)/{proxies,utils,proxy_hooks}.go"


## gen-testdata: Generate test data using Go reflection. Not working yet...
gen-testdata:
	@echo "🔧 Generating test data..."
	@go run . gen-testdata -v



## bin: Build standalone PocketBase server binary
bin: gen
	@echo "🏗️  Building standalone PocketBase server..."
	@mkdir -p $(BIN_DIR)
	go build -o $(BINARY) .
	@echo "✅ Binary: $(BINARY)"
	@echo "💡 This is a standalone PocketBase binary with all commands available"
	@echo "   Try: $(BINARY) --help"

## test: Run all tests (unit + integration)
test:
	@echo "🧪 Running all tests..."
	go test -v ./...

## test-unit: Run unit tests only
test-unit:
	@echo "🧪 Running unit tests..."
	go test -v -short ./...

## test-mcp: Run MCP server tests only
test-mcp:
	@echo "🧪 Running MCP tests..."
	go test -v ./pkg/pbmcp/...




## test-e2e: Build and test PocketBase API endpoints
test-e2e: bin
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

## test-mcp-inspector: Launch MCP Inspector for interactive testing
test-mcp-inspector: bin
	@echo "🔍 Starting MCP Inspector..."
	@echo ""
	@echo "📡 MCP Inspector will launch in your browser"
	@echo "   URL: http://localhost:6274"
	@echo ""
	@echo "💡 You can now:"
	@echo "   - View all registered tools and resources"
	@echo "   - Test tool calls interactively"
	@echo "   - Inspect request/response data"
	@echo ""
	@echo "📋 Starting: npx @modelcontextprotocol/inspector $(BINARY) mcp"
	@echo ""
	npx @modelcontextprotocol/inspector $(BINARY) mcp

## test-mcp-config: Generate Claude Desktop/Code config for MCP testing
test-mcp-config:
	@echo "📋 Claude Desktop/Code Configuration"
	@echo ""
	@echo "Add this to: ~/Library/Application Support/Claude/claude_desktop_config.json"
	@echo ""
	@echo "{"
	@echo "  \"mcpServers\": {"
	@echo "    \"pocketbase\": {"
	@echo "      \"command\": \"$(shell pwd)/$(BINARY)\","
	@echo "      \"args\": [\"mcp\"],"
	@echo "      \"env\": {"
	@echo "        \"PB_DATA_DIR\": \"$(PB_DATA_DIR)\""
	@echo "      }"
	@echo "    }"
	@echo "  }"
	@echo "}"
	@echo ""
	@echo "💡 After updating the config:"
	@echo "   1. Build the binary: make bin"
	@echo "   2. Restart Claude Desktop/Code"
	@echo "   3. Test by asking Claude: 'What collections exist?'"

## vscode-mcp-setup: Setup VSCode/Claude Code MCP configuration
vscode-mcp-setup: bin
	@echo "🔧 Setting up VSCode MCP configuration..."
	@echo ""
	@if [ ! -f .vscode/mcp.json.example ]; then \
		echo "❌ Template not found: .vscode/mcp.json.example"; \
		exit 1; \
	fi
	@if [ -f .vscode/mcp.json ]; then \
		echo "⚠️  .vscode/mcp.json already exists"; \
		read -p "Overwrite? [y/N]: " CONFIRM; \
		if [ "$$CONFIRM" != "y" ] && [ "$$CONFIRM" != "Y" ]; then \
			echo "❌ Cancelled"; \
			exit 1; \
		fi; \
	fi
	@mkdir -p .vscode
	@sed "s|/ABSOLUTE/PATH/TO/WELLKNOWN|$(MAKEFILE_DIR)|g" .vscode/mcp.json.example > .vscode/mcp.json
	@echo "✅ Created .vscode/mcp.json"
	@echo ""
	@echo "📋 Configuration:"
	@cat .vscode/mcp.json
	@echo ""
	@echo "🔄 Next steps:"
	@echo "   1. Restart VSCode or run 'Cmd+Shift+P → Developer: Reload Window'"
	@echo "   2. Verify MCP server: 'Cmd+Shift+P → MCP: List Servers'"
	@echo "   3. Test with Claude: Ask 'What PocketBase collections exist?'"
	@echo ""
	@echo "💡 Troubleshooting:"
	@echo "   - View logs: 'Cmd+Shift+P → MCP: Show Output'"
	@echo "   - Manual test: make mcp"
	@echo "   - Interactive test: make test-mcp-inspector"

## health: Check PocketBase health endpoint (assumes server is running)
health:
	@echo "🏥 Checking PocketBase health..."
	@echo ""
	@curl -s http://localhost:8090/api/health | jq . || \
	curl -s http://localhost:8090/api/health || \
	echo "❌ Health check failed - is PocketBase running? Try: make run"
	@echo ""

## clean: Clean build artifacts and generated files
clean:
	@echo "🧹 Cleaning generated files..."
	rm -rf $(MAKEFILE_DIR)tmp/ $(BIN_DIR) $(DIST_DIR) tests/e2e/generated/
	rm -f $(MODELS)/*.go
	@echo "✅ Cleaned"
	@echo ""
	@echo "💡 Note: .data/ directory is NOT cleaned (persistent data)"
	@echo "   To remove data: rm -rf $(DATA_DIR)"

## kill: Kill processes on ports 8080, 8090 and 8443
kill:
	@echo "🔫 Killing processes on ports 8080, 8090 and 8443..."
	@lsof -ti:8080 | xargs kill -9 2>/dev/null || echo "   No processes on 8080"
	@lsof -ti:8090 | xargs kill -9 2>/dev/null || echo "   No processes on 8090"
	@lsof -ti:8443 | xargs kill -9 2>/dev/null || echo "   No processes on 8443"
	@echo "✅ Ports freed"

## version: Show current PocketBase binary version
version:
	@echo "📋 PocketBase binary version:"
	@if [ ! -f $(BINARY) ]; then \
		echo "❌ Binary not found: $(BINARY)"; \
		echo "   Run: make build"; \
		exit 1; \
	fi
	@strings $(BINARY) | grep "github.com/pocketbase/pocketbase/core.Version=" | head -1 | sed 's/.*Version=/   /' | tr -d '"'

## env-list: List all environment variables and their status
env-list:
	@. ./.env 2>/dev/null || true && go run . env list

## env-validate: Validate required environment variables are set
env-validate:
	@. ./.env 2>/dev/null || true && go run . env validate

## env-example: Generate .env.example from Go registry
env-example:
	@echo "📝 Generating .env.example from Go registry..."
	@go run . env generate-example > .env.example
	@echo "✅ .env.example generated!"

## env-sync: Sync all environment configuration to all files
env-sync: env-sync-dockerfile env-sync-flytoml
	@echo "✅ All environment configuration synced!"

## env-sync-dockerfile: Update Dockerfile env documentation
env-sync-dockerfile:
	@echo "📝 Syncing Dockerfile env docs..."
	@go run . env sync-dockerfile
	@echo "✅ Dockerfile updated"

## env-sync-flytoml: Update fly.toml [env] section
env-sync-flytoml:
	@echo "📝 Syncing fly.toml [env] section..."
	@go run . env sync-flytoml
	@echo "✅ fly.toml updated"

## env-generate-local: Generate .env.local template for development
env-generate-local:
	@echo "📝 Generating .env.local template..."
	@go run . env generate-local
	@echo "✅ .env.local generated"
	@echo "💡 Configure your OAuth credentials before running the server"

## env-generate-production: Generate .env.production template for Fly.io
env-generate-production:
	@echo "📝 Generating .env.production template..."
	@go run . env generate-production
	@echo "✅ .env.production generated"
	@echo "💡 Configure your production OAuth credentials before deploying"

## env-generate-example: Generate .env.example template (safe to commit)
env-generate-example:
	@echo "📝 Generating .env.example template..."
	@go run . env generate-example
	@echo "✅ .env.example generated"
	@echo "💡 This file is safe to commit to version control"

## env-sync-secrets: Merge .env.secrets into .env.local (local development)
env-sync-secrets:
	@echo "🔐 Syncing secrets to .env.local..."
	@test -f .env.secrets || (echo "❌ .env.secrets not found. Copy .env.secrets.example and configure with real credentials" && exit 1)
	@go run . env sync-secrets
	@echo "✅ .env.local generated from .env.secrets"
	@echo "💡 Ready for local development: make run"

## env-sync-secrets-production: Merge .env.secrets into .env.production (Fly.io deployment)
env-sync-secrets-production:
	@echo "🔐 Syncing secrets to .env.production..."
	@test -f .env.secrets || (echo "❌ .env.secrets not found. Copy .env.secrets.example and configure with real credentials" && exit 1)
	@go run . env sync-secrets-production
	@echo "✅ .env.production generated from .env.secrets"
	@echo "💡 Ready to deploy: make fly-secrets"


# ════════════════════════════════════════════════════════════════
# HTTPS Development Certificates (mkcert)
# ════════════════════════════════════════════════════════════════

## certs-init: Initialize local CA (run once per machine)
certs-init:
	@echo "🔐 Initializing local Certificate Authority..."
	@if ! command -v mkcert >/dev/null 2>&1; then \
		echo "❌ mkcert not found!"; \
		echo "   Run: make go-dep"; \
		exit 1; \
	fi
	mkcert -install
	@echo "✅ Local CA installed and trusted!"
	@echo ""
	@echo "📍 CA Root: $$(mkcert -CAROOT)"
	@echo ""
	@echo "💡 Next step:"
	@echo "   make certs-generate"

## certs-generate: Generate HTTPS certificates for localhost + LAN IP
certs-generate:
	@echo "🔑 Generating HTTPS certificates..."
	@if ! command -v mkcert >/dev/null 2>&1; then \
		echo "❌ mkcert not found!"; \
		echo "   Run: make go-dep"; \
		exit 1; \
	fi
	@mkdir -p $(CERTS_DIR)
	@LOCAL_IP=$$(ipconfig getifaddr en0 2>/dev/null || ipconfig getifaddr en1 2>/dev/null || echo "localhost"); \
	echo "   Generating certificates for:"; \
	echo "   • localhost"; \
	echo "   • 127.0.0.1"; \
	echo "   • $$LOCAL_IP"; \
	mkcert -key-file $(CERTS_DIR)/key.pem \
	       -cert-file $(CERTS_DIR)/cert.pem \
	       localhost 127.0.0.1 $$LOCAL_IP
	@echo "✅ Certificates generated:"
	@echo "   • Cert: $(CERTS_DIR)/cert.pem"
	@echo "   • Key:  $(CERTS_DIR)/key.pem"
	@ls -lh $(CERTS_DIR)/
	@echo ""
	@echo "💡 Next step:"
	@echo "   make run-https"

## certs-clean: Remove generated certificates
certs-clean:
	@echo "🧹 Removing certificates..."
	@rm -f $(CERTS_DIR)/cert.pem $(CERTS_DIR)/key.pem
	@echo "✅ Certificates removed"
	@echo ""
	@echo "💡 Note: Local CA still installed."
	@echo "   To remove CA: mkcert -uninstall"

## certs-status: Show certificate and CA status
certs-status:
	@echo "📋 HTTPS Certificate Status"
	@echo ""
	@echo "Local CA:"
	@if command -v mkcert >/dev/null 2>&1; then \
		mkcert -CAROOT | xargs ls -la 2>/dev/null || echo "   Not initialized yet"; \
	else \
		echo "   mkcert not installed (run: make go-dep)"; \
	fi
	@echo ""
	@echo "Generated Certificates:"
	@if [ -d "$(CERTS_DIR)" ] && [ -f "$(CERTS_DIR)/cert.pem" ]; then \
		ls -lh $(CERTS_DIR)/; \
		echo ""; \
		echo "✅ Ready for HTTPS (run: make run-https)"; \
	else \
		echo "   No certificates generated yet"; \
		echo ""; \
		echo "💡 Commands:"; \
		echo "   make certs-init       Initialize local CA"; \
		echo "   make certs-generate   Generate certificates"; \
	fi


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


## update: Update PocketBase binary from GitHub releases (production)
update:
	@echo "⬇️  Updating from GitHub releases..."
	@if [ ! -f $(BINARY) ]; then \
		echo "❌ Binary not found: $(BINARY)"; \
		echo "   Run: make bin"; \
		exit 1; \
	fi
	$(BINARY) update
	@echo "✅ Update complete!"

## test-update-local: Test local update mechanism (from .dist folder)
test-update-local: bin release
	@echo "🧪 Testing local update mechanism..."
	@echo ""
	@echo "📦 Built binary: $(BINARY)"
	@echo "📂 Update source: $(DIST_DIR)"
	@echo ""
	UPDATE_SOURCE=local UPDATE_LOCAL_DIR=$(DIST_DIR) $(BINARY) update
	@echo ""
	@echo "✅ Local update test complete!"
	@echo "💡 To test manually: UPDATE_SOURCE=local $(BINARY) update"

## test-update-github: Test GitHub update mechanism (requires network)
test-update-github: bin
	@echo "🧪 Testing GitHub update mechanism..."
	@echo ""
	@echo "⚠️  This will check for real GitHub releases"
	@echo "📦 Binary: $(BINARY)"
	@echo "🐙 Source: https://github.com/$(GH_OWNER)/$(GH_REPO)/releases"
	@echo ""
	@read -p "Continue? [y/N]: " CONFIRM; \
	if [ "$$CONFIRM" = "y" ] || [ "$$CONFIRM" = "Y" ]; then \
		UPDATE_SOURCE=github $(BINARY) update; \
		echo ""; \
		echo "✅ GitHub update test complete!"; \
	else \
		echo "❌ Cancelled"; \
	fi


# ════════════════════════════════════════════════════════════════
# Fly.io Deployment
# ════════════════════════════════════════════════════════════════

## fly-setup: Complete initial fly.io setup (auth + launch + volume)
fly-setup: fly-auth fly-launch fly-volume fly-secrets
	@echo ""
	@echo "✅ Fly.io setup complete!"
	@echo "   App: $(FLY_APP_NAME)"
	@echo "   Region: $(FLY_REGION)"
	@echo "   Volume: pb_data (1GB)"
	@echo ""
	@echo "Next steps:"
	@echo "  2. Deploy: make fly-deploy"

## fly-auth: Authenticate with fly.io
fly-auth:
	@echo "🔐 Authenticating with fly.io..."
	$(FLY) auth login
	@echo "✅ Authenticated!"

## fly-launch: Initialize fly.io app (run once)
fly-launch:
	@echo "🚀 Launching fly.io app..."
	@if [ -n "$(FLY_APP_NAME)" ]; then \
		echo "📋 App name from fly.toml: $(FLY_APP_NAME)"; \
		echo "🔍 Checking if app exists..."; \
		if $(FLY) apps list 2>&1 | grep -q "$(FLY_APP_NAME)"; then \
			echo "✅ App already exists: $(FLY_APP_NAME)"; \
		else \
			echo "📦 Creating app: $(FLY_APP_NAME)"; \
			$(FLY) apps create $(FLY_APP_NAME) --org personal || \
			$(FLY) apps create $(FLY_APP_NAME) || true; \
			echo "✅ App created!"; \
		fi; \
	else \
		echo "⚠️  fly.toml not found or app name not set"; \
		echo "   Running interactive launch..."; \
		$(FLY) launch --no-deploy; \
	fi

## fly-volume: Create persistent volume for pb_data (1GB)
fly-volume:
	@echo "💾 Creating persistent volume for pb_data..."
	@echo "   App: $(FLY_APP_NAME)"
	@echo "   Region: $(FLY_REGION)"
	$(FLY) volumes create pb_data --size 1 --region $(FLY_REGION) --app $(FLY_APP_NAME) --yes
	@echo "✅ Volume created!"

## fly-secrets: Set environment variables as fly.io secrets (uses .env.production)
fly-secrets: env-sync-secrets-production
	@echo "🔐 Syncing secrets to Fly.io (from .env.production)..."
	@. ./.env.production && go run . pb env export-secrets | $(FLY) secrets import
	@echo "✅ Secrets synced!"
	@echo "💡 Non-secret config is defined in fly.toml [env] section"

## fly-deploy: Deploy to fly.io
fly-deploy:
	@echo "🚀 Deploying to fly.io..."
	$(FLY) deploy
	@echo "✅ Deployed!"
	@echo ""
	@echo "🌐 Your app: https://$(FLY_APP_NAME).fly.dev"
	@echo "🔧 Admin UI: https://$(FLY_APP_NAME).fly.dev/_/"

## fly-status: Check deployment status
fly-status:
	@echo "📊 Checking fly.io status..."
	$(FLY) status

## fly-logs: Tail fly.io logs
fly-logs:
	@echo "📋 Tailing fly.io logs..."
	$(FLY) logs

## fly-ssh: SSH into fly.io machine
fly-ssh:
	@echo "🔌 SSH into fly.io machine..."
	$(FLY) ssh console

## fly-destroy: Destroy fly.io app (WARNING: destructive!)
fly-destroy:
	@echo "⚠️  WARNING: This will DESTROY the fly.io app and ALL data!"
	@read -p "Are you absolutely sure? Type 'yes' to confirm: " CONFIRM; \
	if [ "$$CONFIRM" = "yes" ]; then \
		echo "💥 Destroying fly.io app: $(FLY_APP_NAME)"; \
		$(FLY) apps destroy $(FLY_APP_NAME) --yes; \
	else \
		echo "❌ Cancelled (you must type 'destroy' to confirm)"; \
	fi
