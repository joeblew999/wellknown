# 🪶 Fly.io Multi-App + Secrets Setup Guide

This guide explains how to deploy **Stalwart + Garage + Meilisearch + PocketBase + NATS** on **Fly.io**
with synchronized environment variables and secrets.

---

## 🧩 Architecture Overview

Each service runs as its own Fly app on the same **private 6PN network**:

| App | Purpose | Ports | Notes |
|-----|----------|--------|-------|
| `stalwart-mail` | Mail, CalDAV, CardDAV server | 8080 (HTTPS) | Connects to Garage + Meili |
| `stalwart-garage` | S3-compatible blob store | 3900 | Distributed, persistent storage |
| `stalwart-meili` | Meilisearch full-text index | 7700 | Indexing and search |
| `pocketbase` | Auth + control plane | 8090 | Manages users & tokens |
| `nats-cluster` | Real-time event bus | 4222 | Reactive updates for GUI |

All share credentials and configuration through **Fly Secrets** and a common `.env.shared` file.

---

## ⚙️ Example fly.toml for stalwart-mail

```toml
app = "stalwart-mail"
primary_region = "sin"

[build]
image = "stalwartlabs/mail:latest"

[env]
DOMAIN = "wellknown.dev"
PRIMARY_REGION = "sin"
S3_REGION = "local"

[[services]]
internal_port = 8080
protocol = "tcp"
[[services.ports]]
  handlers = ["tls"]
  port = 443
[[services.ports]]
  handlers = ["http"]
  port = 80
```

Secrets will be injected automatically at runtime.

---

## 🔐 Shared Environment Variables (`.env.shared`)

Create this in your repo root:

```bash
S3_ENDPOINT=http://stalwart-garage.internal:3900
S3_REGION=local
S3_BUCKET=stalwart-blobs
S3_ACCESS_KEY=garage-access
S3_SECRET_KEY=garage-secret

MEILI_URL=http://stalwart-meili.internal:7700
MEILI_MASTER_KEY=meili_master

NATS_URL=nats://nats.internal:4222
DOMAIN=wellknown.dev
PRIMARY_REGION=sin
```

---

## 📦 Makefile for syncing secrets

Add this target to your main `Makefile` or `infra/Makefile`:

```makefile
sync-env:
	for app in stalwart-mail stalwart-garage stalwart-meili pocketbase nats-cluster; do \\
	  echo "🔄 Syncing secrets to $$app..."; \\
	  fly secrets import --app $$app < .env.shared; \\
	done
```

Then run:

```bash
make sync-env
```

Fly will push all environment variables to each app as encrypted secrets.

---

## 🧰 Checking that secrets are synced

```bash
fly ssh console -a stalwart-mail
env | grep S3
```

All apps should show the same shared variables.

---

## 🧠 Tips

- `[env]` values in `fly.toml` are **non-secret defaults** (safe for git).  
- `fly secrets` are **encrypted per app** and override `[env]`.  
- `.env.shared` is your **single source of truth** — version it privately (not public).  
- You can automate syncing from CI/CD with a GitHub Action.

---

## 🚀 Example GitHub Action Workflow

```yaml
name: Sync Fly Secrets

on:
  workflow_dispatch:
  push:
    paths:
      - ".env.shared"

jobs:
  sync:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - name: Fly CLI
        uses: superfly/flyctl-actions/setup-flyctl@master
      - name: Sync env to all apps
        run: make sync-env
        env:
          FLY_API_TOKEN: ${{ secrets.FLY_API_TOKEN }}
```

---

## ✅ Summary

- Use one `.env.shared` for all apps.
- Use `fly secrets import` to keep them synchronized.
- Keep sensitive data out of git.
- Store the master copy securely (PocketBase, Vault, or encrypted file).
- Each app can reference `.internal` hostnames for private Fly network access.

---
