/**
 * Core Wizard Tests - Focus on what we control
 *
 * These tests verify OUR code, not Google's:
 * - Form validation
 * - State management
 * - API endpoints
 * - .env persistence
 * - URL generation
 *
 * We do NOT test:
 * - GCP Console UI (too brittle)
 * - Project creation (out of our control)
 * - API enablement (slow, flaky)
 */

import { test, expect, Page } from '@playwright/test';
import * as fs from 'fs';
import * as path from 'path';

const WIZARD_URL = 'http://localhost:8080/tools/gcp-setup';
const ENV_PATH = path.join(__dirname, '../../.env');

test.describe('GCP Setup Wizard - Core Functionality', () => {

  test.beforeEach(async ({ page }) => {
    // Reset .env before each test
    await page.request.post('/api/gcp-setup/reset');
    await page.goto(WIZARD_URL);
  });

  test.describe('Step 1: Project ID Validation', () => {

    test('rejects project ID that is too short', async ({ page }) => {
      await page.fill('#projectId', 'abc');
      await page.click('button:has-text("Save & Continue")');

      await expect(page.locator('text=6-30 characters')).toBeVisible();
      await expect(page.locator('#status1')).toContainText('Pending');
    });

    test('rejects project ID with uppercase letters', async ({ page }) => {
      await page.fill('#projectId', 'My-Project-123');
      await page.click('button:has-text("Save & Continue")');

      await expect(page.locator('#projectIdError')).toContainText('lowercase');
    });

    test('rejects project ID ending with hyphen', async ({ page }) => {
      await page.fill('#projectId', 'my-project-');
      await page.click('button:has-text("Save & Continue")');

      await expect(page.locator('#projectIdError')).toContainText('cannot end with');
    });

    test('accepts valid project ID', async ({ page }) => {
      await page.fill('#projectId', 'wellknown-test-123');
      await page.fill('#projectName', 'Wellknown Test');
      await page.click('button:has-text("Save & Continue")');

      // Wait for API call to complete and status to update
      await expect(page.locator('#status1')).toContainText('Done', { timeout: 5000 });
    });
  });

  test.describe('Step 2: URL Generation', () => {

    test('generates correct GCP project creation URL', async ({ page }) => {
      // Save project ID first
      await page.fill('#projectId', 'my-test-project');
      await page.fill('#projectName', 'My Test');
      await page.click('button:has-text("Save & Continue")');

      // Verify Step 2 shows the project ID (use specific ID to avoid strict mode)
      await expect(page.locator('#projectIdToCopy')).toContainText('my-test-project');

      // Verify "Open GCP Console" button exists
      const createBtn = page.locator('#step2 >> text=Open GCP Console');
      await expect(createBtn).toBeVisible();
    });

    test('copy button copies project ID to clipboard', async ({ page }) => {
      await page.fill('#projectId', 'copy-test-123');
      await page.fill('#projectName', 'Copy Test');
      await page.click('button:has-text("Save & Continue")');

      // Click copy button
      await page.click('#step2 button:has-text("Copy")');

      // Verify feedback (button text changes to "✓ Copied!")
      await expect(page.locator('button:has-text("✓ Copied!")')).toBeVisible();

      // Verify clipboard content (requires permissions)
      // await expect(await page.evaluate(() => navigator.clipboard.readText()))
      //   .toBe('copy-test-123');
    });
  });

  test.describe('Step 3: API Enable URLs', () => {

    test('generates correct Calendar API URL with project parameter', async ({ page }) => {
      await page.fill('#projectId', 'api-test-project');
      await page.fill('#projectName', 'API Test');
      await page.click('button:has-text("Save & Continue")');

      // Mark Step 2 as done to reveal Step 3
      await page.click('#step2 button:has-text("Mark as Done")');

      // Verify Calendar API button has correct href
      const calendarBtn = page.locator('text=Enable Calendar API');
      await expect(calendarBtn).toBeVisible();

      // Check that clicking opens new tab with correct URL
      const [newPage] = await Promise.all([
        page.context().waitForEvent('page'),
        calendarBtn.click()
      ]);

      expect(newPage.url()).toContain('calendar-json.googleapis.com');
      expect(newPage.url()).toContain('project=api-test-project');

      await newPage.close();
    });

    test('generates correct OAuth2 API URL', async ({ page }) => {
      await page.fill('#projectId', 'oauth-test-123');
      await page.fill('#projectName', 'OAuth Test');
      await page.click('button:has-text("Save & Continue")');
      await page.click('#step2 button:has-text("Mark as Done")');

      const oauth2Btn = page.locator('text=Enable OAuth2 API');

      const [newPage] = await Promise.all([
        page.context().waitForEvent('page'),
        oauth2Btn.click()
      ]);

      expect(newPage.url()).toContain('oauth2.googleapis.com');
      expect(newPage.url()).toContain('project=oauth-test-123');

      await newPage.close();
    });
  });

  test.describe('State Persistence', () => {

    test('saves project ID to .env file', async ({ page }) => {
      await page.fill('#projectId', 'persist-test-123');
      await page.fill('#projectName', 'Persist Test');
      await page.click('button:has-text("Save & Continue")');

      // Wait for API call to complete
      await page.waitForTimeout(500);

      // Verify .env file was updated
      const envContent = fs.readFileSync(ENV_PATH, 'utf-8');
      expect(envContent).toContain('GCP_PROJECT_ID=persist-test-123');
    });

    test('pre-populates form from .env file on page load', async ({ page }) => {
      // Setup: Save via API
      await page.request.post('/api/gcp-setup/save-project', {
        headers: { 'Content-Type': 'application/json' },
        data: JSON.stringify({
          project_id: 'prepopulate-test',
          project_name: 'Prepopulate Test'
        })
      });

      // Test: Reload page and verify form
      await page.reload();

      await expect(page.locator('#projectId')).toHaveValue('prepopulate-test');
      await expect(page.locator('#projectName')).toHaveValue('prepopulate-test');
      await expect(page.locator('#status1')).toContainText('Done');
    });

    test('persists OAuth credentials to .env', async ({ page }) => {
      // Setup steps 1-4 first
      await page.fill('#projectId', 'creds-test');
      await page.fill('#projectName', 'Creds Test');
      await page.click('button:has-text("Save & Continue")');

      // Mark intermediate steps as done
      await page.click('#step2 button:has-text("Mark as Done")');
      await page.click('#step3 button:has-text("Mark as Done")');
      await page.click('#step4 button:has-text("Mark as Done")');

      // Fill credentials
      await page.fill('#clientId', 'test-client-id.apps.googleusercontent.com');
      await page.fill('#clientSecret', 'test-secret-123');
      await page.click('button:has-text("Save Credentials")');

      // Wait for save
      await page.waitForTimeout(500);

      // Verify .env file
      const envContent = fs.readFileSync(ENV_PATH, 'utf-8');
      expect(envContent).toContain('GOOGLE_CLIENT_ID=test-client-id');
      expect(envContent).toContain('GOOGLE_CLIENT_SECRET=test-secret-123');
    });
  });

  test.describe('Reset Functionality', () => {

    test('reset button clears all data and resets steps', async ({ page }) => {
      // Setup: Complete step 1
      await page.fill('#projectId', 'reset-test-123');
      await page.fill('#projectName', 'Reset Test');
      await page.click('button:has-text("Save & Continue")');

      await expect(page.locator('#status1')).toContainText('Done');

      // Click reset (no confirm dialog - instant reset for faster testing)
      await page.click('button:has-text("Reset Setup")');

      // Wait for page reload after reset
      await page.waitForLoadState('load');

      // Verify reset
      await expect(page.locator('#projectId')).toHaveValue('');
      await expect(page.locator('#projectName')).toHaveValue('');
      await expect(page.locator('#status1')).toContainText('Pending');

      // Verify .env was cleared (file may not exist after reset)
      if (fs.existsSync(ENV_PATH)) {
        const envContent = fs.readFileSync(ENV_PATH, 'utf-8');
        expect(envContent).not.toContain('GCP_PROJECT_ID=reset-test');
      }
      // File deletion is also acceptable
      // expect(fs.existsSync(ENV_PATH)).toBe(false);
    });
  });

  test.describe('Delete Project Button', () => {

    test('opens correct IAM settings URL for project deletion', async ({ page }) => {
      await page.fill('#projectId', 'delete-test-123');
      await page.fill('#projectName', 'Delete Test');
      await page.click('button:has-text("Save & Continue")');

      // Click delete button
      const [newPage] = await Promise.all([
        page.context().waitForEvent('page'),
        page.click('button:has-text("Delete")')
      ]);

      expect(newPage.url()).toContain('iam-admin/settings');
      expect(newPage.url()).toContain('project=delete-test-123');

      await newPage.close();
    });

    test('shows warning when trying to delete without project ID', async ({ page }) => {
      // Don't save any project ID

      page.on('dialog', dialog => {
        expect(dialog.message()).toContain('save a project ID first');
        dialog.accept();
      });

      await page.click('button:has-text("Delete")');
    });
  });

  test.describe('Step Progression', () => {

    test('steps progress in correct order', async ({ page }) => {
      // Initially all pending
      await expect(page.locator('#status1')).toContainText('Pending');
      await expect(page.locator('#status2')).toContainText('Pending');
      await expect(page.locator('#status3')).toContainText('Pending');

      // Complete Step 1
      await page.fill('#projectId', 'progress-test');
      await page.fill('#projectName', 'Progress Test');
      await page.click('button:has-text("Save & Continue")');

      await expect(page.locator('#status1')).toContainText('Done');
      await expect(page.locator('#status2')).toContainText('Done'); // Auto-marked

      // Complete Step 3
      await page.click('#step3 button:has-text("Mark as Done")');
      await expect(page.locator('#status3')).toContainText('Done');
    });
  });
});
