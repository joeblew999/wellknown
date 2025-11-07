# wellknown

**Universal Go library for generating and opening deep links across the Google and Apple app ecosystems.**
Pure Go · Zero deps · Deterministic URLs · Cross-platform.

---

## ✨ Overview

`wellknown` lets Go applications and CLIs create **native deep links** and **URL schemes** for common apps such as:

| Category | Google | Apple |
|-----------|---------|--------|
| Calendar | `googlecalendar://render?...` | `calshow:` |
| Maps | `comgooglemaps://?q=` | `maps://?q=` |
| Mail | `mailto:` | `mailto:` |
| Drive / Files | `googledrive://` | `shareddocuments://` |

The library also provides safe fallbacks to open the **web equivalents** when native apps aren't available.

---

## 🧩 Features

- ✅ **Pure Go** — no external dependencies.
- 🧠 **Deterministic**: same input → same output (great for reproducible infra / NATS messages).
- ⚙️ **Cross-platform**: works on macOS, Windows, Linux, iOS, and Android.
- 🕹 **Programmatic & CLI**: embed in binaries or call from shell scripts.
- 🔗 **App-aware**: automatically chooses local URL scheme vs. browser fallback.

---

## 🚀 Getting Started

### Quick Start

```bash
make go-dep    # Install development tools
make run       # Start unified server (API + Demo UI)
```

The server will start on **port 8090** with:

- **Admin UI**: [http://localhost:8090/_/](http://localhost:8090/_/)
- **Demo UI**: [http://localhost:8090/demo/](http://localhost:8090/demo/)
- **API Docs**: [http://localhost:8090/api/](http://localhost:8090/api/)

### Architecture

**Unified Server (Port 8090)**
- PocketBase backend with SQLite
- RESTful API endpoints (`/api/*`)
- Demo & testing UI (`/demo/*`)
- Admin interface (`/_/*`)

### Available Commands

See all commands:
```bash
make help
```

Common tasks:
```bash
make go-dep         # Install development tools
make run            # Start unified server
make gen            # Generate type-safe models from database
make bin            # Build production binary
make test           # Run all tests
make fly-deploy     # Deploy to Fly.io
```

---

## 📋 Migration Notice

**Note**: The standalone server (`wellknown server`) has been merged into the unified server.

All demo features are now available at `/demo/*` routes:

```bash
# Old (deprecated)
wellknown server                    # Port 8080
http://localhost:8080/google/calendar

# New (current)
wellknown pb serve                  # Port 8090
http://localhost:8090/demo/google/calendar
```

See [MIGRATION.md](MIGRATION.md) for full migration guide.

---

## 📚 Documentation

All usage instructions are kept up-to-date in the Makefile. Run `make help` to see available commands and their descriptions.

Additional documentation:
- [MIGRATION.md](MIGRATION.md) - Migration from separate servers
- [CLAUDE.md](CLAUDE.md) - Development rules and guidelines
