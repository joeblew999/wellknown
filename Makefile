.PHONY: help print go-dep go-mod-upgrade gen gen-testdata run bin test health clean kill version release update fly-auth fly-launch fly-volume fly-secrets fly-deploy fly-status fly-logs fly-ssh fly-destroy

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

# PocketBase source directories (version controlled)
PB_CMD_DIR := $(MAKEFILE_DIR)pkg/cmd/pocketbase
# Database migrations (Go)
MIGRATIONS_DIR := $(PB_CMD_DIR)/pb_migrations
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
	@echo "  DATA_DIR        = $(DATA_DIR)"
	@echo "  PB_DATA_DIR     = $(PB_DATA_DIR)"
	@echo "  NATS_DATA_DIR   = $(NATS_DATA_DIR)"
	@echo "  PB_CMD_DIR      = $(PB_CMD_DIR)"
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

## go-dep: Install development tools (pocketbase-gogen, gh, flyctl)
go-dep:
	@echo "📦 Installing development tools..."
	go install github.com/snonky/pocketbase-gogen@v0.7.0
	go install github.com/cli/cli/v2/cmd/gh@latest
	go install github.com/superfly/flyctl@latest
	@echo "✅ Tools installed"
	@echo ""
	@echo "Installed:"
	@echo "  - pocketbase-gogen (PocketBase code generation)"
	@echo "  - gh (GitHub CLI)"
	@echo "  - flyctl (Fly.io CLI)"

## go-mod-tidy: Tidy Go module dependencies
go-mod-tidy:
	go mod tidy

## go-mod-upgrade: Upgrade Go module dependencies
go-mod-upgrade:
	go install github.com/oligot/go-mod-upgrade@latest
	go-mod-upgrade

## run: Run PocketBase server (port 8090)
run:
	@echo "🚀 Starting PocketBase..."
	@echo "Admin UI: http://localhost:8090/_/"
	@echo ""
	@echo "📁 Data directory: $(PB_DATA_DIR)"
	@echo "   PocketBase will auto-create subdirectories as needed:"
	@echo "   - storage/          File uploads"
	@echo "   - backups/          Database backups"
	@echo "   - .pb_temp_to_delete/  Temp files"
	@echo ""
	@echo "⚙️  Configuration (PocketBase best practice):"
	@echo "   1. .env file (dev only - auto-loaded when using 'go run')"
	@echo "   2. Environment variables (production - set in shell/docker/systemd)"
	@echo "   3. Command-line flags (for testing specific values)"
	@echo ""
	@echo "💡 Examples:"
	@echo "   # Use .env file (development)"
	@echo "   make run"
	@echo ""
	@echo "   # Use env vars (production)"
	@echo "   GOOGLE_CLIENT_ID=xxx GOOGLE_CLIENT_SECRET=yyy make run"
	@echo ""
	@echo "   # Use flags"
	@echo "   make run ARGS='serve --http=0.0.0.0:8080'"
	@echo ""
	@mkdir -p $(PB_DATA_DIR) $(NATS_DATA_DIR)
	go run . pb $(ARGS)

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



## bin: Build PocketBase server binary into BIN
bin:
	@echo "🏗️  Building PocketBase server..."
	@mkdir -p $(BIN_DIR)
	go build -o $(BINARY) .
	@echo "✅ Binary: $(BINARY)"

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

## fly-secrets: Set environment variables as fly.io secrets
fly-secrets:
	@echo "🔐 Setting fly.io secrets..."
	@if [ ! -f .env ]; then \
		echo "❌ .env file not found"; \
		echo "   Copy .env.example to .env and configure"; \
		exit 1; \
	fi
	@echo ""
	@echo "📋 Validating required secrets..."
	@. ./.env && \
	MISSING=0; \
	if [ -z "$$GOOGLE_CLIENT_ID" ]; then echo "   ❌ GOOGLE_CLIENT_ID not set"; MISSING=1; fi; \
	if [ -z "$$GOOGLE_CLIENT_SECRET" ]; then echo "   ❌ GOOGLE_CLIENT_SECRET not set"; MISSING=1; fi; \
	if [ -z "$$GOOGLE_REDIRECT_URL" ]; then echo "   ❌ GOOGLE_REDIRECT_URL not set"; MISSING=1; fi; \
	if [ $$MISSING -eq 1 ]; then \
		echo ""; \
		echo "💡 Google OAuth is required. Set these in .env file."; \
		exit 1; \
	fi; \
	echo "   ✅ Required secrets validated"
	@echo ""
	@echo "📤 Syncing secrets to Fly.io..."
	@. ./.env && \
	SECRETS="GOOGLE_CLIENT_ID=\"$$GOOGLE_CLIENT_ID\" GOOGLE_CLIENT_SECRET=\"$$GOOGLE_CLIENT_SECRET\" GOOGLE_REDIRECT_URL=\"$$GOOGLE_REDIRECT_URL\""; \
	if [ -n "$$PB_ADMIN_EMAIL" ]; then \
		SECRETS="$$SECRETS PB_ADMIN_EMAIL=\"$$PB_ADMIN_EMAIL\""; \
	fi; \
	if [ -n "$$PB_ADMIN_PASSWORD" ]; then \
		SECRETS="$$SECRETS PB_ADMIN_PASSWORD=\"$$PB_ADMIN_PASSWORD\""; \
	fi; \
	if [ -n "$$APPLE_TEAM_ID" ] && [ -n "$$APPLE_CLIENT_ID" ] && [ -n "$$APPLE_KEY_ID" ]; then \
		echo "   📱 Apple OAuth credentials found - including in secrets"; \
		SECRETS="$$SECRETS APPLE_TEAM_ID=\"$$APPLE_TEAM_ID\" APPLE_CLIENT_ID=\"$$APPLE_CLIENT_ID\" APPLE_KEY_ID=\"$$APPLE_KEY_ID\""; \
		if [ -n "$$APPLE_PRIVATE_KEY" ]; then \
			echo "   🔑 Using APPLE_PRIVATE_KEY (inline content)"; \
			SECRETS="$$SECRETS APPLE_PRIVATE_KEY=\"$$APPLE_PRIVATE_KEY\""; \
		elif [ -n "$$APPLE_PRIVATE_KEY_PATH" ] && [ -f "$$APPLE_PRIVATE_KEY_PATH" ]; then \
			echo "   🔑 Reading Apple private key from: $$APPLE_PRIVATE_KEY_PATH"; \
			KEY_CONTENT=$$(cat "$$APPLE_PRIVATE_KEY_PATH"); \
			SECRETS="$$SECRETS APPLE_PRIVATE_KEY=\"$$KEY_CONTENT\""; \
		else \
			echo "   ⚠️  Apple private key not found - Apple OAuth will not work"; \
		fi; \
		if [ -n "$$APPLE_REDIRECT_URL" ]; then \
			SECRETS="$$SECRETS APPLE_REDIRECT_URL=\"$$APPLE_REDIRECT_URL\""; \
		fi; \
	fi; \
	if [ -n "$$SMTP_HOST" ] && [ -n "$$SMTP_USERNAME" ] && [ -n "$$SMTP_PASSWORD" ]; then \
		echo "   📧 SMTP credentials found - including in secrets"; \
		SECRETS="$$SECRETS SMTP_HOST=\"$$SMTP_HOST\" SMTP_PORT=\"$$SMTP_PORT\" SMTP_USERNAME=\"$$SMTP_USERNAME\" SMTP_PASSWORD=\"$$SMTP_PASSWORD\""; \
		if [ -n "$$SMTP_FROM_EMAIL" ]; then SECRETS="$$SECRETS SMTP_FROM_EMAIL=\"$$SMTP_FROM_EMAIL\""; fi; \
		if [ -n "$$SMTP_FROM_NAME" ]; then SECRETS="$$SECRETS SMTP_FROM_NAME=\"$$SMTP_FROM_NAME\""; fi; \
	fi; \
	if [ -n "$$S3_BUCKET" ] && [ -n "$$S3_ACCESS_KEY" ] && [ -n "$$S3_SECRET_KEY" ]; then \
		echo "   ☁️  S3 credentials found - including in secrets"; \
		SECRETS="$$SECRETS S3_ENDPOINT=\"$$S3_ENDPOINT\" S3_REGION=\"$$S3_REGION\" S3_BUCKET=\"$$S3_BUCKET\" S3_ACCESS_KEY=\"$$S3_ACCESS_KEY\" S3_SECRET_KEY=\"$$S3_SECRET_KEY\""; \
		if [ -n "$$S3_FORCE_PATH_STYLE" ]; then SECRETS="$$SECRETS S3_FORCE_PATH_STYLE=\"$$S3_FORCE_PATH_STYLE\""; fi; \
	fi; \
	echo ""; \
	eval "$(FLY) secrets set $$SECRETS"
	@echo ""
	@echo "✅ Secrets synced successfully!"
	@echo ""
	@echo "💡 Tip: Non-secret config (SERVER_HOST, SERVER_PORT, PB_DATA_DIR) is in fly.toml [env] section"

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