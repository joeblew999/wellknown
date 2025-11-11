# .src/ - External Reference Code

External repositories for reference patterns and AI agent workflows. **Git-ignored, not part of codebase.**

## Quick Start

```bash
make install  # Clone/update all repos + generate INDEX.md
make status   # Show repo versions and status
make upgrade  # Update to latest tags
```

## Files

- **repos.list** - Repository metadata (what to clone)
- **triggers.list** - Claude trigger mappings (keywords → files)
- **INDEX.md** - Auto-generated catalog (read by Claude)

## How Claude Uses This

Two discovery methods:

### 1. Triggers (Curated)
Keywords → specific files. Edit `triggers.list`:
```
category|priority|triggers|action|path|notes|why
Commits|1|commit,commits|read|.src/datastarui/.claude/commands/commit.md||Workflow description
```

When user says "commit", Claude reads that file immediately.

### 2. Auto-Discovery (Complete Catalog)
Scans all repos for `.claude/` directories and CLAUDE.md files. Auto-generates catalog in INDEX.md.

## repos.list Format

```
name|dir|url|ref|description
pocketbase|pocketbase|https://github.com/pocketbase/pocketbase.git|v0.32.0|PocketBase (official)
```

- **name** - Identifier (for documentation)
- **dir** - Local directory name
- **url** - Git clone URL
- **ref** - Branch/tag (e.g., `main`, `v1.0.0`)
- **description** - Human-readable description

## triggers.list Format

```
category|priority|triggers|action|path|notes|why
```

- **category** - Group name
- **priority** - Order 1-9 (displays as 1️⃣-9️⃣)
- **triggers** - Comma-separated keywords
- **action** - `read` (must read) or `browse` (explore)
- **path** - File/directory path
- **notes** - Optional context
- **why** - Value explanation

**Actions:**
- `read` - Claude must read before responding
- `browse` - Claude explores for patterns

## Workflow

```bash
# Add new repo
echo "name|newrepo|https://github.com/user/repo.git|main|Description" >> repos.list
make install

# Add trigger
echo "Category|5|keyword|read|.src/repo/file.md||Why" >> triggers.list
make index

# Update everything
make upgrade
```

## What Gets Committed

Only meta files:
- `.gitignore` (auto-generated)
- `README.md`
- `Makefile`
- `repos.list`
- `triggers.list`
- `INDEX.md` (auto-generated)

Cloned repos are git-ignored via pattern `*/` in `.gitignore`.
