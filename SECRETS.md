# Git-Tracked Secrets Management with Age Encryption

This project uses **Age encryption** (built into the Go app) to securely store secrets in a public repository. Your real credentials are encrypted in git and automatically decrypted and synced into environment files.

## Why This Approach?

- **Single Source of Truth**: `.env.secrets.age` contains all real credentials
- **Git-Tracked**: Safely stored in git with Age encryption
- **Auto-Decrypt**: Native Go decryption, no external tools needed
- **Auto-Sync**: One command regenerates `.env.local` or `.env.production`
- **No Manual Copying**: Secrets flow automatically
- **Team-Friendly**: Share encrypted secrets with authorized team members
- **Safe for Public Repos**: Military-grade encryption protects credentials
- **Pure Go**: No external CLI dependencies required

## File Overview

| File | Purpose | Git-Tracked? | Encrypted? |
|------|---------|--------------|------------|
| `.env.secrets.example` | Template showing required secrets | ✅ Yes | ❌ No |
| `.env.secrets` | **Plaintext secrets** (local only) | ❌ No | ❌ No |
| `.env.secrets.age` | **Your encrypted credentials** | ✅ Yes | ✅ Yes (Age) |
| `.env.local` | Generated for local dev | ❌ No | ❌ No |
| `.env.production` | Generated for Fly.io | ❌ No | ❌ No |

## One-Time Setup

### 1. Install Age CLI (for encryption only)

```bash
# macOS
brew install age

# Linux (Debian/Ubuntu)
sudo apt-get install age

# Linux (Arch)
sudo pacman -S age

# Or download from: https://github.com/FiloSottile/age/releases
```

### 2. Generate Your Age Identity

```bash
# Create Age identity (private key)
age-keygen -o ~/.ssh/age

# Example output:
# Public key: age1ql3z7hjy54pw3hyww5ayyfg7zqgvc7w3j2elw8zmrj2kg5sfn9aqmcac8p
```

**IMPORTANT**: Save your public key! You'll need it to encrypt secrets.

### 3. Create Your Secrets File

```bash
# Copy the example template
cp .env.secrets.example .env.secrets

# Edit with your real credentials
vim .env.secrets
```

Example `.env.secrets`:
```bash
# Required secrets
GOOGLE_CLIENT_ID=your-real-client-id.apps.googleusercontent.com
GOOGLE_CLIENT_SECRET=your-real-secret
ANTHROPIC_API_KEY=sk-ant-api03-your-real-key...

# Optional secrets (leave empty if not using)
APPLE_TEAM_ID=
SMTP_HOST=
S3_ACCESS_KEY=
```

### 4. Encrypt Your Secrets

```bash
# Encrypt with your public key (replace with YOUR public key)
age -e -r age1ql3z7hjy54pw3hyww5ayyfg7zqgvc7w3j2elw8zmrj2kg5sfn9aqmcac8p \
    .env.secrets > .env.secrets.age
```

### 5. Commit Your Encrypted Secrets

```bash
# Add to git
git add .env.secrets.age
git commit -m "Add encrypted secrets"
git push
```

The file is now encrypted in git! ✅

## Daily Workflow

### Local Development

```bash
# The app automatically decrypts .env.secrets.age and generates .env.local
make env-sync-secrets

# Run the server
make run
```

### Fly.io Deployment

```bash
# Sync secrets to Fly.io (automatically decrypts and generates .env.production)
make fly-secrets

# Deploy
make fly-deploy
```

## How It Works

### Architecture

```
┌──────────────────┐
│  .env.secrets    │ ← Plaintext (local only, gitignored)
│ (real secrets)   │ ← You edit this
└────────┬─────────┘
         │
         │ age -e (manual encryption)
         ↓
┌──────────────────┐
│ .env.secrets.age │ ← Encrypted (git-tracked)
│ (Age encrypted)  │ ← Single source of truth
└────────┬─────────┘
         │
         │ Go app auto-decrypts (filippo.io/age)
         ↓
┌──────────────────┐
│  .env.secrets    │ ← Decrypted in-memory
│ (in memory)      │
└────────┬─────────┘
         │
         │ make env-sync-secrets
         ↓
┌──────────────────┐
│  .env.local      │ ← Auto-generated
│ • localhost URLs │ ← For local development
│ • HTTPS enabled  │
└──────────────────┘

         │ make env-sync-secrets-production
         ↓
┌───────────────────┐
│ .env.production   │ ← Auto-generated
│ • fly.dev URLs    │ ← For Fly.io deployment
│ • HTTPS disabled  │
└────────┬──────────┘
         │
         │ make fly-secrets
         ↓
    ☁️  Fly.io Secrets
```

### What Happens When You Run `make env-sync-secrets`

1. **Detect** `.env.secrets.age` file
2. **Decrypt** using Age identity from `~/.ssh/age`
3. **Generate** `.env.local` template from `pkg/pb/env.go` registry
4. **Merge** secret values into template
5. **Write** `.env.local` (ready for `make run`)

### What Happens When You Run `make fly-secrets`

1. **Auto-runs** `make env-sync-secrets-production`
2. **Detects** `.env.secrets.age` file
3. **Decrypts** using Age identity
4. **Generates** `.env.production` template with fly.dev URLs
5. **Merges** secrets into template
6. **Pushes** secrets to Fly.io via `flyctl secrets import`

## Adding New Environment Variables

When you add a new environment variable to `pkg/pb/env.go`:

```go
{
    Name:        "NEW_API_KEY",
    Description: "API key for new service",
    Required:    true,
    Secret:      true,
    Group:       "New Service",
},
```

**Then:**

1. Add the value to `.env.secrets`:
   ```bash
   NEW_API_KEY=your-real-api-key
   ```

2. Re-encrypt:
   ```bash
   age -e -r YOUR_PUBLIC_KEY .env.secrets > .env.secrets.age
   ```

3. Commit the updated encrypted file:
   ```bash
   git add .env.secrets.age
   git commit -m "Add NEW_API_KEY"
   ```

4. Regenerate environment files:
   ```bash
   make env-sync-secrets              # Local dev
   make env-sync-secrets-production   # Production
   ```

**That's it!** No manual file editing or copying required.

## Team Collaboration

### Adding a New Team Member

1. **New team member generates their Age key**:
   ```bash
   age-keygen -o ~/.ssh/age
   # Save the public key: age1ql3z7hjy54pw3hyww5ayyfg7zqgvc7w3j2elw8zmrj2kg5sfn9aqmcac8p
   ```

2. **They share their public key** with the team (via Slack, email, etc.)

3. **Repository owner re-encrypts for multiple recipients**:
   ```bash
   age -e \
       -r age1ql3z7hjy54pw3hyww5ayyfg7zqgvc7w3j2elw8zmrj2kg5sfn9aqmcac8p \
       -r age1zlnj4x2pqrs7v3k6w5nlqcm8e6vxn5jwu9yq7dfghj8k9l0mn2sqabcdef \
       .env.secrets > .env.secrets.age

   git add .env.secrets.age
   git commit -m "Add team member to encrypted secrets"
   git push
   ```

4. **New team member clones and uses**:
   ```bash
   git clone git@github.com:joeblew999/wellknown.git
   cd wellknown
   make env-sync-secrets  # Auto-decrypts with their ~/.ssh/age key
   ```

Now they can read the encrypted secrets! ✅

### Managing Multiple Recipients

You can encrypt for multiple people at once:

```bash
age -e \
    -r age1person1... \
    -r age1person2... \
    -r age1person3... \
    .env.secrets > .env.secrets.age
```

### Revoking Access

To revoke access, re-encrypt without that person's public key:

```bash
# Only include active team members
age -e -r age1active1... -r age1active2... .env.secrets > .env.secrets.age
git add .env.secrets.age
git commit -m "Revoke access for former team member"
```

## Security Best Practices

✅ **DO:**
- Keep `.env.secrets.age` git-tracked (encrypted)
- Keep `.env.secrets` local only (gitignored)
- Store Age private key safely (`~/.ssh/age` with chmod 600)
- Rotate secrets regularly
- Use Age for this public repository
- Back up your Age private key securely

❌ **DON'T:**
- Commit `.env.secrets` (plaintext) to git
- Commit `.env.local` or `.env.production` (they're gitignored)
- Share your Age private key
- Put secrets in code or comments
- Lose your Age private key (no way to decrypt!)

## Troubleshooting

### "secrets file not found"

```bash
# Make sure you created it from the example
cp .env.secrets.example .env.secrets

# Edit with your real credentials
vim .env.secrets

# Encrypt it
age -e -r YOUR_PUBLIC_KEY .env.secrets > .env.secrets.age
```

### "no Age identities found"

```bash
# Generate an Age key
age-keygen -o ~/.ssh/age

# Or set AGE_IDENTITY environment variable
export AGE_IDENTITY=/path/to/your/age/key
```

### "failed to decrypt"

```bash
# Verify your Age key has access
age -d -i ~/.ssh/age .env.secrets.age

# If that fails, you may not be in the recipient list
# Ask the repository owner to re-encrypt with your public key
```

### "My secrets are visible in git history!"

```bash
# If you accidentally committed unencrypted secrets:
# 1. Rotate all exposed secrets immediately
# 2. Clean git history (requires force push):
git filter-branch --tree-filter 'rm -f .env.secrets' HEAD
git push --force
# 3. Then set up Age encryption properly and recommit
```

## Fly.io and Docker Integration

### Fly.io Secrets

The `make fly-secrets` command automatically:
1. Decrypts `.env.secrets.age`
2. Generates `.env.production` with fly.dev URLs
3. Pushes secrets to Fly.io

```bash
# One command does everything
make fly-secrets
```

### Docker

For Docker builds, you have two options:

#### Option 1: Decrypt Locally (Recommended)

```bash
# Decrypt locally before building
make env-sync-secrets-production

# Build Docker image
docker build -t myapp .
```

#### Option 2: Build-Time Decryption

Add Age decryption to your Dockerfile:

```dockerfile
# Install Age
RUN apk add --no-cache age

# Copy encrypted secrets and identity
COPY .env.secrets.age /app/
COPY --from=secrets ~/.ssh/age /root/.ssh/age

# Decrypt during build (or at runtime)
RUN age -d -i /root/.ssh/age /app/.env.secrets.age > /app/.env.secrets

# Run the app (it will auto-decrypt)
CMD ["./wellknown-pb", "serve"]
```

## Commands Reference

### Local Development
```bash
make env-sync-secrets     # Decrypt and sync to .env.local
make run                  # Start local server
```

### Production Deployment
```bash
make fly-secrets          # Auto-decrypt and sync to Fly.io
make fly-deploy           # Deploy application
```

### Manual Commands (if needed)
```bash
go run . env sync-secrets                # Generate .env.local
go run . env sync-secrets-production     # Generate .env.production
go run . env list                        # Show all env vars
go run . env validate                    # Check required vars
```

### Age Encryption Commands
```bash
# Generate key
age-keygen -o ~/.ssh/age

# Encrypt file
age -e -r YOUR_PUBLIC_KEY .env.secrets > .env.secrets.age

# Decrypt file (manual)
age -d -i ~/.ssh/age .env.secrets.age > .env.secrets

# Encrypt for multiple recipients
age -e -r KEY1 -r KEY2 -r KEY3 .env.secrets > .env.secrets.age
```

## Benefits Summary

| Before | After |
|--------|-------|
| Secrets in private notes | Secrets in git (Age encrypted) |
| Manual copying to .env files | One command auto-sync |
| Risk of outdated credentials | Always in sync with git |
| Hard to share with team | Easy public-key sharing |
| Multiple sources of truth | Single source in .env.secrets.age |
| Public repo unsafe | Public repo with encryption |
| External tool dependency | Native Go decryption |

## Why Age Instead of git-crypt?

- **Native Go Integration**: Decryption built into the app
- **No External Tools**: Just `filippo.io/age` library
- **Modern Crypto**: X25519, ChaCha20-Poly1305
- **Simple**: Public/private key pairs (like SSH)
- **SSH Compatible**: Can use SSH keys for encryption
- **Explicit**: File extension `.age` makes encryption obvious
- **Fast**: Efficient streaming encryption/decryption
- **Trusted**: Created by Filippo Valsorda (Go security team)

## Need Help?

- Age documentation: https://age-encryption.org/
- Age GitHub: https://github.com/FiloSottile/age
- Fly.io secrets: https://fly.io/docs/reference/secrets/
- Project issues: https://github.com/joeblew999/wellknown/issues

---

**Remember**: Never commit unencrypted secrets! Always verify `.env.secrets.age` exists before pushing.
