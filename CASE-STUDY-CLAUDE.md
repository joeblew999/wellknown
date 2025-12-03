# Case Study: Claude MCP Walled Garden - Same Pattern, Different Company

> Discovered immediately after documenting Gmail’s iOS deep linking friction. The irony was not lost on us.

## Context

We had just documented how Google deliberately prevents Gmail deep linking on iOS to maintain lock-in. Claude (the AI assistant) helped write the case study and generated a markdown file.

**Next logical step:** Claude commits the file directly to the GitHub repo.

**Result:** Impossible. Same pattern as Gmail.

## The Problem

**Goal:** Have Claude raise a PR to add `CASE-STUDY.md` to the wellknown repo.

**What exists:**

- GitHub MCP (Model Context Protocol) server exists
- Claude supports MCP integrations
- Other MCPs work fine (Gmail, Calendar, Cloudflare, Reminders)

**What’s blocked:**

- Users cannot enable GitHub MCP on Claude.ai (web/iOS/Android)
- Users cannot add custom MCP servers
- MCP list is curated and locked by Anthropic

## What We Tried

### 1. Check Available Tools

Claude has access to:

- ✅ Gmail (read, search, threads)
- ✅ Google Calendar (events, scheduling)
- ✅ Reminders (create, update, delete)
- ✅ Cloudflare (D1, KV, R2, Workers - read only)
- ❌ GitHub (not available)

### 2. Cloudflare Worker Proxy Idea

**Theory:** Deploy a Worker that wraps GitHub API, call it via existing Cloudflare MCP.

**Reality:** Cloudflare MCP tools are:

- `workers_list` - list workers
- `workers_get_worker` - read worker details
- `workers_get_worker_code` - read source code
- ❌ No `workers_invoke` or `workers_deploy`

Infrastructure management only. No runtime invocation.

### 3. Check User Settings

**Question:** Can users enable additional MCPs?

**Answer:** No. Claude.ai has a locked-down MCP list. No self-service addition of:

- Pre-built MCPs (like GitHub)
- Custom MCP servers
- User-hosted integrations

### 4. GitHub URL Scheme

GitHub supports URL parameters for creating files:

```
https://github.com/OWNER/REPO/new/BRANCH?filename=PATH&value=CONTENT
```

**Limitation:** URL length limits (~2000-8000 chars). Full documents don’t fit.

## The Parallel

|Aspect           |Gmail iOS               |Claude MCP              |
|-----------------|------------------------|------------------------|
|Feature exists   |Deep linking URL schemes|GitHub MCP integration  |
|User benefit     |Open emails from any app|Commit code from chat   |
|Company blocks it|Google doesn’t implement|Anthropic doesn’t expose|
|Claimed reason   |(none given)            |(none given)            |
|Actual reason    |Lock-in / control       |Lock-in / control       |
|Workaround       |iOS Mail + Message-ID   |Claude Code CLI         |

## Why This Matters

### Same Incentive Structure

**Google:** If third-party apps can deep link to Gmail, users might prefer those apps. Google loses the entry point.

**Anthropic:** If users can add any MCP, they might:

- Connect Claude to competitors’ services
- Build workflows that don’t depend on Anthropic’s curated integrations
- Reduce engagement with Anthropic’s preferred partners

### The Curated Garden

Anthropic has clearly built MCP as an open protocol. But on their hosted product:

- They choose which MCPs are available
- They negotiate partnerships (Cloudflare, Google, etc.)
- Users get convenience in exchange for control

This isn’t inherently evil - it’s a business model. But it’s the same trade-off as Gmail: platform convenience vs. user autonomy.

### Where It Breaks Down

We just demonstrated a legitimate workflow:

1. User and Claude collaborate on documentation
1. Documentation is ready to commit
1. User wants Claude to raise a PR
1. Blocked - must copy/paste manually

The copy/paste step exists purely because of platform restrictions, not technical limitations.

## The Wellknown Solution

Wellknown’s architecture anticipates this:

```
┌─────────────────────────────────────────────────────────────┐
│                    User's Wellknown Server                  │
│                    (PocketBase + NATS)                      │
│                                                             │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐        │
│  │ Email Store │  │ GitHub API  │  │ Calendar    │        │
│  │ (Automerge) │  │   Proxy     │  │   Sync      │        │
│  └─────────────┘  └─────────────┘  └─────────────┘        │
│                                                             │
│                    ┌─────────────┐                         │
│                    │  MCP Server │ ← Speaks MCP protocol   │
│                    └──────┬──────┘                         │
└───────────────────────────┼─────────────────────────────────┘
                            │
                            ▼
                    ┌─────────────┐
                    │   Claude    │ ← Connects to YOUR server
                    │  (any UI)   │
                    └─────────────┘
```

**Key insight:** If Claude Code supports custom MCPs, and your wellknown server speaks MCP, then:

- You control what Claude can access
- You add GitHub, Jira, Notion, whatever you need
- No platform gatekeeper

### Implementation Path

1. **Wellknown MCP Server** - Add MCP protocol support to PocketBase
1. **GitHub Integration** - Wrap GitHub API as MCP tools
1. **Claude Code** - Connect to your wellknown MCP server
1. **Workflow restored** - Claude can commit, PR, interact with your repos

## Code Sketch

```go
package mcp

import (
    "context"
    "github.com/pocketbase/pocketbase/core"
)

// MCPServer implements the Model Context Protocol
type MCPServer struct {
    app core.App
    tools map[string]Tool
}

// Tool represents an MCP tool that Claude can invoke
type Tool struct {
    Name        string
    Description string
    Parameters  []Parameter
    Handler     func(ctx context.Context, params map[string]any) (any, error)
}

// RegisterGitHubTools adds GitHub operations as MCP tools
func (s *MCPServer) RegisterGitHubTools(token string) {
    s.tools["github_create_pr"] = Tool{
        Name:        "github_create_pr",
        Description: "Create a pull request with new or modified files",
        Parameters: []Parameter{
            {Name: "repo", Type: "string", Required: true},
            {Name: "branch", Type: "string", Required: true},
            {Name: "title", Type: "string", Required: true},
            {Name: "files", Type: "array", Required: true},
        },
        Handler: func(ctx context.Context, params map[string]any) (any, error) {
            // GitHub API calls here
            return createPullRequest(token, params)
        },
    }
    
    s.tools["github_commit_file"] = Tool{
        Name:        "github_commit_file",
        Description: "Commit a file to a repository",
        Parameters: []Parameter{
            {Name: "repo", Type: "string", Required: true},
            {Name: "path", Type: "string", Required: true},
            {Name: "content", Type: "string", Required: true},
            {Name: "message", Type: "string", Required: true},
        },
        Handler: func(ctx context.Context, params map[string]any) (any, error) {
            return commitFile(token, params)
        },
    }
}
```

## Workarounds (For Now)

1. **Claude Code CLI** - Supports custom MCPs, run locally
1. **Copy/paste** - Claude generates, user commits manually
1. **GitHub web editor** - Claude provides link with partial content pre-filled
1. **Feature request** - Tell Anthropic to open up MCP additions

## Conclusion

Within one hour of documenting Google’s platform lock-in, we hit Anthropic’s platform lock-in.

The tools exist. The protocols are open. The integration is technically trivial. But the platform owner decides what’s allowed.

**Wellknown’s thesis holds:** The only way to guarantee interoperability is to own the infrastructure layer. Don’t ask permission from platforms - build your own gateway that speaks their protocols.

When you own your MCP server, you decide which tools Claude gets.

-----

*Documented: December 2024*  
*Context: Trying to have Claude commit the Gmail case study to GitHub*  
*Irony level: Maximum*
