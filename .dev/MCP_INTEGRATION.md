# PocketBase MCP Integration

This document explains how to integrate PocketBase with Claude (Desktop or VSCode) using the Model Context Protocol (MCP).

## What is MCP?

MCP (Model Context Protocol) allows Claude to directly interact with your PocketBase database. Claude can:
- List all collections
- Query records with filters
- Create, update, and delete records
- View collection schemas
- Perform complex database operations via natural language

## Quick Start

Choose your Claude client:

- **[VSCode/Claude Code Setup](#vscode-claude-code-setup)** (Recommended for development)
- **[Claude Desktop Setup](#claude-desktop-setup)** (Standalone app)

---

## VSCode / Claude Code Setup

### 1. Run Automated Setup

```bash
make vscode-mcp-setup
```

This will:
1. Build the binary (`.bin/wellknown-pb`)
2. Create `.vscode/mcp.json` with your project paths
3. Show next steps

### 2. Reload VSCode

Press `Cmd+Shift+P` (or `Ctrl+Shift+P` on Windows/Linux) and run:
```
Developer: Reload Window
```

### 3. Verify MCP Server

Press `Cmd+Shift+P` and run:
```
MCP: List Servers
```

You should see **wellknown-pocketbase** in the list.

### 4. Test Integration

Ask Claude in the chat:
```
What PocketBase collections exist?
```

### Troubleshooting VSCode Setup

**Server not showing up?**
- Check `.vscode/mcp.json` exists and has absolute paths
- Verify binary exists: `ls -la .bin/wellknown-pb`
- View logs: `Cmd+Shift+P` → `MCP: Show Output`

**Want to reconfigure?**
```bash
# Regenerate config
make vscode-mcp-setup

# Or manually edit
code .vscode/mcp.json
```

**Manual setup (if automated setup fails)**:
1. Copy `.vscode/mcp.json.example` to `.vscode/mcp.json`
2. Replace `/ABSOLUTE/PATH/TO/WELLKNOWN` with your project path
3. Reload VSCode

---

## Claude Desktop Setup

### 1. Build the Binary

```bash
make bin
```

This creates `.bin/wellknown-pb` with the MCP server.

### 2. Configure Claude Desktop

Edit your Claude Desktop configuration file:

**macOS**: `~/Library/Application Support/Claude/claude_desktop_config.json`

**Linux**: `~/.config/Claude/claude_desktop_config.json`

**Windows**: `%APPDATA%\Claude\claude_desktop_config.json`

Add the PocketBase MCP server:

```json
{
  "mcpServers": {
    "pocketbase": {
      "command": "/FULL/PATH/TO/.bin/wellknown-pb",
      "args": ["mcp"],
      "env": {
        "PB_DATA": "/FULL/PATH/TO/.data/pb"
      }
    }
  }
}
```

**Important**: Use **absolute paths** for both `command` and `PB_DATA`.

### 3. Restart Claude Desktop

After saving the configuration, restart Claude Desktop to load the MCP server.

### 4. Verify Connection

In Claude Desktop, you should see a "hammer" icon (🔨) indicating MCP tools are available.

Try asking:
- "List all my PocketBase collections"
- "Show me the users collection schema"
- "Query all records from the accounts collection"

## Available MCP Tools

### 1. `list_collections`

Lists all PocketBase collections with their schemas.

**Example**:
```
Claude: List all my collections
```

### 2. `query_records`

Query records from a collection with optional filtering, sorting, and pagination.

**Parameters**:
- `collection` (required): Collection name
- `filter` (optional): PocketBase filter query
- `sort` (optional): Sort order (e.g., `-created` for descending)
- `limit` (optional): Max records to return (default 50)
- `page` (optional): Page number (default 1)

**Examples**:
```
Claude: Show me all accounts where balance > 1000
Claude: List the 10 most recent transactions
Claude: Query users where email contains 'gmail'
```

### 3. `get_record`

Get a specific record by its ID.

**Parameters**:
- `collection` (required): Collection name
- `record_id` (required): Record ID

**Example**:
```
Claude: Get the account with ID abc123
```

### 4. `create_record`

Create a new record in a collection.

**Parameters**:
- `collection` (required): Collection name
- `data` (required): Record data as JSON

**Example**:
```
Claude: Create a new account with user_id "user123", account_number "ACC001", and balance 1000
```

### 5. `update_record`

Update an existing record.

**Parameters**:
- `collection` (required): Collection name
- `record_id` (required): Record ID
- `data` (required): Fields to update

**Example**:
```
Claude: Update account abc123 to set balance to 2500
```

### 6. `delete_record`

Delete a record from a collection.

**Parameters**:
- `collection` (required): Collection name
- `record_id` (required): Record ID

**Example**:
```
Claude: Delete the transaction with ID xyz789
```

## Available MCP Resources

Resources are read-only views of PocketBase data that Claude can access.

### 1. `pocketbase://schema/collections`

Full schema information for all collections including field types and options.

### 2. `pocketbase://collections`

Simple list of all collection names.

## Development & Testing

### Run MCP Server Directly

```bash
make mcp
```

The server will start on stdio and wait for MCP requests.

### Debug Mode

To see MCP server logs, check Claude Desktop logs:

**macOS**: `~/Library/Logs/Claude/mcp-server-pocketbase.log`

### Testing Without Claude Desktop

You can test the MCP server using the MCP inspector:

```bash
npx @modelcontextprotocol/inspector go run . mcp
```

## Example Workflows

### 1. Database Exploration

```
You: What collections do I have?
Claude: [Uses list_collections tool]
Claude: You have 5 collections: users, accounts, transactions, google_tokens, and _superusers.

You: Show me the accounts collection schema
Claude: [Uses pocketbase://schema/collections resource]
Claude: The accounts collection has these fields:
- user_id (text, required)
- account_number (text, required)
- account_name (text)
- balance (number)
- currency (text)
- is_active (bool)
```

### 2. Data Analysis

```
You: How many transactions do I have in total?
Claude: [Uses query_records tool with transactions collection]
Claude: You have 147 transactions in total.

You: Show me the 5 largest transactions
Claude: [Uses query_records with sort="-amount" and limit=5]
Claude: Here are your 5 largest transactions:
1. $5,000 - Salary Payment
2. $2,500 - Rent Payment
3. $1,200 - Investment
...
```

### 3. Record Management

```
You: Create a test account with $500 balance
Claude: [Uses create_record tool]
Claude: I've created a test account:
- Account Number: ACC123
- Balance: $500
- Status: Active

You: Actually, make that $1000
Claude: [Uses update_record tool]
Claude: Updated! The account now has a balance of $1,000.
```

## Troubleshooting

### MCP Server Not Appearing in Claude Desktop

1. Check Claude Desktop logs for errors
2. Verify the binary path in config is absolute and correct
3. Ensure PB_DATA path exists and is readable
4. Try running `make mcp` manually to test the server

### "Permission Denied" Errors

Make sure the binary is executable:

```bash
chmod +x .bin/wellknown-pb
```

### "Collection Not Found" Errors

Ensure PocketBase has been initialized:

```bash
make run  # Initialize PocketBase first
# Press Ctrl+C after it starts
```

### Claude Can't Query Records

Check that:
1. The PB_DATA path in config points to your actual `.data/pb` directory
2. The database file exists: `.data/pb/data.db`
3. You have records in the collection you're querying

## Security Considerations

⚠️ **Important**: The MCP server runs with full PocketBase access.

**For Development**:
- MCP server is safe for local development
- Runs on your machine only (stdio, not network)
- Only Claude Desktop can access it

**For Production**:
- Do NOT expose MCP server over network
- Consider adding authentication layer
- Limit MCP to read-only operations if needed
- Use separate PocketBase instances for dev/prod

## Architecture

```
┌─────────────────┐
│ Claude Desktop  │
│                 │
│  [MCP Client]   │
└────────┬────────┘
         │ stdio
         │
┌────────▼────────┐
│  MCP Server     │
│  (wellknown)    │
│                 │
│  Tools:         │
│  - list_coll    │
│  - query_recs   │
│  - create_rec   │
│  - update_rec   │
│  - delete_rec   │
└────────┬────────┘
         │
┌────────▼────────┐
│  PocketBase     │
│  Database       │
│  (.data/pb)     │
└─────────────────┘
```

## Next Steps

- Explore your data with Claude
- Use Claude to generate queries and reports
- Let Claude help manage your database
- Build complex workflows with natural language

For more information on MCP, see: https://modelcontextprotocol.io
