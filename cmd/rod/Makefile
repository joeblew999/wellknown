.PHONY: help run dev show build clean clear-cookies fresh fresh-show

help: ## Show this help message
	@echo "Rod Browser Automation - Available commands:"
	@echo ""
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2}'

run: ## Run the automation (headless mode)
	@go run .

show: ## Run the automation with visible browser (GUI mode)
	@go run . -show

fresh: ## Run with fresh login (headless, clears cookies)
	@go run . -clear-cookies

fresh-show: ## Run with fresh login (GUI mode, clears cookies)
	@go run . -show -clear-cookies

dev: ## Run with hot-reload using Air (headless)
	@air

dev-show: ## Run with hot-reload using Air (GUI mode)
	@air -- -show

build: ## Build the binary
	@echo "Building rod automation..."
	@go build -o ../../.bin/rod-automation .
	@echo "✅ Built to .bin/rod-automation"

clean: ## Clean temporary files and state
	@rm -rf tmp/
	@rm -rf .rod/
	@rm -f ../../.bin/rod-automation
	@echo "✅ Cleaned"

install-air: ## Install Air for hot-reload
	@command -v air >/dev/null 2>&1 || go install github.com/cosmtrek/air@latest
	@echo "✅ Air installed"
