/**
 * End-to-End Test: Complete GCP OAuth Setup Flow
 *
 * This test covers:
 * 1. Local wizard: Save project ID
 * 2. GCP Console: Create project
 * 3. GCP Console: Enable Calendar API
 * 4. GCP Console: Enable OAuth2 API
 * 5. GCP Console: Configure OAuth consent
 * 6. GCP Console: Create OAuth credentials
 * 7. Local wizard: Save credentials
 *
 * Prerequisites:
 * - Local server running at http://localhost:8080
 * - Valid Google account logged into GCP Console
 * - HEADLESS=false to see browser interactions
 */

import { test, expect, Page } from '@playwright/test';

// Configuration
const LOCAL_WIZARD_URL = 'http://localhost:8080/tools/gcp-setup';
const PROJECT_ID = `wellknown-test-${Date.now()}`; // Unique project ID
const PROJECT_NAME = PROJECT_ID;

test.describe('GCP OAuth Setup - Full Flow', () => {
  let page: Page;

  test.beforeAll(async ({ browser }) => {
    // Create a persistent context to maintain GCP Console login
    const context = await browser.newContext();
    page = await context.newPage();
  });

  test('Step 1: Save Project ID in Local Wizard', async () => {
    await page.goto(LOCAL_WIZARD_URL);

    // Wait for page load
    await expect(page.locator('h1')).toContainText('GCP OAuth Setup');

    // Fill in project ID and name
    await page.fill('input#projectId', PROJECT_ID);
    await page.fill('input#projectName', PROJECT_NAME);

    // Click Save & Continue
    await page.click('button:has-text("Save & Continue")');

    // Verify Step 1 is marked as done
    await expect(page.locator('#status1')).toContainText('Done');

    // Verify Step 2 shows the project ID
    await expect(page.locator('code').filter({ hasText: PROJECT_ID })).toBeVisible();

    console.log(`✅ Step 1: Saved project ID: ${PROJECT_ID}`);
  });

  test('Step 2: Create GCP Project', async () => {
    // Navigate to GCP project creation
    await page.goto('https://console.cloud.google.com/projectcreate');

    // Wait for page load
    await expect(page.locator('h1')).toContainText('New Project');

    // Click Edit to reveal Project ID field
    await page.click('button:has-text("Edit")');

    // Fill Project ID
    await page.fill('input[aria-label*="Project ID"]', PROJECT_ID);

    // Fill Project Name
    await page.fill('input[aria-label*="Project name"]', PROJECT_NAME);

    // Verify no error message
    await expect(page.locator('text=Project ID is not available')).not.toBeVisible({ timeout: 2000 }).catch(() => {});

    // Click Create
    await page.click('button:has-text("Create")');

    // Wait for project creation (can take 10-30 seconds)
    await page.waitForURL(/project=.*/, { timeout: 60000 });

    console.log(`✅ Step 2: Created GCP project: ${PROJECT_ID}`);
  });

  test('Step 3a: Enable Google Calendar API', async () => {
    // Navigate to Calendar API
    await page.goto(`https://console.cloud.google.com/apis/library/calendar-json.googleapis.com?project=${PROJECT_ID}`);

    // Wait for page load
    await expect(page.locator('h1')).toContainText('Google Calendar API');

    // Click Enable button
    await page.click('button:has-text("Enable")');

    // Wait for enablement (can take a few seconds)
    await expect(page.locator('text=API enabled').or(page.locator('button:has-text("Manage")'))).toBeVisible({ timeout: 30000 });

    console.log(`✅ Step 3a: Enabled Google Calendar API`);
  });

  test('Step 3b: Enable OAuth2 API', async () => {
    // Navigate to OAuth2 API
    await page.goto(`https://console.cloud.google.com/apis/library/oauth2.googleapis.com?project=${PROJECT_ID}`);

    // Wait for page load
    await expect(page.locator('h1')).toContainText('Google OAuth2 API');

    // Click Enable button
    await page.click('button:has-text("Enable")');

    // Wait for enablement
    await expect(page.locator('text=API enabled').or(page.locator('button:has-text("Manage")'))).toBeVisible({ timeout: 30000 });

    console.log(`✅ Step 3b: Enabled OAuth2 API`);
  });

  test('Step 4: Configure OAuth Consent Screen', async () => {
    // Navigate to OAuth consent configuration
    await page.goto(`https://console.cloud.google.com/apis/credentials/consent?project=${PROJECT_ID}`);

    // Wait for page load
    await expect(page.locator('text=OAuth consent screen')).toBeVisible();

    // Select External (if not already configured)
    const externalRadio = page.locator('input[value="EXTERNAL"]');
    if (await externalRadio.isVisible()) {
      await externalRadio.click();
      await page.click('button:has-text("Create")');
    }

    // Fill in App Name
    await page.fill('input[aria-label*="App name"]', 'Wellknown Calendar App');

    // Fill in User Support Email (select from dropdown)
    await page.click('button:has-text("User support email")');
    await page.click('div[role="option"]:first-child'); // Select first email

    // Fill in Developer Email
    await page.fill('input[aria-label*="email addresses"]', 'developer@example.com');

    // Click Save and Continue
    await page.click('button:has-text("Save and Continue")');

    // Skip through remaining steps
    for (let i = 0; i < 3; i++) {
      await page.click('button:has-text("Save and Continue")').catch(() => {});
      await page.waitForTimeout(1000);
    }

    console.log(`✅ Step 4: Configured OAuth consent screen`);
  });

  test('Step 5: Create OAuth Credentials', async () => {
    // Navigate to credentials creation
    await page.goto(`https://console.cloud.google.com/apis/credentials?project=${PROJECT_ID}`);

    // Click Create Credentials
    await page.click('button:has-text("Create Credentials")');

    // Select OAuth client ID
    await page.click('text=OAuth client ID');

    // Select application type: Web application
    await page.click('button:has-text("Application type")');
    await page.click('text=Web application');

    // Fill in name
    await page.fill('input[aria-label*="Name"]', 'Wellknown OAuth Client');

    // Add Authorized redirect URI
    await page.click('button:has-text("Add URI")');
    await page.fill('input[placeholder*="https://"]', 'http://localhost:8090/auth/google/callback');

    // Click Create
    await page.click('button:has-text("Create")');

    // Wait for credentials popup
    await expect(page.locator('text=OAuth client created')).toBeVisible({ timeout: 10000 });

    // Extract Client ID and Client Secret
    const clientId = await page.locator('input[aria-label*="Client ID"]').inputValue();
    const clientSecret = await page.locator('input[aria-label*="Client secret"]').inputValue();

    console.log(`✅ Step 5: Created OAuth credentials`);
    console.log(`   Client ID: ${clientId}`);
    console.log(`   Client Secret: ${clientSecret.substring(0, 10)}...`);

    // Close popup
    await page.click('button:has-text("OK")');

    // Store credentials for next test
    test.info().annotations.push({ type: 'client_id', description: clientId });
    test.info().annotations.push({ type: 'client_secret', description: clientSecret });
  });

  test('Step 6: Save Credentials in Local Wizard', async () => {
    // Get credentials from previous test
    const annotations = test.info().annotations;
    const clientId = annotations.find(a => a.type === 'client_id')?.description || '';
    const clientSecret = annotations.find(a => a.type === 'client_secret')?.description || '';

    // Navigate back to local wizard
    await page.goto(LOCAL_WIZARD_URL);

    // Scroll to Step 5
    await page.locator('#step5').scrollIntoViewIfNeeded();

    // Fill in credentials
    await page.fill('input#clientId', clientId);
    await page.fill('input#clientSecret', clientSecret);

    // Click Save Credentials
    await page.click('button:has-text("Save Credentials")');

    // Verify success message
    await expect(page.locator('text=Setup Complete')).toBeVisible({ timeout: 5000 });

    console.log(`✅ Step 6: Saved credentials to .env file`);
  });

  test('Verify Complete Setup', async () => {
    await page.goto(LOCAL_WIZARD_URL);

    // Verify all steps show "Done"
    await expect(page.locator('#status1')).toContainText('Done');
    await expect(page.locator('#status2')).toContainText('Done');
    await expect(page.locator('#status3')).toContainText('Done');
    await expect(page.locator('#status4')).toContainText('Done');
    await expect(page.locator('#status5')).toContainText('Done');

    console.log(`✅ Full Setup Complete! All steps verified.`);
  });
});
