.PHONY: help install-tools dev build test clean run gcp-setup gcp-setup-cli gcp-setup-dev pb-server pb-build

help:
	@echo "Development:"
	@echo "  make install-tools  - Install dev tools (Air, pocketbase-gogen)"
	@echo "  make dev           - Start dev server with hot-reload"
	@echo "  make build         - Build server binary"
	@echo "  make test          - Run tests"
	@echo "  make run           - Run server"
	@echo "  make clean         - Clean artifacts"
	@echo ""
	@echo "Google Cloud Integration:"
	@echo "  make gcp-setup     - Interactive web-based OAuth setup (NO gcloud needed!)"
	@echo "  make gcp-setup-dev - Start web setup with hot-reload (for development)"
	@echo "  make gcp-setup-cli - Automated CLI setup (requires gcloud CLI)"
	@echo ""
	@echo "Pocketbase Server:"
	@echo "  make pb-server     - Run Pocketbase server (requires .env setup)"
	@echo "  make pb-build      - Build Pocketbase server binary"

install-tools:
	@echo "📦 Installing development tools..."
	# https://github.com/air-verse/air/releases/tag/v1.63.0
	go install github.com/air-verse/air@v1.63.0

	# https://github.com/Snonky/pocketbase-gogen/releases/tag/v0.7.0
	go install github.com/snonky/pocketbase-gogen@v0.7.0
	@echo "✅ Tools installed"

dev:
	@which air > /dev/null || (echo "Run: make install-tools" && exit 1)
	air

build:
	go build -o bin/wellknown-server ./cmd/server

test:
	go test -v ./pkg/...

clean:
	rm -rf tmp/ bin/

run: build
	./bin/wellknown-server

gcp-setup:
	@echo "🚀 Web-Based OAuth Setup Dashboard"
	@echo "===================================="
	@echo ""
	@echo "Starting web server on http://localhost:3030"
	@echo ""
	cd tools/gcp-setup && go run main.go --web

gcp-setup-dev:
	@which air > /dev/null || (echo "Run: make install-tools" && exit 1)
	@echo "🚀 GCP Setup with Hot-Reload (Development Mode)"
	@echo "================================================="
	@echo ""
	@echo "Web server: http://localhost:3030"
	@echo "Watching: *.go, *.html files"
	@echo ""
	cd tools/gcp-setup && air

gcp-setup-cli:
	@echo "🚀 Automated GCP Setup (requires gcloud CLI)"
	@echo "See: tools/gcp-setup/README.md for prerequisites"
	@echo ""
	cd tools/gcp-setup && go run main.go --cli

pb-server:
	@if [ ! -f pb/base/.env ]; then \
		echo "❌ Missing pb/base/.env file"; \
		echo "Run 'make gcp-setup' first or copy pb/base/.env.example"; \
		exit 1; \
	fi
	@echo "🚀 Starting Pocketbase server..."
	@echo "Access at: http://localhost:8090"
	@echo "Admin UI:  http://localhost:8090/_/"
	cd pb/base && source .env && go run main.go serve

pb-build:
	@echo "🏗️  Building Pocketbase server..."
	cd pb/base && go build -o ../../bin/wellknown-pb main.go
	@echo "✅ Binary: bin/wellknown-pb"
