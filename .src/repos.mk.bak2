# ================================================================================
# Repository Management Commands
# ================================================================================
# Manages reference repositories: validation, cloning, syncing, .gitignore generation
# Uses variables: REPOS_FILE, REPOS (defined in Makefile)
#
# Format: dir|url|ref|fork_of|description
#   dir: Local directory name (also used as repository identifier)
#   url: Full Git repository URL (e.g., https://github.com/owner/repo.git)
#   ref: Git ref to track (tag like v1.0.0, or branch like main)
#   fork_of: Empty for upstream repos, directory name of upstream for forks
#   description: Human-readable description
#
# Commands:
#   - doctor:    Validate repos.list format and content
#   - install:   Sync repos to match repos.list (tags: checkout only, branches: checkout + pull)
#   - status:    Show current state vs repos.list (shows if newer tags available)
#   - upgrade:   Update repos.list to latest tags from GitHub, then install
#   - gitignore: Generate .gitignore from repos.list
#
# ================================================================================

# Validates repos.list file exists
define check-repos-file
	[ -f $(REPOS_FILE) ] || { echo "❌ Error: $(REPOS_FILE) not found"; exit 1; }
endef

# ================================================================================
# Shared Functions - DRY principle
# ================================================================================

# Print unified table header (used by install, upgrade, status)
define print-table-header
	@printf "  %-30s %-15s %-18s %-20s %-15s %s / %s\n" "REPOSITORY" "STATUS" "AGE" "OWNER" "FORK" "VERSION" "BRANCH"
	@printf "  %-30s %-15s %-18s %-20s %-15s %s\n" "──────────────────────────────" "───────────────" "──────────────────" "────────────────────" "───────────────" "────────────────"
endef

# Fetch tags in parallel for all installed repos (used by status, upgrade)
define parallel-fetch-tags
	@echo "⏳ Fetching latest version information..."
	@while IFS='|' read -r dir url ref fork_of desc; do \
		[ "$$dir" = "dir" ] && continue; \
		[ -d "$$dir/.git" ] && (cd $$dir && git fetch --tags --quiet 2>/dev/null) & \
	done < $(REPOS_FILE); \
	wait
	@echo ""
endef

# Common shell snippet: Extract metadata from repo
# Sets: owner, fork_display, branch, age
# Usage: $(call extract-repo-metadata) (requires url, fork_of, dir in scope)
define extract-repo-metadata
	owner=$$(echo "$$url" | sed -E 's|https?://github.com/([^/]+)/.*|\1|'); \
	if [ -n "$$fork_of" ]; then fork_display="fork:$$fork_of"; else fork_display="-"; fi; \
	branch=$$(cd $$dir 2>/dev/null && git branch --show-current 2>/dev/null); \
	[ -z "$$branch" ] && branch="-"; \
	age=$$(cd $$dir 2>/dev/null && git log -1 --format="%ar" 2>/dev/null | sed 's/ ago//'); \
	[ -z "$$age" ] && age="-"
endef

## doctor: Check repos.list format and content
doctor:
	@$(call check-repos-file)
	@echo "🩺 Checking Repository Configuration"
	@echo ""
	@printf "  %-30s %-15s %-15s %s\n" "CHECK" "LINE" "STATUS" "DETAILS"
	@printf "  %-30s %-15s %-15s %s\n" "──────────────────────────────" "───────────────" "───────────────" "───────────"
	@errors=0; \
	output=$$(mktemp); \
	trap "rm -f $$output" EXIT; \
	seen_dirs=""; \
	seen_urls=""; \
	line_num=1; \
	while IFS='|' read -r dir url ref fork_of desc; do \
		[ "$$dir" = "dir" ] && line_num=$$((line_num + 1)) && continue; \
		if [ -z "$$dir" ] || [ -z "$$url" ] || [ -z "$$ref" ]; then \
			printf "  ❌ %-27s %-15s %-15s %s\n" "Missing fields" "$$line_num" "error" "dir/url/ref" >> $$output; \
			errors=$$((errors + 1)); \
		fi; \
		if ! echo "$$url" | grep -qE '^https?://'; then \
			printf "  ❌ %-27s %-15s %-15s %s\n" "Invalid URL" "$$line_num" "error" "$$url" >> $$output; \
			errors=$$((errors + 1)); \
		fi; \
		if echo " $$seen_dirs " | grep -q " $$dir "; then \
			printf "  ❌ %-27s %-15s %-15s %s\n" "Duplicate directory" "$$line_num" "error" "$$dir" >> $$output; \
			errors=$$((errors + 1)); \
		fi; \
		seen_dirs="$$seen_dirs $$dir"; \
		if echo " $$seen_urls " | grep -q " $$url "; then \
			printf "  ❌ %-27s %-15s %-15s %s\n" "Duplicate URL" "$$line_num" "error" "$$url" >> $$output; \
			errors=$$((errors + 1)); \
		fi; \
		seen_urls="$$seen_urls $$url"; \
		line_num=$$((line_num + 1)); \
	done < $(REPOS_FILE); \
	if [ $$errors -gt 0 ]; then \
		cat $$output; \
		echo ""; \
		echo "Summary: $$errors error(s) found"; \
		echo "Legend: ❌ error"; \
		exit 1; \
	else \
		total=$$((line_num - 2)); \
		printf "  ✅ %-27s %-15s %-15s %s\n" "Format" "-" "healthy" "$$total repos"; \
		printf "  ✅ %-27s %-15s %-15s %s\n" "URLs" "-" "healthy" "all valid"; \
		printf "  ✅ %-27s %-15s %-15s %s\n" "Directories" "-" "healthy" "no duplicates"; \
		echo ""; \
		echo "Summary: All checks passed"; \
		echo "Legend: ✅ healthy"; \
	fi

## gitignore: Generate .gitignore from repos.list (auto-run by install/upgrade)
gitignore:
	@echo "🔒 Generating .gitignore..."
	@{ \
		echo "# Auto-generated - do not edit manually"; \
		echo "# Regenerate with: make gitignore"; \
		echo ""; \
		echo "# Ignore all cloned repositories"; \
		echo "*/"; \
		echo ""; \
		echo "# Track meta files"; \
		echo "!.gitignore"; \
		echo "!README.md"; \
		echo "!Makefile"; \
		echo "!repos.list"; \
		echo "!stacks.list"; \
		echo "!agent-prompt.md"; \
		echo "!INDEX.md"; \
		echo "!STACK-*.md"; \
		echo "!DATASTAR.md"; \
		echo ""; \
		echo "# Backup files"; \
		echo "*.bak"; \
		echo "*.old"; \
	} > .gitignore
	@echo "✅ Generated .gitignore"

## install: Clone/update repositories from repos.list
install:
	@command -v git >/dev/null 2>&1 || { echo "❌ Error: git is required but not installed"; exit 1; }
	@[ -f $(REPOS_FILE) ] || { echo "❌ Error: $(REPOS_FILE) not found"; exit 1; }
	@$(MAKE) --no-print-directory gitignore
	@$(MAKE) --no-print-directory doctor
	@echo "📦 Synchronizing Reference Repositories"
	@echo ""
	$(call print-table-header)
	@start=$$(date +%s); \
	output=$$(mktemp); errors=$$(mktemp); \
	trap "rm -f $$output $$errors" EXIT; \
	while IFS='|' read -r dir url ref fork_of desc; do \
		[ "$$dir" = "dir" ] && continue; \
		{ $(call sync-one-inline); } & \
	done < $(REPOS_FILE); \
	wait; \
	LC_ALL=C sort -k1.5 $$output; \
	cloned=$$(grep -c "cloned" $$output || true); \
	synced=$$(grep -c "synced" $$output || true); \
	total=$$((cloned + synced)); \
	echo ""; \
	end=$$(date +%s); \
	elapsed=$$((end - start)); \
	if [ -s $$errors ]; then \
		echo "⚠️  Completed with errors ($${elapsed}s)"; \
		echo "Errors:"; \
		cat $$errors; \
		echo ""; \
	fi; \
	echo "Summary: $$total repos total - $$cloned cloned, $$synced synced ($${elapsed}s)"; \
	echo "Legend: ✅ cloned/synced"
	@echo ""

## status: Show current repository versions
status:
	@[ -f $(REPOS_FILE) ] || { echo "❌ Error: $(REPOS_FILE) not found"; exit 1; }
	@echo "📊 Repository Version Status"
	@echo ""
	$(call parallel-fetch-tags)
	$(call print-table-header)
	@output=$$(mktemp); \
	trap "rm -f $$output" EXIT; \
	uptodate=0; outdated=0; wrong=0; notinstalled=0; \
	while IFS='|' read -r dir url ref fork_of desc; do \
		[ "$$dir" = "dir" ] && continue; \
		if [ -d "$$dir/.git" ]; then \
			$(call extract-repo-metadata); \
			case "$$ref" in main|master|develop|head|trunk) \
				is_branch=1; \
				current=$$(cd $$dir && git branch --show-current 2>/dev/null || git rev-parse --short HEAD);; \
			*) \
				is_branch=0; \
				current=$$(cd $$dir && git describe --tags --exact-match 2>/dev/null || git branch --show-current 2>/dev/null || git rev-parse --short HEAD);; \
			esac; \
			latest=$$(cd $$dir && git tag --sort=-version:refname | grep -vE '(rc|beta|alpha|pre)' | head -1 2>/dev/null); \
			if [ "$$current" != "$$ref" ]; then \
				icon="⚠️ "; status="wrong"; wrong=$$((wrong + 1)); \
			elif [ $$is_branch -eq 0 ] && [ -n "$$latest" ] && [ "$$latest" != "$$current" ]; then \
				icon="🔼"; status="outdated"; outdated=$$((outdated + 1)); \
			else \
				icon="✅"; status="up-to-date"; uptodate=$$((uptodate + 1)); \
			fi; \
			if [ $$is_branch -eq 0 ] && [ -n "$$latest" ] && [ "$$latest" != "$$current" ]; then \
				printf "  %s %-27s %-15s %-18s %-20s %-15s %s / %s\n" "$$icon" "$$dir" "$$status" "$$age" "$$owner" "$$fork_display" "$$current (→$$latest)" "$$branch" >> $$output; \
			else \
				printf "  %s %-27s %-15s %-18s %-20s %-15s %s / %s\n" "$$icon" "$$dir" "$$status" "$$age" "$$owner" "$$fork_display" "$$current" "$$branch" >> $$output; \
			fi; \
		else \
			notinstalled=$$((notinstalled + 1)); \
			printf "  ❌ %-27s %-15s %-18s %-20s %-15s %s\n" "$$dir" "not installed" "-" "-" "-" "-" >> $$output; \
		fi; \
	done < $(REPOS_FILE); \
	LC_ALL=C sort -k1.5 $$output; \
	total=$$((uptodate + outdated + wrong + notinstalled)); \
	echo ""; \
	echo "Summary: $$total repos total - $$uptodate up-to-date, $$outdated outdated, $$wrong wrong version, $$notinstalled not installed"; \
	echo "Legend: ✅ up-to-date  🔼 outdated  ⚠️  wrong version  ❌ not installed"

## upgrade: Update repos.list to latest tags and install repositories
upgrade:
	@[ -f $(REPOS_FILE) ] || { echo "❌ Error: $(REPOS_FILE) not found"; exit 1; }
	@$(MAKE) --no-print-directory gitignore
	@echo "⬆️  Repository Upgrade Process"
	@echo ""
	@echo "📝 Step 1/2: Updating Configuration to Latest Versions"
	@echo ""
	@cp $(REPOS_FILE) $(REPOS_FILE).bak
	$(call parallel-fetch-tags)
	$(call print-table-header)
	@tmpfile=$$(mktemp); output=$$(mktemp); \
	updated=0; uptodate=0; branch=0; notag=0; notinstalled=0; \
	trap "rm -f $$tmpfile $$output" EXIT; \
	while IFS='|' read -r dir url ref fork_of desc; do \
		[ "$$dir" = "dir" ] && echo "$$dir|$$url|$$ref|$$fork_of|$$desc" >> $$tmpfile && continue; \
		if [ ! -d "$$dir/.git" ]; then \
			notinstalled=$$((notinstalled + 1)); \
			echo "$$dir|$$url|$$ref|$$fork_of|$$desc" >> $$tmpfile; \
			continue; \
		fi; \
		$(call extract-repo-metadata); \
		case "$$ref" in main|master|develop|head|trunk) \
			printf "  ↩︎  %-27s %-15s %-18s %-20s %-15s %s / %s\n" "$$dir" "on branch" "$$age" "$$owner" "$$fork_display" "$$ref" "$$branch" >> $$output; \
			echo "$$dir|$$url|$$ref|$$fork_of|$$desc" >> $$tmpfile; \
			branch=$$((branch + 1)); \
			continue;; \
		esac; \
		latest=$$(cd $$dir && git tag --sort=-version:refname | grep -vE '(rc|beta|alpha|pre)' | head -1 2>/dev/null); \
		if [ -z "$$latest" ]; then \
			printf "  ↩︎  %-27s %-15s %-18s %-20s %-15s %s / %s\n" "$$dir" "no tags" "$$age" "$$owner" "$$fork_display" "$$ref" "$$branch" >> $$output; \
			echo "$$dir|$$url|$$ref|$$fork_of|$$desc" >> $$tmpfile; \
			notag=$$((notag + 1)); \
			continue; \
		fi; \
		if [ "$$latest" != "$$ref" ]; then \
			printf "  🔼 %-27s %-15s %-18s %-20s %-15s %s / %s\n" "$$dir" "outdated" "$$age" "$$owner" "$$fork_display" "$$ref (→$$latest)" "$$branch" >> $$output; \
			echo "$$dir|$$url|$$latest|$$fork_of|$$desc" >> $$tmpfile; \
			updated=$$((updated + 1)); \
		else \
			printf "  ✅ %-27s %-15s %-18s %-20s %-15s %s / %s\n" "$$dir" "up-to-date" "$$age" "$$owner" "$$fork_display" "$$ref" "$$branch" >> $$output; \
			echo "$$dir|$$url|$$ref|$$fork_of|$$desc" >> $$tmpfile; \
			uptodate=$$((uptodate + 1)); \
		fi; \
	done < $(REPOS_FILE); \
	LC_ALL=C sort -k1.5 $$output; \
	total=$$((updated + uptodate + branch + notag + notinstalled)); \
	echo ""; \
	if [ $$updated -gt 0 ]; then \
		mv $$tmpfile $(REPOS_FILE) && echo "✅ Updated $$updated repo(s)"; \
	else \
		echo "✅ All repos up to date"; \
	fi; \
	echo "Summary: $$total repos total - $$updated upgraded, $$uptodate up-to-date, $$branch on branch, $$notag no tags, $$notinstalled not installed"
	@echo ""
	@echo "Legend: ✅ up-to-date  🔼 outdated  ↩︎  skipped"
	@echo ""
	@echo "📦 Step 2/2: Synchronizing Updated Repositories"
	@echo ""
	@$(MAKE) install

# Sync function for inline use (dir, url, ref, fork_of already set)
define sync-one-inline
	if [ ! -d "$$dir/.git" ]; then \
		if ! git clone --quiet $$url $$dir 2>&1 | grep -v "Cloning"; then \
			echo "Failed to clone $$dir from $$url" >> $$errors; \
		fi; \
		if [ -n "$$ref" ]; then \
			if ! (cd $$dir && git checkout --quiet $$ref 2>&1); then \
				echo "Failed to checkout $$ref in $$dir" >> $$errors; \
			fi; \
		fi; \
		$(call extract-repo-metadata); \
		current=$$(cd $$dir 2>/dev/null && git describe --tags --exact-match 2>/dev/null || git branch --show-current 2>/dev/null || git rev-parse --short HEAD 2>/dev/null); \
		printf "  ✅ %-27s %-15s %-18s %-20s %-15s %s / %s\n" "$$dir" "cloned" "$$age" "$$owner" "$$fork_display" "$${current:-unknown}" "$$branch" >> $$output; \
	else \
		if ! (cd $$dir && git fetch --all --tags --quiet 2>&1); then \
			echo "Failed to fetch in $$dir" >> $$errors; \
		fi; \
		if [ -n "$$ref" ]; then \
			if ! (cd $$dir && git checkout --quiet $$ref 2>&1); then \
				echo "Failed to checkout $$ref in $$dir" >> $$errors; \
			fi; \
			case "$$ref" in main|master|develop|head|trunk) \
				if ! (cd $$dir && git pull --quiet 2>&1); then \
					echo "Failed to pull in $$dir" >> $$errors; \
				fi;; \
			esac; \
		fi; \
		$(call extract-repo-metadata); \
		current=$$(cd $$dir 2>/dev/null && git describe --tags --exact-match 2>/dev/null || git branch --show-current 2>/dev/null || git rev-parse --short HEAD 2>/dev/null); \
		printf "  ✅ %-27s %-15s %-18s %-20s %-15s %s / %s\n" "$$dir" "synced" "$$age" "$$owner" "$$fork_display" "$${current:-unknown}" "$$branch" >> $$output; \
	fi
endef
