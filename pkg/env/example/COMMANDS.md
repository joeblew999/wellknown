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
# 1. Essential - Setup (creates .env.local template - DO NOT COMMIT YET)
go run . clean              # Remove old files (optional)
go run . setup              # Registry → .env.local

# 2. Template Generation - Create LOCAL secrets file
go run . generate-secrets > .env.secrets.local  # Registry → .env.secrets.local
# Edit .env.secrets.local with LOCAL values:
#   DATABASE_URL=postgres://localhost:5432/myapp_dev
#   STRIPE_API_KEY=sk_test_xxxxx

go run . sync-secrets       # Auto-uses .env.secrets.local → .env.local
go run . validate           # Check all required vars

# 3. Age Encryption - Secure for git
go run . age-keygen         # Generate key (one-time) - NEVER COMMIT .age/key.txt!
go run . age-encrypt        # Encrypts all: .env.local, .env.secrets.local → *.age

# ✅ GIT CHECKPOINT: Encrypted files are safe
git add .env.local.age .env.secrets.local.age
git commit -m "chore: update encrypted local environment"
git push

# 4. File Sync - Update deployment configs
go run . dockerfile-sync    # Registry → Dockerfile
go run . fly-sync           # Registry → fly.toml [env]
go run . compose-sync       # Registry → docker-compose.yml

# ✅ GIT CHECKPOINT: Config files are safe (no secrets, just structure)
git add Dockerfile fly.toml docker-compose.yml
git commit -m "sync: update deployment configs from registry"
git push
```

### Production Deployment (Fly.io)

```bash
# 1. Essential - Setup Production (creates .env.production template - DO NOT COMMIT YET)
go run . setup-prod         # Registry → .env.production
go run . generate-secrets > .env.secrets.production  # Registry → .env.secrets.production
# Edit .env.secrets.production with PRODUCTION values:
#   DATABASE_URL=postgres://production.db.internal:5432/myapp_prod
#   STRIPE_API_KEY=sk_live_xxxxx

go run . sync-secrets-prod  # Auto-uses .env.secrets.production → .env.production
go run . validate           # Check all required vars

# 2. Age Encryption - Secure production secrets
go run . age-encrypt        # Encrypts all: .env.production, .env.secrets.production → *.age

# ✅ GIT CHECKPOINT: Encrypted production files are safe
git add .env.production.age .env.secrets.production.age
git commit -m "chore: update encrypted production environment"
git push

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
vim registry.go

# 2. Sync deployment configs (safe - no secrets yet)
go run . dockerfile-sync    # Registry → Dockerfile
go run . fly-sync           # Registry → fly.toml
go run . compose-sync       # Registry → docker-compose.yml

# ✅ GIT CHECKPOINT 1: Registry and config files (no secrets)
git add registry.go Dockerfile fly.toml docker-compose.yml
git commit -m "feat: update environment registry"
git push

# 3. Update local environment
go run . setup                               # Registry → .env.local
go run . generate-secrets > .env.secrets.local  # Update local secrets template
# Edit .env.secrets.local with local values
go run . sync-secrets                        # .env.secrets.local → .env.local

# 4. Update production environment
go run . setup-prod                                   # Registry → .env.production
go run . generate-secrets > .env.secrets.production  # Update production secrets template
# Edit .env.secrets.production with production values
go run . sync-secrets-prod                           # .env.secrets.production → .env.production

# 5. Validate and encrypt
go run . validate           # Check all vars
go run . age-encrypt        # Encrypt all .env files → *.age

# ✅ GIT CHECKPOINT 2: Encrypted environments (safe)
git add *.age
git commit -m "chore: update encrypted environments"
git push

# 6. Deploy to production
go run . fly-secrets-import # Sync production secrets
go run . fly-deploy         # Deploy
go run . fly-status         # Verify
```

## CI/CD Workflow (GitHub Actions)

```yaml
# .github/workflows/deploy.yml
name: Deploy to Production

on:
  push:
    branches: [main]

jobs:
  deploy:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3

      # 1. Install age for decryption
      - name: Install age
        run: |
          wget https://github.com/FiloSottile/age/releases/download/v1.1.1/age-v1.1.1-linux-amd64.tar.gz
          tar xzf age-v1.1.1-linux-amd64.tar.gz
          sudo mv age/age /usr/local/bin/

      # 2. Restore age key from GitHub Secret
      - name: Setup age key
        run: |
          mkdir -p .age
          echo "${{ secrets.AGE_KEY }}" > .age/key.txt
          chmod 600 .age/key.txt

      # 3. Decrypt production environment
      - name: Decrypt production secrets
        run: |
          cd pkg/env/example
          go run . age-decrypt  # .env.production.age → .env.production

      # 4. Deploy to Fly.io
      - name: Deploy to Fly.io
        run: |
          cd pkg/env/example
          go run . fly-secrets-import
          go run . fly-deploy
        env:
          FLY_API_TOKEN: ${{ secrets.FLY_API_TOKEN }}
```

**Setting up GitHub Secrets:**

1. Copy your age key: `cat .age/key.txt`
2. Go to: GitHub repo → Settings → Secrets and variables → Actions
3. Add secret: `AGE_KEY` = [paste content from .age/key.txt]
4. Add secret: `FLY_API_TOKEN` = [your Fly.io API token]

**What gets committed to git:**
- ✅ `.env.production.age` (encrypted, safe)
- ✅ `.github/workflows/deploy.yml` (workflow, safe)
- ❌ `.age/key.txt` (NEVER commit - stored in GitHub Secrets)
- ❌ `.env.production` (NEVER commit - decrypted at runtime in CI)

## File Structure

```
# Committed to git (safe)
registry.go              # ✅ Single source of truth
helpers.go               # ✅ Template generators
main.go                  # ✅ CLI commands
Dockerfile               # ✅ With ENV docs (no secrets)
fly.toml                 # ✅ With [env] section (no secrets)
docker-compose.yml       # ✅ With environment (no secrets)
.env.local.age           # ✅ Encrypted local (SAFE to commit)
.env.production.age      # ✅ Encrypted production (SAFE to commit)
.env.secrets.local.age   # ✅ Encrypted local secrets (SAFE to commit)
.env.secrets.production.age  # ✅ Encrypted production secrets (SAFE to commit)

# Gitignored (NEVER commit)
.env.local               # ❌ Real local credentials
.env.production          # ❌ Real production credentials
.env.secrets.local       # ❌ Real local secrets
.env.secrets.production  # ❌ Real production secrets
.age/key.txt             # ❌ Encryption key (store in password manager!)
```

## RULE: Registry → Generate → Fill → Encrypt → Deploy

**Forward engineering enforced:**
- Registry defines all variables (local + production)
- Templates generated from registry
- User fills environment-specific values
- Encrypt for git safety
- Deploy to target environment
