# wellknown

https://github.com/joeblew999/wellknown

**Own your data, own your AI, own your routing, and optionally use Big Tech as enhancement services.**

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

#### 1.5. The Self-Syncing Foundation: Automerge CRDTs

**Here's what makes this truly powerful**: Your data isn't stored on "a server" - it's **distributed across all your devices** using [Automerge CRDTs](https://github.com/joeblew99/automerge-wazero-example).

**What are CRDTs?**
Conflict-free Replicated Data Types (CRDTs) are data structures that automatically sync and merge changes across multiple devices without conflicts. Think "Git for your data" but automatic and real-time.

**How it works:**

```
Your Phone (Automerge) ←→ Your Laptop (Automerge) ←→ Your Server (Automerge)
         ↓                        ↓                          ↓
    Calendar events          Email drafts               Contact cards
    (full copy)              (full copy)                (full copy)
```

**Every device has a complete copy**. Changes sync peer-to-peer when online, work offline, and automatically merge without conflicts.

**Key Capabilities:**

1. **Offline-First** - Edit emails, calendar events, contacts offline. Syncs when you reconnect.
2. **No Single Point of Failure** - If your server goes down, your phone still has everything.
3. **Git-like Version Control** - Full revision history of all changes. Undo anything, see who changed what, branch and merge data.
4. **Automatic Conflict Resolution** - Two devices edit the same contact? Automerge merges intelligently.
5. **Peer-to-Peer Sync** - Devices sync directly, or via your server, or via any relay. Your choice.

**Real-World Example: Email System**

Traditional (Gmail):
```
Your Phone → Gmail Server (owns everything) ← Your Laptop
```

Wellknown + Automerge:
```
Your Phone (full mailbox) ←→ Your Laptop (full mailbox) ←→ Your Server (full mailbox)
                                                              ↓ (optional)
                                                         Gmail (mirror copy)
```

- Compose email offline on phone → syncs to laptop via Automerge
- Send via your SMTP server → optionally mirror to Gmail for compatibility
- Full revision history: see every draft, every edit, recover anything
- If Gmail bans you, you still have everything

**Implementation:**

We use [automerge-wazero](https://github.com/joeblew99/automerge-wazero-example), a WebAssembly implementation of Automerge that runs in Go. This gives you:

- **Email**: IMAP/SMTP data structures in Automerge, syncs across all devices
- **Calendar**: CalDAV events in Automerge, works offline, syncs everywhere
- **Contacts**: CardDAV entries in Automerge, distributed address book
- **Files**: Content-addressed storage with Automerge metadata

**NATS + Automerge: The Perfect Pairing**

Here's why this combination is powerful:

**The Problem with Traditional Sync:**
- Traditional systems (like databases with replication) require messages to arrive **in order**
- If message #5 arrives before message #4, the system breaks or needs complex coordination
- Air-gapped operation is impossible - you need constant connectivity

**NATS: Fast Signal Distribution**
[NATS](https://nats.io) is a lightweight messaging system that pushes signals anywhere:
- "Device X has new data available"
- "Sync request from phone"
- "Calendar event updated"

But NATS doesn't guarantee message ordering across the network. Messages can arrive out of order.

**Automerge: Order-Independent Merging**
This is where Automerge CRDTs shine:
- **Messages can arrive in ANY order** - Automerge will merge them correctly
- **100% air-gapped operation** - Edit for days offline, sync when network returns
- **No coordination needed** - Each device independently merges changes

**How They Work Together:**

```
1. You edit a contact on your phone (air-gapped, no network)
2. Your laptop edits the same contact (also offline)
3. Your server is running but can't reach either device
4. Network comes back online
5. NATS signals: "phone has updates" → server
6. NATS signals: "laptop has updates" → server
7. Automerge sync: Changes merge regardless of order
8. Result: All three devices converge to the same state
```

**Traditional sync would fail** because:
- Messages arrive out of order
- Devices were offline for extended periods
- Multiple devices edited the same data

**Automerge + NATS succeeds** because:
- NATS provides fast signaling (when network is available)
- Automerge handles out-of-order merging (always works)
- Air-gapped operation is natural, not a failure mode

**Real-World Scenario:**

```
Day 1: Take your phone on a plane (offline for 12 hours)
       - Edit 50 emails, update 10 calendar events, modify 5 contacts

Day 2: Land and turn on WiFi
       - NATS signals your server: "phone has updates"
       - Automerge syncs all changes (order doesn't matter)
       - Meanwhile, your laptop made changes too
       - Everything merges correctly, no conflicts

Result: All devices have the same state, even though:
        - Changes happened offline
        - Messages arrived out of order
        - Multiple devices edited simultaneously
```

#### 1.6. Local-First AI: 100% Offline Intelligence

**The final piece**: Your devices don't just store and sync data - they have **local AI** that works completely offline using [Yzma](https://github.com/hybridgroup/yzma).

**Why Local AI Matters:**

Traditional AI (ChatGPT, Claude API, etc.):
- Requires internet connectivity
- Sends your data to Big Tech servers
- Costs per API call
- Subject to rate limits and censorship
- Privacy risk: they see everything you send

**Yzma: 100% Offline AI**

[Yzma](https://github.com/hybridgroup/yzma) is a WebAssembly-based AI runtime that runs locally:
- **Vision models**: Process images, video, 3D scenes offline
- **Language models**: Text generation, summarization, Q&A without internet
- **Multi-modal**: Text + images + 3D data together
- **Zero cloud dependency**: Everything runs on your device

**The Hybrid Approach:**

```
Your Device (Local AI via Yzma) → Fast, private, always available
         ↓ (optional, when you need more power)
Gateway AI (Claude, GPT-4, etc.) → Expensive, powerful, cloud-based
```

You decide when to use local AI vs. gateway AI:
- Quick tasks, privacy-sensitive: Local AI
- Complex reasoning, latest models: Gateway AI (optional)

**AI is the new Gateway drug**. Just like wellknown URIs make Big Tech platforms optional distribution channels, local AI makes Big Tech AI models optional enhancement services. You start with local, use gateway when beneficial, but you're never dependent.

**3D Vision System: The Tesla Approach**

We're building AI vision and 3D systems around Yzma because **3D AI has vast implications**:

**Tesla Analogy:**
- Tesla cars have multiple cameras (analog world)
- Sensor fusion combines camera feeds into unified 3D model
- Real-time 3D "cockpit" of the environment
- All processing happens locally in the car

**Our Implementation:**
```
Multiple Input Sources (cameras, sensors, 2D images)
         ↓
Yzma Vision Models (local processing)
         ↓
3D Reconstruction + Sensor Fusion
         ↓
Real-time 3D Understanding (offline-capable)
```

**Use Cases:**

**2D Vision (Images/Video):**
- OCR on documents (offline, private)
- Image classification and search
- Object detection in photos
- Video analysis and summarization

**3D Vision (Spatial Understanding):**
- Room scanning and 3D reconstruction
- AR/VR applications with spatial awareness
- Multi-camera sensor fusion
- Real-time environmental mapping

**AI + Your Data (Automerge):**
- AI analyzes your emails, calendar, contacts (locally, privately)
- Suggests actions, drafts responses, finds patterns
- Everything stays on your device unless you choose to sync
- Full revision history of AI's changes (via Automerge)

**Example Workflow:**

```
1. Take photo of business card (offline)
2. Local AI (Yzma) extracts name, email, phone (OCR)
3. Creates contact in Automerge (local)
4. Syncs to your devices via NATS
5. Optionally mirrors to Google Contacts
6. AI drafts email intro using local LLM
7. If you need better prose, send to Gateway AI (Claude)
8. Send email via your SMTP server
```

**Zero cloud dependency** for steps 1-6. Gateway AI (step 7) is optional.

**Why 3D AI Matters:**

The analog world is 3D. Traditional AI works with:
- Text (1D sequences)
- Images (2D planes)
- Video (2D + time)

But humans live in **3D space + time**. To build truly intelligent systems that understand the physical world:

- **Autonomous systems** need 3D understanding (robots, drones, vehicles)
- **AR/VR** needs real-time 3D scene analysis
- **Smart spaces** need spatial awareness (IoT, home automation)
- **Content creation** needs 3D capture and understanding

We're building the infrastructure for **local-first 3D AI** that:
- Runs on your devices (phones, laptops, edge servers)
- Works offline (no cloud dependency)
- Respects privacy (your data never leaves)
- Syncs via Automerge (distributed, version-controlled)
- Uses Gateway AI only when needed (optional enhancement)

**The Complete Power Stack:**

```
Layer 5: Local AI (Yzma) + Optional Gateway AI (Claude/GPT-4)
         ↓
Layer 4: NATS (signal distribution when network available)
         ↓
Layer 3: Wellknown Gateway (routing: which app opens your links?)
         ↓
Layer 2: Automerge CRDT (order-independent data sync)
         ↓
Layer 1: Optional Mirrors (publish to Gmail/YouTube/etc for reach)
```

**You own the entire stack**: data, AI, routing, and optionally use Big Tech services when beneficial.

This architecture gives you what Git gave to code: distributed ownership, full history, and independence from any single server. Plus the ability to work 100% air-gapped and sync later. Plus **local AI that works offline and understands 3D space** - something impossible with traditional cloud-dependent systems.

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

### How Publishing to Old Gateways Works

Here's the critical insight: **You can still publish TO their platforms, you just don't START there.**

**The Flow:**
1. **Your system is the source** - Video lives on your server (Peertube, R2, your VPS)
2. **You generate the wellknown URI** - `wellknown://video/abc123` points to YOUR gateway
3. **Your gateway decides routing** - Send iOS users to Apple, Android to YouTube, web to your player
4. **You also upload TO YouTube/Twitch** - Use their APIs to publish copies for discovery
5. **Your wellknown links are everywhere** - Social media, email, your website, QR codes
6. **Users come to YOUR gateway first** - You capture analytics, offer your experience, then redirect if needed

**Example: Video Workflow**
```bash
# Upload to YOUR server
curl -X POST yourserver.com/api/videos -F file=@myvideo.mp4
# Returns: wellknown://video/abc123

# Your system auto-publishes to YouTube via API
youtube-upload --title "My Video" --url "wellknown://video/abc123" myvideo.mp4

# Share the wellknown link everywhere
# Users hit YOUR gateway → you track → redirect to best platform
```

### Why This Is Asymmetric Power

This architecture gives you **leverage**:

| Capability | Traditional (Captive) | Wellknown (Sovereign) |
|------------|----------------------|----------------------|
| **Move platforms** | Hard/impossible | Easy - just update routing |
| **Multi-platform** | Manual cross-posting | Automatic via your gateway |
| **Analytics** | Their data, their rules | Your data, complete picture |
| **Monetization** | Their ads, their cut | Your choice, your revenue |
| **Censorship risk** | Total (they own you) | Partial (they're just mirrors) |
| **API changes** | Break your integration | You adapt your gateway, users unaffected |

**The key advantage**: You can publish TO Google/Apple/YouTube, but they can't force you to STAY.

- If YouTube changes their terms → remove them from your routing, users still work
- If you find a better platform → add it to your gateway, users still work
- If you want to go fully independent → disable redirects, users still work

Your wellknown URIs are **portable**. Their platform URIs are **prisons**.

### Real-World Benefits

**For individuals:**
- Post to YouTube for reach, but own your subscriber relationships
- Share calendar links that work everywhere, data lives on your CalDAV server
- Email from your domain, fallback to Gmail UX when convenient

**For businesses:**
- Brand owns the customer relationship, not the platform
- Can switch CDNs/platforms without breaking user links
- Multi-platform presence without multi-platform lock-in

**For communities:**
- Self-hosted Mastodon/Peertube, but discoverable via Big Tech mirrors
- Exit strategy built-in from day one
- Network effects without platform dependency

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

**For Developers**: See [.dev/](.dev/) for internal documentation, deployment guides, and AI agent configuration.
