# ================================================================================
# INDEX.md Generation Command
# ================================================================================
# Generates INDEX.md documentation from repository state and trigger mappings
# Uses variables: REPOS_FILE, INDEX_FILE (defined in Makefile)
# Uses shared define blocks: check-repos-file (from forks.mk)
#
# Process:
#   1. Reads repos.list for repository metadata
#   2. Reads index.list for trigger-to-file mappings
#   3. Queries git state (branches, tags) for version info
#   4. Generates markdown with trigger categories and file paths
#
# Output: INDEX.md (auto-trigger guide for Claude agents)
#
# ================================================================================

## index: Generate INDEX.md from current repository state
index:
	@$(call check-repos-file)
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
	while IFS='|' read -r name dir url ref fork_url desc; do \
		[ "$$name" = "name" ] && continue; \
		version_info=""; \
		fork_badge=""; \
		if [ -d "$$dir/.git" ]; then \
			current=$$(cd $$dir 2>/dev/null && git describe --tags --exact-match 2>/dev/null || git branch --show-current 2>/dev/null || git rev-parse --short HEAD 2>/dev/null); \
			version_info=" ($$current)"; \
			if [ -n "$$fork_url" ]; then \
				fork_badge=" 🍴"; \
			fi; \
		fi; \
		echo "- **$$dir/**$$version_info$$fork_badge - $$desc"; \
	done < $(REPOS_FILE); \
	echo ""; \
	echo "---"; \
	echo ""; \
	echo "## 🎯 When to Use These Files"; \
	echo ""; \
	echo "**AUTO-TRIGGER**: Read specified files IMMEDIATELY when user query matches triggers."; \
	echo ""; \
	if [ -f $(INDEX_FILE) ]; then \
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
				done < $(INDEX_FILE); \
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
		done < $(INDEX_FILE); \
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
	while IFS='|' read -r name dir url ref fork_url desc; do \
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
