.PHONY: help build serve serve-http certs-info certs-generate certs-regenerate test clean

# Paths
BIN_DIR := ../../.bin
BINARY := $(BIN_DIR)/pdfform

.DEFAULT_GOAL := help

## help: Show this help message
help:
	@echo "════════════════════════════════════════════════════════════════"
	@echo "  PDF Form CLI - Development Commands"
	@echo "════════════════════════════════════════════════════════════════"
	@echo ""
	@grep -E '^##' $(MAKEFILE_LIST) | sed 's/^## //' | awk -F: '{printf "  %-20s %s\n", $$1, $$2}'
	@echo ""
	@echo "════════════════════════════════════════════════════════════════"
	@echo "💡 Quick Start:"
	@echo "   make build         Build the pdfform CLI"
	@echo "   make serve         Start HTTPS server (recommended)"
	@echo "   make certs-info    Show certificate information"
	@echo "════════════════════════════════════════════════════════════════"

## build: Build pdfform CLI tool
build:
	@echo "🏗️  Building pdfform CLI..."
	@mkdir -p $(BIN_DIR)
	cd cmd/pdfform && go build -o $(BINARY) .
	@echo "✅ Binary: $(BINARY)"
	@echo "💡 Try: $(BINARY) --help"

## serve: Build and run pdfform HTTPS web server
serve: build
	@echo "🌐 Starting pdfform HTTPS web server..."
	@echo "🔒 The server will auto-generate certificates using mkcert"
	@echo ""
	$(BINARY) serve

## serve-http: Build and run pdfform HTTP web server (not recommended for mobile)
serve-http: build
	@echo "🌐 Starting pdfform HTTP web server..."
	@echo "⚠️  HTTP mode - not recommended for mobile devices"
	@echo ""
	$(BINARY) serve --http

## certs-info: Show certificate information
certs-info: build
	@$(BINARY) certs info

## certs-generate: Generate HTTPS certificates
certs-generate: build
	@$(BINARY) certs generate

## certs-regenerate: Regenerate HTTPS certificates
certs-regenerate: build
	@$(BINARY) certs regenerate

## test: Run all tests
test:
	@echo "🧪 Running tests..."
	go test -v ./...

## test-unit: Run unit tests only
test-unit:
	@echo "🧪 Running unit tests..."
	go test -v -short ./...

## clean: Remove built binaries and generated files
clean:
	@echo "🧹 Cleaning..."
	@rm -f $(BINARY)
	@rm -rf .data/certs/
	@rm -rf .data/downloads/
	@rm -rf .data/templates/
	@rm -rf .data/outputs/
	@rm -rf .data/temp/
	@echo "✅ Clean complete"

## run-example-basic: Run basic usage example
run-example-basic:
	@echo "📋 Running basic usage example..."
	cd examples/basic-usage && go run main.go

## run-example-config: Run configuration example
run-example-config:
	@echo "⚙️  Running configuration example..."
	cd examples/case-management && go run main.go
