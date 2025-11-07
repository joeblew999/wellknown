# pkg/env - Registry-Driven Environment Management

**NOTE: Keep this README minimal - JUST the flow as text. No diagrams, no examples, no API docs.**

## Tab Completion (Optional - No Shell Scripts!)

**Build binary to .bin folder (no file completion conflicts):**
```bash
go build -o .bin/env-tool .
```

**Enable completion (Bash):**
```bash
PROG=env-tool source <(.bin/env-tool --init-completion bash)
```

**Enable completion (Zsh):**
```bash
PROG=env-tool source <(.bin/env-tool --init-completion zsh)
```

**Enable completion (Fish):**
```bash
.bin/env-tool --init-completion fish | source
```

**Usage:**
```bash
.bin/env-tool fly-<TAB>        # Tab completion works!
.bin/env-tool --dir /path validate  # Run from anywhere
```

**Note:** Binary in `.bin/` folder avoids shell file completion conflicts with `fly.toml` and `flyio.go`

## THE FLOW (What a developer does)

### First-Time Setup

**0. Clean (optional)**
```bash
go run . clean
```
Remove old generated files.

**1. Edit registry**
```go
// registry.go - SINGLE SOURCE OF TRUTH
var AppEnvVars = []env.EnvVar{
    {Name: "SERVER_PORT", Default: "8080", Required: true, Group: "Server"},
}
```
Registry defines all variables. This is the ONLY place to add/modify vars.

**2. Setup Age encryption (one-time)**
```bash
go run . age-keygen  # DO NOT commit .age/key.txt
```
Generate encryption key for git-safe secrets.

**3. Generate templates from registry**
```bash
go run . setup            # Registry → .env.local
go run . generate-secrets # Registry → .env.secrets (secrets only)
```
Both files generated FROM registry (forward).

**4. Fill in secrets**
```bash
# Edit .env.secrets with real values
DATABASE_URL=postgres://localhost/mydb
STRIPE_API_KEY=sk_test_xxxxx
```
.env.secrets was generated from registry - now fill real values.

**5. Merge secrets**
```bash
go run . sync-secrets  # .env.secrets → .env.local
```
Merge filled secrets into .env.local.

**6. Validate**
```bash
go run . validate
```
Check all required vars are set.

**7. Encrypt**
```bash
go run . age-encrypt  # .env.local → .env.local.age
```
Encrypt .env.local so it's safe to commit to git.

**8. Sync deployment configs (optional)**
```bash
go run . dockerfile-sync  # Registry → Dockerfile env docs
go run . fly-sync         # Registry → fly.toml [env]
go run . compose-sync     # Registry → docker-compose.yml
```
Keep config files in sync with registry.

**9. Deploy to Fly.io (optional, no Makefile needed)**
```bash
go run . fly-install        # Install flyctl (one-time)
go run . fly-auth           # Login to Fly.io
go run . fly-launch         # Create app (reads fly.toml)
go run . fly-volume         # Create volume
go run . fly-secrets-import # Registry → secrets sync
go run . fly-deploy         # Deploy
go run . fly-status         # Check status
go run . fly-logs           # View logs
```
Or export for other platforms:
```bash
go run . export k8s      # Kubernetes
go run . export docker   # Docker
```

### Returning Developer (pulling from git)

**1. Decrypt**
```bash
go run . age-decrypt  # .env.local.age → .env.local
```

**2. Validate**
```bash
go run . validate
```

**3. Export**
```bash
go run . export k8s
```

### Daily Workflow (after registry changes)

```bash
# Edit registry.go
go run . setup            # Registry → .env.local
go run . generate-secrets # Registry → .env.secrets
# Edit .env.secrets with new values
go run . sync-secrets     # Merge into .env.local
go run . dockerfile-sync  # Registry → Dockerfile
go run . fly-sync         # Registry → fly.toml
go run . compose-sync     # Registry → docker-compose.yml
go run . validate         # Check all required vars
go run . age-encrypt      # Encrypt for git
git add .env.local.age
```

**Deploy changes to Fly.io:**
```bash
go run . fly-secrets-import  # Sync new secrets
go run . fly-deploy          # Deploy
go run . fly-status          # Verify
```

## RULE: Registry → Generate → Fill → Merge → Sync → Encrypt → Deploy

**Forward engineering enforced:**
- `.env.secrets` generated from registry
- User fills real values
- `sync-secrets` merges into .env.local
- `dockerfile-sync`, `fly-sync`, `compose-sync` update config files
- ALL files originated from registry (single source of truth)
