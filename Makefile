# Makefile for wellknown project
.PHONY: help install-tools dev build test clean run

# Default target
help:
	@echo "wellknown - Deep links library for Google and Apple ecosystems"
	@echo ""
	@echo "Usage:"
	@echo "  make install-tools  Install development tools (Air, etc.)"
	@echo "  make dev           Start development server with hot-reload (Air)"
	@echo "  make build         Build the server binary"
	@echo "  make test          Run all tests"
	@echo "  make clean         Clean build artifacts"
	@echo "  make run           Run the server (without hot-reload)"
	@echo ""

# Install development tools (uses modern 'tool' directive from go.mod)
install-tools:
	@echo "Installing development tools from go.mod..."
	@echo "(Using Go 1.23+ 'tool' directive - keeps go.mod clean!)"
	go tool -n 2>&1 | grep -q "air" && echo "✓ Air already available" || go install github.com/air-verse/air@latest
	@echo "✓ Tools ready!"
	@echo ""
	@echo "Run 'make dev' to start development server with hot-reload"

# Start development server with hot-reload
dev:
	@which air > /dev/null || (echo "Air not installed. Run 'make install-tools' first." && exit 1)
	@echo "Starting development server with hot-reload..."
	@echo "💻 Local:  http://localhost:8080"
	@echo "📱 Mobile: Check server output for mobile URL"
	@echo ""
	air

# Build the server binary
build:
	@echo "Building server..."
	go build -o bin/wellknown-server ./cmd/server
	@echo "✓ Binary built: bin/wellknown-server"

# Run all tests
test:
	@echo "Running tests..."
	go test -v ./pkg/...

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	rm -rf tmp/ bin/
	@echo "✓ Clean complete"

# Run the server without hot-reload
run: build
	@echo "Starting server..."
	./bin/wellknown-server
