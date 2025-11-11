# Agent Discovery System

Two complementary approaches for Claude to find and use agent files:

## 1. Trigger-Based (Curated)

**File**: `triggers.list`
**Output**: Top section of INDEX.md (🎯 When to Use These Files)

**How it works:**
- You manually map keywords → specific files
- Keywords trigger automatic file reading
- Best for common workflows you want Claude to always use

**Example:**
```
Commits|1|commit,commits|read|.src/datastarui/.claude/commands/commit.md||Workflow description
```

When user says "help me commit", Claude automatically reads that file.

**Use for:**
- Critical workflows (commits, reviews, planning)
- Files Claude should read before responding
- Enforcing consistent patterns

## 2. Auto-Discovery (Complete Catalog)

**Source**: Scans all repos for `.claude/` directories and CLAUDE.md files
**Output**: Bottom section of INDEX.md (📁 Auto-Discovered Agent Files)

**How it works:**
- Makefile automatically finds all agent files
- Extracts descriptions from YAML frontmatter
- Creates browsable catalog with collapsible sections

**Use for:**
- Exploring available commands/agents
- Finding specialized tools
- Understanding repo-specific workflows

## How They Work Together

```
User Query: "help me commit my code"
     ↓
1. TRIGGER MATCH (Top section)
   - Finds keyword: "commit"
   - Says: "→ Read immediately: .src/datastarui/.claude/commands/commit.md"
   - Claude reads file and applies workflow
     ↓
2. CATALOG REFERENCE (Bottom section)
   - User can browse datastarui section
   - Sees 18 commands, 7 agents available
   - Discovers related: ci_commit, describe_pr, etc.
```

## Editing

### Add New Trigger
```bash
# Edit triggers.list
vim triggers.list

# Add line
NewCategory|9|keyword1,keyword2|read|path/to/file.md||Why this matters

# Regenerate
make index
```

### Auto-Discovery is Automatic
Just add `.claude/` directories to any repo in repos.list:
- Commands go in `.claude/commands/*.md`
- Agents go in `.claude/agents/*.md`
- Add `description:` in YAML frontmatter for better catalog entries

## Best Practices

**Triggers (curated)**:
- 5-10 core workflows max
- Keywords users naturally say
- Must-read files only

**Auto-Discovery (complete)**:
- Let it find everything
- Add descriptions to files for better discoverability
- Organize with CLAUDE.md at repo root

**Result**: Claude has both curated shortcuts AND complete catalog.
