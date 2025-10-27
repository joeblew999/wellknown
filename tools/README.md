# Tools Directory

**Purpose**: Development and testing tools that are NOT part of the core library.

## Why This Directory Exists

The `tools/` directory contains:
- Setup scripts
- Development utilities
- Testing infrastructure
- Build tools

These are **not library code** - they're tools to help develop and test the library.

## Directory Structure

```
tools/
├── gcp-setup/               # Google Cloud Project setup
│   ├── main.go             # Automates GCP project creation
│   └── README.md           # Setup instructions
├── pocketbase-gogen/       # Type-safe code generator for Pocketbase
│   └── README.md           # Usage guide
└── README.md               # This file
```

## Tools Available

### `gcp-setup/` - Google Cloud Project Setup

**What**: Automates Google Cloud Project creation and API enablement
**When to use**: Setting up Calendar API integration for testing
**Command**: `make gcp-setup` or `cd tools/gcp-setup && go run main.go`

Creates a GCP project with:
- Google Calendar API enabled
- OAuth2 API enabled
- Instructions for OAuth credential setup

**See**: `tools/gcp-setup/README.md` for details

### `pocketbase-gogen/` - Type-Safe Code Generator

**What**: Generates type-safe Go accessor structures from Pocketbase collections
**When to use**: Working with complex PB schemas, want type safety and autocomplete
**Install**: `go install github.com/snonky/pocketbase-gogen@latest`

Converts raw Pocketbase records into typed Go structs:
- Type-safe getters and setters
- IDE autocomplete support
- Custom methods with typed data
- Optional hooks and utils

**See**: `tools/pocketbase-gogen/README.md` for usage examples

## Future Tools

Potential tools to add:
- `schema-validator/` - Validate JSON schemas
- `test-data-generator/` - Generate test event data
- `oauth-token-manager/` - Manage OAuth tokens for testing
- `api-tester/` - Test against live APIs

## Not in tools/

These directories are NOT tools:
- `pkg/` - Library code
- `cmd/` - Application binaries
- `docs/` - Documentation

Tools are specifically for **development/testing**, not production use.
