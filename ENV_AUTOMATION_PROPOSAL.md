# Environment Variable Automation Proposal

**Date**: 2025-11-06
**Status**: 🎯 Proposal - Ready for Implementation
**Goal**: Leverage Go env registry to eliminate manual sync between code, Dockerfile, Makefile, and env files

---

## 🎉 Current State: Already Excellent!

Your [pkg/pb/env.go](pkg/pb/env.go) is a **phenomenal** single source of truth:

```go
type EnvVar struct {
    Name        string  // Environment variable name
    Description string  // Human-readable description
    Required    bool    // Is this variable required?
    Secret      bool    // Should this be in Fly.io secrets (vs fly.toml)?
    Default     string  // Default value
    Group       string  // Logical grouping
}
```

**Existing Capabilities**:
- ✅ Single source of truth (`AllEnvVars` registry)
- ✅ Auto-generates `.env.example` via `GenerateEnvExample()`
- ✅ Exports secrets via `ExportSecretsFormat()` for `flyctl secrets import`
- ✅ Validates required vars via `ValidateEnv()`
- ✅ Lists all vars via `ListEnvVars()`
- ✅ Type-safe getters (`GetString()`, `GetInt()`, `GetBool()`)

---

## 🚀 Proposed Improvements

### Problem: Manual Synchronization

When you add/remove/change env vars, you must manually update:
1. ❌ `pkg/pb/env.go` (registry)
2. ❌ `Dockerfile` (lines 28-66 documentation)
3. ❌ `.env.local` (local development)
4. ❌ `.env.production` (production deployment)
5. ❌ `fly.toml` ([env] section for non-secrets)

**This is error-prone and violates DRY principles.**

---

## 💡 Solution: Code-Generated Everything

Make `pkg/pb/env.go` the **only** place you define env vars. Everything else auto-generates.

### Architecture

```
pkg/pb/env.go (Single Source of Truth)
       │
       ├──> go run . env generate-dockerfile-docs  → Dockerfile env docs
       ├──> go run . env generate-env-local         → .env.local template
       ├──> go run . env generate-env-production    → .env.production template
       ├──> go run . env generate-flytoml-env       → fly.toml [env] section
       ├──> go run . env export-secrets             → (already exists)
       ├──> go run . env validate                   → (already exists)
       └──> go run . env list                       → (already exists)
```

---

## 📋 Implementation Plan

### 1. New CLI Commands (Add to `main.go`)

#### 1.1 `env sync-dockerfile`
**Purpose**: Auto-update Dockerfile env documentation

```go
syncDockerfileCmd := &cobra.Command{
    Use:   "sync-dockerfile",
    Short: "Sync environment variable documentation to Dockerfile",
    Long: `Updates the Dockerfile environment variable section with current registry.
Preserves Dockerfile structure, only updates the env vars comment block.`,
    RunE: func(cmd *cobra.Command, args []string) error {
        return wellknown.SyncDockerfileEnvDocs("Dockerfile")
    },
}
```

**Implementation**: `pkg/pb/env.go`
```go
// SyncDockerfileEnvDocs updates Dockerfile env documentation
func SyncDockerfileEnvDocs(dockerfilePath string) error {
    // 1. Read Dockerfile
    // 2. Find marker comments (e.g., # BEGIN AUTO-GENERATED ENV DOCS)
    // 3. Replace section with GenerateDockerfileEnvDocs()
    // 4. Write back
}

// GenerateDockerfileEnvDocs generates Dockerfile-style env var docs
func GenerateDockerfileEnvDocs() string {
    var sb strings.Builder
    sb.WriteString("# ================================================================\n")
    sb.WriteString("# Environment Variables (injected at runtime by Fly.io)\n")
    sb.WriteString("# AUTO-GENERATED from pkg/pb/env.go - DO NOT EDIT MANUALLY\n")
    sb.WriteString("# To update: go run . env sync-dockerfile\n")
    sb.WriteString("# ================================================================\n\n")

    // Separate secrets vs non-secrets
    sb.WriteString("# Required (set via fly.toml [env] section):\n")
    for _, v := range AllEnvVars {
        if !v.Secret && v.Required {
            sb.WriteString(fmt.Sprintf("#   %s=%s  # %s\n", v.Name, v.Default, v.Description))
        }
    }

    sb.WriteString("\n# Required (set via Fly.io secrets):\n")
    for _, v := range AllEnvVars {
        if v.Secret && v.Required {
            sb.WriteString(fmt.Sprintf("#   %s  # %s\n", v.Name, v.Description))
        }
    }

    sb.WriteString("\n# Optional (set via Fly.io secrets if needed):\n")
    for _, v := range AllEnvVars {
        if v.Secret && !v.Required {
            sb.WriteString(fmt.Sprintf("#   %s  # %s\n", v.Name, v.Description))
        }
    }

    sb.WriteString("\n# Sync secrets: make fly-secrets\n")
    sb.WriteString("# ================================================================\n")

    return sb.String()
}
```

---

#### 1.2 `env sync-flytoml`
**Purpose**: Auto-update fly.toml [env] section with non-secret vars

```go
syncFlyTomlCmd := &cobra.Command{
    Use:   "sync-flytoml",
    Short: "Sync non-secret environment variables to fly.toml",
    Long: `Updates fly.toml [env] section with non-secret environment variables.
Only includes variables where Secret=false.`,
    RunE: func(cmd *cobra.Command, args []string) error {
        return wellknown.SyncFlyTomlEnv("fly.toml")
    },
}
```

**Implementation**: `pkg/pb/env.go`
```go
// SyncFlyTomlEnv updates fly.toml [env] section
func SyncFlyTomlEnv(flytomlPath string) error {
    // 1. Read fly.toml
    // 2. Find [env] section
    // 3. Replace with GenerateFlyTomlEnv()
    // 4. Write back
}

// GenerateFlyTomlEnv generates fly.toml [env] section
func GenerateFlyTomlEnv() string {
    var sb strings.Builder
    sb.WriteString("[env]\n")
    sb.WriteString("  # PocketBase configuration (non-secret)\n")
    sb.WriteString("  # Secrets (OAuth, SMTP, etc.) are set via: make fly-secrets\n")

    for _, v := range AllEnvVars {
        if !v.Secret && v.Default != "" {
            sb.WriteString(fmt.Sprintf("  %s = \"%s\"\n", v.Name, v.Default))
        }
    }

    return sb.String()
}
```

---

#### 1.3 `env generate-local`
**Purpose**: Generate `.env.local` template for development

```go
generateLocalCmd := &cobra.Command{
    Use:   "generate-local",
    Short: "Generate .env.local template",
    Long: `Generates .env.local template with development-specific defaults.
Includes HTTPS_ENABLED=true and localhost OAuth URLs.`,
    RunE: func(cmd *cobra.Command, args []string) error {
        content := wellknown.GenerateEnvLocal()
        return os.WriteFile(".env.local", []byte(content), 0644)
    },
}
```

**Implementation**: `pkg/pb/env.go`
```go
// GenerateEnvLocal generates .env.local for development
func GenerateEnvLocal() string {
    var sb strings.Builder

    sb.WriteString("# ================================================================\n")
    sb.WriteString("# Wellknown Environment Variables - LOCAL DEVELOPMENT\n")
    sb.WriteString("# AUTO-GENERATED from pkg/pb/env.go\n")
    sb.WriteString("# To update: go run . env generate-local\n")
    sb.WriteString("# ================================================================\n\n")

    groups := GetVarsByGroup()
    groupNames := []string{
        "Server",
        "Google OAuth",
        "HTTPS (Development)",
        // ... rest of groups
    }

    for _, groupName := range groupNames {
        vars := groups[groupName]
        sb.WriteString(fmt.Sprintf("# ----------------------------------------------------------------\n"))
        sb.WriteString(fmt.Sprintf("# %s\n", groupName))
        sb.WriteString(fmt.Sprintf("# ----------------------------------------------------------------\n"))

        for _, v := range vars {
            sb.WriteString(fmt.Sprintf("# %s\n", v.Description))

            if v.Required {
                sb.WriteString("# REQUIRED\n")
            }

            // Development-specific defaults
            switch v.Name {
            case "GOOGLE_REDIRECT_URL":
                sb.WriteString(fmt.Sprintf("%s=https://localhost:8443/auth/google/callback\n", v.Name))
            case "HTTPS_ENABLED":
                sb.WriteString(fmt.Sprintf("%s=true\n", v.Name))
            default:
                if v.Default != "" {
                    sb.WriteString(fmt.Sprintf("%s=%s\n", v.Name, v.Default))
                } else {
                    sb.WriteString(fmt.Sprintf("%s=\n", v.Name))
                }
            }
            sb.WriteString("\n")
        }
    }

    return sb.String()
}
```

---

#### 1.4 `env generate-production`
**Purpose**: Generate `.env.production` template for Fly.io

```go
generateProductionCmd := &cobra.Command{
    Use:   "generate-production",
    Short: "Generate .env.production template",
    Long: `Generates .env.production template with production-specific defaults.
Includes HTTPS_ENABLED=false (Fly.io handles TLS) and production OAuth URLs.`,
    RunE: func(cmd *cobra.Command, args []string) error {
        content := wellknown.GenerateEnvProduction()
        return os.WriteFile(".env.production", []byte(content), 0644)
    },
}
```

**Implementation**: Similar to `GenerateEnvLocal()` but with production defaults.

---

#### 1.5 `env validate-required`
**Purpose**: Enhanced validation with helpful error messages

```go
validateRequiredCmd := &cobra.Command{
    Use:   "validate-required",
    Short: "Validate all required environment variables are set",
    Long: `Checks if all required environment variables are set.
Provides helpful error messages with examples for missing vars.`,
    RunE: func(cmd *cobra.Command, args []string) error {
        return wellknown.ValidateEnvWithHelp()
    },
}
```

**Implementation**: `pkg/pb/env.go`
```go
// ValidateEnvWithHelp provides detailed validation errors
func ValidateEnvWithHelp() error {
    var missing []EnvVar
    for _, v := range GetRequiredVars() {
        if os.Getenv(v.Name) == "" {
            missing = append(missing, v)
        }
    }

    if len(missing) > 0 {
        var sb strings.Builder
        sb.WriteString("❌ Missing required environment variables:\n\n")

        for _, v := range missing {
            sb.WriteString(fmt.Sprintf("  %s\n", v.Name))
            sb.WriteString(fmt.Sprintf("    Description: %s\n", v.Description))
            if v.Secret {
                sb.WriteString("    Type: SECRET (set via: make fly-secrets)\n")
            }
            sb.WriteString("\n")
        }

        return fmt.Errorf(sb.String())
    }

    return nil
}
```

---

### 2. Makefile Integration

Update `Makefile` to use Go commands instead of manual file editing:

```makefile
## env-sync: Sync environment configuration to all files
env-sync: env-sync-dockerfile env-sync-flytoml env-generate-local env-generate-production
	@echo "✅ All environment configuration synced!"

## env-sync-dockerfile: Update Dockerfile env documentation
env-sync-dockerfile:
	@echo "📝 Syncing Dockerfile env docs..."
	@go run . env sync-dockerfile
	@echo "✅ Dockerfile updated"

## env-sync-flytoml: Update fly.toml [env] section
env-sync-flytoml:
	@echo "📝 Syncing fly.toml [env] section..."
	@go run . env sync-flytoml
	@echo "✅ fly.toml updated"

## env-generate-local: Generate .env.local template
env-generate-local:
	@echo "📝 Generating .env.local template..."
	@go run . env generate-local
	@echo "✅ .env.local generated"

## env-generate-production: Generate .env.production template
env-generate-production:
	@echo "📝 Generating .env.production template..."
	@go run . env generate-production
	@echo "✅ .env.production generated"

## env-validate: Validate required environment variables
env-validate:
	@go run . env validate-required

## env-list: List all environment variables and their status
env-list:
	@go run . env list
```

---

### 3. Pre-Commit Hook (Optional but Recommended)

Create `.git/hooks/pre-commit`:

```bash
#!/bin/bash
# Pre-commit hook to ensure env vars are synced

echo "🔍 Checking environment variable synchronization..."

# Run env sync (dry-run check)
if ! make env-sync --dry-run > /dev/null 2>&1; then
    echo "❌ Environment variables out of sync!"
    echo "   Run: make env-sync"
    exit 1
fi

echo "✅ Environment variables synchronized"
exit 0
```

---

### 4. CI/CD Integration

#### GitHub Actions Workflow (`.github/workflows/env-check.yml`)

```yaml
name: Environment Config Check

on: [push, pull_request]

jobs:
  check-env-sync:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3

      - name: Set up Go
        uses: actions/setup-go@v4
        with:
          go-version: '1.25'

      - name: Check Dockerfile sync
        run: |
          make env-sync-dockerfile
          git diff --exit-code Dockerfile || (echo "Dockerfile out of sync! Run: make env-sync" && exit 1)

      - name: Check fly.toml sync
        run: |
          make env-sync-flytoml
          git diff --exit-code fly.toml || (echo "fly.toml out of sync! Run: make env-sync" && exit 1)

      - name: Validate env vars
        run: make env-validate || echo "Warning: Required env vars not set (expected in CI)"
```

---

## 🎯 Benefits

### Before (Manual Sync)
```
Developer adds new OAuth provider (e.g., GitHub OAuth)
  ↓
1. Update pkg/pb/env.go (add GITHUB_CLIENT_ID, GITHUB_CLIENT_SECRET)
2. Update Dockerfile (manually add to env docs)
3. Update .env.local (manually add with localhost values)
4. Update .env.production (manually add with production values)
5. Update fly.toml if non-secret
6. Remember to run make fly-secrets

❌ Error-prone: Easy to forget steps 2-5
❌ Time-consuming: 5-10 minutes of manual editing
❌ Drift risk: Docs fall out of sync over time
```

### After (Automated)
```
Developer adds new OAuth provider
  ↓
1. Update pkg/pb/env.go ONLY (add to AllEnvVars registry)
2. Run: make env-sync

✅ Automatic: All files updated in 2 seconds
✅ No errors: Single source of truth
✅ Always in sync: CI/CD enforces synchronization
```

---

## 📊 Comparison Matrix

| Task | Before (Manual) | After (Automated) | Time Saved |
|------|----------------|-------------------|------------|
| Add new env var | Edit 4-5 files | Edit 1 file + `make env-sync` | 80% |
| Update description | Edit 4-5 files | Edit 1 file + `make env-sync` | 90% |
| Remove env var | Edit 4-5 files + test | Edit 1 file + `make env-sync` | 85% |
| Verify sync | Manual inspection | `make env-validate` | 95% |
| Onboard new dev | Explain 5 files | Point to `pkg/pb/env.go` | 70% |

---

## 🚀 Migration Path

### Phase 1: Add CLI Commands (No Breaking Changes)
1. Add new CLI commands to `main.go`
2. Implement generator functions in `pkg/pb/env.go`
3. Test manually: `go run . env sync-dockerfile`

### Phase 2: Update Makefile
1. Add new Makefile targets
2. Update existing targets to use Go commands
3. Document in README

### Phase 3: Add Marker Comments (Safe Refactor)
1. Update `Dockerfile` with marker comments:
   ```dockerfile
   # BEGIN AUTO-GENERATED ENV DOCS
   # ... existing docs ...
   # END AUTO-GENERATED ENV DOCS
   ```
2. Update `fly.toml` with marker comments
3. Run `make env-sync` to verify

### Phase 4: Enable CI/CD Checks
1. Add GitHub Actions workflow
2. Test on feature branch
3. Enable on main branch

### Phase 5: Documentation
1. Update `README.md` with new workflow
2. Add migration guide for existing `.env` files
3. Update `DEPLOYMENT_ANALYSIS.md` with automation details

---

## 🔍 Example: Adding New OAuth Provider

### Before
```bash
# Step 1: Edit pkg/pb/env.go
vim pkg/pb/env.go
# ... add GITHUB_CLIENT_ID, GITHUB_CLIENT_SECRET ...

# Step 2: Edit Dockerfile
vim Dockerfile
# ... scroll to line 28 ...
# ... manually add GitHub OAuth section ...

# Step 3: Edit .env.local
vim .env.local
# ... manually add GITHUB_CLIENT_ID=, GITHUB_CLIENT_SECRET= ...

# Step 4: Edit .env.production
vim .env.production
# ... manually add production GitHub OAuth values ...

# Step 5: Pray you didn't miss anything
```

### After
```bash
# Step 1: Edit pkg/pb/env.go ONLY
vim pkg/pb/env.go
# ... add to AllEnvVars registry:
{
    Name:        "GITHUB_CLIENT_ID",
    Description: "GitHub OAuth client ID",
    Required:    false,
    Secret:      true,
    Group:       "GitHub OAuth",
},
{
    Name:        "GITHUB_CLIENT_SECRET",
    Description: "GitHub OAuth client secret",
    Required:    false,
    Secret:      true,
    Group:       "GitHub OAuth",
},

# Step 2: Sync everything
make env-sync

# Done! All files updated automatically:
# ✅ Dockerfile env docs updated
# ✅ .env.local template updated
# ✅ .env.production template updated
# ✅ Ready for: make fly-secrets
```

---

## 🛡️ Safety Features

### 1. Dry-Run Mode
```go
syncDockerfileCmd.Flags().BoolP("dry-run", "n", false, "Preview changes without writing")
```

### 2. Backup Before Sync
```go
func SyncDockerfileEnvDocs(dockerfilePath string) error {
    // 1. Create backup: Dockerfile.backup
    os.WriteFile(dockerfilePath+".backup", data, 0644)

    // 2. Perform sync
    // 3. Verify result
    // 4. Remove backup on success
}
```

### 3. Validation After Sync
```go
func SyncDockerfileEnvDocs(dockerfilePath string) error {
    // ... sync logic ...

    // Verify file is valid
    if err := ValidateDockerfile(dockerfilePath); err != nil {
        return fmt.Errorf("sync produced invalid Dockerfile: %w", err)
    }
}
```

---

## 📚 Additional Features

### 1. Environment Diff Tool
```bash
# Compare current environment with registry
go run . env diff

Output:
Environment Diff
================

✅ Set (3):
  GOOGLE_CLIENT_ID
  GOOGLE_CLIENT_SECRET
  SERVER_PORT

❌ Missing Required (1):
  GOOGLE_REDIRECT_URL

⚠️  Set but not in registry (1):
  LEGACY_API_KEY (consider removing)
```

### 2. Environment Migration Tool
```bash
# Migrate existing .env to new format
go run . env migrate .env

Output:
Migrating .env to new format...
  ✅ Preserved: GOOGLE_CLIENT_ID
  ✅ Preserved: GOOGLE_CLIENT_SECRET
  ⚠️  Removed: LEGACY_API_KEY (not in registry)
  ✅ Added: PB_DATA_DIR (with default: .data/pb)
✅ Migration complete! Backup saved to: .env.backup
```

### 3. Environment Documentation Generator
```bash
# Generate markdown documentation
go run . env generate-docs > ENVIRONMENT_VARIABLES.md

Output: ENVIRONMENT_VARIABLES.md
# Environment Variables

## Google OAuth
**Required for OAuth authentication**

### GOOGLE_CLIENT_ID
- **Description**: Google OAuth client ID
- **Required**: Yes
- **Secret**: Yes
- **How to get**: [Google Cloud Console](https://console.cloud.google.com/apis/credentials)

... (rest of vars)
```

---

## 🎉 Final Workflow

### Daily Development
```bash
# Add new env var
vim pkg/pb/env.go      # Edit AllEnvVars registry

# Sync everything
make env-sync

# Verify
make env-validate

# Commit
git add pkg/pb/env.go Dockerfile fly.toml .env.local .env.production
git commit -m "feat: add GitHub OAuth support"
```

### Deployment
```bash
# Sync production config
make env-sync

# Deploy
make fly-secrets       # Uses ExportSecretsFormat() - already works!
make fly-deploy

# Verify
make fly-logs
```

---

## 🚦 Implementation Priority

### High Priority (Do First)
1. ✅ `env sync-dockerfile` - Most manual work currently
2. ✅ `env generate-local` - Helps new developers onboard
3. ✅ `env generate-production` - Reduces production deployment errors

### Medium Priority (Nice to Have)
4. ⭐ `env sync-flytoml` - Automates fly.toml updates
5. ⭐ Pre-commit hook - Prevents accidental drift

### Low Priority (Future Enhancement)
6. 💡 `env diff` - Debugging tool
7. 💡 `env migrate` - One-time migration helper
8. 💡 `env generate-docs` - Documentation generation

---

## 📝 Next Steps

1. **Review this proposal** - Does this solve your pain points?
2. **Prioritize features** - Which commands are most valuable?
3. **Implementation plan** - Should I start with Phase 1?

**Estimated Implementation Time**:
- Phase 1 (CLI commands): 2-3 hours
- Phase 2 (Makefile): 30 minutes
- Phase 3 (Marker comments): 30 minutes
- Phase 4 (CI/CD): 1 hour
- **Total**: 4-5 hours for complete automation

---

## ❓ Questions for You

1. **Which phase should we start with?** (Recommend: Phase 1 + 2)
2. **Any additional env vars to add?** (e.g., DATABASE_URL, REDIS_URL?)
3. **Prefer interactive mode?** (e.g., `make env-sync` prompts before overwriting)
4. **Want environment-specific validation?** (e.g., local requires HTTPS_ENABLED=true)

Let me know and I'll start implementing! 🚀
