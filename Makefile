.PHONY: help go-dep mod-upgrade dev build test test-e2e gen-testdata clean run pb-server pb-dev pb-build pb-clean pb-release pb-update pb-version pb-gen-template pb-gen-models pb-debug

## help: Show available make targets
help:
	@echo "Wellknown Development:"
	@echo ""
	@grep -E '^##' $(MAKEFILE_LIST) | sed 's/^## /  make /' | column -t -s ':'

.DEFAULT_GOAL := help

## go-dep: Install development tools (Air, pocketbase-gogen, gh)
go-dep:
	@echo "📦 Installing development tools..."
	# https://github.com/air-verse/air/releases/tag/v1.63.0
	go install github.com/air-verse/air@v1.63.0

	# https://github.com/Snonky/pocketbase-gogen/releases/tag/v0.7.0
	go install github.com/snonky/pocketbase-gogen@v0.7.0

	# https://github.com/cli/cli
	go install github.com/cli/cli/v2/cmd/gh@latest
	@echo "✅ Tools installed (air, pocketbase-gogen, gh)"

## mod-upgrade: Upgrade Go module dependencies
mod-upgrade:
	# https://github.com/oligot/go-mod-upgrade/releases/tag/v0.12.0
	go install github.com/oligot/go-mod-upgrade@latest
	go-mod-upgrade

## dev: Start dev server with hot-reload (requires Air)
dev:
	@which air > /dev/null || (echo "Run: make go-dep" && exit 1)
	air

## build: Build server binary
build:
	go build -o .bin/wellknown-server ./cmd/server

## test: Run Go unit tests
test:
	go test -v ./pkg/...

## gen-testdata: Generate test data using Go reflection
gen-testdata:
	@echo "🔧 Generating test data using Go reflection..."
	@go run ./cmd/testdata-gen -v

## test-e2e: Run data-driven calendar E2E tests (Go + Playwright)
test-e2e: gen-testdata
	@echo "🧪 Running data-driven calendar E2E tests (Go + Playwright)..."
	@echo "   Step 1: Generated test data with Go reflection ✅"
	@echo "   Step 2: Running Playwright tests..."
	@cd tests && bun run playwright test calendar-generated --reporter=line --workers=1

## clean: Clean build artifacts and generated test files
clean:
	rm -rf tmp/ .bin/ .dist/ tests/e2e/generated/

## run: Build and run server
run: build
	./.bin/wellknown-server

