# AI Agent Documentation System

> **Working Directory**: `/Users/apple/workspace/go/src/github.com/joeblew999/wellknown/.src/`
> All repository paths in STACK-*.md files are relative to this directory.

This directory contains reference repositories organized by tech stack to help AI agents quickly find relevant code patterns and examples.

## Purpose

When working on code, AI agents should:
1. Check `.src/INDEX.md` BEFORE web searches
2. Browse stack-specific docs (STACK-*.md) for relevant patterns
3. Read actual code in referenced repositories for implementation details

## Stack Categories

Each stack groups related technologies with:
- **Tags**: Technology names (pocketbase, datastar, via, hugo, etc.)
- **Repos**: Actual codebases to reference
- **Patterns**: Common file locations and code patterns
- **When to Use**: Scenarios where this stack is relevant

## Common Triggers

### PocketBase Backend
**Triggers**: `pocketbase`
→ Check **STACK-pocketbase.md**

### Datastar/Hypermedia
**Triggers**: `datastar`, `via`
→ Check **STACK-datastar.md**

### Go HTML Components
**Triggers**: `gomponents`
→ Check **STACK-gomponents.md**

### Cloudflare Workers
**Triggers**: `cloudflare`, `workers`
→ Check **STACK-cloudflare.md**

### AI/LLM Integration
**Triggers**: `mcp`, `model context protocol`, `anthropic`, `claude`
→ Check **STACK-mcp.md** or **STACK-ai.md**

### Static Site Generation
**Triggers**: `hugo`
→ Check **STACK-hugo.md**

### JSON Schema Validation
**Triggers**: `jsonschema`
→ Check **STACK-jsonschema.md**

### NATS Messaging
**Triggers**: `nats`
→ Check **STACK-nats.md**

## File Structure Patterns

### PocketBase Projects
```
pkg/
  cmd/pocketbase/main.go        # Entry point
  pb_migrations/                # Database migrations (auto-generated)
  pb_data/                      # SQLite database and uploads
```

### Via Projects
```
examples/                       # Working examples
  {name}/
    main.go                     # Example entry point
    {name}.go                   # Core logic
```

### Datastar Projects
```
components/                     # Reusable UI components
examples/                       # Demo applications
```

### Go Module Projects
```
go.mod                          # Dependencies
go.sum                          # Checksums
Makefile                        # Build commands
README.md                       # Documentation
CLAUDE.md                       # AI agent context (if present)
```

## How to Use This System

1. **User asks about a technology**
   - Parse their message for trigger keywords (technology names)
   - Open relevant STACK-*.md file
   - Read the referenced repos for patterns

2. **User asks to implement something**
   - Identify which stack(s) are relevant
   - Check "Common Patterns" in stack docs
   - Browse actual code in repos for examples

3. **User has an error**
   - Check if error relates to a known stack
   - Look for similar patterns in repos
   - Reference CLAUDE.md files for context

4. **Always prefer local examples over web searches**
   - These repos are curated and version-locked
   - They match the project's actual dependencies
   - They contain working, tested code

## Maintenance

- Run `make agent` to regenerate documentation
- Edit `stacks.list` to add/remove/reorganize stacks
- Edit `repos.list` to add/remove repositories
- This file (agent-prompt.md) documents the system itself
