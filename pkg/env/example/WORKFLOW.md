# Simplified Workflow Guide

**3 Commands. 3 Phases. Clear separation of USER ACTIONS vs AUTOMATION.**

This guide shows you the streamlined workflow using the new automation commands:
- `sync-registry`
- `sync-environments`
- `finalize`

---

## Quick Start (First Time Setup)

```bash
# 1. Edit your registry (USER ACTION)
vim registry.go

# 2. Sync from registry (AUTOMATION)
go run . sync-registry

# 3. Fill in secrets (USER ACTION)
vim .env.secrets.local
vim .env.secrets.production

# 4. Sync environments (AUTOMATION)
go run . sync-environments

# 5. Finalize for git (AUTOMATION)
go run . finalize

# 6. Commit and deploy
git commit -m "feat: initial environment setup"
git push
```

**That's it!** 6 steps instead of 16+.

---

## The 3-Phase Workflow

```
┌─────────────────────────────────────────┐
│ Phase 1: USER EDITS REGISTRY            │
│ - Edit registry.go (add/remove vars)    │
│ - Run: go run . sync-registry           │
│   (syncs configs + shows what to edit)  │
└─────────────────────────────────────────┘
                  ↓
┌─────────────────────────────────────────┐
│ Phase 2: USER EDITS SECRETS             │
│ - Edit .env.secrets.local               │
│ - Edit .env.secrets.production          │
│ - Run: go run . sync-environments       │
│   (merges secrets, validates)           │
└─────────────────────────────────────────┘
                  ↓
┌─────────────────────────────────────────┐
│ Phase 3: AUTOMATION                      │
│ - Run: go run . finalize                │
│   (encrypt, git add, show commit msg)   │
└─────────────────────────────────────────┘
```

---

## Command Details

### 1. `sync-registry` - After Editing Registry

**What it does:**
1. Syncs deployment configs (Dockerfile, fly.toml, docker-compose.yml)
2. Updates `.env.local` template
3. Updates `.env.production` template
4. Creates `.env.secrets.local` if missing
5. Creates `.env.secrets.production` if missing
6. Shows summary of changes

**When to use:** Every time you edit `registry.go`

**Output:**
```bash
$ go run . sync-registry

🔄 Syncing from registry...

📝 Step 1/5: Syncing deployment configs
   ✅ Synced Dockerfile
   ✅ Synced fly.toml
   ✅ Synced docker-compose.yml

📝 Step 2/5: Updating .env.local template
   ✅ Updated .env.local

📝 Step 3/5: Updating .env.production template
   ✅ Updated .env.production

📝 Step 4/5: Checking secrets templates
   ✅ Created .env.secrets.local
   ✅ Created .env.secrets.production

📝 Step 5/5: Summary
   Registry: 6 total variables (3 secrets)

✅ Registry synced successfully!

📝 NEXT: Edit secrets files with real values:
   - .env.secrets.local (for local development)
   - .env.secrets.production (for production)

Then run: go run . sync-environments
```

### 2. `sync-environments` - After Editing Secrets

**What it does:**
1. Merges `.env.secrets.local` → `.env.local`
2. Merges `.env.secrets.production` → `.env.production`
3. Validates all required variables are set
4. Shows what changed

**When to use:** After editing `.env.secrets.local` or `.env.secrets.production`

**Output:**
```bash
$ go run . sync-environments

🔄 Syncing environments from secrets...

📝 Step 1/3: Syncing local environment
   ✅ Merged 3 secrets → .env.local

📝 Step 2/3: Syncing production environment
   ✅ Merged 3 secrets → .env.production

📝 Step 3/3: Validating environments
   ✅ All required variables set

✅ Environments synced successfully!

📝 NEXT: Encrypt and commit:
   go run . finalize
```

### 3. `finalize` - Prepare for Git

**What it does:**
1. Checks for age encryption key (generates if missing)
2. Encrypts all environment files → `.age` versions
3. Runs `git add *.age`
4. Shows commit message suggestion

**When to use:** Before committing changes to git

**Output:**
```bash
$ go run . finalize

🔒 Finalizing for git commit...

📝 Step 1/3: Checking encryption key
   ✅ Found key at .age/key.txt

📝 Step 2/3: Encrypting environment files
   ✅ .env.local → .env.local.age
   ✅ .env.production → .env.production.age
   ✅ .env.secrets.local → .env.secrets.local.age
   ✅ .env.secrets.production → .env.secrets.production.age

📝 Step 3/3: Preparing git commit
   ✅ Added 4 files to git

✅ Finalized successfully!

📝 NEXT: Commit and push:
   git commit -m "chore: update encrypted environments"
   git push

⚠️  REMINDER:
   - .age files are SAFE to commit
   - NEVER commit plaintext .env files
   - NEVER commit .age/key.txt
```

---

## Daily Workflow Examples

### Adding a New Variable

```bash
# 1. Edit registry (USER ACTION)
vim registry.go  # Add: DATABASE_POOL_SIZE

# 2. Sync from registry
go run . sync-registry
# Output: "Added DATABASE_POOL_SIZE to templates"
#         "Update .env.secrets.local and .env.secrets.production"

# 3. Fill in values (USER ACTION)
vim .env.secrets.local       # Add: DATABASE_POOL_SIZE=10
vim .env.secrets.production  # Add: DATABASE_POOL_SIZE=100

# 4. Sync environments
go run . sync-environments
# Output: "Merged DATABASE_POOL_SIZE"
#         "local: 10 → .env.local"
#         "prod: 100 → .env.production"

# 5. Finalize
go run . finalize

# 6. Commit
git commit -m "feat: add database pool size configuration"
git push
```

### Updating Secret Values

```bash
# 1. Edit secrets directly (USER ACTION)
vim .env.secrets.production  # Change: STRIPE_API_KEY=sk_live_new_key

# 2. Sync environments
go run . sync-environments

# 3. Finalize
go run . finalize

# 4. Commit
git commit -m "chore: rotate Stripe API key"
git push
```

### After Pulling from Git

```bash
git pull

# Decrypt environments
go run . age-decrypt

# Verify everything works
go run . validate
```

---

## Comparison: Old vs New Workflow

### OLD WAY (16 commands)

```bash
# Initial setup
go run . setup
go run . generate-secrets > .env.secrets.local
# Edit .env.secrets.local
go run . sync-secrets
go run . dockerfile-sync
go run . fly-sync
go run . compose-sync
go run . setup-prod
go run . generate-secrets > .env.secrets.production
# Edit .env.secrets.production
go run . sync-secrets-prod
go run . age-keygen
go run . age-encrypt
git add *.age
git commit -m "..."
git push
```

### NEW WAY (6 steps)

```bash
vim registry.go                      # 1. USER ACTION
go run . sync-registry               # 2. AUTOMATION
vim .env.secrets.*                   # 3. USER ACTION
go run . sync-environments           # 4. AUTOMATION
go run . finalize                    # 5. AUTOMATION
git commit -m "..." && git push      # 6. USER ACTION
```

**Result:** 16 commands → 6 steps (62% reduction)

---

## Benefits

### 1. Clear USER vs SYSTEM Actions

**USER ACTIONS** are always:
- Edit registry.go
- Edit .env.secrets.* files
- Git commit/push

**SYSTEM ACTIONS** are always:
- sync-registry
- sync-environments
- finalize

No confusion about when to do what!

### 2. Smart Feedback

Each command tells you:
- What it did
- What changed
- What to do next

### 3. Idempotent & Safe

- Safe to run multiple times
- Won't overwrite existing secrets files
- Shows diffs before making changes
- Validates before proceeding

### 4. Backwards Compatible

All original commands still work:
- `setup`, `setup-prod`
- `generate-*`
- `sync-secrets`, `sync-secrets-prod`
- `age-*`
- `fly-*`

Use the new workflow commands for speed, or use individual commands for fine-grained control.

---

## Advanced Scenarios

### Migrating Existing Setup

If you're already using the old workflow:

```bash
# You already have .env.secrets.local and .env.secrets.production
# Just start using the new commands:

go run . sync-environments  # Syncs existing secrets
go run . finalize           # Encrypts and prepares for git
```

### Local-Only Development

If you only work locally:

```bash
vim registry.go
go run . sync-registry
vim .env.secrets.local  # Only edit local
go run . sync-environments
# Skip finalize if you don't need encryption
```

### Production-Only Changes

If you only need to update production:

```bash
vim .env.secrets.production  # Just edit production
go run . sync-environments   # Syncs both, but only production changed
go run . finalize
```

---

## Troubleshooting

### "No secrets file found"

**Problem:** Ran `sync-environments` before creating secrets files.

**Solution:**
```bash
go run . sync-registry  # This creates the templates
vim .env.secrets.local
vim .env.secrets.production
go run . sync-environments
```

### "Missing required variables"

**Problem:** Some required variables don't have values.

**Solution:**
```bash
# Check which variables are missing
go run . validate

# Fill them in
vim .env.secrets.local       # or .env.secrets.production
go run . sync-environments
```

### "No age key found"

**Problem:** Trying to finalize without an encryption key.

**Solution:**
```bash
# finalize will offer to create one
go run . finalize
# Answer "y" when prompted

# Or create manually
go run . age-keygen
go run . finalize
```

---

## File Reference

### What Gets Generated

```
# After sync-registry:
.env.local              # ❌ DO NOT COMMIT (will have secrets after sync-environments)
.env.production         # ❌ DO NOT COMMIT (will have secrets after sync-environments)
.env.secrets.local      # ❌ DO NOT COMMIT (plaintext secrets)
.env.secrets.production # ❌ DO NOT COMMIT (plaintext secrets)

# After finalize:
.env.local.age          # ✅ SAFE TO COMMIT (encrypted)
.env.production.age     # ✅ SAFE TO COMMIT (encrypted)
.env.secrets.local.age  # ✅ SAFE TO COMMIT (encrypted)
.env.secrets.production.age  # ✅ SAFE TO COMMIT (encrypted)
```

### What to Commit

```bash
# ALWAYS commit these:
git add registry.go *.age
git commit -m "feat: update environment configuration"

# NEVER commit these:
# .env.local
# .env.production
# .env.secrets.local
# .env.secrets.production
# .age/key.txt
```

---

## Next Steps

1. **Try it:** Run through the Quick Start above
2. **Read:** Check [README.md](README.md) for full command reference
3. **Deploy:** Use `fly-*` commands for Fly.io deployment
4. **Customize:** Edit `workflow.go` if you need custom workflows

**Questions?** The `finalize` command provides safety reminders and commit guidance.
