# ================================================================================
# Fork Workflow Commands
# ================================================================================
# Manages fork-based development workflow: context, branches, commits, sync
# Uses variables: FORK_FILE, REPOS_FILE (defined in Makefile)
# Uses optional parameters: REPO, BRANCH, MSG
#
# Context system:
#   - fork.list stores current working repo to eliminate REPO= flags
#   - Commands read REPO parameter first, fall back to context file
#   - Use 'make fork-context REPO=name' to set context once
#
# ================================================================================

# ────────────────────────────────────────────────────────────────────────────────
# Shared Define Blocks (used by multiple .mk modules)
# ────────────────────────────────────────────────────────────────────────────────

# Validates repos.list file exists
# Used by: repos.mk (doctor, install, upgrade), forks.mk (all commands), index.mk
# Usage: $(call check-repos-file)
define check-repos-file
	[ -f $(REPOS_FILE) ] || { echo "❌ Error: $(REPOS_FILE) not found"; exit 1; }
endef

# Gets target repo from REPO parameter or fork.list context
# Used by: forks.mk fork commands (branch, commit, save, switch, guide)
# Sets: TARGET_REPO shell variable (exported)
# Usage: $(call get-target-repo,command-name)
define get-target-repo
	target_repo="$(REPO)"; \
	if [ -z "$$target_repo" ] && [ -f $(FORK_FILE) ]; then \
		target_repo=$$(grep "^current_repo=" $(FORK_FILE) | cut -d= -f2); \
	fi; \
	if [ -z "$$target_repo" ]; then \
		echo "❌ Error: No repository specified"; \
		echo ""; \
		echo "Either:"; \
		echo "  1. Set context: make fork-context REPO=via"; \
		echo "  2. Specify repo: make $(1) REPO=via"; \
		exit 1; \
	fi; \
	export TARGET_REPO=$$target_repo
endef

# Determines git base branch (main or master) from remote
# Used by: forks.mk (branch, guide, sync commands)
# Returns: Branch name string
# Usage: base_branch=$$($(call get-base-branch,$$dir))
define get-base-branch
	cd $(1) && git remote show origin | grep "HEAD branch" | cut -d ":" -f 2 | tr -d ' '
endef

# ────────────────────────────────────────────────────────────────────────────────
# Interactive Prompt Helpers
# ────────────────────────────────────────────────────────────────────────────────

# Prompts user to select a repository with fork configured
# Returns: Selected repo name via stdout
# Usage: selected=$$($(call prompt-for-repo))
define prompt-for-repo
	tmpfile=$$(mktemp); \
	trap "rm -f $$tmpfile" EXIT; \
	echo "📦 Available repositories with forks:" >&2; \
	echo "" >&2; \
	idx=1; \
	while IFS='|' read -r name dir url ref fork_url desc; do \
		[ "$$name" = "name" ] && continue; \
		[ -z "$$fork_url" ] && continue; \
		echo "$$idx|$$name" >> $$tmpfile; \
		printf "  %d) %s\n" $$idx "$$name" >&2; \
		idx=$$((idx + 1)); \
	done < $(REPOS_FILE); \
	total=$$((idx - 1)); \
	echo "" >&2; \
	printf "Select repository (1-%d): " $$total >&2; \
	read -r choice </dev/tty; \
	selected=$$(grep "^$$choice|" $$tmpfile | cut -d'|' -f2); \
	if [ -z "$$selected" ]; then \
		echo "❌ Invalid selection" >&2; \
		exit 1; \
	fi; \
	echo "$$selected"
endef

# Prompts user to select existing branch or create new one
# Param: Directory path
# Returns: Selected/new branch name via stdout
# Usage: selected=$$($(call prompt-for-branch,$$dir))
define prompt-for-branch
	tmpfile=$$(mktemp); \
	trap "rm -f $$tmpfile" EXIT; \
	echo "🌿 Branches in $(1):" >&2; \
	echo "" >&2; \
	printf "  %d) %s\n" 0 "<create new branch>" >&2; \
	idx=1; \
	cd $(1) && git branch 2>/dev/null | sed 's/^[* ]*//' | while read -r branch; do \
		echo "$$idx|$$branch" >> $$tmpfile; \
		printf "  %d) %s\n" $$idx "$$branch" >&2; \
		idx=$$((idx + 1)); \
	done; \
	total=$$(wc -l < $$tmpfile | tr -d ' '); \
	echo "" >&2; \
	printf "Select branch (0 for new, 1-%s for existing): " $$total >&2; \
	read -r choice </dev/tty; \
	if [ "$$choice" = "0" ]; then \
		printf "Enter new branch name: " >&2; \
		read -r new_branch </dev/tty; \
		if [ -z "$$new_branch" ]; then \
			echo "❌ Branch name cannot be empty" >&2; \
			exit 1; \
		fi; \
		echo "$$new_branch"; \
	else \
		selected=$$(grep "^$$choice|" $$tmpfile | cut -d'|' -f2); \
		if [ -z "$$selected" ]; then \
			echo "❌ Invalid selection" >&2; \
			exit 1; \
		fi; \
		echo "$$selected"; \
	fi
endef

# Prompts user for commit message
# Returns: Message via stdout
# Usage: msg=$$($(call prompt-for-message))
define prompt-for-message
	printf "Commit message: " >&2; \
	read -r msg </dev/tty; \
	if [ -z "$$msg" ]; then \
		echo "❌ Commit message cannot be empty" >&2; \
		exit 1; \
	fi; \
	echo "$$msg"
endef

# Prompts user for branch name with validation
# Returns: Branch name via stdout
# Usage: branch_name=$$($(call prompt-for-branch-name))
define prompt-for-branch-name
	printf "Branch name: " >&2; \
	read -r branch_name </dev/tty; \
	branch_name=$$(echo "$$branch_name" | tr -d ' \t\n\r'); \
	if [ -z "$$branch_name" ]; then \
		echo "❌ Branch name required" >&2; \
		exit 1; \
	fi; \
	echo "$$branch_name"
endef

# ────────────────────────────────────────────────────────────────────────────────
# Fork Workflow Commands
# ────────────────────────────────────────────────────────────────────────────────

## fork-context: Set or show current repository context (prompts if REPO not provided)
fork-context:
	@$(call check-repos-file)
	@if [ -n "$(REPO)" ]; then \
		echo "current_repo=$(REPO)" > $(FORK_FILE); \
		echo "✅ Fork context set to: $(REPO)"; \
		echo ""; \
		echo "Now you can use fork commands without REPO=:"; \
		echo "  make fork-save"; \
		echo "  make fork-branch"; \
		echo "  make fork-guide"; \
	elif [ -f $(FORK_FILE) ] && [ "$$(cat $(FORK_FILE) 2>/dev/null)" != "" ]; then \
		echo "📍 Current fork context:"; \
		cat $(FORK_FILE); \
		echo ""; \
		echo "To change context: make fork-context"; \
	else \
		selected=$$($(call prompt-for-repo)); \
		echo "current_repo=$$selected" > $(FORK_FILE); \
		echo ""; \
		echo "✅ Fork context set to: $$selected"; \
		echo ""; \
		echo "Now you can use fork commands without any flags:"; \
		echo "  make fork-save"; \
		echo "  make fork-branch"; \
		echo "  make fork-guide"; \
	fi

## fork-branch: Create/switch branch (prompts if BRANCH not provided)
fork-branch:
	@$(call check-repos-file)
	@$(call get-target-repo,fork-branch); \
	found=0; \
	while IFS='|' read -r name dir url ref fork_url desc; do \
		[ "$$name" = "name" ] && continue; \
		if [ "$$name" != "$$target_repo" ]; then continue; fi; \
		found=1; \
		if [ -z "$$fork_url" ]; then \
			echo "❌ Error: $$target_repo does not have a fork configured"; \
			exit 1; \
		fi; \
		if [ ! -d "$$dir/.git" ]; then \
			echo "❌ Error: $$dir not cloned yet"; \
			echo "Run: make install"; \
			exit 1; \
		fi; \
		branch_name="$(BRANCH)"; \
		if [ -z "$$branch_name" ]; then \
			branch_name=$$($(call prompt-for-branch-name)); \
			echo ""; \
		fi; \
		current_branch=$$(cd $$dir && git branch --show-current); \
		if (cd $$dir && git show-ref --verify --quiet refs/heads/$$branch_name); then \
			echo "🔄 Switching to existing branch: $$branch_name"; \
			cd $$dir && git checkout $$branch_name; \
			echo ""; \
			echo "✅ Now on branch: $$branch_name"; \
		else \
			echo "🌿 Creating new branch: $$branch_name"; \
			echo ""; \
			base_branch=$$($(call get-base-branch,$$dir)); \
			echo "📍 Current branch: $$current_branch"; \
			echo "📍 Base branch: $$base_branch"; \
			echo ""; \
			if [ "$$current_branch" != "$$base_branch" ]; then \
				if (cd $$dir && git diff --quiet && git diff --cached --quiet); then \
					echo "🔄 Switching to $$base_branch..."; \
					(cd $$dir && git checkout $$base_branch --quiet); \
				else \
					echo "❌ Error: You have uncommitted changes on $$current_branch"; \
					echo ""; \
					echo "Options:"; \
					echo "  1. Commit them: make fork-commit"; \
					echo "  2. Stash them: cd $$dir && git stash"; \
					exit 1; \
				fi; \
			fi; \
			echo "🔄 Updating $$base_branch from origin..."; \
			(cd $$dir && git pull origin $$base_branch --quiet); \
			echo "✅ Creating branch: $$branch_name"; \
			(cd $$dir && git checkout -b $$branch_name --quiet); \
			echo ""; \
			echo "✅ Success! Branch $$branch_name created from $$base_branch"; \
		fi; \
		echo ""; \
		echo "Next steps:"; \
		echo "  1. Make your changes"; \
		echo "  2. Save: make fork-save"; \
		break; \
	done < $(REPOS_FILE); \
	if [ $$found -eq 0 ]; then \
		echo "❌ Error: Repository $$target_repo not found in repos.list"; \
		exit 1; \
	fi

## fork-commit: Commit changes (prompts for MSG if not provided)
fork-commit:
	@$(call check-repos-file)
	@$(call get-target-repo,fork-commit); \
	msg="$(MSG)"; \
	if [ -z "$$msg" ]; then \
		msg=$$($(call prompt-for-message)); \
		echo ""; \
	fi; \
	echo "💾 Committing Changes"; \
	echo ""; \
	found=0; \
	while IFS='|' read -r name dir url ref fork_url desc; do \
		[ "$$name" = "name" ] && continue; \
		if [ "$$name" != "$$target_repo" ]; then continue; fi; \
		found=1; \
		if [ ! -d "$$dir/.git" ]; then \
			echo "❌ Error: $$dir not cloned yet"; \
			exit 1; \
		fi; \
		current_branch=$$(cd $$dir && git branch --show-current); \
		base_branch=$$($(call get-base-branch,$$dir)); \
		if [ "$$current_branch" = "$$base_branch" ]; then \
			echo "⚠️  Warning: You're on $$base_branch (main branch)"; \
			echo ""; \
			echo "What do you want to do?"; \
			echo ""; \
			echo "  1) Create new branch and commit changes"; \
			echo "  2) Just create a branch (stash changes for now)"; \
			echo "  3) Cancel"; \
			echo ""; \
			printf "Select: "; \
			read choice </dev/tty; \
			choice=$$(echo "$$choice" | tr -d ' \t\n\r'); \
			echo ""; \
			case $$choice in \
				1) \
					branch_name=$$($(call prompt-for-branch-name)); \
					cd $$dir && git checkout -b $$branch_name; \
					echo "✅ Created branch: $$branch_name"; \
					echo ""; \
					;; \
				2) \
					branch_name=$$($(call prompt-for-branch-name)); \
					cd $$dir && git stash; \
					cd $$dir && git checkout -b $$branch_name; \
					echo "✅ Created branch: $$branch_name (changes stashed)"; \
					echo ""; \
					exit 0; \
					;; \
				*) \
					echo "Cancelled"; \
					exit 1; \
					;; \
			esac; \
		fi; \
		echo "📍 Branch: $$current_branch"; \
		echo ""; \
		if [ -z "$$(cd $$dir && git status --porcelain)" ]; then \
			echo "⚠️  Warning: No changes to commit"; \
			exit 0; \
		fi; \
		echo "📝 Changed files:"; \
		(cd $$dir && git status --short); \
		echo ""; \
		echo "🔄 Staging all changes..."; \
		(cd $$dir && git add -A); \
		echo "✅ Committing with message: $$msg"; \
		(cd $$dir && git commit -m "$$msg" --quiet); \
		echo ""; \
		echo "✅ Success! Changes committed to $$current_branch"; \
		echo ""; \
		echo "Next: Push with 'make fork-push'"; \
		break; \
	done < $(REPOS_FILE); \
	if [ $$found -eq 0 ]; then \
		echo "❌ Error: Repository $$target_repo not found"; \
		exit 1; \
	fi

## fork-save: Save and push in one command (prompts for MSG if not provided)
fork-save:
	@$(call check-repos-file)
	@$(call get-target-repo,fork-save); \
	msg="$(MSG)"; \
	if [ -z "$$msg" ]; then \
		msg=$$($(call prompt-for-message)); \
		echo ""; \
	fi; \
	echo "🚀 Quick Save & Push"; \
	echo ""; \
	found=0; \
	while IFS='|' read -r name dir url ref fork_url desc; do \
		[ "$$name" = "name" ] && continue; \
		if [ "$$name" != "$$target_repo" ]; then continue; fi; \
		found=1; \
		if [ -z "$$fork_url" ]; then \
			echo "❌ Error: $$target_repo does not have a fork configured"; \
			exit 1; \
		fi; \
		if [ ! -d "$$dir/.git" ]; then \
			echo "❌ Error: $$dir not cloned yet"; \
			exit 1; \
		fi; \
		current_branch=$$(cd $$dir && git branch --show-current); \
		base_branch=$$($(call get-base-branch,$$dir)); \
		if [ "$$current_branch" = "$$base_branch" ]; then \
			echo "⚠️  Warning: You're on $$base_branch (main branch)"; \
			echo ""; \
			echo "What do you want to do?"; \
			echo ""; \
			echo "  1) Create new branch and save changes"; \
			echo "  2) Just create a branch (stash changes for now)"; \
			echo "  3) Cancel"; \
			echo ""; \
			printf "Select: "; \
			read choice </dev/tty; \
			choice=$$(echo "$$choice" | tr -d ' \t\n\r'); \
			echo ""; \
			case $$choice in \
				1) \
					branch_name=$$($(call prompt-for-branch-name)); \
					cd $$dir && git checkout -b $$branch_name; \
					echo "✅ Created branch: $$branch_name"; \
					echo ""; \
					;; \
				2) \
					branch_name=$$($(call prompt-for-branch-name)); \
					cd $$dir && git stash; \
					cd $$dir && git checkout -b $$branch_name; \
					echo "✅ Created branch: $$branch_name (changes stashed)"; \
					echo ""; \
					exit 0; \
					;; \
				*) \
					echo "Cancelled"; \
					exit 1; \
					;; \
			esac; \
		fi; \
		echo "📍 Branch: $$current_branch"; \
		echo ""; \
		if [ -z "$$(cd $$dir && git status --porcelain)" ]; then \
			echo "⚠️  Warning: No changes to save"; \
			exit 0; \
		fi; \
		echo "📝 Changed files:"; \
		(cd $$dir && git status --short); \
		echo ""; \
		echo "🔄 Step 1/3: Staging changes..."; \
		(cd $$dir && git add -A); \
		echo "✅ Staged"; \
		echo ""; \
		echo "🔄 Step 2/3: Committing..."; \
		(cd $$dir && git commit -m "$$msg" --quiet); \
		echo "✅ Committed: $$msg"; \
		echo ""; \
		echo "🔄 Step 3/3: Pushing to fork..."; \
		if ! (cd $$dir && git remote get-url fork >/dev/null 2>&1); then \
			echo "⚠️  Fork remote not configured, adding it..."; \
			(cd $$dir && git remote add fork $$fork_url); \
		fi; \
		(cd $$dir && git push fork $$current_branch --quiet 2>&1); \
		echo "✅ Pushed to fork/$$current_branch"; \
		echo ""; \
		fork_owner=$$(echo "$$fork_url" | sed 's|https://github.com/\([^/]*\)/.*|\1|'); \
		fork_repo=$$(echo "$$fork_url" | sed 's|https://github.com/[^/]*/\([^.]*\).*|\1|'); \
		origin_owner=$$(echo "$$url" | sed 's|https://github.com/\([^/]*\)/.*|\1|'); \
		origin_repo=$$(echo "$$url" | sed 's|https://github.com/[^/]*/\([^.]*\).*|\1|'); \
		pr_url="https://github.com/$$origin_owner/$$origin_repo/compare/main...$$fork_owner:$$fork_repo:$$current_branch?expand=1"; \
		echo "✅ Success! All changes saved and pushed"; \
		echo ""; \
		echo "🔗 Create PR: $$pr_url"; \
		break; \
	done < $(REPOS_FILE); \
	if [ $$found -eq 0 ]; then \
		echo "❌ Error: Repository $$target_repo not found in repos.list"; \
		exit 1; \
	fi

## fork-switch: Switch to a different branch safely (prompts if BRANCH not provided)
fork-switch:
	@$(call check-repos-file)
	@$(call get-target-repo,fork-switch); \
	echo "🔄 Switching Branch"; \
	echo ""; \
	found=0; \
	while IFS='|' read -r name dir url ref fork_url desc; do \
		[ "$$name" = "name" ] && continue; \
		if [ "$$name" != "$$target_repo" ]; then continue; fi; \
		found=1; \
		if [ ! -d "$$dir/.git" ]; then \
			echo "❌ Error: $$dir not cloned yet"; \
			exit 1; \
		fi; \
		target_branch="$(BRANCH)"; \
		if [ -z "$$target_branch" ]; then \
			target_branch=$$($(call prompt-for-branch,$$dir)); \
			echo ""; \
		fi; \
		current_branch=$$(cd $$dir && git branch --show-current); \
		echo "📍 Current branch: $$current_branch"; \
		echo "📍 Target branch: $$target_branch"; \
		echo ""; \
		if [ "$$current_branch" = "$$target_branch" ]; then \
			echo "✅ Already on $$target_branch"; \
			exit 0; \
		fi; \
		if ! (cd $$dir && git diff --quiet && git diff --cached --quiet); then \
			echo "⚠️  Warning: You have uncommitted changes"; \
			echo ""; \
			(cd $$dir && git status --short); \
			echo ""; \
			echo "Options:"; \
			echo "  1. Save them first: make fork-save MSG=\"WIP\""; \
			echo "  2. Discard them: cd $$dir && git reset --hard"; \
			echo "  3. Stash them: cd $$dir && git stash"; \
			exit 1; \
		fi; \
		if ! (cd $$dir && git show-ref --verify --quiet refs/heads/$$target_branch); then \
			echo "❌ Error: Branch $$target_branch does not exist"; \
			echo ""; \
			echo "Available local branches:"; \
			(cd $$dir && git branch); \
			echo ""; \
			echo "Create it with: make fork-branch BRANCH=$$target_branch"; \
			exit 1; \
		fi; \
		echo "🔄 Switching to $$target_branch..."; \
		(cd $$dir && git checkout $$target_branch --quiet); \
		echo "✅ Switched to $$target_branch"; \
		break; \
	done < $(REPOS_FILE); \
	if [ $$found -eq 0 ]; then \
		echo "❌ Error: Repository $$target_repo not found in repos.list"; \
		exit 1; \
	fi

## fork-list: List all branches (local and remote)
fork-list:
	@$(call check-repos-file)
	@$(call get-target-repo,fork-list); \
	echo "📋 Branch List"; \
	echo ""; \
	found=0; \
	while IFS='|' read -r name dir url ref fork_url desc; do \
		[ "$$name" = "name" ] && continue; \
		if [ "$$name" != "$$target_repo" ]; then continue; fi; \
		found=1; \
		if [ ! -d "$$dir/.git" ]; then \
			echo "❌ Error: $$dir not cloned yet"; \
			exit 1; \
		fi; \
		current_branch=$$(cd $$dir && git branch --show-current); \
		echo "📍 Current branch: $$current_branch"; \
		echo ""; \
		echo "🌿 Local branches:"; \
		(cd $$dir && git branch -v); \
		echo ""; \
		if [ -n "$$fork_url" ]; then \
			if (cd $$dir && git remote get-url fork >/dev/null 2>&1); then \
				echo "🍴 Fork branches:"; \
				(cd $$dir && git fetch fork --quiet 2>/dev/null || true); \
				(cd $$dir && git branch -r | grep fork/ || echo "  No fork branches yet"); \
				echo ""; \
			fi; \
		fi; \
		echo "🌍 Origin branches:"; \
		(cd $$dir && git fetch origin --quiet 2>/dev/null || true); \
		(cd $$dir && git branch -r | grep origin/ | head -10); \
		break; \
	done < $(REPOS_FILE); \
	if [ $$found -eq 0 ]; then \
		echo "❌ Error: Repository $$target_repo not found in repos.list"; \
		exit 1; \
	fi

## fork-guide: Show current state and suggest next steps
fork-guide:
	@$(call check-repos-file)
	@$(call get-target-repo,fork-guide); \
	echo "🧭 Fork Workflow Guide"; \
	echo ""; \
	found=0; \
	while IFS='|' read -r name dir url ref fork_url desc; do \
		[ "$$name" = "name" ] && continue; \
		if [ "$$name" != "$$target_repo" ]; then continue; fi; \
		found=1; \
		if [ -z "$$fork_url" ]; then \
			echo "❌ No fork configured for $$target_repo"; \
			echo ""; \
			echo "Setup steps:"; \
			echo "  1. Fork the repository on GitHub"; \
			echo "  2. Add fork URL to repos.list (5th column)"; \
			echo "  3. Run: make fork-add"; \
			exit 0; \
		fi; \
		if [ ! -d "$$dir/.git" ]; then \
			echo "❌ Repository not cloned yet"; \
			echo ""; \
			echo "Next step:"; \
			echo "  make install"; \
			exit 0; \
		fi; \
		current_branch=$$(cd $$dir && git branch --show-current); \
		base_branch=$$($(call get-base-branch,$$dir)); \
		has_changes=0; \
		if [ -n "$$(cd $$dir && git status --porcelain)" ]; then \
			has_changes=1; \
		fi; \
		is_ahead=0; \
		is_behind=0; \
		if (cd $$dir && git remote get-url fork >/dev/null 2>&1); then \
			(cd $$dir && git fetch fork --quiet 2>/dev/null); \
			if (cd $$dir && git rev-parse fork/$$current_branch >/dev/null 2>&1); then \
				ahead=$$(cd $$dir && git rev-list --count fork/$$current_branch..$$current_branch 2>/dev/null || echo "0"); \
				behind=$$(cd $$dir && git rev-list --count $$current_branch..fork/$$current_branch 2>/dev/null || echo "0"); \
				[ "$$ahead" -gt 0 ] && is_ahead=1; \
				[ "$$behind" -gt 0 ] && is_behind=1; \
			fi; \
		fi; \
		echo "📍 Current State:"; \
		echo "  Repository: $$dir"; \
		echo "  Branch: $$current_branch"; \
		if [ "$$current_branch" = "$$base_branch" ]; then \
			echo "  Location: Main branch"; \
		else \
			echo "  Location: Feature branch"; \
		fi; \
		if [ $$has_changes -eq 1 ]; then \
			echo "  Changes: Yes (uncommitted)"; \
			changed_count=$$(cd $$dir && git status --porcelain | wc -l | tr -d ' '); \
			echo "  Files changed: $$changed_count"; \
		else \
			echo "  Changes: No"; \
		fi; \
		if [ $$is_ahead -eq 1 ]; then \
			echo "  Fork status: Ahead by $$ahead commits (need to push)"; \
		elif [ $$is_behind -eq 1 ]; then \
			echo "  Fork status: Behind by $$behind commits"; \
		else \
			if (cd $$dir && git rev-parse fork/$$current_branch >/dev/null 2>&1); then \
				echo "  Fork status: In sync"; \
			else \
				echo "  Fork status: Branch not on fork yet"; \
			fi; \
		fi; \
		echo ""; \
		echo "🎯 What To Do Next:"; \
		echo ""; \
		if [ "$$current_branch" = "$$base_branch" ]; then \
			if [ $$has_changes -eq 1 ]; then \
				echo "⚠️  You're on main with uncommitted changes!"; \
				echo ""; \
				echo "Recommended: Create a feature branch first"; \
				echo "  make fork-branch BRANCH=your-feature-name"; \
				echo "  (This will ask you to commit or stash changes first)"; \
			else \
				echo "✅ Ready to start new work"; \
				echo ""; \
				echo "Next step: Create a feature branch"; \
				echo "  make fork-branch BRANCH=your-feature-name"; \
				echo ""; \
				echo "Example branch names:"; \
				echo "  • add-nats-state"; \
				echo "  • fix-session-bug"; \
				echo "  • improve-docs"; \
			fi; \
		else \
			if [ $$has_changes -eq 1 ]; then \
				echo "✅ You're on feature branch '$$current_branch' with changes"; \
				echo ""; \
				echo "Choose one:"; \
				echo ""; \
				echo "Option 1 (Recommended): Save everything in one command"; \
				echo "  make fork-save MSG=\"describe your changes\""; \
				echo "  → Stages + commits + pushes + gives you PR link"; \
				echo ""; \
				echo "Option 2: Commit only (don't push yet)"; \
				echo "  make fork-commit MSG=\"describe your changes\""; \
				echo "  → Later: make fork-push"; \
			elif [ $$is_ahead -eq 1 ]; then \
				echo "✅ You have $$ahead commit(s) ready to push"; \
				echo ""; \
				echo "Next step: Push to your fork"; \
				echo "  make fork-push"; \
				echo "  → Creates PR link for you"; \
			elif (cd $$dir && git rev-parse fork/$$current_branch >/dev/null 2>&1); then \
				echo "✅ Everything is saved and pushed"; \
				echo ""; \
				echo "Your options:"; \
				echo ""; \
				echo "1. Continue working on this branch"; \
				echo "   • Edit files"; \
				echo "   • make fork-save MSG=\"more changes\""; \
				echo ""; \
				echo "2. Switch to another branch"; \
				echo "   • make fork-list  (see all branches)"; \
				echo "   • make fork-switch BRANCH=other-branch"; \
				echo ""; \
				echo "3. Go back to main"; \
				echo "   • make fork-switch BRANCH=$$base_branch"; \
			else \
				echo "✅ Clean working tree, branch not on fork yet"; \
				echo ""; \
				echo "No changes to push. Either:"; \
				echo ""; \
				echo "1. Make some changes and commit:"; \
				echo "   • Edit files"; \
				echo "   • make fork-save MSG=\"describe changes\""; \
				echo ""; \
				echo "2. Switch back to main:"; \
				echo "   • make fork-switch BRANCH=$$base_branch"; \
			fi; \
		fi; \
		echo ""; \
		echo "📚 Quick Reference:"; \
		echo "  make fork-guide      ← Show this guide"; \
		echo "  make fork-list       ← List all branches"; \
		echo "  make fork-status     ← Check sync status"; \
		break; \
	done < $(REPOS_FILE); \
	if [ $$found -eq 0 ]; then \
		echo "❌ Error: Repository $$target_repo not found in repos.list"; \
		exit 1; \
	fi


## fork: Interactive workflow - handles everything with prompts
fork:
	@$(call check-repos-file)
	@$(call get-target-repo,fork); \
	found=0; \
	while IFS='|' read -r name dir url ref fork_url desc; do \
		[ "$$name" = "name" ] && continue; \
		if [ "$$name" != "$$target_repo" ]; then continue; fi; \
		found=1; \
		if [ -z "$$fork_url" ]; then \
			echo "❌ Error: $$target_repo does not have a fork configured"; \
			exit 1; \
		fi; \
		if [ ! -d "$$dir/.git" ]; then \
			echo "❌ Error: $$dir not cloned yet"; \
			echo "Run: make install"; \
			exit 1; \
		fi; \
		current_branch=$$(cd $$dir && git branch --show-current); \
		if [ -z "$$current_branch" ]; then \
			echo "❌ Error: Detached HEAD state detected"; \
			echo ""; \
			echo "You're not on any branch. Fix with:"; \
			echo "  cd $$dir && git checkout main"; \
			exit 1; \
		fi; \
		base_branch=$$($(call get-base-branch,$$dir)); \
		has_changes=0; \
		[ -n "$$(cd $$dir && git status --porcelain)" ] && has_changes=1; \
		echo "🎯 Fork Workflow"; \
		echo ""; \
		echo "📍 Current State:"; \
		printf "  Repository: \033[1;32m%s\033[0m\n" "$$target_repo"; \
		printf "  Branch: \033[1;33m%s\033[0m\n" "$$current_branch"; \
		if [ $$has_changes -eq 1 ]; then \
			changed_count=$$(cd $$dir && git status --porcelain | wc -l | tr -d ' '); \
			printf "  Changes: \033[1;31m%s file(s)\033[0m\n" "$$changed_count"; \
		else \
			echo "  Changes: None"; \
		fi; \
		echo ""; \
		echo "What do you want to do?"; \
		echo ""; \
		if [ "$$current_branch" = "$$base_branch" ]; then \
			if [ $$has_changes -eq 1 ]; then \
				echo "  1) Create branch and save changes"; \
				echo "  2) Discard changes and start fresh"; \
				echo "  3) Just create a branch (stash changes)"; \
				echo "  4) Cancel"; \
			else \
				echo "  1) Start new work (create branch)"; \
				echo "  2) Switch to existing branch"; \
				echo "  3) View all branches"; \
				echo "  4) Cancel"; \
			fi; \
		else \
			if [ $$has_changes -eq 1 ]; then \
				echo "  1) Save changes (commit + push)"; \
				echo "  2) Just commit (don't push yet)"; \
				echo "  3) Discard changes"; \
				echo "  4) Switch to different branch"; \
				echo "  5) Back to main"; \
				echo "  6) Cancel"; \
			else \
				echo "  1) Continue working (stay here)"; \
				echo "  2) Switch to different branch"; \
				echo "  3) Back to main"; \
				echo "  4) View all branches"; \
				echo "  5) Cancel"; \
			fi; \
		fi; \
		echo ""; \
		printf "Select: "; \
		read choice </dev/tty; \
		choice=$$(echo "$$choice" | tr -d ' \t\n\r'); \
		echo ""; \
		if [ "$$current_branch" = "$$base_branch" ]; then \
			if [ $$has_changes -eq 1 ]; then \
				case $$choice in \
					1) \
						branch_name=$$($(call prompt-for-branch-name)); \
						cd $$dir && git checkout -b $$branch_name; \
						msg=$$($(call prompt-for-message)); \
						cd $$dir && git add -A && git commit -m "$$msg"; \
						cd $$dir && git push fork $$branch_name; \
						echo ""; \
						echo "✅ Branch created, changes saved and pushed!"; \
						;; \
					2) \
						cd $$dir && git reset --hard; \
						cd $$dir && git clean -fd; \
						echo "✅ Changes discarded"; \
						;; \
					3) \
						branch_name=$$($(call prompt-for-branch-name)); \
						cd $$dir && git stash; \
						cd $$dir && git checkout -b $$branch_name; \
						echo "✅ Branch created, changes stashed"; \
						;; \
					*) \
						echo "Cancelled"; \
						;; \
				esac; \
			else \
				case $$choice in \
					1) \
						branch_name=$$($(call prompt-for-branch-name)); \
						if (cd $$dir && git show-ref --verify --quiet refs/heads/$$branch_name); then \
							echo "⚠️  Branch $$branch_name already exists, switching to it"; \
							cd $$dir && git checkout $$branch_name; \
						else \
							cd $$dir && git checkout -b $$branch_name; \
						fi; \
						echo "✅ Ready to work on $$branch_name"; \
						;; \
					2) \
						branch_name=$$($(call prompt-for-branch,$$dir)); \
						cd $$dir && git checkout $$branch_name; \
						echo "✅ Switched to $$branch_name"; \
						;; \
					3) \
						$(MAKE) fork-list REPO=$$target_repo; \
						;; \
					*) \
						echo "Cancelled"; \
						;; \
				esac; \
			fi; \
		else \
			if [ $$has_changes -eq 1 ]; then \
				case $$choice in \
					1) \
						msg=$$($(call prompt-for-message)); \
						cd $$dir && git add -A && git commit -m "$$msg"; \
						cd $$dir && git push fork $$current_branch; \
						echo ""; \
						echo "✅ Changes saved and pushed!"; \
						fork_owner=$$(echo "$$fork_url" | sed 's|https://github.com/\([^/]*\)/.*|\1|'); \
						fork_repo=$$(echo "$$fork_url" | sed 's|https://github.com/[^/]*/\([^.]*\).*|\1|'); \
						origin_owner=$$(echo "$$url" | sed 's|https://github.com/\([^/]*\)/.*|\1|'); \
						origin_repo=$$(echo "$$url" | sed 's|https://github.com/[^/]*/\([^.]*\).*|\1|'); \
						echo ""; \
						echo "🔗 Create PR: https://github.com/$$origin_owner/$$origin_repo/compare/main...$$fork_owner:$$fork_repo:$$current_branch?expand=1"; \
						;; \
					2) \
						msg=$$($(call prompt-for-message)); \
						cd $$dir && git add -A && git commit -m "$$msg"; \
						echo "✅ Changes committed (not pushed)"; \
						;; \
					3) \
						cd $$dir && git reset --hard; \
						cd $$dir && git clean -fd; \
						echo "✅ Changes discarded"; \
						;; \
					4) \
						branch_name=$$($(call prompt-for-branch,$$dir)); \
						cd $$dir && git checkout $$branch_name; \
						echo "✅ Switched to $$branch_name"; \
						;; \
					5) \
						cd $$dir && git checkout $$base_branch; \
						echo "✅ Back on $$base_branch"; \
						;; \
					*) \
						echo "Cancelled"; \
						;; \
				esac; \
			else \
				case $$choice in \
					1) \
						echo "✅ Ready! Make your changes, then run 'make fork' again"; \
						;; \
					2) \
						branch_name=$$($(call prompt-for-branch,$$dir)); \
						cd $$dir && git checkout $$branch_name; \
						echo "✅ Switched to $$branch_name"; \
						;; \
					3) \
						cd $$dir && git checkout $$base_branch; \
						echo "✅ Back on $$base_branch"; \
						;; \
					4) \
						$(MAKE) fork-list REPO=$$target_repo; \
						;; \
					*) \
						echo "Cancelled"; \
						;; \
				esac; \
			fi; \
		fi; \
		break; \
	done < $(REPOS_FILE); \
	if [ $$found -eq 0 ]; then \
		echo "❌ Error: Repository $$target_repo not found"; \
		exit 1; \
	fi
