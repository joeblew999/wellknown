# MANDATORY RULES - READ BEFORE EVERY TASK

**Purpose**: Critical rules for AI agents working on this codebase

---

## 🔴 CRITICAL: THREE 9s NOT TWO!

**MOST COMMON ERROR:** Using `joeblew99` (TWO 9s) instead of `joeblew999` (THREE 9s)

✅ CORRECT: `github.com/joeblew999/wellknown`
❌ WRONG: `github.com/joeblew99/wellknown`

**If you see TWO 9s anywhere, STOP and fix it immediately!**

---

## ⚡ Pre-Task Checklist

Before doing ANYTHING, verify:

- [ ] Using `joeblew999` (THREE 9s)?
- [ ] Using `make <target>` (not direct code execution)?
- [ ] Will update Makefile if changing code structure?
- [ ] Will update `.gitignore` if adding generated/build files?
- [ ] User EXPLICITLY asked to commit?

---

## 🚨 Critical Rules (ZERO EXCEPTIONS)

### 1. Module Name & Path
- **Module:** `github.com/joeblew999/wellknown`
- **File Path:** `/Users/apple/workspace/go/src/github.com/joeblew999/wellknown`
- **Always THREE 9s:** `999` not `99`

### 2. Always Use Makefile
- ✅ DO: `make run`, `make build`, `make gen`
- ❌ DON'T: `go run ./pkg/cmd/pocketbase`, direct code execution
- The Makefile is the single source of truth

### 3. Keep Code and Makefile in Sync
- When moving files → Update Makefile paths immediately
- When adding features → Add/update Makefile targets

### 4. Never Auto-Commit
- ✅ DO: Ask "Should I commit these changes?"
- ❌ DON'T: Commit automatically or "helpfully"
- User controls version history, not AI

### 5. Manage .gitignore Proactively
- ✅ DO: Update `.gitignore` when adding new build artifacts, generated files, or directories
- ✅ DO: Add patterns for: compiled binaries, generated code, temp files, IDE files, logs, databases
- ✅ DO: Keep `.gitignore` organized with comments explaining sections
- ❌ DON'T: Let generated files or build artifacts get committed
- **Common patterns to ignore:**
  - Build outputs: `dist/`, `*.exe`, `*.so`, `*.dylib`
  - Generated code: `pb_migrations/`, `*_gen.go`
  - PocketBase data: `pb_data/` (except types)
  - Dependencies: `vendor/`, `node_modules/`
  - IDE files: `.vscode/`, `.idea/`, `*.swp`
  - Temp files: `*.tmp`, `*.log`, `.DS_Store`

---

## 📚 Reference Code (.src folder)

**ALWAYS check `.src/` BEFORE web searches!**

**Path:** `/Users/apple/workspace/go/src/github.com/joeblew999/wellknown/.src/`

### Agent File Discovery

**Check `.src/INDEX.md` when user mentions:**
Commits • Reviews • Planning • Exploration • PocketBase • UI • AI • Schemas

**Flow:**
```
User: "help me commit" → INDEX.md → Read trigger file → Apply workflow
```

**Configured via:** `.src/triggers.list` (edit there, run `make index`)
**Auto-generated:** INDEX.md updates after `make install`/`make upgrade`

**Keep `.src/Makefile` in sync with root Makefile**

---

## ❌ Common Violations

1. Using `joeblew99` (TWO 9s) instead of `joeblew999` (THREE 9s)
2. Running code directly instead of `make <target>`
3. Changing code structure without updating Makefile
4. Creating git commits without being asked
5. Web searches before checking `.src/`
6. Adding generated files without updating `.gitignore`

---

## 🎯 Why These Rules Exist

- **THREE 9s rule**: Prevents Go from fetching wrong remote packages
- **Makefile enforcement**: Ensures consistency, prevents breaking changes
- **Code/Makefile sync**: Keeps documentation and implementation aligned
- **No auto-commits**: User controls version history
- **.src/ first**: Faster, more reliable, project-specific examples
- **`.gitignore` management**: Prevents accidental commits of generated/temp files, keeps repo clean
