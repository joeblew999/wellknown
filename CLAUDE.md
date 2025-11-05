# MANDATORY RULES - READ BEFORE EVERY TASK

**Purpose**: Critical rules for AI agents working on this codebase

---

## ⚡ Pre-Task Checklist (VERIFY BEFORE EVERY ACTION)

Before doing ANYTHING, verify:

- [ ] Am I using `github.com/joeblew999/wellknown` (999 not 99)?
- [ ] Am I using Makefile (`make <target>`) instead of calling code directly?
- [ ] If changing code structure, will I update the Makefile?
- [ ] Did the user EXPLICITLY ask me to commit? (If no, DO NOT commit)

---

## 🚨 Critical Rules (ALL ARE MANDATORY - ZERO EXCEPTIONS)

### Rule 1: Module Name
**MUST USE:** `github.com/joeblew999/wellknown` (⚠️ THREE 9s: `999`)

✅ **DO:** `github.com/joeblew999/wellknown`
❌ **DON'T:** `github.com/joeblew99/wellknown` (only two 9s)

### Rule 2: File Path
**MUST USE:** `/Users/apple/workspace/go/src/github.com/joeblew999/wellknown`

✅ **DO:** Use the exact path above
❌ **DON'T:** Use any variation or shortened path

### Rule 3: Always Use Makefile
**MUST:** Run ALL operations through Makefile

✅ **DO:** `make run`, `make build`, `make gen`
❌ **DON'T:** `go run ./cmd/...`, `go run .`, direct code execution

### Rule 4: Keep Code and Makefile in Sync
**MUST:** Update Makefile when changing code structure

✅ **DO:** When moving files, update Makefile paths immediately
❌ **DON'T:** Change code structure without updating Makefile targets

### Rule 5: Never Auto-Commit
**MUST:** Wait for explicit user request before committing

✅ **DO:** Ask "Should I commit these changes?"
❌ **DON'T:** Commit automatically or "helpfully"

---

## ❌ Common Violations (NEVER DO THESE)

1. Running `go run ./pkg/cmd/pocketbase` instead of `make run`
2. Using `joeblew99` (two 9s) instead of `joeblew999` (three 9s)
3. Creating git commits without being asked
4. Changing file locations without updating Makefile paths
5. Calling code directly instead of using Makefile targets

---

## ⚙️ How to Work with This Codebase

1. **Check rules:** Verify the Pre-Task Checklist above
2. **Use Makefile:** Run `make help` to see available commands
3. **Keep in sync:** When changing code, update Makefile
4. **Ask before committing:** Never commit without explicit user request

The Makefile is the single source of truth for how to run this system.

---

## 📚 Reference Code (.src folder)

**ALWAYS check `.src/` BEFORE web searches** - it's faster, more reliable, and project-specific.

### ⚠️ CRITICAL: Use CORRECT path with THREE 9s!

**CORRECT `.src/` path (THREE 9s):**
```
/Users/apple/workspace/go/src/github.com/joeblew999/wellknown/.src/
```

**WRONG paths (TWO 9s - NEVER USE):**
```
/Users/apple/workspace/go/src/github.com/joeblew99/wellknown/.src/  ❌ WRONG!
```

✅ **DO:** Check `.src/` folder FIRST for examples, templates, and reference implementations
✅ **DO:** Use `.src/` when planning, researching, or looking for code patterns
✅ **DO:** ALWAYS use the path with THREE 9s: `joeblew999` not `joeblew99`
✅ **DO:** Maintain `.src/Makefile` alongside the root Makefile (keep them in sync)
✅ **DO:** Update `.src/Makefile` when updating root Makefile with new patterns
❌ **DON'T:** Default to web searches without checking `.src/` first
❌ **DON'T:** Skip `.src/` when researching how to implement features
❌ **DON'T:** EVER use `joeblew99` (two 9s) in ANY path

**When to use `.src/`:**
- Planning new features → Check `/Users/apple/workspace/go/src/github.com/joeblew999/wellknown/.src/` for similar patterns
- Need code examples → Look in `/Users/apple/workspace/go/src/github.com/joeblew999/wellknown/.src/` first
- Want to understand project structure → Review `.src/` reference code (THREE 9s!)
- Looking for Makefile patterns → Check `.src/Makefile` (reference Makefile for Claude)
- Need implementation guidance → Review `.src/` before web searches

**REMEMBER: It's `999` not `99` - THREE NINES!**

---

## 🎯 Why These Rules Exist

- **Module name typo**: Prevents Go from trying to fetch wrong remote packages
- **Makefile enforcement**: Ensures consistency and prevents breaking changes
- **Code/Makefile sync**: Keeps documentation and implementation aligned
- **No auto-commits**: User controls version history, not AI
