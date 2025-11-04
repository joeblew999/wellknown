.PHONY: help debug server dev build clean release update version gen-template gen-models

# Paths relative to this Makefile (pb/)
MAKEFILE_DIR := $(patsubst %/,%,$(dir $(abspath $(lastword $(MAKEFILE_LIST)))))
PROJECT_ROOT := $(abspath $(MAKEFILE_DIR)/..)

# Directories
BASE_DIR := $(MAKEFILE_DIR)/cmd
MIGRATIONS_DIR := $(BASE_DIR)/pb_migrations
CODEGEN_DIR := $(MAKEFILE_DIR)/codegen
BIN_DIR := $(PROJECT_ROOT)/.bin
DIST_DIR := $(PROJECT_ROOT)/.dist

# Files
BINARY := $(BIN_DIR)/wellknown-pb
DATA := $(BASE_DIR)/pb_data
MIGRATIONS := $(MIGRATIONS_DIR)
TEMPLATE := $(CODEGEN_DIR)/_templates/google_tokens.go
MODELS := $(CODEGEN_DIR)/models

# GitHub
GH_OWNER := joeblew999
GH_REPO := wellknown

## help: Show available make targets
help:
	@echo "Pocketbase Server:"
	@echo ""
	@grep -E '^##' $(MAKEFILE_LIST) | sed 's/^## /  make /' | column -t -s ':'

.DEFAULT_GOAL := help

## debug: Show all Makefile variables and paths
debug:
	@echo "=== Makefile Debug Info ==="
	@echo ""
	@echo "📁 Paths:"
	@echo "  MAKEFILE_DIR  = $(MAKEFILE_DIR)"
	@echo "  PROJECT_ROOT  = $(PROJECT_ROOT)"
	@echo ""
	@echo "📂 Directories:"
	@echo "  BASE_DIR        = $(BASE_DIR)"
	@echo "  MIGRATIONS_DIR  = $(MIGRATIONS_DIR)"
	@echo "  CODEGEN_DIR     = $(CODEGEN_DIR)"
	@echo "  BIN_DIR         = $(BIN_DIR)"
	@echo "  DIST_DIR        = $(DIST_DIR)"
	@echo ""
	@echo "📄 Files:"
	@echo "  BINARY        = $(BINARY)"
	@echo "  DATA          = $(DATA)"
	@echo "  MIGRATIONS    = $(MIGRATIONS)"
	@echo "  TEMPLATE      = $(TEMPLATE)"
	@echo "  MODELS        = $(MODELS)"
	@echo ""
	@echo "🐙 GitHub:"
	@echo "  GH_OWNER      = $(GH_OWNER)"
	@echo "  GH_REPO       = $(GH_REPO)"
	@echo ""
	@echo "✅ File Checks:"
	@if [ -d "$(BASE_DIR)" ]; then echo "  ✓ BASE_DIR exists"; else echo "  ✗ BASE_DIR missing"; fi
	@if [ -d "$(MIGRATIONS_DIR)" ]; then echo "  ✓ MIGRATIONS_DIR exists"; else echo "  ✗ MIGRATIONS_DIR missing"; fi
	@if [ -d "$(CODEGEN_DIR)" ]; then echo "  ✓ CODEGEN_DIR exists"; else echo "  ✗ CODEGEN_DIR missing"; fi
	@if [ -d "$(BIN_DIR)" ]; then echo "  ✓ BIN_DIR exists"; else echo "  ✗ BIN_DIR missing"; fi
	@if [ -f "$(BINARY)" ]; then echo "  ✓ BINARY exists"; else echo "  ✗ BINARY missing (run 'make build')"; fi
	@if [ -d "$(MODELS)" ]; then echo "  ✓ MODELS dir exists"; else echo "  ✗ MODELS dir missing"; fi

## gen-template: Generate template from schema (edit before gen-models)
gen-template:
	@echo "📝 [Step 1/4] Generating template from PocketBase data..."
	@if [ ! -d "$(DATA)" ]; then \
		echo "❌ $(DATA) not found!"; \
		echo "   Run 'make server' once to initialize PocketBase"; \
		exit 1; \
	fi
	@mkdir -p $(dir $(TEMPLATE))
	pocketbase-gogen template $(DATA) $(TEMPLATE) --package models
	@echo "✅ Template: $(TEMPLATE)"
	@echo "   [Step 2/4] Edit template, then run: make gen-models"

## gen-models: Generate type-safe models from template
gen-models:
	@echo "🔧 [Step 3/4] Generating type-safe models from template..."
	@if [ ! -f "$(TEMPLATE)" ]; then \
		echo "❌ Template not found. Run: make gen-template"; \
		exit 1; \
	fi
	pocketbase-gogen generate $(TEMPLATE) $(MODELS)/proxies.go --package models --utils --hooks
	@echo "✅ Models: $(MODELS)/{proxies,utils,proxy_hooks}.go"

## kill-port: Kill any process using port 8090
kill-port:
	@echo "🔫 Killing processes on port 8090..."
	@lsof -ti:8090 | xargs kill -9 2>/dev/null || echo "   No processes found on port 8090"
	@echo "✅ Port 8090 freed"

## server: Run Pocketbase server (auto-creates collections)
server:
	@echo "🚀 Starting PocketBase server..."
	@echo "Admin UI: http://localhost:8090/_/"
	@if [ -f $(BASE_DIR)/.env ]; then \
		cd $(BASE_DIR) && source .env && go run . serve; \
	else \
		echo "⚠️  No .env - OAuth disabled. Copy $(BASE_DIR)/.env.example"; \
		cd $(BASE_DIR) && go run . serve; \
	fi

## dev: Run Pocketbase with hot-reload (Air)
dev:
	@echo "🔥 Starting PocketBase with hot-reload (Air)..."
	@which air > /dev/null || (echo "❌ Air not installed. Run: make install-tools" && exit 1)
	@echo "Admin UI: http://localhost:8090/_/"
	@if [ -f $(BASE_DIR)/.env ]; then \
		echo "✅ Using .env for OAuth"; \
	else \
		echo "⚠️  No .env - OAuth disabled. Copy $(BASE_DIR)/.env.example"; \
	fi
	air

## build: Build Pocketbase server binary
build:
	@echo "🏗️  Building PocketBase server..."
	@mkdir -p $(BIN_DIR)
	cd $(BASE_DIR) && go build -o $(BINARY) .
	@echo "✅ Binary: $(BINARY)"

## clean: Clean generated models and temp files
clean:
	@echo "🧹 Cleaning PocketBase generated files..."
	rm -rf $(MAKEFILE_DIR)/tmp $(MAKEFILE_DIR)/build-errors.log
	rm -f $(MODELS)/*.go
	@echo "✅ Cleaned: tmp/, codegen/models/*.go, build-errors.log"
	@echo "💡 Tip: Run 'make gen-models' to regenerate type-safe models"

## release: Build & create GitHub release (multi-platform)
release:
	@echo "🚀 Creating GitHub release for PocketBase server..."
	@if ! command -v gh >/dev/null 2>&1; then \
		echo "⚙️  Installing GitHub CLI..."; \
		go install github.com/cli/cli/v2/cmd/gh@latest; \
		if ! command -v gh >/dev/null 2>&1; then \
			echo "❌ GitHub CLI install failed or not in PATH"; \
			echo "   Ensure ~/go/bin is in your PATH"; \
			exit 1; \
		fi; \
		echo "✅ GitHub CLI installed"; \
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
	if [ -z "$$VERSION" ]; then \
		echo "❌ Version required"; \
		exit 1; \
	fi; \
	echo "📦 Building for multiple platforms (parallel) with version $$VERSION..."; \
	mkdir -p $(DIST_DIR); \
	LDFLAGS="-X github.com/pocketbase/pocketbase/core.Version=$$VERSION"; \
	echo "  - darwin/arm64..." && ( cd $(BASE_DIR) && CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build -ldflags "$$LDFLAGS" -o $(DIST_DIR)/wellknown-pb-darwin-arm64 . ) & \
	echo "  - darwin/amd64..." && ( cd $(BASE_DIR) && CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -ldflags "$$LDFLAGS" -o $(DIST_DIR)/wellknown-pb-darwin-amd64 . ) & \
	echo "  - linux/amd64..." && ( cd $(BASE_DIR) && CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags "$$LDFLAGS" -o $(DIST_DIR)/wellknown-pb-linux-amd64 . ) & \
	echo "  - linux/arm64..." && ( cd $(BASE_DIR) && CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -ldflags "$$LDFLAGS" -o $(DIST_DIR)/wellknown-pb-linux-arm64 . ) & \
	echo "  - windows/amd64..." && ( cd $(BASE_DIR) && CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -ldflags "$$LDFLAGS" -o $(DIST_DIR)/wellknown-pb-windows-amd64.exe . ) & \
	echo "  - windows/arm64..." && ( cd $(BASE_DIR) && CGO_ENABLED=0 GOOS=windows GOARCH=arm64 go build -ldflags "$$LDFLAGS" -o $(DIST_DIR)/wellknown-pb-windows-arm64.exe . ) & \
	wait && \
	echo "✅ All builds complete!" && \
	echo "📦 Creating ZIP archives for ghupdate..." && \
	cd $(DIST_DIR) && \
	cp wellknown-pb-darwin-arm64 wellknown-pb && zip wellknown-pb_darwin_arm64.zip wellknown-pb && rm wellknown-pb && \
	cp wellknown-pb-darwin-amd64 wellknown-pb && zip wellknown-pb_darwin_amd64.zip wellknown-pb && rm wellknown-pb && \
	cp wellknown-pb-linux-amd64 wellknown-pb && zip wellknown-pb_linux_amd64.zip wellknown-pb && rm wellknown-pb && \
	cp wellknown-pb-linux-arm64 wellknown-pb && zip wellknown-pb_linux_arm64.zip wellknown-pb && rm wellknown-pb && \
	cp wellknown-pb-windows-amd64.exe wellknown-pb.exe && zip wellknown-pb_windows_amd64.zip wellknown-pb.exe && rm wellknown-pb.exe && \
	cp wellknown-pb-windows-arm64.exe wellknown-pb.exe && zip wellknown-pb_windows_arm64.zip wellknown-pb.exe && rm wellknown-pb.exe && \
	cd $(PROJECT_ROOT) && \
	echo "🏷️  Creating git tag $$VERSION..." && \
	git tag -a "$$VERSION" -m "Release $$VERSION" && \
	git push origin "$$VERSION" && \
	echo "📝 Creating GitHub release $$VERSION..." && \
	gh release create "$$VERSION" $(DIST_DIR)/wellknown-pb*.zip \
		--title "PocketBase Server $$VERSION" \
		--notes "Release $$VERSION of wellknown PocketBase server" && \
	echo "" && \
	echo "✅ Release $$VERSION created!" && \
	echo "🔗 https://github.com/$(GH_OWNER)/$(GH_REPO)/releases/tag/$$VERSION" && \
	echo "⬇️  Update: $(BINARY) update"

## update: Update binary from GitHub releases
update:
	@echo "⬇️  Updating from GitHub releases..."
	@if [ ! -f $(BINARY) ]; then \
		echo "❌ Binary not found: $(BINARY)"; \
		echo "   Run: make build"; \
		exit 1; \
	fi
	$(BINARY) update
	@echo "✅ Update complete!"

## version: Show current binary version
version:
	@echo "📋 PocketBase binary version:"
	@if [ ! -f $(BINARY) ]; then \
		echo "❌ Binary not found: $(BINARY)"; \
		echo "   Run: make build"; \
		exit 1; \
	fi
	@strings $(BINARY) | grep "github.com/pocketbase/pocketbase/core.Version=" | head -1 | sed 's/.*Version=/   /' | tr -d '"'
