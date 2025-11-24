# Simple repository sync Makefile
# Data files:
#   - repos.list: Repository metadata (dir|url|ref|fork_of|description)
#   - stacks.list: Tech stack groupings (stack|repos|description)

REPOS_FILE := repos.list
STACKS_FILE := stacks.list
SHELL := /bin/bash

# Read repo data once (using standard Unix tools)
REPOS := $(shell tail -n +2 $(REPOS_FILE) | cut -d'|' -f1)

.PHONY: help doctor gitignore install status upgrade agent test
.DEFAULT_GOAL := help

## help: Show available commands
help:
	@echo ""
	@printf "\033[1;36m%s\033[0m\n" "Repository Management:"
	@echo ""
	@printf "  \033[1;32m%-20s\033[0m %s\n" "install" "Clone and sync repos from repos.list"
	@printf "  \033[1;32m%-20s\033[0m %s\n" "status" "Show current vs repos.list state"
	@printf "  \033[1;32m%-20s\033[0m %s\n" "upgrade" "Update repos.list to latest tags"
	@printf "  \033[1;32m%-20s\033[0m %s\n" "doctor" "Validate repos.list format"
	@printf "  \033[1;32m%-20s\033[0m %s\n" "test" "Run repository tests"
	@echo ""
	@printf "\033[1;36m%s\033[0m\n" "Other:"
	@echo ""
	@printf "  \033[1;32m%-20s\033[0m %s\n" "agent" "Generate stack documentation"
	@echo ""

# Include module files
include repos.mk
include agent.mk

## test: Run all tests
test:
	@echo "╔══════════════════════════════════════════════════════════════════╗"
	@echo "║                    Running All Tests                              ║"
	@echo "╚══════════════════════════════════════════════════════════════════╝"
	@echo ""
	@echo "━━━ Running repository tests (real GitHub repos) ━━━"
	@./test-repos.sh
	@echo ""
	@echo "╔══════════════════════════════════════════════════════════════════╗"
	@echo "║                    ✅ All Tests Passed!                           ║"
	@echo "╚══════════════════════════════════════════════════════════════════╝"
