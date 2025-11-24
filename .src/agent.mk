# ================================================================================
# Agent Documentation Generator
# ================================================================================
# Generates INDEX.md and stack-specific .md files for AI agents
# Uses: repos.list, stacks.list, agent-prompt.md
#
# Process:
#   1. Cleans up old STACK-*.md files
#   2. Reads repos.list for repository metadata
#   3. Reads stacks.list for tech stack groupings with tags
#   4. Generates INDEX.md (overview with tag-based navigation)
#   5. Generates STACK-{name}.md for each tech stack
#
# Output: INDEX.md + STACK-*.md files
# ================================================================================

STACKS_FILE := stacks.list
PROMPT_FILE := agent-prompt.md

## agent: Generate INDEX.md and stack documentation
agent:
	@$(call check-repos-file)
	@echo "📝 Generating agent documentation..."
	@$(MAKE) --no-print-directory agent-clean
	@$(MAKE) --no-print-directory agent-index
	@$(MAKE) --no-print-directory agent-stacks
	@echo "✅ Generated INDEX.md and stack docs"

## agent-clean: Remove old STACK-*.md files before regeneration
agent-clean:
	@rm -f STACK-*.md
	@echo "  Cleaned old stack files"

## agent-index: Generate INDEX.md overview
agent-index:
	@{ \
	echo "# Tech Stack Reference for AI Agents"; \
	echo ""; \
	echo "> **Auto-generated** from repos.list and stacks.list"; \
	echo "> Last updated: $$(date '+%Y-%m-%d %H:%M:%S')"; \
	echo "> Run \`make agent\` to regenerate"; \
	echo ""; \
	echo "Quick reference for finding patterns in curated repositories. **Always check here before web searches.**"; \
	echo ""; \
	echo "---"; \
	echo ""; \
	echo "## 🏷️  Browse by Tag"; \
	echo ""; \
	all_tags=$$(mktemp); \
	while IFS='|' read -r stack repos tags desc; do \
		[ "$$stack" = "stack" ] && continue; \
		echo "$$tags" | tr ',' '\n' >> $$all_tags; \
	done < $(STACKS_FILE); \
	sort -u $$all_tags | while read -r tag; do \
		[ -z "$$tag" ] && continue; \
		echo -n "**$$tag**: "; \
		matching_stacks=""; \
		while IFS='|' read -r stack repos tags desc; do \
			[ "$$stack" = "stack" ] && continue; \
			if echo "$$tags" | grep -q "$$tag"; then \
				[ -n "$$matching_stacks" ] && matching_stacks="$$matching_stacks, "; \
				matching_stacks="$$matching_stacks[$$stack](STACK-$$stack.md)"; \
			fi; \
		done < $(STACKS_FILE); \
		echo "$$matching_stacks"; \
	done; \
	rm -f $$all_tags; \
	echo ""; \
	echo "---"; \
	echo ""; \
	echo "## 📚 Tech Stacks"; \
	echo ""; \
	while IFS='|' read -r stack repos tags desc; do \
		[ "$$stack" = "stack" ] && continue; \
		repo_count=$$(echo "$$repos" | tr ',' '\n' | wc -l | tr -d ' '); \
		echo "### [$$stack](STACK-$$stack.md)"; \
		echo ""; \
		echo "$$desc"; \
		echo ""; \
		echo "- **Tags**: \`$$(echo "$$tags" | sed 's/,/\`, \`/g')\`"; \
		echo "- **Repos**: $$repo_count"; \
		echo "- [View details →](STACK-$$stack.md)"; \
		echo ""; \
	done < $(STACKS_FILE); \
	echo "---"; \
	echo ""; \
	echo "## 📦 All Repositories"; \
	echo ""; \
	while IFS='|' read -r dir url ref fork_of desc; do \
		[ "$$dir" = "dir" ] && continue; \
		version_info=""; \
		if [ -d "$$dir/.git" ]; then \
			current=$$(cd $$dir 2>/dev/null && git describe --tags --exact-match 2>/dev/null || git branch --show-current 2>/dev/null || git rev-parse --short HEAD 2>/dev/null); \
			version_info=" (\`$$current\`)"; \
		fi; \
		fork_badge=""; \
		if [ -n "$$fork_of" ]; then \
			fork_badge=" 🍴"; \
		fi; \
		echo "- **$$dir/**$$version_info$$fork_badge - $$desc"; \
	done < $(REPOS_FILE); \
	echo ""; \
	echo "---"; \
	echo ""; \
	echo "## 📖 How to Use"; \
	echo ""; \
	echo "1. **Browse by Tag** - Find stacks by technology keyword above"; \
	echo "2. **Browse by Stack** - Click stack names for detailed docs"; \
	echo "3. **Find Patterns** - Stack docs show file locations and code patterns"; \
	echo "4. **Read agent-prompt.md** - Full system documentation"; \
	echo ""; \
	echo "**Always check .src/ before web searches!**"; \
	echo ""; \
	echo "---"; \
	echo ""; \
	echo "## 🔧 Commands"; \
	echo ""; \
	echo "\`\`\`bash"; \
	echo "make install   # Clone/sync all repos"; \
	echo "make status    # Show repo versions"; \
	echo "make agent     # Regenerate this documentation"; \
	echo "\`\`\`"; \
	} > INDEX.md

## agent-stacks: Generate individual stack .md files
agent-stacks:
	@while IFS='|' read -r stack repos tags desc; do \
		[ "$$stack" = "stack" ] && continue; \
		{ \
		echo "# $$stack Stack"; \
		echo ""; \
		echo "$$desc"; \
		echo ""; \
		echo "**Tags**: \`$$(echo "$$tags" | sed 's/,/\`, \`/g')\`"; \
		echo ""; \
		echo "---"; \
		echo ""; \
		echo "## 📦 Repositories"; \
		echo ""; \
		echo "$$repos" | tr ',' '\n' | while read -r repo; do \
			while IFS='|' read -r dir url ref fork_of repo_desc; do \
				[ "$$dir" = "dir" ] && continue; \
				if [ "$$dir" = "$$repo" ]; then \
					version=""; \
					if [ -d "$$dir/.git" ]; then \
						current=$$(cd $$dir 2>/dev/null && git describe --tags --exact-match 2>/dev/null || git branch --show-current 2>/dev/null || git rev-parse --short HEAD 2>/dev/null); \
						version=" (\`$$current\`)"; \
					fi; \
					fork_note=""; \
					if [ -n "$$fork_of" ]; then \
						fork_note=" - 🍴 Fork of $$fork_of"; \
					fi; \
					echo "### $$dir/$$version"; \
					echo ""; \
					echo "$$repo_desc$$fork_note"; \
					echo ""; \
					echo "- **URL**: $$url"; \
					echo "- **Ref**: \`$$ref\`"; \
					if [ -f "$$dir/README.md" ]; then \
						first_line=$$(head -1 "$$dir/README.md" 2>/dev/null | sed 's/^# //; s/^## //'); \
						echo "- **[README.md]($$dir/README.md)** - $$first_line"; \
					fi; \
					if [ -f "$$dir/CLAUDE.md" ]; then \
						first_line=$$(head -1 "$$dir/CLAUDE.md" 2>/dev/null | sed 's/^# //; s/^## //'); \
						echo "- **[CLAUDE.md]($$dir/CLAUDE.md)** - $$first_line"; \
					fi; \
					if [ -f "$$dir/go.mod" ]; then \
						module=$$(grep '^module ' "$$dir/go.mod" 2>/dev/null | awk '{print $$2}'); \
						echo "- **Go Module**: \`$$module\`"; \
					fi; \
					echo ""; \
				fi; \
			done < $(REPOS_FILE); \
		done; \
		echo "---"; \
		echo ""; \
		echo "## 🎯 When to Use"; \
		echo ""; \
		echo "Use this stack when working with: \`$$(echo "$$tags" | sed 's/,/\`, \`/g')\`"; \
		echo ""; \
		echo "---"; \
		echo ""; \
		echo "## 🔍 Common Patterns"; \
		echo ""; \
		case "$$stack" in \
			backend) \
				echo "### PocketBase Projects"; \
				echo ""; \
				echo "\`\`\`"; \
				echo "pkg/"; \
				echo "  cmd/pocketbase/main.go        # Entry point"; \
				echo "  pb_migrations/                # Database migrations"; \
				echo "  pb_data/                      # SQLite DB and uploads"; \
				echo "\`\`\`"; \
				echo ""; \
				echo "### Key Files"; \
				echo ""; \
				echo "- \`main.go\` - App initialization, hooks, routes"; \
				echo "- \`pb_migrations/*.go\` - Auto-generated migration files"; \
				echo "- \`pb_data/data.db\` - SQLite database"; \
				;; \
			hypermedia|components) \
				echo "### Component Structure"; \
				echo ""; \
				echo "\`\`\`"; \
				echo "components/                   # Reusable UI components"; \
				echo "examples/                     # Demo applications"; \
				echo "\`\`\`"; \
				echo ""; \
				echo "### Key Patterns"; \
				echo ""; \
				echo "- Type-safe HTML generation"; \
				echo "- Server-sent events for reactivity"; \
				echo "- Component composition"; \
				;; \
			*) \
				echo "Check README.md files in repos above for patterns."; \
				;; \
		esac; \
		echo ""; \
		echo "---"; \
		echo ""; \
		echo "## 🔗 Related Stacks"; \
		echo ""; \
		related_found=0; \
		while IFS='|' read -r other_stack other_repos other_tags other_desc; do \
			[ "$$other_stack" = "stack" ] && continue; \
			[ "$$other_stack" = "$$stack" ] && continue; \
			common_repos=$$(comm -12 <(echo "$$repos" | tr ',' '\n' | sort) <(echo "$$other_repos" | tr ',' '\n' | sort) | wc -l | tr -d ' '); \
			common_tags=$$(comm -12 <(echo "$$tags" | tr ',' '\n' | sort) <(echo "$$other_tags" | tr ',' '\n' | sort) | wc -l | tr -d ' '); \
			if [ $$common_repos -gt 0 ] || [ $$common_tags -gt 0 ]; then \
				echo "- [$$other_stack](STACK-$$other_stack.md) - $$other_desc"; \
				related_found=1; \
			fi; \
		done < $(STACKS_FILE); \
		if [ $$related_found -eq 0 ]; then \
			echo "None"; \
		fi; \
		echo ""; \
		echo "---"; \
		echo ""; \
		echo "[← Back to Index](INDEX.md) | [View agent-prompt.md](agent-prompt.md)"; \
		} > "STACK-$$stack.md"; \
		echo "  Generated STACK-$$stack.md"; \
	done < $(STACKS_FILE)
