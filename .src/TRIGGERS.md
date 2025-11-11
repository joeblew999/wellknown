# triggers.list Configuration

Configures how Claude automatically discovers agent files. Pipe-delimited CSV:

```
category|priority|triggers|action|path|notes|why
```

## Columns

- **category** - Group name (e.g., `Git Operations`)
- **priority** - Order 1-9 (maps to 1️⃣-9️⃣)
- **triggers** - Comma-separated keywords (e.g., `commit,git commit`)
- **action** - `read` (must read) or `browse` (explore)
- **path** - Relative path to file/directory
- **notes** - Optional context shown in parentheses
- **why** - Value explanation (shown once per category)

## Example

```
Git Operations|1|commit,git commit|read|.src/datastarui/.claude/commands/commit.md||No Claude attribution
Code Review & PRs|2|review,PR|read|.src/datastarui/.claude/commands/local_review.md|for worktree|Review workflow
```

Generates:

```markdown
### 1️⃣ Git Operations
**Triggers**: `commit`, `git commit`

→ **Read immediately**: `.src/datastarui/.claude/commands/commit.md`

**Why**: No Claude attribution
```

## Quick Tips

1. **Specific triggers** - `commit,git commit` not `git,code`
2. **Add notes** - Context helps Claude know what to focus on
3. **Group by category** - Multiple rows same category = merged triggers
4. **Deduplicated** - Triggers auto-sorted and deduplicated per category
5. **Test changes** - Run `make index` to see result

## Actions

- `read` - Claude must read before responding
- `browse` - Claude explores for patterns

## Workflow

```bash
# Edit triggers
vim triggers.list

# Regenerate index
make index

# Verify
head -100 INDEX.md
```
