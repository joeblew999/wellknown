# .src/ - External Reference Code

This directory contains external source code repositories for reference and learning purposes. **These are NOT part of our codebase** and are git-ignored.

## Purpose

- Study implementation patterns from other projects
- Extract useful techniques without direct dependencies
- Keep reference code locally accessible for offline work
- Avoid polluting our codebase with external code
- Enable AI assistants (Claude) to access agent files and documentation from reference repositories

## Usage

```bash
make          # Show available commands (default)
make doctor   # Check repos.list for errors
make install  # Clone/update all repos (parallel, ~2s) + auto-generate INDEX.md
make status   # Show repo status with versions, age, and descriptions
make upgrade  # Update repos.list to latest tags and install repositories
make index    # Manually regenerate INDEX.md (auto-runs after install/upgrade)
```

## Data Files

### repos.list - Repository Metadata

Pipe-delimited format:
```
name|dir|url|ref|description
pocketbase|pocketbase|https://github.com/pocketbase/pocketbase.git|v0.32.0|PocketBase (official)
```

Fields:
- **name**: Repository identifier (not used, for documentation)
- **dir**: Local directory name
- **url**: Git clone URL
- **ref**: Branch/tag to checkout (e.g., `main`, `v1.0.0`)
- **description**: Human-readable description

**Note**: `.gitignore` is auto-generated from this file via `make gitignore` (runs automatically during `make install` and `make upgrade`)

### triggers.list - Claude Trigger Config

Maps keywords → agent files. Format: `category|priority|triggers|action|path|notes|why`

Example:
```
Commits|1|commit,commits|read|.src/datastarui/.claude/commands/commit.md||No attribution
```

See [TRIGGERS.md](TRIGGERS.md) for details.

## Agent Files

Many reference repositories contain AI assistant configurations:

- **CLAUDE.md**: Project-specific instructions for Claude
- **.claude/**: Slash commands, hooks, and agent configurations
- **.claud/**: Alternative naming convention (some projects)

**INDEX.md is auto-generated** after `make install` and `make upgrade` to provide an always-current catalog of:
- All repositories with current versions
- CLAUDE.md files with project-specific instructions
- Available slash commands (`.claude/commands/`)
- Available agents (`.claude/agents/`)

Claude reads INDEX.md to discover patterns, workflows, and integration examples across all reference repositories.