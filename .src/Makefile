# Repository definitions
# Format: REPO_NAME = directory_name|git_url|ref|description
# ref can be: tag (v0.31.0), branch (main), commit hash, or empty for default branch
REPO_POCKETBASE := pocketbase|https://github.com/pocketbase/pocketbase.git|v0.31.0|PocketBase (official source - matches go.mod version)
REPO_POCKETBASE_HA := pocketbase-ha|https://github.com/litesql/pocketbase-ha.git|v0.0.6|PocketBase-HA (High Availability implementation)
REPO_GOGEN := pocketbase-gogen|https://github.com/Snonky/pocketbase-gogen.git|v0.7.0|pocketbase-gogen (type-safe code generator - matches Makefile version)
REPO_PRESENTATOR := presentator|https://github.com/presentator/presentator.git|v3.4.27|Presentator (PocketBase library pattern example)
REPO_JSONSCHEMA := jsonschema|https://github.com/santhosh-tekuri/jsonschema.git|v6.0.2|jsonschema (validation library - matches go.mod v6.0.2)
REPO_GOPOCJSON := goPocJsonSchemaForm|https://github.com/warlockxins/goPocJsonSchemaForm.git|main|goPocJsonSchemaForm (dynamic form generation example)
REPO_DATASTARUI := datastarui|https://github.com/coreycole/datastarui.git|main|DatastarUI (Go/templ shadcn/ui components with Datastar)

# Auto-computed variables
# Extract directory names from all REPO_* variables
ALL_REPOS := $(foreach var,$(filter REPO_%,$(.VARIABLES)),$(word 1,$(subst |, ,$($(var)))))

# Extract target names from REPO_* variable names (used for pattern rule invocation)
# Example: REPO_POCKETBASE -> pocketbase, REPO_POCKETBASE_HA -> pocketbase-ha
REPO_NAMES := $(foreach var,$(filter REPO_%,$(.VARIABLES)),$(shell echo $(var) | sed 's/^REPO_//' | tr '[:upper:]' '[:lower:]' | tr '_' '-'))

# Auto-generate .PHONY declaration (pattern rules don't need explicit .PHONY)
.PHONY: help sync-all list print clean status

# ============================================================================
# Parameterized Functions
# ============================================================================

# Helper to extract field from REPO_* variable
# Usage: $(call get-field,VARIABLE,FIELD_NUMBER)
# Fields: 1=dir, 2=url, 3=ref, 4=description
get-field = $(word $(2),$(subst |, ,$($(1))))

# Extract specific fields from repo variables
get-dir = $(call get-field,$(1),1)
get-url = $(call get-field,$(1),2)
get-ref = $(call get-field,$(1),3)
get-desc = $(call get-field,$(1),4)

# Convert target suffix to REPO_* variable name
# Examples: pocketbase -> REPO_POCKETBASE, pocketbase-ha -> REPO_POCKETBASE_HA
target-to-var = REPO_$(shell echo $(1) | tr '[:lower:]' '[:upper:]' | tr '-' '_')

# Template for syncing a repository to specified ref (idempotent clone/update)
# Parameters: $(1) = REPO_* variable name (e.g., REPO_POCKETBASE)
# This replaces both clone-repo-template and update-repo-template
define sync-repo-template
	@REF=$(call get-ref,$(1)); \
	DIR=$(call get-dir,$(1)); \
	URL=$(call get-url,$(1)); \
	if [ ! -d "$$DIR" ]; then \
		if [ -n "$$REF" ]; then \
			echo "📦 Cloning $$DIR @ $$REF..."; \
			if git clone --branch $$REF --depth 1 $$URL $$DIR 2>/dev/null; then \
				echo "✅ $$DIR synced to $$REF"; \
			else \
				git clone $$URL $$DIR && cd $$DIR && git checkout $$REF && cd ..; \
				echo "✅ $$DIR synced to $$REF"; \
			fi; \
		else \
			echo "📦 Cloning $$DIR..."; \
			git clone $$URL $$DIR; \
			echo "✅ $$DIR cloned"; \
		fi; \
	else \
		if [ -n "$$REF" ]; then \
			echo "🔄 Syncing $$DIR to $$REF..."; \
			cd $$DIR && git fetch --all --tags 2>/dev/null && git checkout $$REF && git pull origin $$REF 2>/dev/null || true; \
			CURRENT=$$(git describe --tags --always 2>/dev/null); \
			echo "✅ $$DIR synced to $$CURRENT"; \
		else \
			echo "🔄 Updating $$DIR..."; \
			cd $$DIR && git pull; \
			echo "✅ $$DIR updated"; \
		fi; \
	fi
endef

# ============================================================================
# Main Targets
# ============================================================================

.DEFAULT_GOAL := sync-all

## help: Show available make targets
help:
	@echo "════════════════════════════════════════════════════════════════"
	@echo "  Reference Code Management"
	@echo "════════════════════════════════════════════════════════════════"
	@echo ""
	@grep -E '^##' $(MAKEFILE_LIST) | sed 's/^## /  make /' | column -t -s ':'
	@echo ""
	@echo "════════════════════════════════════════════════════════════════"
	@echo "💡 Common Commands:"
	@echo "   make              Sync all repositories (default)"
	@echo "   make list         Show repository status and versions"
	@echo "   make help         Show this help"
	@echo ""
	@echo "Advanced: make sync-<name>, make clean, make status, make print"
	@echo "════════════════════════════════════════════════════════════════"

# print: Show all Makefile variables (for debugging) - Hidden from help
print:
	@echo "=== Repository Definitions (dir|url|ref|description) ==="
	@$(foreach var,$(sort $(filter-out REPO_NAMES,$(filter REPO_%,$(.VARIABLES)))),\
		printf "%-18s = %s\n" "$(var)" "$($(var))";)
	@echo ""
	@echo "=== Computed Variables ==="
	@echo "ALL_REPOS          = $(ALL_REPOS)"
	@echo "REPO_NAMES         = $(REPO_NAMES)"

## sync-all: Sync all repositories to specified refs (in parallel)
sync-all:
	@echo "🔄 Syncing all repositories in parallel..."
	@$(MAKE) -j $(words $(REPO_NAMES)) $(foreach repo,$(REPO_NAMES),sync-$(repo))
	@echo "✅ All repositories synced!"

# sync-REPO: Sync specific repository (use make list to see available repos) - Hidden pattern rule
sync-%:
	@$(eval REPO_VAR := $(call target-to-var,$*))
	@$(if $($(REPO_VAR)),,$(error ❌ Unknown repo: $*. Run 'make list-repos' to see available repositories))
	$(call sync-repo-template,$(REPO_VAR))

## list: Show repository status, versions, and available repos
list:
	@echo "════════════════════════════════════════════════════════════════"
	@echo "  📦 Reference Repositories"
	@echo "════════════════════════════════════════════════════════════════"
	@echo ""
	@$(foreach var,$(sort $(filter-out REPO_NAMES,$(filter REPO_%,$(.VARIABLES)))),\
		DIR=$(call get-dir,$(var)); \
		REF=$(call get-ref,$(var)); \
		if [ -d "$$DIR" ]; then \
			CURRENT_BRANCH=$$(cd $$DIR && git branch --show-current 2>/dev/null); \
			CURRENT_COMMIT=$$(cd $$DIR && git rev-parse --short HEAD 2>/dev/null); \
			CURRENT_TAG=$$(cd $$DIR && git describe --tags --exact-match 2>/dev/null); \
			if [ -n "$$CURRENT_TAG" ]; then \
				DISPLAY="$$CURRENT_TAG"; \
				if [ "$$REF" = "$$CURRENT_TAG" ]; then SYNCED="yes"; else SYNCED="no"; fi; \
			elif [ -n "$$CURRENT_BRANCH" ]; then \
				DISPLAY="$$CURRENT_BRANCH ($$CURRENT_COMMIT)"; \
				if [ "$$REF" = "$$CURRENT_BRANCH" ]; then SYNCED="yes"; else SYNCED="no"; fi; \
			else \
				DISPLAY="detached@$$CURRENT_COMMIT"; \
				if [ "$$REF" = "$$CURRENT_COMMIT" ]; then SYNCED="yes"; else SYNCED="no"; fi; \
			fi; \
			if [ "$$SYNCED" = "yes" ]; then \
				printf "  ✅ %-25s %s\n" "$$DIR" "$$DISPLAY"; \
			else \
				printf "  ⚠️  %-25s %s (expected %s)\n" "$$DIR" "$$DISPLAY" "$$REF"; \
			fi; \
		else \
			printf "  ❌ %-25s %s\n" "$$DIR" "not cloned"; \
		fi;)
	@echo ""
	@echo "Usage: make sync-<name>  Example: make sync-pocketbase"
	@echo "       make              Sync all repositories"

# clean: Remove all cloned repositories (WARNING: destructive!) - Hidden from help
clean:
	@echo "⚠️  This will delete all cloned reference repositories"
	@read -p "Are you sure? [y/N]: " confirm; \
	if [ "$$confirm" = "y" ] || [ "$$confirm" = "Y" ]; then \
		rm -rf $(ALL_REPOS); \
		echo "✅ All repositories removed"; \
	else \
		echo "❌ Cancelled"; \
	fi

# status: Show git status and latest commit for all cloned repositories - Hidden from help
status:
	@echo "📊 Repository Status:"
	@echo ""
	@for dir in $(ALL_REPOS); do \
		if [ -d "$$dir" ]; then \
			echo "=== $$dir ==="; \
			cd $$dir && git status --short && git log --oneline -1 && cd ..; \
			echo ""; \
		fi; \
	done
