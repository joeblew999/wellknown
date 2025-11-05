# MCP Testing Guide

This document describes how to test the PocketBase MCP (Model Context Protocol) server integration.

## Test Hierarchy

```
┌─────────────────────────────────────┐
│ Level 3: Claude Desktop/Code (E2E) │  Real-world usage
├─────────────────────────────────────┤
│ Level 2: MCP Inspector (Manual)    │  Interactive debugging
├─────────────────────────────────────┤
│ Level 1: Go Tests (Automated)      │  In-memory transport
└─────────────────────────────────────┘
```

## Quick Start

```bash
# Run all MCP tests
make test-mcp

# Launch interactive inspector
make test-mcp-inspector

# Generate Claude Desktop config
make test-mcp-config
```

---

## Level 1: Automated Go Tests

**Purpose**: Fast, automated validation of MCP protocol implementation

**Location**: `pkg/pbmcp/server_test.go`

### Running Tests

```bash
# All MCP tests
make test-mcp

# All tests in the project
make test

# Unit tests only (short flag)
make test-unit

# Specific test
go test -v ./pkg/pbmcp -run TestServerCreation
```

### What's Tested

✅ **Server Creation** - Server initializes correctly
✅ **Tool Registration** - All 6 CRUD tools register
✅ **Resource Registration** - Schema and collections resources
✅ **List Collections** - Returns collection metadata
✅ **Query Records** - Filters, sorts, paginates
✅ **Create Records** - Inserts new data

### Test Pattern

All tests use **in-memory transports** from the MCP SDK:

```go
serverTransport, clientTransport := mcp.NewInMemoryTransports()

server := pbmcp.NewServer(app)
client := mcp.NewClient(testImpl, nil)

// Connect both sides
serverSession, _ := server.server.Connect(ctx, serverTransport, nil)
clientSession, _ := client.Connect(ctx, clientTransport, nil)

// Test tools
result, _ := clientSession.CallTool(ctx, &mcp.CallToolParams{
    Name: "list_collections",
    Arguments: map[string]any{},
})
```

**Advantages**:
- ⚡ Fast (milliseconds)
- 🤖 Fully automated
- 🔁 Repeatable
- 🚀 No external dependencies
- ✅ Perfect for CI/CD

---

## Level 2: Interactive Testing with MCP Inspector

**Purpose**: Visual debugging and interactive tool testing

**Official Tool**: [@modelcontextprotocol/inspector](https://github.com/modelcontextprotocol/inspector)

### Launch Inspector

```bash
make test-mcp-inspector
```

This will:
1. Build the binary (`make bin`)
2. Launch MCP Inspector via `npx`
3. Open browser at `http://localhost:6274`

### Inspector Features

📋 **View All Tools**
- Names, descriptions, schemas
- Input/output validation

🔧 **Interactive Testing**
- Call tools with custom arguments
- See request/response JSON
- Test error handling

📊 **Protocol Validation**
- Verify MCP compliance
- Check message formats
- Inspect transport layer

### Manual Installation

If you prefer to install globally:

```bash
npm install -g @modelcontextprotocol/inspector

# Then run directly
mcp-inspector /path/to/wellknown-pb mcp
```

### Example Usage

1. **Select a tool** from the left panel
2. **Fill in arguments** in the form
3. **Execute** and see results
4. **Inspect** the raw JSON-RPC messages

**Screenshot**: (Inspector UI shows tools list, input form, and response viewer)

---

## Level 3: Claude Desktop/Code Integration

**Purpose**: Real-world end-to-end testing with Claude

### Step 1: Build Binary

```bash
make bin
```

Outputs: `.bin/wellknown-pb`

### Step 2: Generate Config

```bash
make test-mcp-config
```

This displays the JSON config you need to add.

### Step 3: Update Claude Config

**macOS Location**:
```
~/Library/Application Support/Claude/claude_desktop_config.json
```

**Add This Section**:
```json
{
  "mcpServers": {
    "pocketbase": {
      "command": "/full/path/to/.bin/wellknown-pb",
      "args": ["mcp"],
      "env": {
        "PB_DATA_DIR": "/full/path/to/.data/pb"
      }
    }
  }
}
```

**Important**: Use **absolute paths** for both `command` and `PB_DATA_DIR`.

### Step 4: Restart Claude

- Claude Desktop: Quit and reopen the application
- Claude Code (VSCode): Reload window (Cmd+Shift+P → "Reload Window")

### Step 5: Test Integration

Ask Claude:

```
"What PocketBase collections exist?"
```

Expected behavior:
1. Claude calls `list_collections` tool
2. Returns collection names and schemas
3. You can then query, create, update records

### Example Interactions

**Query Records**:
```
"Show me all users in the test_users collection"
```

**Create Record**:
```
"Create a new user named Alice with email alice@example.com"
```

**Get Schema**:
```
"What fields does the test_users collection have?"
```

### Troubleshooting

#### Claude doesn't see the MCP server

**Symptoms**:
- No PocketBase tools available
- Claude says "I don't have access to that"

**Solutions**:
1. Check config path is correct (use absolute paths)
2. Verify binary exists: `ls -la .bin/wellknown-pb`
3. Verify binary is executable: `chmod +x .bin/wellknown-pb`
4. Check Claude logs (if available)
5. Restart Claude Desktop/Code

#### Tool calls fail

**Symptoms**:
- Claude tries to call tool but gets errors
- Timeout or connection errors

**Solutions**:
1. Test with MCP Inspector first
2. Check PocketBase database exists at `PB_DATA_DIR`
3. Verify collections exist: `make run` then visit admin UI
4. Check env vars are set correctly

#### Permission errors

**Symptoms**:
- "Permission denied" errors
- Database access failures

**Solutions**:
1. Check file permissions on `.data/` directory
2. Verify PocketBase can write to `PB_DATA_DIR`
3. Run `mkdir -p .data/pb` to ensure directory exists

---

## Test Data Setup

All tests use **in-memory SQLite** databases created by PocketBase's test utilities.

### Test Collections

**test_users**:
- `name` (text, required)
- `email` (text, required)
- `age` (number)

**test_posts**:
- `title` (text, required)
- `content` (text)
- `published` (bool)

### Creating Test Data

```go
import "github.com/joeblew999/wellknown/pkg/pbmcp/testutil"

app, _ := testutil.NewTestApp()
defer testutil.CleanupTestApp(app)

// Create a user
testutil.CreateTestRecord(app, "test_users", map[string]any{
    "name": "Alice",
    "email": "alice@example.com",
    "age": 30,
})
```

---

## Available Tools (v1.0)

| Tool | Description | Status |
|------|-------------|--------|
| `list_collections` | List all collections with schemas | ✅ Tested |
| `query_records` | Query with filter, sort, pagination | ✅ Tested |
| `get_record` | Get single record by ID | ⚠️  Implemented |
| `create_record` | Create new record | ✅ Tested |
| `update_record` | Update existing record | ⚠️  Implemented |
| `delete_record` | Delete record | ⚠️  Implemented |

Legend:
- ✅ Tested: Has automated tests
- ⚠️  Implemented: Works but needs tests
- 🚧 Planned: Not yet implemented

---

## Available Resources (v1.0)

| URI | Description | Status |
|-----|-------------|--------|
| `pocketbase://collections` | List of collection names | ✅ Tested |
| `pocketbase://schema/collections` | Full collection schemas | ⚠️  Implemented |

---

## CI/CD Integration

### GitHub Actions Example

```yaml
name: MCP Tests

on: [push, pull_request]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: '1.23'

      - name: Run MCP Tests
        run: make test-mcp
```

### Pre-commit Hook

```bash
#!/bin/sh
# .git/hooks/pre-commit

echo "Running MCP tests..."
make test-mcp

if [ $? -ne 0 ]; then
    echo "❌ MCP tests failed. Commit aborted."
    exit 1
fi
```

---

## Development Workflow

### TDD Cycle for New Tools

1. **Write test** in `pkg/pbmcp/server_test.go`
2. **Run test** with `go test -v ./pkg/pbmcp -run TestNewTool`
3. **Implement tool** in `pkg/pbmcp/tools.go`
4. **Verify** test passes
5. **Test manually** with MCP Inspector
6. **Test E2E** with Claude Desktop/Code

### Adding New Test Fixtures

1. Add fixture type to `pkg/pbmcp/testutil/fixtures.go`
2. Create helper function
3. Use in tests

Example:
```go
// fixtures.go
type TestComment struct {
    PostID  string
    Content string
    Author  string
}

func GetTestComments() []TestComment { /* ... */ }
func CommentToMap(comment TestComment) map[string]any { /* ... */ }
```

---

## Performance Benchmarks

All tests run in **< 1 second** total:

```bash
$ make test-mcp
🧪 Running MCP tests...
go test -v ./pkg/pbmcp/...
=== RUN   TestServerCreation
--- PASS: TestServerCreation (0.04s)
=== RUN   TestToolRegistration
--- PASS: TestToolRegistration (0.03s)
=== RUN   TestResourceRegistration
--- PASS: TestResourceRegistration (0.03s)
=== RUN   TestListCollectionsTool
--- PASS: TestListCollectionsTool (0.03s)
=== RUN   TestQueryRecordsTool
--- PASS: TestQueryRecordsTool (0.03s)
=== RUN   TestCreateRecordTool
--- PASS: TestCreateRecordTool (0.03s)
PASS
ok      github.com/joeblew999/wellknown/pkg/pbmcp      0.551s
```

---

## Debugging Tips

### Enable Verbose Logging

In `pkg/pbmcp/server.go`, add logging:

```go
import "log/slog"

func (s *Server) handleQueryRecords(...) {
    slog.Info("Query records", "collection", input.Collection, "filter", input.Filter)
    // ... rest of implementation
}
```

### Inspect Transport Messages

Use MCP Inspector's "Network" tab to see raw JSON-RPC messages.

### Test Individual Tools

```bash
# Test only query_records
go test -v ./pkg/pbmcp -run TestQueryRecordsTool
```

### Use PocketBase Admin UI

```bash
make run
# Open http://localhost:8090/_/
# Inspect collections, records, logs
```

---

## Common Issues

### "Collection not found"

**Cause**: Test tried to access non-existent collection

**Fix**: Ensure collection is created in test setup:
```go
testutil.CreateTestCollections(app)
```

### "Failed to unmarshal"

**Cause**: Tool output doesn't match expected structure

**Fix**: Check that handler returns correct output type:
```go
return nil, ListCollectionsOutput{Collections: result}, nil
```

### "Transport closed"

**Cause**: Server or client session closed prematurely

**Fix**: Use `defer clientSession.Close()` after connection

---

## Future Improvements

### Planned Tests
- [ ] Update record tool
- [ ] Delete record tool
- [ ] Get record tool
- [ ] Resource reading (schema, collections)
- [ ] Error handling (invalid input, missing collections)
- [ ] Concurrent tool calls
- [ ] Large result sets (pagination)

### Planned Features
- [ ] Batch operations (create multiple records)
- [ ] Transaction support
- [ ] Real-time subscriptions
- [ ] File upload/download tools
- [ ] Advanced query filters (relations, expand)

---

## References

- [MCP Specification](https://spec.modelcontextprotocol.io/)
- [MCP Go SDK](https://github.com/modelcontextprotocol/go-sdk)
- [MCP Inspector](https://github.com/modelcontextprotocol/inspector)
- [PocketBase Documentation](https://pocketbase.io/docs/)
- [Claude Desktop MCP Guide](https://docs.claude.com/claude-desktop)

---

## Getting Help

### Issues or Questions?

1. Check this document first
2. Review existing tests in `pkg/pbmcp/server_test.go`
3. Run MCP Inspector to debug interactively
4. Open an issue with:
   - Test output (`make test-mcp`)
   - Inspector screenshots
   - Claude Desktop config (sanitized)
   - Error messages

### Contributing Tests

Tests are welcome! Follow the existing patterns in `server_test.go`:

1. Use in-memory transports
2. Create minimal test data
3. Test both success and error cases
4. Keep tests fast and focused
