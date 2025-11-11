# Simple repository sync Makefile
# Data files:
#   - repos.list: Repository metadata (name|dir|url|ref|fork_url|description)
#   - index.list: Trigger-to-file mappings for Claude (category|priority|triggers|action|path|why)
#   - fork.list: Fork workflow state (current_repo=...)

REPOS_FILE := repos.list
INDEX_FILE := index.list
FORK_FILE := fork.list
SHELL := /bin/bash

# Read repo data once (using standard Unix tools)
REPOS := $(shell tail -n +2 $(REPOS_FILE) | cut -d'|' -f2)

.PHONY: help doctor gitignore install status upgrade index test fork-add fork-sync fork-push fork-status fork-context fork fork-start fork-continue fork-finish
.DEFAULT_GOAL := help

## help: Show available commands
help:
	@echo ""
	@printf "\033[1;36m%s\033[0m\n" "Git Workflow (3 commands):"
	@echo ""
	@printf "  \033[1;32m%-20s\033[0m %s\n" "fork-start" "Start new work"
	@printf "  \033[1;32m%-20s\033[0m %s\n" "fork-continue" "Save progress"
	@printf "  \033[1;32m%-20s\033[0m %s\n" "fork-finish" "Done, back to main"
	@printf "  \033[1;32m%-20s\033[0m %s\n" "fork" "Interactive menu"
	@echo ""
	@printf "\033[1;36m%s\033[0m\n" "Repository Management:"
	@echo ""
	@printf "  \033[1;32m%-20s\033[0m %s\n" "install" "Clone all repos"
	@printf "  \033[1;32m%-20s\033[0m %s\n" "upgrade" "Update repos to latest"
	@printf "  \033[1;32m%-20s\033[0m %s\n" "status" "Show repo status"
	@printf "  \033[1;32m%-20s\033[0m %s\n" "doctor" "Check repos.list"
	@echo ""
	@printf "\033[1;36m%s\033[0m\n" "Fork Management:"
	@echo ""
	@printf "  \033[1;32m%-20s\033[0m %s\n" "fork-context" "Set working repo (one-time)"
	@printf "  \033[1;32m%-20s\033[0m %s\n" "fork-add" "Add fork remotes"
	@printf "  \033[1;32m%-20s\033[0m %s\n" "fork-sync" "Sync with upstream"
	@printf "  \033[1;32m%-20s\033[0m %s\n" "fork-push" "Push to fork"
	@printf "  \033[1;32m%-20s\033[0m %s\n" "fork-status" "Fork sync status"
	@echo ""
	@printf "\033[1;36m%s\033[0m\n" "Other:"
	@echo ""
	@printf "  \033[1;32m%-20s\033[0m %s\n" "test" "Run all tests"
	@printf "  \033[1;32m%-20s\033[0m %s\n" "index" "Update INDEX.md"
	@echo ""

# Include module files
include repos.mk
include forks.mk
include index.mk

## test: Run all tests (scenarios + expect scripts)
test:
	@echo "╔══════════════════════════════════════════════════════════════════╗"
	@echo "║                    Running All Tests                              ║"
	@echo "╚══════════════════════════════════════════════════════════════════╝"
	@echo ""
	@echo "━━━ Running scenario tests ━━━"
	@./test-scenarios.sh
	@echo ""
	@echo "━━━ Running baseline tests ━━━"
	@./test-fork-baseline.exp
	@echo ""
	@echo "━━━ Running prompt tests ━━━"
	@./test-prompts.exp
	@echo ""
	@echo "╔══════════════════════════════════════════════════════════════════╗"
	@echo "║                    ✅ All Tests Passed!                           ║"
	@echo "╚══════════════════════════════════════════════════════════════════╝"
