# pkg/env - Registry-Driven Environment Management

**NOTE: Keep this README minimal - JUST the flow as text. No diagrams, no examples, no API docs.**

## Quick Start

```bash
go run . --help  # See all commands organized by category
```

## The Registry (Single Source of Truth)

```go
// registry.go - Edit this ONCE, use everywhere
var AppEnvVars = []env.EnvVar{
    {Name: "SERVER_PORT", Default: "8080", Required: true, Group: "Server"},
    {Name: "DATABASE_URL", Secret: true, Required: true, Group: "Database"},
}
```

Registry defines ALL environment variables for ALL environments (local + production).

## Two Environments, Same Registry

### Local Development

```bash
# 1. Essential - Setup
go run . clean              # Remove old files (optional)
go run . setup              # Registry → .env.local

# 2. Template Generation - Create secrets file
go run . generate-secrets   # Registry → .env.secrets (template)
# Edit .env.secrets with LOCAL values:
#   DATABASE_URL=postgres://localhost/mydb
#   STRIPE_API_KEY=sk_test_xxxxx

go run . sync-secrets       # .env.secrets → .env.local
go run . validate           # Check all required vars

# 3. Age Encryption - Secure for git
go run . age-keygen         # Generate key (one-time)
go run . age-encrypt        # .env.local → .env.local.age
git add .env.local.age      # Safe to commit!

# 4. File Sync - Update deployment configs
go run . dockerfile-sync    # Registry → Dockerfile
go run . fly-sync           # Registry → fly.toml [env]
go run . compose-sync       # Registry → docker-compose.yml
```

### Production Deployment (Fly.io)

```bash
# 1. Template Generation - Create production env
go run . generate-prod      # Registry → .env.production (template)
# Edit .env.production with PRODUCTION values:
#   DATABASE_URL=postgres://myapp.internal/production
#   STRIPE_API_KEY=sk_live_xxxxx

# 2. Age Encryption - Secure production secrets
go run . age-encrypt        # Encrypts .env.production → .env.production.age
git add .env.production.age # Safe to commit!

# 3. Fly.io Deployment - Full deployment flow
go run . fly-install        # Install flyctl (one-time)
go run . fly-auth           # Login to Fly.io
go run . fly-launch         # Create app (reads fly.toml)
go run . fly-volume         # Create persistent volume
go run . fly-secrets-import # Import from .env.production
go run . fly-deploy         # Deploy!
go run . fly-status         # Check deployment
go run . fly-logs           # View logs

# 4. Export Formats - Alternative deployment targets
go run . export k8s         # Kubernetes ConfigMap/Secret
go run . export docker      # Docker Compose format
go run . export systemd     # Systemd EnvironmentFile
```

## Returning Developer (Pulling from Git)

```bash
# Decrypt local environment
go run . age-decrypt        # .env.local.age → .env.local
go run . validate

# OR decrypt production
go run . age-decrypt        # .env.production.age → .env.production
```

## Daily Workflow (After Registry Changes)

```bash
# 1. Edit registry.go (add/remove/modify variables)

# 2. Update local environment
go run . setup              # Registry → .env.local
go run . generate-secrets   # Registry → .env.secrets
# Edit .env.secrets with new values
go run . sync-secrets       # Merge into .env.local

# 3. Update production environment
go run . generate-prod      # Registry → .env.production
# Edit .env.production with production values

# 4. Sync deployment configs
go run . dockerfile-sync    # Registry → Dockerfile
go run . fly-sync           # Registry → fly.toml
go run . compose-sync       # Registry → docker-compose.yml

# 5. Validate and encrypt
go run . validate           # Check all vars
go run . age-encrypt        # Encrypt both .env.local and .env.production
git add .env.local.age .env.production.age

# 6. Deploy to production
go run . fly-secrets-import # Sync production secrets
go run . fly-deploy         # Deploy
go run . fly-status         # Verify
```

## File Structure

```
registry.go              # Single source of truth
.env.local               # Local dev values (gitignored)
.env.local.age           # Encrypted local (committed)
.env.production          # Production values (gitignored)
.env.production.age      # Encrypted production (committed)
.env.secrets             # Secrets template (gitignored)
.age/key.txt             # Encryption key (gitignored, DO NOT COMMIT!)
```

## RULE: Registry → Generate → Fill → Encrypt → Deploy

**Forward engineering enforced:**
- Registry defines all variables (local + production)
- Templates generated from registry
- User fills environment-specific values
- Encrypt for git safety
- Deploy to target environment
