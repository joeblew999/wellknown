# .src/ - Reference Code & Fork Workflow

External repositories for reference + Makefile-based fork workflow.

## Quick Start

```bash
# 1. Clone reference repos
make install

# 2. Set fork context (only once!)
make fork-context REPO=via

# 3. Use fork workflow (no flags needed!)
make fork                # Interactive menu
make fork-guide          # What should I do next?
make fork-save MSG="..."  # Save and push changes
```

## Fork Workflow Commands

**🎯 Core Workflow (3 commands you'll use daily):**
- `make fork-start` - Start new work (creates branch)
- `make fork-continue` - Save progress (commit + push)
- `make fork-finish` - Done with work (back to main)

**🧭 Navigation (when you need it):**
- `make fork-switch` - Switch branches
- `make fork-sync` - Get latest from upstream
- `make fork` - Interactive menu (explore all 16 states)

**📚 Info (read-only helpers):**
- `make fork-guide` - Show current state + next steps
- `make fork-list` - List all branches

**Cognitive Load: 3 commands** (start, continue, finish)
Learn the rest progressively as needed.

All commands work without parameters (prompt when needed). Override with: `REPO=name`, `BRANCH=name`, `MSG="message"`.

## Example Workflows

### Standard Feature Development (3 commands)
```bash
make fork-start              # Creates branch (prompts for name)
# Make your changes...
make fork-continue           # Saves progress (prompts for msg)
# Repeat as needed...
make fork-finish             # Back to main, cleanup
```

### Quick Fix (2 commands)
```bash
make fork-start              # Creates branch
# Edit files...
make fork-continue           # Saves & pushes (gives PR link)
make fork-finish             # Back to main
```

### Working on Multiple Branches
```bash
make fork-start              # Start feature A
# Work on A...
make fork-continue           # Save A
make fork-switch             # Switch to feature B
# Work on B...
make fork-continue           # Save B
make fork-finish             # Done with B, back to main
```

## Reference Repos

```bash
make install         # Clone all repos
make upgrade         # Update all repos
make status          # Show status of all repos
```

## Testing

```bash
make test            # Run all tests (30 seconds)

# Individual test suites:
./test-scenarios.sh          # State machine tests (7 scenarios)
./test-fork-baseline.exp     # Regression tests
./test-prompts.exp           # Prompt functionality

# Development testing:
make fork-branch DRY_RUN=1   # See prompts without executing
```

**Test coverage:**
- 7 scenario tests (state-dependent logic + git operations)
- 5 baseline tests (regression prevention)
- 2 prompt tests (interactive functionality)
- DRY_RUN mode for rapid UX validation

## Files

**Core:**
- `Makefile` - Main entry point
- `repos.mk`, `forks.mk`, `index.mk` - Modular command sets
- `repos.list` - Repository metadata
- `fork.list` - Current fork context
- `index.list` - Trigger mappings for Claude agent discovery

**Auto-generated:**
- `.gitignore`, `INDEX.md`

**Tests:**
- `test-scenarios.sh` - Comprehensive scenario testing
- `test-fork-baseline.exp` - Regression tests
- `test-prompts.exp` - Prompt validation

## repos.list Format

```
name|dir|url|ref|fork_url|description
via|via|https://github.com/go-via/via.git|main|https://github.com/joeblew999/via.git|Via framework
```

## What Gets Committed

Meta files only. Cloned repos ignored via `*/` pattern in `.gitignore`.

## Fork Workflow Features

- ✅ Interactive menus with numbered options
- ✅ Shows which command is running (learn as you use)
- ✅ Always displays repo context
- ✅ Prevents commits to main branch
- ✅ Detects uncommitted changes
- ✅ Creates branches automatically when switching
- ✅ Generates PR links after push
- ✅ No manual git commands needed

## How Testing Works

**Three-tier approach:**

1. **DRY_RUN** - Test prompts instantly without git operations
2. **Scenarios** - Test state machine (main/feature × clean/dirty)
3. **Expect** - Test user journeys end-to-end

Run `make test` to execute all tiers (~30 seconds).

## Common Questions

**Q: Which command should I use most?**
A: `make fork` (interactive) or `make fork-save MSG="..."` (direct)

**Q: How do I know what to do next?**
A: Run `make fork-guide` - it tells you exactly what to do

**Q: Can I work on multiple repos?**
A: Yes! Use `make fork-context REPO=name` to switch between repos

**Q: What if I'm on the wrong branch?**
A: `make fork-switch` or `make fork-list` to see/switch branches

**Q: Do I need to learn git?**
A: No! The Makefile handles all git operations

## Safety Features

- ❌ Cannot commit to main (shows menu with options)
- ⚠️  Warns about uncommitted changes before switching
- ✅ Detects detached HEAD state
- 📍 Always shows current repo + branch
- 🔍 Validates all inputs

See `make help` for all available commands.
