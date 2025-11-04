# Rod Browser Automation

Automates Google Cloud OAuth credential setup using headless Chrome.

## What It Does

This tool uses [Rod](https://github.com/go-rod/rod) (a Chrome DevTools Protocol driver) to automate the tedious process of creating OAuth credentials in Google Cloud Console.

**Problem**: Creating OAuth credentials manually requires:
- Logging into Google Cloud Console
- Navigating through multiple pages
- Clicking buttons, filling forms
- Copying Client ID and Secret to `.env` file

**Solution**: This tool does all of that automatically!

## Quick Start

```bash
# Headless mode (no GUI, suitable for CI/automation)
make run

# GUI mode (watch the browser automation happen)
make show

# Build standalone binary
make build

# Run the binary
../../.bin/rod-automation
../../.bin/rod-automation -show  # GUI mode
```

## Features

- ✅ **Cookie Persistence**: Saves session cookies to skip repeated logins
- ✅ **Headless by Default**: No GUI needed, runs in background
- ✅ **GUI Mode**: Use `-show` flag to watch automation in action
- ✅ **Auto-Downloads Chromium**: First run downloads required browser
- 🚧 **TODO**: Actual credential creation automation (skeleton ready!)

## Development

### Hot Reload with Air

```bash
# Install Air (first time only)
make install-air

# Run with hot-reload (headless)
make dev

# Run with hot-reload (GUI mode)
make dev-show
```

Air will automatically rebuild and restart when you edit `main.go`.

### Project Structure

```
cmd/rod/
├── main.go           # Main automation code
├── go.mod           # Dependencies
├── Makefile         # Build & run commands
├── .air.toml        # Air hot-reload config
├── README.md        # This file
└── tmp/             # Air temp files (gitignored)
```

## How It Works

1. **Launch Browser**: Starts headless Chromium
2. **Load Cookies**: Restores session from previous run
3. **Navigate**: Goes to Google Cloud Console credentials page
4. **Automate** (TODO): Clicks buttons, fills forms, extracts credentials
5. **Save Cookies**: Persists session for next run
6. **Close**: Cleanly shuts down browser

## Local State Directory

All automation state is stored locally in the `.rod/` directory:
```
cmd/rod/.rod/
├── cookies.json      # Session cookies (allows skipping login)
└── automation.log    # Complete log of all runs
```

This directory is gitignored and can be cleaned with `make clean`.

## Next Steps

The skeleton is working! Next steps:
1. Add actual automation logic in the TODO section
2. Click "CREATE CREDENTIALS" button
3. Select "OAuth client ID"
4. Fill form (Application type, Authorized redirect URIs)
5. Extract generated Client ID and Secret
6. Write to `../../pb/base/.env` file

## CLI Flags

- `-show`: Show browser GUI (default: headless)
- `-clear-cookies`: Clear saved session cookies before running (force fresh login)

Examples:
```bash
go run main.go -show                    # GUI mode
go run main.go -clear-cookies           # Headless with fresh login
go run main.go -show -clear-cookies     # GUI with fresh login
```

## Makefile Commands

```bash
make help          # Show all commands
make run           # Run headless (uses saved cookies)
make show          # Run with GUI (uses saved cookies)
make fresh         # Run headless with fresh login (clears cookies)
make fresh-show    # Run with GUI and fresh login (clears cookies)
make dev           # Hot-reload (headless)
make dev-show      # Hot-reload with GUI
make build         # Build binary to ../../.bin/rod-automation
make clean         # Clean temp files
make install-air   # Install Air for hot-reload
```

## Related Files

- [pb/GCP_SETUP.md](../../pb/GCP_SETUP.md) - Manual setup guide
- [pb/OAUTH_SETUP.md](../../pb/OAUTH_SETUP.md) - OAuth overview
- [pb/base/.env.example](../../pb/base/.env.example) - Environment template

## Resources

- Rod Documentation: https://go-rod.github.io/
- Rod GitHub: https://github.com/go-rod/rod
- Chrome DevTools Protocol: https://chromedevtools.github.io/devtools-protocol/
