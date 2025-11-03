.PHONY: help install-tools dev build test test-e2e test-e2e-calendar gen-testdata clean run pb-server pb-dev pb-build pb-clean pb-release pb-update pb-version pb-gen-template pb-gen-models

help:
	@echo "Development:"
	@echo "  make install-tools     - Install dev tools (Air, pocketbase-gogen, gh)"
	@echo "  make dev              - Start dev server with hot-reload"
	@echo "  make build            - Build server binary"
	@echo "  make test             - Run Go unit tests"
	@echo "  make test-e2e         - Run Playwright E2E tests (fast core tests)"
	@echo "  make test-e2e-calendar - Run data-driven calendar tests (Go + Playwright)"
	@echo "  make gen-testdata     - Generate test data using Go reflection"
	@echo "  make run              - Run server"
	@echo "  make clean            - Clean artifacts"
	@echo ""
	@echo "NOTE: GCP OAuth setup is now integrated into the main server!"
	@echo "      Access it at: http://localhost:8080/tools/gcp-setup"
	@echo ""
	@echo "Pocketbase Server:"
	@echo "  make pb-server         - Run Pocketbase server (auto-creates collections)"
	@echo "  make pb-dev            - Run Pocketbase with hot-reload (Air)"
	@echo "  make pb-build          - Build Pocketbase server binary"
	@echo "  make pb-clean          - Clean generated models and temp files"
	@echo "  make pb-release        - Build & create GitHub release (multi-platform)"
	@echo "  make pb-update         - Update binary from GitHub releases"
	@echo "  make pb-version        - Show current binary version"
	@echo "  make pb-gen-template   - Generate template from schema (edit before gen-models)"
	@echo "  make pb-gen-models     - Generate type-safe models from template"

install-tools:
	@echo "📦 Installing development tools..."
	# https://github.com/air-verse/air/releases/tag/v1.63.0
	go install github.com/air-verse/air@v1.63.0

	# https://github.com/Snonky/pocketbase-gogen/releases/tag/v0.7.0
	go install github.com/snonky/pocketbase-gogen@v0.7.0

	# https://github.com/cli/cli
	go install github.com/cli/cli/v2/cmd/gh@latest
	@echo "✅ Tools installed (air, pocketbase-gogen, gh)"

upgrade:
	# https://github.com/oligot/go-mod-upgrade/releases/tag/v0.12.0
	go install github.com/oligot/go-mod-upgrade@latest
	go-mod-upgrade

dev:
	@which air > /dev/null || (echo "Run: make install-tools" && exit 1)
	air

build:
	go build -o .bin/wellknown-server ./cmd/server

test:
	go test -v ./pkg/...

gen-testdata:
	@echo "🔧 Generating test data using Go reflection..."
	@go run ./cmd/testdata-gen -v

test-e2e: gen-testdata dev
	@echo "🧪 Running data-driven calendar E2E tests (Go + Playwright)..."
	@echo "   Step 1: Generated test data with Go reflection ✅"
	@echo "   Step 2: Running Playwright tests..."
	@cd tests && bun run playwright test calendar-generated --reporter=line --workers=1

clean:
	rm -rf tmp/ .bin/ .dist/ tests/e2e/generated/

run: build
	./.bin/wellknown-server

### PB

# Variables
PB_DIR := pb/base
PB_BINARY := .bin/wellknown-pb
PB_DATA := $(PB_DIR)/pb_data
PB_TEMPLATE := pb/_templates/google_tokens.go
PB_MODELS := pb/models
DIST_DIR := .dist
GH_OWNER := joeblew999
GH_REPO := wellknown

pb-gen-template:
	@echo "📝 [Step 1/4] Generating template from PocketBase data..."
	@if [ ! -d "$(PB_DATA)" ]; then \
		echo "❌ $(PB_DATA) not found!"; \
		echo "   Run 'make pb-server' once to initialize PocketBase"; \
		exit 1; \
	fi
	@mkdir -p $(dir $(PB_TEMPLATE))
	pocketbase-gogen template $(PB_DATA) $(PB_TEMPLATE) --package models
	@echo "✅ Template: $(PB_TEMPLATE)"
	@echo "   [Step 2/4] Edit template, then run: make pb-gen-models"

pb-gen-models:
	@echo "🔧 [Step 3/4] Generating type-safe models from template..."
	@if [ ! -f "$(PB_TEMPLATE)" ]; then \
		echo "❌ Template not found. Run: make pb-gen-template"; \
		exit 1; \
	fi
	pocketbase-gogen generate $(PB_TEMPLATE) $(PB_MODELS)/proxies.go --package models --utils --hooks
	@echo "✅ Models: $(PB_MODELS)/{proxies,utils,proxy_hooks}.go"


pb-server:
	@echo "🚀 Starting PocketBase server..."
	@echo "Admin UI: http://localhost:8090/_/"
	@if [ -f $(PB_DIR)/.env ]; then \
		cd $(PB_DIR) && source .env && go run main.go serve; \
	else \
		echo "⚠️  No .env - OAuth disabled. Copy $(PB_DIR)/.env.example"; \
		cd $(PB_DIR) && go run main.go serve; \
	fi

pb-dev:
	@echo "🔥 Starting PocketBase with hot-reload (Air)..."
	@which air > /dev/null || (echo "❌ Air not installed. Run: make install-tools" && exit 1)
	@echo "Admin UI: http://localhost:8090/_/"
	@if [ -f $(PB_DIR)/.env ]; then \
		echo "✅ Using .env for OAuth"; \
	else \
		echo "⚠️  No .env - OAuth disabled. Copy $(PB_DIR)/.env.example"; \
	fi
	cd pb && air

pb-build:
	@echo "🏗️  Building PocketBase server..."
	cd $(PB_DIR) && go build -o ../../$(PB_BINARY) main.go
	@echo "✅ Binary: $(PB_BINARY)"

pb-clean:
	@echo "🧹 Cleaning PocketBase generated files..."
	rm -rf pb/tmp pb/models pb/build-errors.log
	@echo "✅ Cleaned: pb/tmp/, pb/models/, pb/build-errors.log"
	@echo "💡 Tip: Run 'make pb-gen-models' to regenerate type-safe models"

pb-release:
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
	echo "  - darwin/arm64..." && ( cd $(PB_DIR) && CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build -ldflags "$$LDFLAGS" -o ../../$(DIST_DIR)/wellknown-pb-darwin-arm64 . ) & \
	echo "  - darwin/amd64..." && ( cd $(PB_DIR) && CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -ldflags "$$LDFLAGS" -o ../../$(DIST_DIR)/wellknown-pb-darwin-amd64 . ) & \
	echo "  - linux/amd64..." && ( cd $(PB_DIR) && CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags "$$LDFLAGS" -o ../../$(DIST_DIR)/wellknown-pb-linux-amd64 . ) & \
	echo "  - linux/arm64..." && ( cd $(PB_DIR) && CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -ldflags "$$LDFLAGS" -o ../../$(DIST_DIR)/wellknown-pb-linux-arm64 . ) & \
	echo "  - windows/amd64..." && ( cd $(PB_DIR) && CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -ldflags "$$LDFLAGS" -o ../../$(DIST_DIR)/wellknown-pb-windows-amd64.exe . ) & \
	echo "  - windows/arm64..." && ( cd $(PB_DIR) && CGO_ENABLED=0 GOOS=windows GOARCH=arm64 go build -ldflags "$$LDFLAGS" -o ../../$(DIST_DIR)/wellknown-pb-windows-arm64.exe . ) & \
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
	cd .. && \
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
	echo "⬇️  Update: $(PB_BINARY) update"

pb-update:
	@echo "⬇️  Updating from GitHub releases..."
	@if [ ! -f $(PB_BINARY) ]; then \
		echo "❌ Binary not found: $(PB_BINARY)"; \
		echo "   Run: make pb-build"; \
		exit 1; \
	fi
	$(PB_BINARY) update
	@echo "✅ Update complete!"

pb-version:
	@echo "📋 PocketBase binary version:"
	@if [ ! -f $(PB_BINARY) ]; then \
		echo "❌ Binary not found: $(PB_BINARY)"; \
		echo "   Run: make pb-build"; \
		exit 1; \
	fi
	@strings $(PB_BINARY) | grep "github.com/pocketbase/pocketbase/core.Version=" | head -1 | sed 's/.*Version=/   /' | tr -d '"'
