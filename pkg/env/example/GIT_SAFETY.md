# Git Safety Guide - When to Commit What

## The Critical Question: What's Safe to Commit?

### ✅ SAFE TO COMMIT (Track in Git)

```bash
registry.go                    # Source of truth - ALWAYS commit
helpers.go                     # Template generators - ALWAYS commit
main.go                        # CLI commands - ALWAYS commit
Dockerfile                     # With synced ENV comments - ALWAYS commit
fly.toml                       # With synced [env] section - ALWAYS commit
docker-compose.yml             # With synced environment - ALWAYS commit

# Encrypted files (*.age) - SAFE because they're encrypted
.env.local.age                 # ✅ Encrypted local env
.env.production.age            # ✅ Encrypted production env
.env.secrets.local.age         # ✅ Encrypted local secrets
.env.secrets.production.age    # ✅ Encrypted production secrets
```

### ❌ NEVER COMMIT (Gitignored)

```bash
# Plaintext environment files - CONTAIN REAL SECRETS
.env.local                     # ❌ Real local credentials
.env.production                # ❌ Real production credentials
.env.secrets.local             # ❌ Real local secrets
.env.secrets.production        # ❌ Real production secrets
.env.secrets                   # ❌ Legacy secrets file

# Age encryption key - MUST STAY PRIVATE
.age/key.txt                   # ❌ NEVER COMMIT THIS!
.age-key.txt                   # ❌ NEVER COMMIT THIS!
```

## Git Commit Checkpoints

### Checkpoint 1: After Registry Changes

```bash
# Edit registry.go
vim registry.go

# Sync deployment configs (these are safe - no secrets)
go run . dockerfile-sync
go run . fly-sync
go run . compose-sync

# ✅ SAFE TO COMMIT - No secrets yet!
git add registry.go Dockerfile fly.toml docker-compose.yml
git commit -m "feat: update environment registry"
git push
```

**Why safe?** These files only contain:
- Variable names and structure
- Default values (non-secrets)
- Comments and documentation
- NO actual credentials

### Checkpoint 2: After Filling Secrets and Encrypting

```bash
# Generate and fill secrets locally
go run . setup
go run . generate-secrets > .env.secrets.local
# Edit .env.secrets.local with REAL LOCAL values
go run . sync-secrets

# Generate and fill production secrets
go run . setup-prod
go run . generate-secrets > .env.secrets.production
# Edit .env.secrets.production with REAL PRODUCTION values
go run . sync-secrets-prod

# Encrypt BEFORE committing
go run . age-keygen          # One-time: generates .age/key.txt
go run . age-encrypt         # Encrypts all 4 files

# ✅ SAFE TO COMMIT - Encrypted files only!
git add .env.local.age .env.production.age
git add .env.secrets.local.age .env.secrets.production.age
git commit -m "chore: update encrypted environments"
git push
```

**Why safe?** The `.age` files are:
- Encrypted with age (filippo.io/age)
- Useless without the `.age/key.txt` file
- The key is gitignored and NEVER committed

## Age Key Management

### The Age Key is Your Master Secret

```bash
.age/key.txt               # THIS IS YOUR MASTER PASSWORD
```

**CRITICAL RULES:**
1. ❌ NEVER commit `.age/key.txt` to git
2. ❌ NEVER share it in Slack/email/tickets
3. ✅ DO store it in a password manager (1Password, LastPass, etc.)
4. ✅ DO share it securely with team members (encrypted channels)

### For Individual Developers

**Store your age key in a password manager:**
```bash
# After running: go run . age-keygen
cat .age/key.txt
# Copy this to your password manager as "Project XYZ - Age Key"
```

**On a new machine:**
```bash
# Retrieve from password manager and save
mkdir -p .age
echo "AGE-SECRET-KEY-1..." > .age/key.txt
chmod 600 .age/key.txt

# Now you can decrypt
go run . age-decrypt
```

### For Teams - Option 1: Shared Age Key

**One key for the whole team (simpler but less secure):**

1. Generate key once: `go run . age-keygen`
2. Share `.age/key.txt` content via secure channel (Signal, 1Password shared vault)
3. Each team member saves it to their local `.age/key.txt`
4. Everyone can encrypt/decrypt with the same key

**Pros:** Simple, one key to manage
**Cons:** If one person leaks it, everyone's secrets are exposed

### For Teams - Option 2: Multiple Recipients (More Secure)

**Encrypt for multiple team members (more secure, more complex):**

This requires modifying `cmdAgeEncrypt()` to support multiple recipients:

```go
// Each team member generates their own keypair
alice: age-keygen -o alice-key.txt  # age1alice...
bob:   age-keygen -o bob-key.txt    # age1bob...

// Collect public keys (safe to share)
alice-public.txt: age1alice123...
bob-public.txt:   age1bob456...

// Encrypt for multiple recipients
age -r age1alice123... -r age1bob456... -o .env.local.age .env.local
```

**Pros:** Each person has their own key, revocation possible
**Cons:** More complex to set up

**For this example, we're using Option 1 (shared key) for simplicity.**

## CI/CD Workflow

### GitHub Actions Example

**The age key becomes a GitHub Secret:**

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

      # Critical: Install age CLI
      - name: Install age
        run: |
          wget https://github.com/FiloSottile/age/releases/download/v1.1.1/age-v1.1.1-linux-amd64.tar.gz
          tar xzf age-v1.1.1-linux-amd64.tar.gz
          sudo mv age/age /usr/local/bin/

      # Critical: Restore age key from GitHub Secret
      - name: Setup age key
        run: |
          mkdir -p .age
          echo "${{ secrets.AGE_KEY }}" > .age/key.txt
          chmod 600 .age/key.txt

      # Now decrypt works
      - name: Decrypt production environment
        run: |
          cd pkg/env/example
          go run . age-decrypt
          # This creates .env.production from .env.production.age

      # Deploy with real secrets
      - name: Deploy to Fly.io
        run: |
          cd pkg/env/example
          go run . fly-secrets-import  # Uses .env.production
          go run . fly-deploy
        env:
          FLY_API_TOKEN: ${{ secrets.FLY_API_TOKEN }}
```

**Setting up GitHub Secrets:**

```bash
# 1. Copy your age key
cat .age/key.txt
# Output: AGE-SECRET-KEY-1QYQSZQGPQYQSZQGP...

# 2. Go to GitHub: Settings → Secrets and variables → Actions → New repository secret
# Name: AGE_KEY
# Value: [paste the AGE-SECRET-KEY-1... content]
```

**Why this works:**
- `.env.production.age` is committed (encrypted, safe)
- GitHub Secret `AGE_KEY` stores the decryption key (secure)
- CI decrypts at runtime to get real credentials
- Fly.io deployment gets real production secrets
- No plaintext secrets ever in git

### Alternative: Don't Commit Production Secrets at All

**Some teams prefer to NEVER commit production secrets to git, even encrypted:**

```yaml
# .github/workflows/deploy.yml
jobs:
  deploy:
    steps:
      # Don't decrypt from git - build from GitHub Secrets directly
      - name: Create production env from secrets
        run: |
          cd pkg/env/example
          go run . setup-prod  # Creates template

          # Fill secrets from GitHub Secrets (not from git)
          cat << EOF > .env.production
          DATABASE_URL=${{ secrets.PROD_DATABASE_URL }}
          STRIPE_API_KEY=${{ secrets.PROD_STRIPE_API_KEY }}
          SENDGRID_API_KEY=${{ secrets.PROD_SENDGRID_API_KEY }}
          SERVER_PORT=8080
          LOG_LEVEL=warn
          EOF

      - name: Deploy to Fly.io
        run: |
          cd pkg/env/example
          go run . fly-secrets-import
          go run . fly-deploy
```

**Pros:**
- Production secrets NEVER in git (even encrypted)
- Each secret managed individually in GitHub UI

**Cons:**
- Have to manually sync GitHub Secrets when registry changes
- Can't version control production config changes
- More GitHub Secrets to manage

**Recommendation:** Use encrypted `.age` files for small teams; use GitHub Secrets directly for larger orgs with compliance requirements.

## Summary: Commit Safety Checklist

**Before every commit, verify:**

```bash
# ✅ What WILL be committed (encrypted files are safe)
git status --porcelain | grep -E "\.age$"
.env.local.age
.env.production.age
.env.secrets.local.age
.env.secrets.production.age

# ❌ What MUST NOT be committed (plaintext secrets)
git status --porcelain | grep -E "^\?\?" | grep -E "\.(env|secrets)" | grep -v "\.age$"
# Should output NOTHING! If you see .env.local or .env.secrets, DON'T COMMIT!

# ❌ Verify age key is gitignored
git check-ignore .age/key.txt
.age/key.txt  # Should show it's ignored

# If git check-ignore shows nothing, STOP! Your key is not gitignored!
```

**Safe commit flow:**

```bash
# 1. Make changes to registry/helpers/config files
git add registry.go helpers.go Dockerfile fly.toml

# 2. Fill and encrypt secrets
go run . age-encrypt

# 3. Add encrypted files
git add *.age

# 4. Verify no plaintext secrets staged
git diff --cached --name-only | grep -E "\.env\." | grep -v "\.age$"
# Should be EMPTY! If not, unstage them: git reset HEAD .env.local

# 5. Safe to commit
git commit -m "feat: update environment configuration"
git push
```

## Recovery Scenarios

### Lost Age Key

**If you lose `.age/key.txt`, you CANNOT decrypt your `.age` files.**

**Recovery options:**
1. Restore from password manager (why you should store it there)
2. Restore from team member's copy (if using shared key)
3. Regenerate all secrets from scratch (painful but works)

### Accidentally Committed Plaintext Secrets

**If you committed `.env.local` or `.env.production` with real secrets:**

```bash
# ⚠️ DANGER: This rewrites git history!
git filter-repo --path .env.local --invert-paths
git filter-repo --path .env.production --invert-paths
git filter-repo --path .env.secrets.local --invert-paths
git filter-repo --path .env.secrets.production --invert-paths

# Force push (coordinate with team first!)
git push --force

# THEN: Rotate ALL compromised secrets immediately!
# - Database passwords
# - API keys
# - Stripe keys
# - Everything in those files
```

**Prevention is better:** Use pre-commit hooks to block plaintext secrets.

## Pre-Commit Hook (Recommended)

Create `.git/hooks/pre-commit`:

```bash
#!/bin/bash
# Prevent committing plaintext secrets

if git diff --cached --name-only | grep -E "^\.env\.(local|production|secrets)$|^\.env\.secrets\.(local|production)$" | grep -v "\.age$"; then
  echo "❌ ERROR: Attempting to commit plaintext secrets!"
  echo "   Plaintext .env files must NOT be committed"
  echo "   Did you mean to add the .age files instead?"
  echo ""
  echo "   Run: go run . age-encrypt"
  echo "   Then: git add *.age"
  exit 1
fi

if git diff --cached --name-only | grep -E "\.age/key\.txt|\.age-key\.txt"; then
  echo "❌ ERROR: Attempting to commit age encryption key!"
  echo "   The .age/key.txt file must NEVER be committed"
  echo "   This would expose all your encrypted secrets!"
  exit 1
fi

exit 0
```

```bash
chmod +x .git/hooks/pre-commit
```

**Now git will BLOCK you from committing secrets!**
