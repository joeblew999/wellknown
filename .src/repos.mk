# ================================================================================
# Repository Management Commands
# ================================================================================
# Manages reference repositories: validation, cloning, syncing, .gitignore generation
# Uses variables: REPOS_FILE, REPOS (defined in Makefile)
# Uses shared define blocks: check-repos-file (from forks.mk)
#
# Commands:
#   - doctor:    Validate repos.list format and content
#   - install:   Clone missing repositories
#   - status:    Show sync status for all repos
#   - upgrade:   Pull latest changes for all repos
#   - gitignore: Generate .gitignore from repos.list
#
# ================================================================================

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
	while IFS='|' read -r name dir url ref fork_url desc; do \
		[ "$$name" = "name" ] && line_num=$$((line_num + 1)) && continue; \
		if [ -z "$$name" ] || [ -z "$$dir" ] || [ -z "$$url" ] || [ -z "$$ref" ]; then \
			printf "  ❌ %-27s %-15s %-15s %s\n" "Missing fields" "$$line_num" "error" "name/dir/url/ref" >> $$output; \
			errors=$$((errors + 1)); \
		fi; \
		if ! echo "$$url" | grep -qE '^https?://'; then \
			printf "  ❌ %-27s %-15s %-15s %s\n" "Invalid URL" "$$line_num" "error" "$$url" >> $$output; \
			errors=$$((errors + 1)); \
		fi; \
		if [ -n "$$fork_url" ] && ! echo "$$fork_url" | grep -qE '^https?://'; then \
			printf "  ❌ %-27s %-15s %-15s %s\n" "Invalid fork URL" "$$line_num" "error" "$$fork_url" >> $$output; \
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
		echo "!index.list"; \
		echo "!INDEX.md"; \
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
	@start=$$(date +%s); \
	printf "  %-30s %-15s %-15s %s\n" "REPOSITORY" "VERSION" "STATUS" "TYPE"; \
	printf "  %-30s %-15s %-15s %s\n" "──────────────────────────────" "───────────────" "───────────────" "────"; \
	output=$$(mktemp); errors=$$(mktemp); \
	trap "rm -f $$output $$errors" EXIT; \
	while IFS='|' read -r name dir url ref fork_url desc; do \
		[ "$$name" = "name" ] && continue; \
		{ $(call sync-one-inline); } & \
	done < $(REPOS_FILE); \
	wait; \
	LC_ALL=C sort -t' ' -k3,3 $$output; \
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
	echo "Legend: ✅ success"
	@echo ""
	@$(MAKE) --no-print-directory index

## status: Show current repository versions
status:
	@[ -f $(REPOS_FILE) ] || { echo "❌ Error: $(REPOS_FILE) not found"; exit 1; }
	@echo "📊 Repository Version Status"
	@echo ""
	@echo "⏳ Fetching latest version information..."
	@while IFS='|' read -r name dir url ref fork_url desc; do \
		[ "$$name" = "name" ] && continue; \
		[ -d "$$dir/.git" ] && (cd $$dir && git fetch --tags --quiet 2>/dev/null) & \
	done < $(REPOS_FILE); \
	wait
	@echo ""
	@printf "  %-30s %-15s %-15s %s\n" "REPOSITORY" "VERSION" "STATUS" "AGE"
	@printf "  %-30s %-15s %-15s %s\n" "──────────────────────────────" "───────────────" "───────────────" "───"
	@output=$$(mktemp); \
	trap "rm -f $$output" EXIT; \
	uptodate=0; outdated=0; wrong=0; notinstalled=0; \
	while IFS='|' read -r name dir url ref fork_url desc; do \
		[ "$$name" = "name" ] && continue; \
		if [ -d "$$dir/.git" ]; then \
			current=$$(cd $$dir && git describe --tags --exact-match 2>/dev/null || git branch --show-current 2>/dev/null || git rev-parse --short HEAD); \
			latest=$$(cd $$dir && git tag --sort=-version:refname | grep -vE '(rc|beta|alpha|pre)' | head -1 2>/dev/null); \
			age=$$(cd $$dir && git log -1 --format="%ar" 2>/dev/null | sed 's/ ago//'); \
			case "$$ref" in main|master|develop|head|trunk) \
				is_branch=1;; \
			*) \
				is_branch=0;; \
			esac; \
			if [ "$$current" != "$$ref" ]; then \
				icon="⚠️ "; status="wrong"; wrong=$$((wrong + 1)); \
			elif [ $$is_branch -eq 0 ] && [ -n "$$latest" ] && [ "$$latest" != "$$current" ]; then \
				icon="🔼"; status="outdated"; outdated=$$((outdated + 1)); \
			else \
				icon="✅"; status="up-to-date"; uptodate=$$((uptodate + 1)); \
			fi; \
			if [ $$is_branch -eq 0 ] && [ -n "$$latest" ] && [ "$$latest" != "$$current" ]; then \
				printf "  %s %-27s %-15s %-15s %s\n" "$$icon" "$$dir" "$$current (→$$latest)" "$$status" "$$age" >> $$output; \
			else \
				printf "  %s %-27s %-15s %-15s %s\n" "$$icon" "$$dir" "$$current" "$$status" "$$age" >> $$output; \
			fi; \
		else \
			notinstalled=$$((notinstalled + 1)); \
			printf "  ❌ %-27s %-15s %-15s %s\n" "$$dir" "not cloned" "not installed" "-" >> $$output; \
		fi; \
	done < $(REPOS_FILE); \
	LC_ALL=C sort -t' ' -k3,3 $$output; \
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
	@echo "⏳ Fetching latest version information..."
	@while IFS='|' read -r name dir url ref fork_url desc; do \
		[ "$$name" = "name" ] && continue; \
		[ -d "$$dir/.git" ] && (cd $$dir && git fetch --tags --quiet 2>/dev/null) & \
	done < $(REPOS_FILE); \
	wait
	@echo ""
	@printf "  %-30s %-15s %-15s %s\n" "REPOSITORY" "VERSION" "STATUS" "UPGRADE"
	@printf "  %-30s %-15s %-15s %s\n" "──────────────────────────────" "───────────────" "───────────────" "───────"
	@tmpfile=$$(mktemp); output=$$(mktemp); \
	updated=0; uptodate=0; branch=0; notag=0; notinstalled=0; \
	trap "rm -f $$tmpfile $$output" EXIT; \
	while IFS='|' read -r name dir url ref fork_url desc; do \
		[ "$$name" = "name" ] && echo "$$name|$$dir|$$url|$$ref|$$fork_url|$$desc" >> $$tmpfile && continue; \
		if [ ! -d "$$dir/.git" ]; then \
			notinstalled=$$((notinstalled + 1)); \
			echo "$$name|$$dir|$$url|$$ref|$$fork_url|$$desc" >> $$tmpfile; \
			continue; \
		fi; \
		case "$$ref" in main|master|develop|head|trunk) \
			printf "  ↩︎  %-27s %-15s %-15s %s\n" "$$dir" "$$ref" "on branch" "-" >> $$output; \
			echo "$$name|$$dir|$$url|$$ref|$$fork_url|$$desc" >> $$tmpfile; \
			branch=$$((branch + 1)); \
			continue;; \
		esac; \
		latest=$$(cd $$dir && git tag --sort=-version:refname | grep -vE '(rc|beta|alpha|pre)' | head -1 2>/dev/null); \
		if [ -z "$$latest" ]; then \
			printf "  ↩︎  %-27s %-15s %-15s %s\n" "$$dir" "$$ref" "no tags" "-" >> $$output; \
			echo "$$name|$$dir|$$url|$$ref|$$fork_url|$$desc" >> $$tmpfile; \
			notag=$$((notag + 1)); \
			continue; \
		fi; \
		if [ "$$latest" != "$$ref" ]; then \
			printf "  🔼 %-27s %-15s %-15s %s\n" "$$dir" "$$ref" "outdated" "$$latest" >> $$output; \
			echo "$$name|$$dir|$$url|$$latest|$$fork_url|$$desc" >> $$tmpfile; \
			updated=$$((updated + 1)); \
		else \
			printf "  ✅ %-27s %-15s %-15s %s\n" "$$dir" "$$ref" "up-to-date" "-" >> $$output; \
			echo "$$name|$$dir|$$url|$$ref|$$fork_url|$$desc" >> $$tmpfile; \
			uptodate=$$((uptodate + 1)); \
		fi; \
	done < $(REPOS_FILE); \
	LC_ALL=C sort -t' ' -k3,3 $$output; \
	total=$$((updated + uptodate + branch + notag + notinstalled)); \
	echo ""; \
	if [ $$updated -gt 0 ]; then \
		mv $$tmpfile $(REPOS_FILE) && echo "✅ Updated $$updated repo(s)"; \
	else \
		echo "✅ All repos up to date"; \
	fi; \
	echo "Summary: $$total repos total - $$updated upgraded, $$uptodate up-to-date, $$branch on branch, $$notag no tags, $$notinstalled not installed"
	@echo ""
	@echo "Legend: ✅ up-to-date  🔼 upgraded  ↩︎  skipped"
	@echo ""
	@echo "📦 Step 2/2: Synchronizing Updated Repositories"
	@echo ""
	@$(MAKE) install

## fork-add: Add fork remote to repositories with fork_url
fork-add:
	@[ -f $(REPOS_FILE) ] || { echo "❌ Error: $(REPOS_FILE) not found"; exit 1; }
	@echo "🍴 Adding Fork Remotes"
	@echo ""
	@printf "  %-30s %-15s %s\n" "REPOSITORY" "STATUS" "FORK URL"
	@printf "  %-30s %-15s %s\n" "──────────────────────────────" "───────────────" "────────"
	@added=0; skipped=0; exists=0; \
	while IFS='|' read -r name dir url ref fork_url desc; do \
		[ "$$name" = "name" ] && continue; \
		if [ -z "$$fork_url" ]; then \
			skipped=$$((skipped + 1)); \
			continue; \
		fi; \
		if [ ! -d "$$dir/.git" ]; then \
			printf "  ⚠️  %-27s %-15s %s\n" "$$dir" "not cloned" "$$fork_url"; \
			continue; \
		fi; \
		if (cd $$dir && git remote get-url fork >/dev/null 2>&1); then \
			current_fork=$$(cd $$dir && git remote get-url fork); \
			if [ "$$current_fork" = "$$fork_url" ]; then \
				printf "  ✅ %-27s %-15s %s\n" "$$dir" "exists" "$$fork_url"; \
				exists=$$((exists + 1)); \
			else \
				(cd $$dir && git remote set-url fork $$fork_url); \
				printf "  🔄 %-27s %-15s %s\n" "$$dir" "updated" "$$fork_url"; \
				added=$$((added + 1)); \
			fi; \
		else \
			(cd $$dir && git remote add fork $$fork_url); \
			printf "  ✅ %-27s %-15s %s\n" "$$dir" "added" "$$fork_url"; \
			added=$$((added + 1)); \
		fi; \
	done < $(REPOS_FILE); \
	echo ""; \
	total=$$((added + exists)); \
	echo "Summary: $$total forks configured - $$added added/updated, $$exists already exists"; \
	echo "Legend: ✅ success  🔄 updated  ⚠️  warning"

## fork-sync: Sync fork remotes (fetch from upstream, push to fork)
fork-sync:
	@[ -f $(REPOS_FILE) ] || { echo "❌ Error: $(REPOS_FILE) not found"; exit 1; }
	@$(MAKE) --no-print-directory fork-add
	@echo ""
	@echo "🔄 Syncing Forks with Upstream"
	@echo ""
	@printf "  %-30s %-15s %s\n" "REPOSITORY" "STATUS" "DETAILS"
	@printf "  %-30s %-15s %s\n" "──────────────────────────────" "───────────────" "───────"
	@synced=0; skipped=0; \
	while IFS='|' read -r name dir url ref fork_url desc; do \
		[ "$$name" = "name" ] && continue; \
		if [ -z "$$fork_url" ]; then \
			skipped=$$((skipped + 1)); \
			continue; \
		fi; \
		if [ ! -d "$$dir/.git" ]; then \
			printf "  ⚠️  %-27s %-15s %s\n" "$$dir" "not cloned" "-"; \
			continue; \
		fi; \
		current_branch=$$(cd $$dir && git branch --show-current); \
		if (cd $$dir && git fetch origin --quiet 2>&1); then \
			if (cd $$dir && git push fork $$current_branch --quiet 2>&1); then \
				printf "  ✅ %-27s %-15s %s\n" "$$dir" "synced" "$$current_branch → fork"; \
				synced=$$((synced + 1)); \
			else \
				printf "  ❌ %-27s %-15s %s\n" "$$dir" "push failed" "$$current_branch"; \
			fi; \
		else \
			printf "  ❌ %-27s %-15s %s\n" "$$dir" "fetch failed" "-"; \
		fi; \
	done < $(REPOS_FILE); \
	echo ""; \
	echo "Summary: $$synced forks synced"; \
	echo "Legend: ✅ success  ❌ error  ⚠️  warning"

## fork-push: Push current branch to fork remote
fork-push:
	@[ -f $(REPOS_FILE) ] || { echo "❌ Error: $(REPOS_FILE) not found"; exit 1; }
	@echo "📤 Pushing to Fork Remotes"
	@echo ""
	@printf "  %-30s %-15s %s\n" "REPOSITORY" "BRANCH" "STATUS"
	@printf "  %-30s %-15s %s\n" "──────────────────────────────" "───────────────" "──────"
	@pushed=0; skipped=0; \
	while IFS='|' read -r name dir url ref fork_url desc; do \
		[ "$$name" = "name" ] && continue; \
		if [ -z "$$fork_url" ]; then \
			skipped=$$((skipped + 1)); \
			continue; \
		fi; \
		if [ ! -d "$$dir/.git" ]; then \
			printf "  ⚠️  %-27s %-15s %s\n" "$$dir" "-" "not cloned"; \
			continue; \
		fi; \
		current_branch=$$(cd $$dir && git branch --show-current); \
		if [ -z "$$current_branch" ]; then \
			printf "  ⚠️  %-27s %-15s %s\n" "$$dir" "-" "detached HEAD"; \
			continue; \
		fi; \
		if (cd $$dir && git push fork $$current_branch --quiet 2>&1); then \
			printf "  ✅ %-27s %-15s %s\n" "$$dir" "$$current_branch" "pushed"; \
			pushed=$$((pushed + 1)); \
		else \
			printf "  ❌ %-27s %-15s %s\n" "$$dir" "$$current_branch" "failed"; \
		fi; \
	done < $(REPOS_FILE); \
	echo ""; \
	echo "Summary: $$pushed branches pushed to fork"; \
	echo "Legend: ✅ success  ❌ error  ⚠️  warning"

## fork-status: Show fork sync status
fork-status:
	@[ -f $(REPOS_FILE) ] || { echo "❌ Error: $(REPOS_FILE) not found"; exit 1; }
	@echo "📊 Fork Status"
	@echo ""
	@printf "  %-30s %-15s %-20s %s\n" "REPOSITORY" "BRANCH" "AHEAD/BEHIND" "STATUS"
	@printf "  %-30s %-15s %-20s %s\n" "──────────────────────────────" "───────────────" "────────────────────" "──────"
	@has_forks=0; \
	while IFS='|' read -r name dir url ref fork_url desc; do \
		[ "$$name" = "name" ] && continue; \
		if [ -z "$$fork_url" ]; then continue; fi; \
		has_forks=1; \
		if [ ! -d "$$dir/.git" ]; then \
			printf "  ⚠️  %-27s %-15s %-20s %s\n" "$$dir" "-" "-" "not cloned"; \
			continue; \
		fi; \
		if ! (cd $$dir && git remote get-url fork >/dev/null 2>&1); then \
			printf "  ⚠️  %-27s %-15s %-20s %s\n" "$$dir" "-" "-" "no fork remote"; \
			continue; \
		fi; \
		current_branch=$$(cd $$dir && git branch --show-current); \
		if [ -z "$$current_branch" ]; then \
			printf "  ⚠️  %-27s %-15s %-20s %s\n" "$$dir" "-" "-" "detached HEAD"; \
			continue; \
		fi; \
		(cd $$dir && git fetch fork --quiet 2>/dev/null); \
		if (cd $$dir && git rev-parse fork/$$current_branch >/dev/null 2>&1); then \
			ahead=$$(cd $$dir && git rev-list --count fork/$$current_branch..$$current_branch 2>/dev/null || echo "0"); \
			behind=$$(cd $$dir && git rev-list --count $$current_branch..fork/$$current_branch 2>/dev/null || echo "0"); \
			if [ "$$ahead" -eq 0 ] && [ "$$behind" -eq 0 ]; then \
				printf "  ✅ %-27s %-15s %-20s %s\n" "$$dir" "$$current_branch" "up to date" "synced"; \
			elif [ "$$ahead" -gt 0 ] && [ "$$behind" -eq 0 ]; then \
				printf "  🔼 %-27s %-15s %-20s %s\n" "$$dir" "$$current_branch" "ahead $$ahead" "push needed"; \
			elif [ "$$ahead" -eq 0 ] && [ "$$behind" -gt 0 ]; then \
				printf "  🔽 %-27s %-15s %-20s %s\n" "$$dir" "$$current_branch" "behind $$behind" "pull needed"; \
			else \
				printf "  ⚠️  %-27s %-15s %-20s %s\n" "$$dir" "$$current_branch" "±$$ahead/$$behind" "diverged"; \
			fi; \
		else \
			printf "  📤 %-27s %-15s %-20s %s\n" "$$dir" "$$current_branch" "-" "not on fork"; \
		fi; \
	done < $(REPOS_FILE); \
	if [ $$has_forks -eq 0 ]; then \
		echo "  No repositories with fork_url configured"; \
		echo ""; \
		echo "Add fork URLs to repos.list to enable fork management"; \
	fi; \
	echo ""; \
	echo "Legend: ✅ synced  🔼 ahead  🔽 behind  ⚠️  diverged  📤 not on fork"

# Sync function for inline use (dir, url, ref already set)
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
		current=$$(cd $$dir 2>/dev/null && git describe --tags --exact-match 2>/dev/null || git branch --show-current 2>/dev/null || git rev-parse --short HEAD 2>/dev/null); \
		case "$$ref" in main|master|develop|head|trunk) type="branch";; *) type="tag";; esac; \
		printf "  ✅ %-27s %-15s %-15s %s\n" "$$dir" "$${current:-unknown}" "cloned" "$$type" >> $$output; \
	else \
		if ! (cd $$dir && git fetch --all --tags --quiet 2>&1); then \
			echo "Failed to fetch in $$dir" >> $$errors; \
		fi; \
		if [ -n "$$ref" ]; then \
			if ! (cd $$dir && git checkout --quiet $$ref 2>&1); then \
				echo "Failed to checkout $$ref in $$dir" >> $$errors; \
			fi; \
		fi; \
		if ! (cd $$dir && git pull --quiet 2>&1); then \
			echo "Failed to pull in $$dir" >> $$errors; \
		fi; \
		current=$$(cd $$dir 2>/dev/null && git describe --tags --exact-match 2>/dev/null || git branch --show-current 2>/dev/null || git rev-parse --short HEAD 2>/dev/null); \
		case "$$ref" in main|master|develop|head|trunk) type="branch";; *) type="tag";; esac; \
		printf "  ✅ %-27s %-15s %-15s %s\n" "$$dir" "$${current:-unknown}" "synced" "$$type" >> $$output; \
	fi
endef
