/**
 * Calendar Test Helpers - DRY utilities for data-driven testing
 *
 * These helpers work for ANY calendar platform (Google, Apple, etc.)
 * by using reflection-style introspection of the data structure.
 */

import { Page, expect } from '@playwright/test';
import * as fs from 'fs';
import * as path from 'path';

// Constants - Single source of truth for selectors and text
export const SELECTORS = {
  EXAMPLE_CARD: '.showcase-item',
  GENERATE_BUTTON: 'button:has-text("Generate & Open")',
  SCHEMA_MESSAGE: 'text=This form is dynamically generated from JSON Schema',
  DATA_SOURCE_INFO: (platform: string, appType: string) => `pkg/${platform}/${appType}/data-examples.json`,
} as const;

export const TEXT = {
  CLICK_EXAMPLE_MESSAGE: 'Click any example below',
  TITLE_LABEL: 'Title',
  START_LABEL: 'Start',
  END_LABEL: 'End',
} as const;

// Type definitions
export interface Example {
  name: string;
  description: string;
  data: Record<string, any>;
}

export interface ExamplesData {
  examples: Example[];
}

export interface PlatformConfig {
  platform: 'google' | 'apple';
  appType: 'calendar' | 'maps';
  baseUrl: string;
}

/**
 * Load examples data from JSON file using reflection
 * Works for any platform/appType combination
 */
export function loadExamplesData(config: PlatformConfig): ExamplesData {
  const jsonPath = path.join(
    __dirname,
    `../../../pkg/${config.platform}/${config.appType}/data-examples.json`
  );

  if (!fs.existsSync(jsonPath)) {
    throw new Error(`Examples data not found: ${jsonPath}`);
  }

  const data = JSON.parse(fs.readFileSync(jsonPath, 'utf-8'));

  // Validate structure using reflection
  if (!data.examples || !Array.isArray(data.examples)) {
    throw new Error(`Invalid examples data structure in ${jsonPath}`);
  }

  return data as ExamplesData;
}

/**
 * Navigate to examples page for any platform
 */
export async function navigateToExamplesPage(page: Page, config: PlatformConfig): Promise<void> {
  const url = `${config.baseUrl}/${config.platform}/${config.appType}/examples`;
  await page.goto(url);
  await page.waitForLoadState('networkidle');
}

/**
 * Navigate to custom form page for any platform
 */
export async function navigateToCustomForm(page: Page, config: PlatformConfig): Promise<void> {
  const url = `${config.baseUrl}/${config.platform}/${config.appType}`;
  await page.goto(url);
  await page.waitForLoadState('networkidle');
}

/**
 * Verify examples page loaded correctly (generic for all platforms)
 */
export async function verifyExamplesPageLoaded(page: Page, config: PlatformConfig): Promise<void> {
  // Verify data source info box
  const dataSourceText = SELECTORS.DATA_SOURCE_INFO(config.platform, config.appType);
  await expect(page.locator(`text=${dataSourceText}`)).toBeVisible();

  // Verify click message
  await expect(page.locator(`text=${TEXT.CLICK_EXAMPLE_MESSAGE}`)).toBeVisible();
}

/**
 * Verify correct number of examples are displayed
 */
export async function verifyExampleCount(page: Page, expectedCount: number): Promise<void> {
  await page.waitForSelector(SELECTORS.EXAMPLE_CARD, { state: 'visible' });
  const cards = page.locator(SELECTORS.EXAMPLE_CARD);
  await expect(cards).toHaveCount(expectedCount);
}

/**
 * Verify a single example card renders correctly
 * Uses reflection to check all data fields dynamically
 */
export async function verifyExampleCard(page: Page, example: Example): Promise<void> {
  const card = page.locator(SELECTORS.EXAMPLE_CARD, { hasText: example.name });
  await expect(card).toBeVisible();

  // Verify description
  if (example.description) {
    await expect(card).toContainText(example.description);
  }

  // Dynamically verify all data fields that should be visible
  // This uses reflection-style iteration over the data object
  for (const [key, value] of Object.entries(example.data)) {
    // Only check primitive values that would be displayed as text
    if (typeof value === 'string' && !key.includes('recurrence') && !key.includes('attendees')) {
      await expect(card).toContainText(value);
    }
  }

  // Verify button
  await expect(card.locator(SELECTORS.GENERATE_BUTTON)).toBeVisible();
}

/**
 * Verify custom form has schema-driven fields
 */
export async function verifyCustomFormLoaded(page: Page): Promise<void> {
  await expect(page.locator(SELECTORS.SCHEMA_MESSAGE)).toBeVisible();

  // All calendar forms should have these basic fields
  await expect(page.locator(`label:has-text("${TEXT.TITLE_LABEL}")`)).toBeVisible();
  await expect(page.locator(`label:has-text("${TEXT.START_LABEL}")`)).toBeVisible();
  await expect(page.locator(`label:has-text("${TEXT.END_LABEL}")`)).toBeVisible();
}

/**
 * Fill calendar form with data (works for any platform)
 * Uses reflection to dynamically fill all fields based on data object
 */
export async function fillCalendarForm(page: Page, data: Record<string, any>): Promise<void> {
  // Dynamically fill all fields based on data object keys
  for (const [fieldName, value] of Object.entries(data)) {
    if (typeof value === 'string') {
      const input = page.locator(`input[name="${fieldName}"]`);
      const inputCount = await input.count();

      if (inputCount > 0) {
        await input.fill(value);
      }
    }
    // Handle checkboxes
    else if (typeof value === 'boolean') {
      const checkbox = page.locator(`input[name="${fieldName}"][type="checkbox"]`);
      const checkboxCount = await checkbox.count();

      if (checkboxCount > 0) {
        if (value) {
          await checkbox.check();
        } else {
          await checkbox.uncheck();
        }
      }
    }
    // Handle nested objects (recurrence, attendees, etc.) - to be implemented
  }
}

/**
 * Submit form and verify it handled correctly
 * Platform-agnostic - works for both Google and Apple
 */
export async function submitFormAndVerify(page: Page, config: PlatformConfig): Promise<void> {
  await page.click('button[type="submit"]');
  await page.waitForLoadState('networkidle');

  // Verify we're still on the app (not an error page)
  expect(page.url()).toContain(config.baseUrl);

  // Platform-specific checks
  if (config.platform === 'google') {
    // Google Calendar redirects to external URL or success page
    // We verify the URL was at least processed
    expect(page.url()).toContain(`/${config.platform}/${config.appType}`);
  } else if (config.platform === 'apple') {
    // Apple Calendar may show download link or stay on form
    // Verify no errors occurred
    const errorMessage = page.locator('text=Error');
    const errorCount = await errorMessage.count();
    expect(errorCount).toBe(0);
  }
}

/**
 * Get advanced features from examples data using reflection
 * Returns examples that have specific feature types
 */
export function getExamplesWithFeature(examples: Example[], featureName: string): Example[] {
  return examples.filter(example => {
    // Recursively search for feature in data object
    return hasFeature(example.data, featureName);
  });
}

/**
 * Recursively check if data object has a feature (reflection-style)
 */
function hasFeature(obj: any, featureName: string): boolean {
  if (obj === null || obj === undefined) return false;

  // Direct key match
  if (obj[featureName] !== undefined) return true;

  // Recursive search in nested objects
  for (const value of Object.values(obj)) {
    if (typeof value === 'object' && hasFeature(value, featureName)) {
      return true;
    }
  }

  return false;
}

/**
 * Validate example data against expected schema (reflection-based validation)
 */
export function validateExampleStructure(example: Example): void {
  // Required fields
  if (!example.name || typeof example.name !== 'string') {
    throw new Error(`Invalid example: name is required and must be a string`);
  }

  if (!example.data || typeof example.data !== 'object') {
    throw new Error(`Invalid example: data is required and must be an object`);
  }

  // Description is optional but must be string if present
  if (example.description !== undefined && typeof example.description !== 'string') {
    throw new Error(`Invalid example: description must be a string`);
  }
}

/**
 * Get all field names from examples data (reflection)
 * Useful for dynamic test generation
 */
export function getAllFieldNames(examples: Example[]): Set<string> {
  const fields = new Set<string>();

  examples.forEach(example => {
    Object.keys(example.data).forEach(key => fields.add(key));
  });

  return fields;
}
