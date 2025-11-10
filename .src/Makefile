# Repository metadata lives in repos.csv (name,dir,url,ref,description)
REPOS_FILE := repos.csv
REPOS_MK := .repos.mk
$(if $(wildcard $(REPOS_FILE)),,$(error Missing $(REPOS_FILE). Please create it.))

$(REPOS_MK): $(REPOS_FILE)
	@awk -F',' 'NR==1 {next} \
	{ name=$$1; dir=$$2; url=$$3; ref=$$4; desc=$$5; var=name; gsub(/[^A-Za-z0-9]/,"_",var); var=toupper(var); \
	  printf "REPO_%s := %s|%s|%s|%s\n", var, dir, url, ref, desc; \
	  rv = (rv ? rv " REPO_" var : "REPO_" var); \
	  rn = (rn ? rn " " name : name); \
	  ar = (ar ? ar " " dir : dir); \
	} \
	END { \
	  printf "REPO_VARS := %s\n", rv; \
	  printf "REPO_NAMES := %s\n", rn; \
	  printf "ALL_REPOS := %s\n", ar; \
	}' $(REPOS_FILE) > $(REPOS_MK)

include $(REPOS_MK)

# Auto-computed variables
# Ensure derived values exist even if repos file was empty
REPO_VARS ?=
REPO_NAMES ?=
ALL_REPOS ?=
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
	@$(foreach var,$(sort $(REPO_VARS)),\
		printf "    • %-20s %s\n" "$(var)" "$($(var))";)
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
	@$(eval REPO_VAR := $(call target-to-var,$*))
	@$(if $($(REPO_VAR)),,$(error ❌ Unknown repo: $*. Run 'make list-repos' to see available repositories))
	$(call sync-repo-template,$(REPO_VAR))

## list: Show repository status, versions, and available repos
list:
	@echo "════════════════════════════════════════════════════════════════"
	@echo "  📦 Reference Repositories"
	@echo "════════════════════════════════════════════════════════════════"
	@echo ""
	@$(foreach var,$(sort $(REPO_VARS)),\
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
	@$(foreach var,$(sort $(REPO_VARS)),\
		DIR=$(call get-dir,$(var)); \
		REF=$(call get-ref,$(var)); \
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
update: sync
	@echo "════════════════════════════════════════════════════════════════"
	@echo "  🔁 Update Repository Tags"
	@echo "════════════════════════════════════════════════════════════════"
	@echo ""
	@cp $(REPOS_FILE) $(REPOS_FILE).bak
	@tmp=$$(mktemp); \
	TMPFILE="$$tmp" TAG_BRANCH_IGNORE="$(TAG_BRANCH_IGNORE)" awk -F',' '\
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
			printf("  ⚠️  %s: directory missing (run make sync)\n", dir) > "/dev/stderr"; \
		}else if (is_branch(ref)){ \
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
		printf("%s,%s,%s,%s,%s\n", name, dir, url, newref, desc) >> tmp; \
	} \
	END{close(tmp)}' $(REPOS_FILE); \
	if cmp -s $$tmp $(REPOS_FILE); then \
		rm $$tmp; \
		echo ""; \
		echo "No tag updates detected."; \
	else \
		mv $$tmp $(REPOS_FILE); \
		echo ""; \
		echo "$(REPOS_FILE) updated with latest tags."; \
		$(MAKE) $(REPOS_MK) >/dev/null; \
	fi
