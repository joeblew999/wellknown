# Simple repository sync Makefile
# Data files:
#   - repos.list: Repository metadata (name|dir|url|ref|description)
#   - triggers.list: Trigger-to-file mappings for Claude (category|priority|triggers|action|path|why)

REPOS_FILE := repos.list
TRIGGERS_FILE := triggers.list
SHELL := /bin/bash

# Read repo data once (using standard Unix tools)
REPOS := $(shell tail -n +2 $(REPOS_FILE) | cut -d'|' -f2)

.PHONY: help doctor install status upgrade index
.DEFAULT_GOAL := help

## help: Show available commands
help:
	@echo "Targets:"
	@grep -E '^##' $(firstword $(MAKEFILE_LIST)) | sed 's/^## /  /' | column -t -s ':'

## doctor: Check repos.list format and content
doctor:
	@[ -f $(REPOS_FILE) ] || { echo "❌ Error: $(REPOS_FILE) not found"; exit 1; }
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
	while IFS='|' read -r name dir url ref desc; do \
		[ "$$name" = "name" ] && line_num=$$((line_num + 1)) && continue; \
		if [ -z "$$name" ] || [ -z "$$dir" ] || [ -z "$$url" ] || [ -z "$$ref" ]; then \
			printf "  ❌ %-27s %-15s %-15s %s\n" "Missing fields" "$$line_num" "error" "name/dir/url/ref" >> $$output; \
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
		echo "!triggers.list"; \
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
	while IFS='|' read -r name dir url ref desc; do \
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
	@while IFS='|' read -r name dir url ref desc; do \
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
	while IFS='|' read -r name dir url ref desc; do \
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
	@while IFS='|' read -r name dir url ref desc; do \
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
	while IFS='|' read -r name dir url ref desc; do \
		[ "$$name" = "name" ] && echo "$$name|$$dir|$$url|$$ref|$$desc" >> $$tmpfile && continue; \
		if [ ! -d "$$dir/.git" ]; then \
			notinstalled=$$((notinstalled + 1)); \
			echo "$$name|$$dir|$$url|$$ref|$$desc" >> $$tmpfile; \
			continue; \
		fi; \
		case "$$ref" in main|master|develop|head|trunk) \
			printf "  ↩︎  %-27s %-15s %-15s %s\n" "$$dir" "$$ref" "on branch" "-" >> $$output; \
			echo "$$name|$$dir|$$url|$$ref|$$desc" >> $$tmpfile; \
			branch=$$((branch + 1)); \
			continue;; \
		esac; \
		latest=$$(cd $$dir && git tag --sort=-version:refname | grep -vE '(rc|beta|alpha|pre)' | head -1 2>/dev/null); \
		if [ -z "$$latest" ]; then \
			printf "  ↩︎  %-27s %-15s %-15s %s\n" "$$dir" "$$ref" "no tags" "-" >> $$output; \
			echo "$$name|$$dir|$$url|$$ref|$$desc" >> $$tmpfile; \
			notag=$$((notag + 1)); \
			continue; \
		fi; \
		if [ "$$latest" != "$$ref" ]; then \
			printf "  🔼 %-27s %-15s %-15s %s\n" "$$dir" "$$ref" "outdated" "$$latest" >> $$output; \
			echo "$$name|$$dir|$$url|$$latest|$$desc" >> $$tmpfile; \
			updated=$$((updated + 1)); \
		else \
			printf "  ✅ %-27s %-15s %-15s %s\n" "$$dir" "$$ref" "up-to-date" "-" >> $$output; \
			echo "$$name|$$dir|$$url|$$ref|$$desc" >> $$tmpfile; \
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

## index: Generate INDEX.md from current repository state
index:
	@[ -f $(REPOS_FILE) ] || { echo "❌ Error: $(REPOS_FILE) not found"; exit 1; }
	@echo "📝 Generating INDEX.md from repository data..."
	@{ \
	echo "# Reference Repository Index"; \
	echo ""; \
	echo "> **Auto-generated** from repos.list and discovered agent files"; \
	echo "> Last updated: $$(date '+%Y-%m-%d %H:%M:%S')"; \
	echo "> Run \`make index\` to regenerate"; \
	echo ""; \
	echo "Quick reference guide for finding patterns and examples in \`.src/\` repositories."; \
	echo ""; \
	echo "---"; \
	echo ""; \
	echo "## 📦 Available Repositories"; \
	echo ""; \
	while IFS='|' read -r name dir url ref desc; do \
		[ "$$name" = "name" ] && continue; \
		version_info=""; \
		if [ -d "$$dir/.git" ]; then \
			current=$$(cd $$dir 2>/dev/null && git describe --tags --exact-match 2>/dev/null || git branch --show-current 2>/dev/null || git rev-parse --short HEAD 2>/dev/null); \
			version_info=" ($$current)"; \
		fi; \
		echo "- **$$dir/**$$version_info - $$desc"; \
	done < $(REPOS_FILE); \
	echo ""; \
	echo "---"; \
	echo ""; \
	echo "## 🎯 When to Use These Files"; \
	echo ""; \
	echo "**AUTO-TRIGGER**: Read specified files IMMEDIATELY when user query matches triggers."; \
	echo ""; \
	if [ -f $(TRIGGERS_FILE) ]; then \
		prev_category=""; \
		category_why=""; \
		emoji_map="1:1️⃣ 2:2️⃣ 3:3️⃣ 4:4️⃣ 5:5️⃣ 6:6️⃣ 7:7️⃣ 8:8️⃣ 9:9️⃣"; \
		while IFS='|' read -r category priority triggers action path notes why; do \
			[ "$$category" = "category" ] && continue; \
			[[ "$$category" =~ ^# ]] && continue; \
			if [ "$$category" != "$$prev_category" ]; then \
				if [ -n "$$prev_category" ] && [ -n "$$category_why" ]; then \
					echo ""; \
					echo "**Why**: $$category_why"; \
					echo ""; \
					echo "---"; \
					echo ""; \
				elif [ -n "$$prev_category" ]; then \
					echo ""; \
					echo "---"; \
					echo ""; \
				fi; \
				emoji=$$(echo "$$emoji_map" | tr ' ' '\n' | grep "^$$priority:" | cut -d: -f2); \
				echo "### $${emoji} $$category"; \
				all_triggers=""; \
				tmpfile=$$(mktemp); \
				while IFS='|' read -r cat2 pri2 trigs2 act2 path2 notes2 why2; do \
					[ "$$cat2" = "category" ] && continue; \
					[[ "$$cat2" =~ ^# ]] && continue; \
					[ "$$cat2" = "$$category" ] && echo "$$trigs2" >> $$tmpfile; \
				done < $(TRIGGERS_FILE); \
				all_triggers=$$(cat $$tmpfile | tr ',' '\n' | sort -u | tr '\n' ',' | sed 's/,$$//'); \
				rm -f $$tmpfile; \
				echo "**Triggers**: \`$$(echo "$$all_triggers" | sed 's/,/\`, \`/g')\`"; \
				echo ""; \
				category_why="$$why"; \
				prev_category="$$category"; \
			fi; \
			case "$$action" in \
				read) \
					[ -n "$$notes" ] && echo "→ **Read immediately**: \`$$path\` ($$notes)" || echo "→ **Read immediately**: \`$$path\`";; \
				browse) \
					[ -n "$$notes" ] && echo "→ **Browse for patterns**: \`$$path\` ($$notes)" || echo "→ **Browse for patterns**: \`$$path\`";; \
				*) \
					[ -n "$$notes" ] && echo "→ **Check**: \`$$path\` ($$notes)" || echo "→ **Check**: \`$$path\`";; \
			esac; \
			echo ""; \
		done < $(TRIGGERS_FILE); \
		if [ -n "$$category_why" ]; then \
			echo "**Why**: $$category_why"; \
		fi; \
		echo ""; \
	else \
		echo "⚠️  No triggers.list found - using default configuration"; \
		echo ""; \
	fi; \
	echo "---"; \
	echo ""; \
	echo "## 📖 How to Use This Index"; \
	echo ""; \
	echo "1. **Trigger Match** → Use trigger keywords above to find relevant files"; \
	echo "2. **Browse Catalog** → Explore all available agent files below"; \
	echo "3. **Read CLAUDE.md** → Each repo's CLAUDE.md has project-specific context"; \
	echo ""; \
	echo "---"; \
	echo ""; \
	echo "## 📁 Auto-Discovered Agent Files"; \
	echo ""; \
	echo "All \`.claude/\` directories and CLAUDE.md files found in reference repositories:"; \
	echo ""; \
	while IFS='|' read -r name dir url ref desc; do \
		[ "$$name" = "name" ] && continue; \
		if [ ! -d "$$dir/.git" ]; then continue; fi; \
		has_agents=0; \
		if [ -f "$$dir/CLAUDE.md" ]; then has_agents=1; fi; \
		if [ -d "$$dir/.claude/commands" ] && [ $$(find "$$dir/.claude/commands" -name "*.md" 2>/dev/null | wc -l | tr -d ' ') -gt 0 ]; then has_agents=1; fi; \
		if [ -d "$$dir/.claude/agents" ] && [ $$(find "$$dir/.claude/agents" -name "*.md" 2>/dev/null | wc -l | tr -d ' ') -gt 0 ]; then has_agents=1; fi; \
		if [ $$has_agents -eq 0 ]; then continue; fi; \
		echo "### $$dir/ - $$desc"; \
		echo ""; \
		if [ -f "$$dir/CLAUDE.md" ]; then \
			first_line=$$(head -1 "$$dir/CLAUDE.md" | sed 's/^# //; s/^## //'); \
			if [ -n "$$first_line" ] && [ "$$first_line" != "CLAUDE.md" ]; then \
				echo "**[CLAUDE.md]($$dir/CLAUDE.md)** - $$first_line"; \
			else \
				echo "**[CLAUDE.md]($$dir/CLAUDE.md)** - Project context and architecture"; \
			fi; \
			echo ""; \
		fi; \
		cmd_count=0; agent_count=0; \
		if [ -d "$$dir/.claude/commands" ] || [ -d "$$dir/.claud/commands" ]; then \
			cmd_dir="$$dir/.claude/commands"; \
			[ -d "$$dir/.claud/commands" ] && cmd_dir="$$dir/.claud/commands"; \
			cmd_count=$$(find "$$cmd_dir" -name "*.md" 2>/dev/null | wc -l | tr -d ' '); \
		fi; \
		if [ -d "$$dir/.claude/agents" ] || [ -d "$$dir/.claud/agents" ]; then \
			agent_dir="$$dir/.claude/agents"; \
			[ -d "$$dir/.claud/agents" ] && agent_dir="$$dir/.claud/agents"; \
			agent_count=$$(find "$$agent_dir" -name "*.md" 2>/dev/null | wc -l | tr -d ' '); \
		fi; \
		if [ $$cmd_count -gt 0 ] || [ $$agent_count -gt 0 ]; then \
			echo "**Available**: $$cmd_count commands, $$agent_count agents"; \
			echo ""; \
			if [ $$cmd_count -gt 0 ]; then \
				cmd_dir="$$dir/.claude/commands"; \
				[ -d "$$dir/.claud/commands" ] && cmd_dir="$$dir/.claud/commands"; \
				echo "<details><summary>Commands ($$cmd_count)</summary>"; \
				echo ""; \
				find "$$cmd_dir" -name "*.md" 2>/dev/null | sort | while read -r f; do \
					cmd_name=$$(basename "$$f" .md); \
					desc=$$(grep -m1 '^description:' "$$f" 2>/dev/null | sed 's/^description: *//; s/^description://'); \
					if [ -n "$$desc" ]; then \
						echo "- [\`$$cmd_name\`]($$f) - $$desc"; \
					else \
						echo "- [\`$$cmd_name\`]($$f)"; \
					fi; \
				done; \
				echo ""; \
				echo "</details>"; \
				echo ""; \
			fi; \
			if [ $$agent_count -gt 0 ]; then \
				agent_dir="$$dir/.claude/agents"; \
				[ -d "$$dir/.claud/agents" ] && agent_dir="$$dir/.claud/agents"; \
				echo "<details><summary>Agents ($$agent_count)</summary>"; \
				echo ""; \
				find "$$agent_dir" -name "*.md" 2>/dev/null | sort | while read -r f; do \
					agent_name=$$(basename "$$f" .md); \
					desc=$$(grep -m1 '^description:' "$$f" 2>/dev/null | sed 's/^description: *//; s/^description://'); \
					if [ -n "$$desc" ]; then \
						echo "- [\`$$agent_name\`]($$f) - $$desc"; \
					else \
						echo "- [\`$$agent_name\`]($$f)"; \
					fi; \
				done; \
				echo ""; \
				echo "</details>"; \
				echo ""; \
			fi; \
		fi; \
		echo ""; \
	done < $(REPOS_FILE); \
	echo "---"; \
	echo ""; \
	echo "## 🔍 Discovery Commands"; \
	echo ""; \
	echo "\`\`\`bash"; \
	echo "# Sync all repositories and regenerate this index"; \
	echo "make install"; \
	echo ""; \
	echo "# Check repository status"; \
	echo "make status"; \
	echo ""; \
	echo "# Manually regenerate this index"; \
	echo "make index"; \
	echo "\`\`\`"; \
	echo ""; \
	echo "---"; \
	echo ""; \
	echo "**Always check \`.src/\` before web searches!**"; \
	} > INDEX.md
	@echo "✅ Generated INDEX.md"

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
