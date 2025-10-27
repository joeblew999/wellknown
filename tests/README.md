# Wellknown E2E Tests

End-to-end Playwright tests for the GCP OAuth setup wizard.

## Structure

```
tests/
├── e2e/
│   └── gcp-setup-flow.spec.ts    # Full GCP setup workflow
├── playwright.config.ts           # Playwright configuration
├── package.json                   # Node.js dependencies
└── README.md                      # This file
```

## Setup

1. **Install dependencies with Bun (LOCAL to project - no global pollution):**
   ```bash
   cd tests
   bun install
   bun run install:browsers
   ```

   **Why Bun?**
   - ⚡ 10-100x faster than npm
   - 📦 Everything stays in `tests/node_modules/` - zero OS pollution!
   - 🔒 No global installs, no system-wide changes

2. **Start the local server** (in another terminal):
   ```bash
   cd ..
   make dev
   ```

3. **Login to GCP Console** (first time only):
   - Run tests with `--headed` flag
   - Browser will open - manually login to your Google account
   - Close browser after login
   - Subsequent runs will reuse the login session

## Running Tests

### Run all tests (headless):
```bash
bun test
```

### Run with UI (see browser):
```bash
bun run test:headed
```

### Run with Playwright UI mode (best for development):
```bash
bun run test:ui
```

### Debug mode (step through tests):
```bash
bun run test:debug
```

### View test report:
```bash
bun run test:report
```

### Generate tests interactively:
```bash
bun run codegen
```

## Test Flow

The main test (`gcp-setup-flow.spec.ts`) covers the complete OAuth setup:

1. **Step 1**: Save Project ID in local wizard
2. **Step 2**: Create GCP Project
3. **Step 3a**: Enable Google Calendar API
4. **Step 3b**: Enable OAuth2 API
5. **Step 4**: Configure OAuth consent screen
6. **Step 5**: Create OAuth credentials
7. **Step 6**: Save credentials back to local wizard
8. **Verification**: Confirm all steps marked as "Done"

## Important Notes

### GCP Login Session
- Tests require you to be logged into GCP Console
- First run will open browser for manual login
- Playwright stores session for subsequent runs
- If tests fail with login errors, delete `playwright/.auth/` and login again

### Project ID Uniqueness
- Tests generate unique project IDs using timestamp: `wellknown-test-{timestamp}`
- This prevents conflicts with existing projects
- You can manually set PROJECT_ID in the test file if needed

### Test Timeouts
- Default timeout: 2 minutes per test
- GCP operations (project creation, API enablement) can take 10-60 seconds
- If tests timeout, increase `timeout` in `playwright.config.ts`

### Cleanup
Tests do NOT automatically delete created GCP projects. To clean up:
```bash
# List all projects
gcloud projects list

# Delete a test project
gcloud projects delete wellknown-test-XXXXXXXXXX
```

## Troubleshooting

### Port 8080 already in use
```bash
# Kill process on port 8080
lsof -ti:8080 | xargs kill -9

# Start server again
make dev
```

### Browser not opening
```bash
# Reinstall browsers
bun run playwright install --force
```

### Test failing on "Select a project" dialog
- Project doesn't exist yet - Step 2 didn't complete
- Check GCP Console manually to verify project creation
- May need to increase timeout for project creation

### Credentials not saving
- Check `.env` file exists and is writable
- Verify server API endpoint `/api/gcp-setup/save-creds` is working
- Check server logs for errors

## CI/CD Integration

To run in CI (GitHub Actions, etc.):

```yaml
- name: Install dependencies
  run: |
    cd tests
    npm ci
    npx playwright install --with-deps

- name: Run E2E tests
  run: |
    cd tests
    npm test
  env:
    CI: true

- name: Upload test results
  if: always()
  uses: actions/upload-artifact@v3
  with:
    name: playwright-report
    path: tests/playwright-report/
```

**Note**: CI requires GCP service account for authentication (not implemented in current tests).

## Development Tips

### Record new tests:
```bash
bun run codegen
# Browser opens - perform actions
# Playwright generates test code
```

### Debug failing tests:
```bash
bun run test:debug
# Playwright Inspector opens
# Step through each action
```

### Update selectors:
If GCP Console UI changes, update selectors in test file:
- Use Playwright Inspector to find new selectors
- Prefer `data-testid` or `aria-label` over CSS classes
- Use `getByRole()` for accessibility

### Screenshots and videos:
Tests automatically capture:
- Screenshots on failure: `test-results/`
- Videos on failure: `test-results/`
- Traces for debugging: `test-results/`

## Resources

- [Playwright Documentation](https://playwright.dev)
- [Best Practices](https://playwright.dev/docs/best-practices)
- [Debugging Tests](https://playwright.dev/docs/debug)
- [CI/CD Integration](https://playwright.dev/docs/ci)
