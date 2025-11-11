# Reference Repository Index

> **Auto-generated** from repos.list and discovered agent files
> Last updated: 2025-11-11 10:37:13
> Run `make index` to regenerate

Quick reference guide for finding patterns and examples in `.src/` repositories.

---

## 📦 Available Repositories

- **pocketbase/** (v0.32.0) - PocketBase (official source - matches go.mod version)
- **pocketbase-ha/** (v0.0.9) - PocketBase-HA (High Availability implementation)
- **pocketbase-gogen/** (v0.7.0) - pocketbase-gogen (type-safe code generator - matches Makefile version)
- **presentator/** (v3.4.28) - Presentator (PocketBase library pattern example)
- **jsonschema/** (v6.0.2) - jsonschema (validation library - matches go.mod v6.0.2)
- **goPocJsonSchemaForm/** (main) - goPocJsonSchemaForm (dynamic form generation example)
- **datastarui/** (main) - DatastarUI (Go/templ shadcn/ui components with Datastar)
- **gic/** (v0.5.0) - gic (Git + Claude AI commit message generator with MCP server mode)
- **mcp-go-sdk/** (v1.1.0) - MCP Go SDK (Model Context Protocol implementation - stable version)
- **anthropic-go-sdk/** (v1.17.0) - Anthropic Go SDK (Claude API official client - stable version)
- **peanats/** (v0.20.2) - peanats (NATS integration for PocketBase)
- **terraform-provider-nsc/** (v0.13.2) - terraform-provider-nsc (Terraform provider for NATS Security CLI)
- **datastar-go/** (v1.0.3) - Datastar Go SDK (v1.0.3 release)
- **northstar/** (main) - northstar (reference Datastar + Templ example)
- **via/** (main) - via (Go web framework)

---

## 🎯 When to Use These Files

**AUTO-TRIGGER**: Read specified files IMMEDIATELY when user query matches triggers.

### 1️⃣ Commits
**Triggers**: `commit`, `commits`, `committing`, `git commit`

→ **Read immediately**: `.src/datastarui/.claude/commands/commit.md`


**Why**: Workflow: review changes → plan commits → execute without Claude attribution

---

### 2️⃣ Code Review
**Triggers**: `code review`, `describe`, `merge request`, `PR`, `pr description`, `pull request`, `review`, `reviewing`, `summarize changes`

→ **Read immediately**: `.src/datastarui/.claude/commands/local_review.md` (worktree-based)

→ **Read immediately**: `.src/datastarui/.claude/commands/describe_pr.md` (PR templates)


**Why**: Workflow: setup worktree → install deps → launch session

---

### 3️⃣ Planning
**Triggers**: `architecture`, `breakdown`, `design`, `execution`, `implement`, `implementing`, `plan`, `planning`

→ **Read immediately**: `.src/datastarui/.claude/commands/create_plan.md` (planning phase)

→ **Read immediately**: `.src/datastarui/.claude/commands/implement_plan.md` (execution phase)


**Why**: Workflow: research context → interactive planning → success criteria

---

### 4️⃣ Datastar
**Triggers**: `api`, `app`, `build`, `create`, `datastar`, `example`, `pure go`, `reactive`, `sdk`, `sse`, `via`, `web framework`

→ **Read immediately**: `.src/DATASTAR.md`

→ **Browse for patterns**: `.src/datastarui/components/` (20 components)

→ **Browse for patterns**: `.src/northstar/features/` (6 examples)

→ **Browse for patterns**: `.src/datastar-go/datastar/` (v1.0.3 SDK)

→ **Browse for patterns**: `.src/via/internal/examples/` (7 examples)


**Why**: Guide: compares 4 approaches (via/northstar/components/SDK) with examples and recommendations

---

### 5️⃣ PocketBase
**Triggers**: `backend`, `cluster`, `codegen`, `database`, `events`, `ha`, `nats`, `pb`, `pocketbase`, `types`

→ **Browse for patterns**: `.src/pocketbase/` (v0.32.0)

→ **Read immediately**: `.src/peanats/CLAUDE.md`

→ **Browse for patterns**: `.src/pocketbase-ha/` (HA setup)

→ **Browse for patterns**: `.src/pocketbase-gogen/` (generator)


**Why**: Pattern source: migrations, stores, hooks, events, core/app.go

---

### 6️⃣ AI
**Triggers**: `ai`, `anthropic`, `claude`, `commit messages`, `git`, `mcp`, `protocol`, `server`, `streaming`, `tools`

→ **Browse for patterns**: `.src/anthropic-go-sdk/examples/` (v1.17.0)

→ **Browse for patterns**: `.src/mcp-go-sdk/examples/` (v1.1.0)

→ **Browse for patterns**: `.src/gic/` (MCP mode)


**Why**: Pattern source: streaming, tools, vision, prompt caching, messages API

---

### 7️⃣ Schema
**Triggers**: `dynamic ui`, `forms`, `json schema`, `schema to ui`, `validation`

→ **Browse for patterns**: `.src/jsonschema/` (v6.0.2)

→ **Browse for patterns**: `.src/goPocJsonSchemaForm/` (POC)

**Why**: Pattern source: schema compilation, validation, custom formats, draft 2020-12

---

## 📖 How to Use This Index

1. **Trigger Match** → Use trigger keywords above to find relevant files
2. **Browse Catalog** → Explore all available agent files below
3. **Read CLAUDE.md** → Each repo's CLAUDE.md has project-specific context

---

## 📁 Auto-Discovered Agent Files

All `.claude/` directories and CLAUDE.md files found in reference repositories:

### datastarui/ - DatastarUI (Go/templ shadcn/ui components with Datastar)

**[CLAUDE.md](datastarui/CLAUDE.md)** - Project context and architecture

**Available**: 18 commands, 7 agents

<details><summary>Commands (18)</summary>

- [`ci_commit`](datastarui/.claude/commands/ci_commit.md) - Create git commits for session changes with clear, atomic messages
- [`ci_describe_pr`](datastarui/.claude/commands/ci_describe_pr.md) - Generate comprehensive PR descriptions following repository templates
- [`commit`](datastarui/.claude/commands/commit.md) - Create git commits with user approval and no Claude attribution
- [`create_handoff`](datastarui/.claude/commands/create_handoff.md) - Create handoff document for transferring work to another session
- [`create_plan`](datastarui/.claude/commands/create_plan.md) - Create detailed implementation plans with thorough research and iteration
- [`create_worktree`](datastarui/.claude/commands/create_worktree.md) - Create worktree and launch implementation session for a plan
- [`debug`](datastarui/.claude/commands/debug.md) - Debug issues by investigating logs, database state, and git history
- [`describe_pr`](datastarui/.claude/commands/describe_pr.md) - Generate comprehensive PR descriptions following repository templates
- [`founder_mode`](datastarui/.claude/commands/founder_mode.md) - Create Linear ticket and PR for experimental features after implementation
- [`implement_plan`](datastarui/.claude/commands/implement_plan.md) - Implement technical plans from thoughts/shared/plans with verification
- [`iterate_plan`](datastarui/.claude/commands/iterate_plan.md) - Iterate on existing implementation plans with thorough research and updates
- [`linear`](datastarui/.claude/commands/linear.md) - Manage Linear tickets - create, update, comment, and follow workflow patterns
- [`local_review`](datastarui/.claude/commands/local_review.md) - Set up worktree for reviewing colleague's branch
- [`oneshot_plan`](datastarui/.claude/commands/oneshot_plan.md) - Execute ralph plan and implementation for a ticket
- [`oneshot`](datastarui/.claude/commands/oneshot.md) - Research ticket and launch planning session
- [`research_codebase`](datastarui/.claude/commands/research_codebase.md) - Research codebase comprehensively using parallel sub-agents
- [`resume_handoff`](datastarui/.claude/commands/resume_handoff.md) - Resume work from handoff document with context analysis and validation
- [`validate_plan`](datastarui/.claude/commands/validate_plan.md) - Validate implementation against plan, verify success criteria, identify issues

</details>

<details><summary>Agents (7)</summary>

- [`codebase-analyzer`](datastarui/.claude/agents/codebase-analyzer.md) - Analyzes codebase implementation details. Call the codebase-analyzer agent when you need to find detailed information about specific components. As always, the more detailed your request prompt, the better! :)
- [`codebase-locator`](datastarui/.claude/agents/codebase-locator.md) - Locates files, directories, and components relevant to a feature or task. Call `codebase-locator` with human language prompt describing what you're looking for. Basically a "Super Grep/Glob/LS tool" — Use it if you find yourself desiring to use one of these tools more than once.
- [`codebase-pattern-finder`](datastarui/.claude/agents/codebase-pattern-finder.md) - codebase-pattern-finder is a useful subagent_type for finding similar implementations, usage examples, or existing patterns that can be modeled after. It will give you concrete code examples based on what you're looking for! It's sorta like codebase-locator, but it will not only tell you the location of files, it will also give you code details!
- [`playwright-component-tester`](datastarui/.claude/agents/playwright-component-tester.md) - Use this agent when you need to test a DatastarUI component using Playwright browser automation. This agent will navigate to the component's demo page, interact with the component, check for console errors, verify functionality, and provide a concise summary of any issues found. Examples: <example>Context: User has just finished implementing a new date picker component and wants to verify it works correctly. user: "I just finished the date picker component, can you test it?" assistant: "I'll use the playwright-component-tester agent to test your date picker component and verify it's working correctly."</example> <example>Context: User is debugging a select component that seems to have issues with keyboard navigation. user: "The select component keyboard navigation seems broken, can you check what's wrong?" assistant: "Let me use the playwright-component-tester agent to test the select component's keyboard navigation and identify any issues."</example> <example>Context: User wants to verify all variants of a button component are rendering correctly. user: "Can you test all the button variants to make sure they're working?" assistant: "I'll use the playwright-component-tester agent to test all button variants and verify they're rendering and functioning correctly."</example>
- [`thoughts-analyzer`](datastarui/.claude/agents/thoughts-analyzer.md) - The research equivalent of codebase-analyzer. Use this subagent_type when wanting to deep dive on a research topic. Not commonly needed otherwise.
- [`thoughts-locator`](datastarui/.claude/agents/thoughts-locator.md) - Discovers relevant documents in thoughts/ directory (We use this for all sorts of metadata storage!). This is really only relevant/needed when you're in a reseaching mood and need to figure out if we have random thoughts written down that are relevant to your current research task. Based on the name, I imagine you can guess this is the `thoughts` equivilent of `codebase-locator`
- [`web-search-researcher`](datastarui/.claude/agents/web-search-researcher.md) - Do you find yourself desiring information that you don't quite feel well-trained (confident) on? Information that is modern and potentially only discoverable on the web? Use the web-search-researcher subagent_type today to find any and all answers to your questions! It will research deeply to figure out and attempt to answer your questions! If you aren't immediately satisfied you can get your money back! (Not really - but you can re-run web-search-researcher with an altered prompt in the event you're not satisfied the first time)

</details>


### peanats/ - peanats (NATS integration for PocketBase)

**[CLAUDE.md](peanats/CLAUDE.md)** - Project context and architecture


### terraform-provider-nsc/ - terraform-provider-nsc (Terraform provider for NATS Security CLI)

**[CLAUDE.md](terraform-provider-nsc/CLAUDE.md)** - Project context and architecture


---

## 🔍 Discovery Commands

```bash
# Sync all repositories and regenerate this index
make install

# Check repository status
make status

# Manually regenerate this index
make index
```

---

**Always check `.src/` before web searches!**
