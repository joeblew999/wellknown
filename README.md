# wellknown

**Universal Go library for generating and opening deep links across the Google and Apple app ecosystems.**
Pure Go · Zero deps · Deterministic URLs · Cross-platform.

## WHY

### The Problem: Platform Lock-In

Today, most people are trapped in Big Tech ecosystems. When you use Gmail, Google Calendar, YouTube, or Apple Maps as your primary system, you:
- Don't own your data or user relationships
- Can't easily migrate to alternatives
- Are subject to their rules, algorithms, and business decisions
- Pay with your privacy and attention

### The Solution: Reverse the Relationship

**wellknown** enables you to flip the script: **Make Big Tech platforms work for YOU, not the other way around.**

#### 1. Technical Mechanism: URI Schemas as Your Open Gateway

URI schemas (like `mailto:`, `webcal://`, `maps://`) are standardized protocols that apps understand. By building your own URI schema gateway:

- **Your system becomes the source of truth** - All data lives on infrastructure YOU control (self-hosted or your chosen provider)
- **Interoperability is built-in** - URI schemas work across all platforms (iOS, Android, web, desktop)
- **You decide the routing** - When someone clicks a calendar link, YOUR system decides whether to open Apple Calendar, Google Calendar, or your own app

#### 2. Business Benefit: Own the User Relationship

With wellknown, you can implement this strategy:

**Example: Video Content**
1. Host your videos on YOUR infrastructure (self-hosted Peertube, Cloudflare R2, or your own cloud)
2. Publish COPIES to YouTube and Twitch for distribution and discovery
3. Use wellknown URI schemas in all your links (`wellknown://video/abc123`)
4. When users click links, they come to YOUR platform first
5. You control the experience, analytics, and user data
6. Big Tech platforms become free distribution channels instead of landlords

**This works for everything:**
- **Video**: Host on your server, mirror to YouTube/Twitch for reach
- **Email**: Own your mail server, integrate with Gmail for compatibility
- **Calendar**: Your CalDAV server, sync to Google/Apple for convenience
- **Maps**: Your geographic data, fallback to Google/Apple Maps when needed
- **Contacts**: Your CardDAV server, sync to platform address books
- **Files**: Your storage, selective sharing to Google Drive/Dropbox

#### 3. Philosophical Principle: Data Sovereignty

This isn't just about technology—it's about **digital autonomy**:

- **You own your content** - It lives on infrastructure you control
- **You own your audience** - Direct relationships, not mediated by algorithms
- **You get network effects without surrender** - Publish to big platforms for reach, but they don't own you
- **You can leave anytime** - No lock-in, because you were never locked in

### The Wellknown Advantage

Traditional approach:
```
User → YouTube (owns everything) → Your content (captive)
```

Wellknown approach:
```
User → Your Gateway → Your System (primary)
                   ↳→ YouTube (mirror for discovery)
```

**You control the front door.** Big Tech becomes optional infrastructure, not a prison.

This is how the early web worked—distributed, interoperable, user-owned. Wellknown brings that spirit back using modern URI schemas and self-hosting tools.

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
