# Testing Strategy for GCP OAuth Setup Wizard

## Philosophy: Test What You Control

### ✅ We Test (High ROI)
- **Our wizard logic** - validation, state management, UI updates
- **Our API endpoints** - `/api/gcp-setup/*`
- **Our .env file handling** - persistence, formatting
- **Our URL generation** - correct project IDs, proper encoding

### ❌ We Don't Test (Low ROI, High Maintenance)
- **GCP Console UI** - Google changes it constantly
- **GCP API behavior** - project creation, API enablement
- **Third-party authentication** - OAuth flows, token validation

## Test Suite Structure

```
tests/
├── e2e/
│   ├── wizard-core.spec.ts        # Core functionality (PREFERRED)
│   └── gcp-setup-flow.spec.ts     # Full E2E (reference only)
```

### wizard-core.spec.ts (Recommended)

**What it tests:**
- ✅ Form validation rules
- ✅ Project ID format checking
- ✅ State persistence to .env
- ✅ Form pre-population
- ✅ URL generation correctness
- ✅ Copy button functionality
- ✅ Reset functionality
- ✅ Step progression logic

**Why it's better:**
- **Fast** - no external dependencies
- **Reliable** - tests our code, not Google's
- **Maintainable** - breaks only when we change our code
- **Deterministic** - same results every time

**Test coverage:**
```typescript
✓ Validates project ID length (6-30 chars)
✓ Rejects uppercase letters
✓ Rejects trailing hyphens
✓ Saves to .env correctly
✓ Pre-populates form on page load
✓ Generates correct GCP URLs with project parameter
✓ Copy button copies to clipboard
✓ Reset clears all data
✓ Delete button opens correct IAM page
✓ Steps progress in correct order
```

### gcp-setup-flow.spec.ts (Reference)

**What it tests:**
- Complete flow including GCP Console interactions
- Project creation, API enablement, OAuth setup

**Why it's NOT recommended:**
- ❌ **Brittle** - breaks when Google changes UI
- ❌ **Slow** - project creation takes 30-60 seconds
- ❌ **Flaky** - timing issues, rate limits
- ❌ **Manual** - requires user login to GCP
- ❌ **Expensive** - creates real GCP projects (cleanup required)

**When to use:**
- Manual testing before major releases
- Documenting the full workflow
- Debugging integration issues

## Running Tests

### Quick Test (Recommended)
```bash
cd tests
bun test wizard-core  # Fast, reliable core tests
```

### Full E2E (Use sparingly)
```bash
bun test gcp-setup-flow  # Slow, requires GCP login
```

## Test Development Workflow

### 1. Write Tests First (TDD)
```bash
# Create new test
bun run codegen  # Record interactions
# Edit generated code to focus on our logic
```

### 2. Run Tests During Development
```bash
bun run test:ui  # Visual feedback
bun run test:headed  # See browser
```

### 3. Debug Failures
```bash
bun run test:debug  # Step through with Playwright Inspector
```

## What To Test Next

### High Priority
- [ ] Concurrent access (multiple users editing .env)
- [ ] Invalid .env file formats (recovery)
- [ ] Network errors (API endpoint failures)
- [ ] Browser back/forward button behavior

### Medium Priority
- [ ] Mobile responsive design
- [ ] Accessibility (keyboard navigation)
- [ ] Error message clarity
- [ ] Help text completeness

### Low Priority
- [ ] Different browsers (Firefox, Safari)
- [ ] Slow network conditions
- [ ] Session timeout handling

## Common Pitfalls

### ❌ Don't Do This
```typescript
// Testing Google's UI (brittle!)
test('GCP Console creates project', async ({ page }) => {
  await page.goto('https://console.cloud.google.com/projectcreate');
  await page.click('button.mdc-button--raised');  // BREAKS when Google changes CSS!
  // ...
});
```

### ✅ Do This Instead
```typescript
// Test our URL generation (stable!)
test('generates correct project creation URL', async ({ page }) => {
  await page.fill('#projectId', 'my-project');
  await page.click('button:has-text("Save")');

  const createBtn = page.locator('a:has-text("Open GCP Console")');
  await expect(createBtn).toHaveAttribute('href', /projectcreate/);
});
```

## Test Data Management

### Use Dynamic IDs
```typescript
// Good - unique every time
const PROJECT_ID = `test-${Date.now()}`;

// Bad - conflicts with previous runs
const PROJECT_ID = 'my-test-project';
```

### Clean Up After Tests
```typescript
test.afterEach(async ({ page }) => {
  // Reset .env
  await page.request.post('/api/gcp-setup/reset');
});
```

### Use Test Fixtures
```typescript
// tests/fixtures/env-data.ts
export const validProjectData = {
  project_id: 'wellknown-test-123',
  project_name: 'Wellknown Test'
};
```

## CI/CD Integration

### GitHub Actions Example
```yaml
name: E2E Tests

on: [push, pull_request]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3

      - uses: oven-sh/setup-bun@v1

      - name: Install dependencies
        run: |
          cd tests
          bun install
          bun run install:browsers

      - name: Start server
        run: make dev &

      - name: Run core tests
        run: cd tests && bun test wizard-core

      - name: Upload test results
        if: always()
        uses: actions/upload-artifact@v3
        with:
          name: test-results
          path: tests/test-results/
```

## Metrics

### Test Execution Time
- **Core tests**: ~30 seconds (fast!)
- **Full E2E**: ~5-10 minutes (slow)

### Test Coverage Goal
- **API endpoints**: 100%
- **Validation logic**: 100%
- **UI interactions**: 80%
- **GCP Console**: 0% (intentionally)

## Resources

- [Playwright Best Practices](https://playwright.dev/docs/best-practices)
- [Testing Trophy](https://kentcdodds.com/blog/write-tests) - focus on integration tests
- [Testing Library Philosophy](https://testing-library.com/docs/guiding-principles/) - test user behavior
