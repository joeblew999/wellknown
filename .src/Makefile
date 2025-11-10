# Repository metadata lives in repos.list (pipe-delimited: name|dir|url|ref|description)
REPOS_FILE := repos.list
$(if $(wildcard $(REPOS_FILE)),,$(error Missing $(REPOS_FILE). Please create it.))

# Basic projections reused throughout the Makefile
REPO_NAMES := $(shell awk -F'|' 'NR>1 {print $$1}' $(REPOS_FILE))
ALL_REPOS := $(shell awk -F'|' 'NR>1 {print $$2}' $(REPOS_FILE))
SORTED_REPO_NAMES := $(sort $(REPO_NAMES))

# Control whether `make status` fetches remote tags before reporting (1=yes, 0=no)
STATUS_FETCH_TAGS ?= 1

# Refs treated as branches/commits (never auto-bumped to newest tag)
TAG_BRANCH_IGNORE := main master develop development dev head trunk default

# Auto-generate .PHONY declaration (pattern rules don't need explicit .PHONY)
.PHONY: help sync list print clean status update

# ============================================================================
# Parameterized Functions
# ============================================================================

# Helper to extract fields from repos.list on demand
# Usage: $(call repo-field,repo_name,field_index)
repo-field = $(shell awk -F'|' -v target="$(1)" -v idx=$(2) 'NR>1 && $$1==target {print $$idx; exit}' $(REPOS_FILE))

get-dir = $(call repo-field,$(1),2)
get-url = $(call repo-field,$(1),3)
get-ref = $(call repo-field,$(1),4)
get-desc = $(call repo-field,$(1),5)

# Template for syncing a repository to specified ref (idempotent clone/update)
# Parameters: $(1) = REPO_* variable name (e.g., REPO_POCKETBASE)
# This replaces both clone-repo-template and update-repo-template
define sync-repo-template
	@REF=$(call get-ref,$(1)); \
	DIR=$(call get-dir,$(1)); \
	URL=$(call get-url,$(1)); \
	printf "  🔄 %-25s syncing...\n" "$$DIR"; \
	if [ ! -d "$$DIR" ]; then \
		if [ -n "$$REF" ]; then \
			if git clone --quiet --branch $$REF --depth 1 $$URL $$DIR 2>/dev/null; then true; else git clone --quiet $$URL $$DIR 2>/dev/null && (cd $$DIR && git checkout --quiet $$REF 2>/dev/null); fi; \
		else \
			git clone --quiet $$URL $$DIR 2>/dev/null; \
		fi; \
	fi; \
	if [ -n "$$REF" ]; then \
		if ! (cd $$DIR && git fetch --all --tags --prune --quiet 2>/dev/null && git checkout --quiet $$REF 2>/dev/null && git pull --quiet origin $$REF 2>/dev/null); then \
			printf "  ⚠️  %-25s failed to sync, recloning...\n" "$$DIR"; \
			rm -rf "$$DIR"; \
			git clone --quiet $$URL $$DIR 2>/dev/null; \
			(cd $$DIR && git checkout --quiet $$REF 2>/dev/null); \
		fi; \
	else \
		if ! (cd $$DIR && git pull --quiet 2>/dev/null); then \
			printf "  ⚠️  %-25s pull failed, recloning...\n" "$$DIR"; \
			rm -rf "$$DIR"; \
			git clone --quiet $$URL $$DIR 2>/dev/null; \
		fi; \
	fi; \
	CURRENT_BRANCH=$$(cd $$DIR && git branch --show-current 2>/dev/null); \
	CURRENT_COMMIT=$$(cd $$DIR && git rev-parse --short HEAD 2>/dev/null); \
	CURRENT_TAG=$$(cd $$DIR && git describe --tags --exact-match 2>/dev/null); \
	if [ -n "$$CURRENT_TAG" ]; then \
		DISPLAY="$$CURRENT_TAG"; \
		if [ -n "$$REF" ] && [ "$$REF" != "$$CURRENT_TAG" ]; then ICON="⚠️ "; EXTRA=" (expected $$REF)"; else ICON="✅"; EXTRA=""; fi; \
	elif [ -n "$$CURRENT_BRANCH" ]; then \
		DISPLAY="$$CURRENT_BRANCH ($$CURRENT_COMMIT)"; \
		if [ -n "$$REF" ] && [ "$$REF" != "$$CURRENT_BRANCH" ]; then ICON="⚠️ "; EXTRA=" (expected $$REF)"; else ICON="✅"; EXTRA=""; fi; \
	else \
		DISPLAY="detached@$$CURRENT_COMMIT"; \
		if [ -n "$$REF" ] && [ "$$REF" != "$$CURRENT_COMMIT" ]; then ICON="⚠️ "; EXTRA=" (expected $$REF)"; else ICON="✅"; EXTRA=""; fi; \
	fi; \
	printf "  %s %-25s %s%s\n" "$$ICON" "$$DIR" "$$DISPLAY" "$$EXTRA"
endef

# ============================================================================
# Main Targets
# ============================================================================

.DEFAULT_GOAL := sync

## help: Show available make targets
help:
	@echo "Targets:"
	@grep -E '^##' $(firstword $(MAKEFILE_LIST)) | sed 's/^## /make /' | awk -F': ' '{printf "  %-18s %s\n", $$1, $$2}'

# print: Show all Makefile variables (for debugging) - Hidden from help
print:
	@echo "════════════════════════════════════════════════════════════════"
	@echo "  🧱 Make Variables"
	@echo "════════════════════════════════════════════════════════════════"
	@echo ""
	@echo "  Repository definitions (dir|url|ref|description):"
	@$(foreach repo,$(sort $(REPO_NAMES)),\
		printf "    • %-20s %s|%s|%s|%s\n" "$(repo)" "$(call get-dir,$(repo))" "$(call get-url,$(repo))" "$(call get-ref,$(repo))" "$(call get-desc,$(repo))";)
	@echo ""
	@echo "  Computed variables:"
	@printf "    • %-20s %s\n" "ALL_REPOS" "$(ALL_REPOS)"
	@printf "    • %-20s %s\n" "REPO_NAMES" "$(REPO_NAMES)"
	@printf "    • %-20s %s\n" "SORTED_REPO_NAMES" "$(SORTED_REPO_NAMES)"
	@printf "    • %-20s %s\n" "STATUS_FETCH_TAGS" "$(STATUS_FETCH_TAGS)"

## sync: Sync all repositories (parallel)
sync:
	@echo "════════════════════════════════════════════════════════════════"
	@echo "  🔄 Sync Repositories"
	@echo "════════════════════════════════════════════════════════════════"
	@echo ""
	@echo "  Mode: parallel"
	@$(MAKE) -j $(words $(SORTED_REPO_NAMES)) $(foreach repo,$(SORTED_REPO_NAMES),sync-$(repo))
	@echo ""
	@echo "✅ All repositories synced!"

# sync-REPO: Sync specific repository (use make list to see available repos) - Hidden pattern rule
sync-%:
	@if [ -z "$(filter $*,$(REPO_NAMES))" ]; then \
		echo "❌ Unknown repo: $*"; exit 1; \
	fi
	$(call sync-repo-template,$*)

## list: Show repository status, versions, and available repos
list:
	@echo "════════════════════════════════════════════════════════════════"
	@echo "  📦 Reference Repositories"
	@echo "════════════════════════════════════════════════════════════════"
	@echo ""
	@$(foreach repo,$(sort $(REPO_NAMES)),\
		DIR=$(call get-dir,$(repo)); \
		REF=$(call get-ref,$(repo)); \
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
	@echo "       make / make sync  Sync all repositories"

# clean: Remove all cloned repositories (WARNING: destructive!) - Hidden from help
clean:
	@echo "════════════════════════════════════════════════════════════════"
	@echo "  🧹 Clean Repositories"
	@echo "════════════════════════════════════════════════════════════════"
	@echo ""
	@echo "⚠️  This will delete all cloned reference repositories"
	@read -p "Are you sure? [y/N]: " confirm; \
	if [ "$$confirm" = "y" ] || [ "$$confirm" = "Y" ]; then \
		rm -rf $(ALL_REPOS); \
		echo "✅ All repositories removed"; \
	else \
		echo "❌ Cancelled"; \
	fi

## status: Show git status, last commit, and latest tags for all repos
status:
	@echo "════════════════════════════════════════════════════════════════"
	@echo "  📊 Repository Status"
	@echo "════════════════════════════════════════════════════════════════"
	@echo ""
	@if [ "$(STATUS_FETCH_TAGS)" != "0" ]; then \
		echo "  ⏳ Refreshing tags in parallel..."; \
		for dir in $(sort $(ALL_REPOS)); do \
			if [ -d "$$dir" ]; then \
				( cd $$dir && git fetch --tags --prune --quiet >/dev/null 2>&1 ) & \
			fi; \
		done; \
		wait; \
		echo ""; \
	fi
	@$(foreach repo,$(sort $(REPO_NAMES)),\
		DIR=$(call get-dir,$(repo)); \
		REF=$(call get-ref,$(repo)); \
		if [ -d "$$DIR" ]; then \
			BRANCH=$$(cd $$DIR && git branch --show-current 2>/dev/null); \
			COMMIT=$$(cd $$DIR && git rev-parse --short HEAD 2>/dev/null); \
			TAG=$$(cd $$DIR && git describe --tags --exact-match 2>/dev/null); \
			STATE=$$(cd $$DIR && git status --short); \
			LAST=$$(cd $$DIR && git log --oneline -1 2>/dev/null); \
			LATEST_TAG=""; \
			REF_LOW=$$(echo "$$REF" | tr '[:upper:]' '[:lower:]'); \
			if ! printf "%s" "$(TAG_BRANCH_IGNORE)" | tr ' ' '\n' | grep -qi "^$$REF_LOW$$"; then \
				LATEST_TAG=$$(cd $$DIR && git tag --sort=-creatordate 2>/dev/null | head -1); \
				if [ -z "$$LATEST_TAG" ]; then \
					LATEST_TAG=$$(cd $$DIR && git describe --tags --abbrev=0 2>/dev/null); \
				fi; \
			fi; \
			if [ -n "$$TAG" ]; then \
				DISPLAY="$$TAG"; \
			elif [ -n "$$BRANCH" ]; then \
				DISPLAY="$$BRANCH ($$COMMIT)"; \
			else \
				DISPLAY="detached@$$COMMIT"; \
			fi; \
			if [ -z "$$STATE" ]; then \
				printf "  ✅ %-25s %s\n" "$$DIR" "$$DISPLAY"; \
				printf "       • Last commit: %s\n" "$$LAST"; \
				printf "       • Changes: clean\n"; \
			else \
				printf "  ⚠️  %-25s %s (dirty)\n" "$$DIR" "$$DISPLAY"; \
				printf "       • Last commit: %s\n" "$$LAST"; \
				printf "       • Changes:\n"; \
				echo "$$STATE" | sed 's/^/         - /'; \
			fi; \
			if [ -n "$$LATEST_TAG" ]; then \
				if [ -n "$$TAG" ] && [ "$$TAG" = "$$LATEST_TAG" ]; then \
					printf "       • Latest tag: %s (current)\n" "$$LATEST_TAG"; \
				else \
					printf "       • ⚠️ Latest tag: %s (update available)\n" "$$LATEST_TAG"; \
				fi; \
			else \
				printf "       • Latest tag: n/a (tracking %s)\n" "$$REF"; \
			fi; \
		else \
			printf "  ❌ %-25s %s\n" "$$DIR" "not cloned"; \
		fi; \
	)

## update: Update repo refs to latest tags (updates repos.csv)
update:
	@echo "════════════════════════════════════════════════════════════════"
	@echo "  🔁 Update Repository Tags"
	@echo "════════════════════════════════════════════════════════════════"
	@echo ""
	@cp $(REPOS_FILE) $(REPOS_FILE).bak
	@tmp=$$(mktemp); \
	TMPFILE="$$tmp" TAG_BRANCH_IGNORE="$(TAG_BRANCH_IGNORE)" awk -F'|' '\
	function trim(s){gsub(/^[ \t]+|[ \t]+$$/, "", s); return s} \
	function is_hex_commit(ref,   i,c){ \
		if (length(ref) < 7 || length(ref) > 40) return 0; \
		for (i=1; i<=length(ref); i++){ \
			c=tolower(substr(ref,i,1)); \
			if (index("0123456789abcdef", c)==0) return 0; \
		} \
		return 1; \
	} \
	function is_branch(ref,   low){ \
		ref=trim(ref); \
		if (ref=="") return 1; \
		low=tolower(ref); \
		if (low in branch_map) return 1; \
		if (index(ref,"/")>0) return 1; \
		if (is_hex_commit(ref)) return 1; \
		return 0; \
	} \
	function dir_exists(dir){ \
		cmd="[ -d \"" dir "\" ]"; \
		return system(cmd " >/dev/null 2>&1")==0; \
	} \
	function fetch_tags(dir){ \
		system("cd \"" dir "\" && git fetch --tags --prune --quiet >/dev/null 2>&1"); \
	} \
	function latest_tag(dir, fallback,    cmd, line, tag){ \
		tag=""; \
		cmd="cd \"" dir "\" && git tag --sort=-creatordate"; \
		while ((cmd | getline line) > 0){ \
			if (line ~ /^[ \t]*$$/) continue; \
			if (line ~ /(rc|beta)/) continue; \
			tag=line; \
			break; \
		} \
		close(cmd); \
		if (tag=="") tag=fallback; \
		return tag; \
	} \
	BEGIN{ \
		tmp=ENVIRON["TMPFILE"]; \
		split(ENVIRON["TAG_BRANCH_IGNORE"], arr, /[ \t]+/); \
		for (i in arr){ if (arr[i]!="") branch_map[tolower(arr[i])]=1; } \
	} \
	{ \
		if (NR==1){ print $$0 >> tmp; next; } \
		name=trim($$1); dir=trim($$2); url=trim($$3); ref=trim($$4); desc=$$5; \
		newref=ref; \
		if (!dir_exists(dir)){ \
			printf("  ⚠️  %s: cloning...\n", dir) > "/dev/stderr"; \
			system("git clone --quiet \"" url "\" \"" dir "\" >/dev/null 2>&1 || git clone \"" url "\" \"" dir "\""); \
		} \
		if (is_branch(ref)){ \
			printf("  ↩︎ %s: ref \"%s\" treated as branch/commit\n", dir, ref==""?"default":ref) > "/dev/stderr"; \
		}else{ \
			fetch_tags(dir); \
			candidate=latest_tag(dir, ref); \
			if (candidate!="" && candidate!=ref){ \
				newref=candidate; \
				printf("  🔼 %s: %s -> %s\n", dir, ref, newref) > "/dev/stderr"; \
			}else{ \
				printf("  ✅ %s: %s\n", dir, ref) > "/dev/stderr"; \
			} \
		} \
		printf("%s|%s|%s|%s|%s\n", name, dir, url, newref, desc) >> tmp; \
	} \
	END{close(tmp)}' $(REPOS_FILE); \
	if cmp -s $$tmp $(REPOS_FILE); then \
		rm $$tmp; \
		echo ""; \
		echo "No tag updates detected."; \
	else \
		mv $$tmp $(REPOS_FILE); \
		echo ""; \
		echo "$(REPOS_FILE) updated with latest tags. Running sync..."; \
		$(MAKE) $(REPOS_MK) >/dev/null; \
		$(MAKE) sync; \
	fi
