/**
 * Schema-Reflective Test Runner
 *
 * This test file REFLECTS over JSON Schemas at runtime and dynamically
 * generates test cases. NO code generation needed!
 *
 * Benefits:
 * - Always in sync with schema.json (reads it live)
 * - Change schema → Tests update automatically
 * - Single source of truth (the schema itself)
 * - No build/generate step needed
 */

import { test, expect } from '@playwright/test';
import * as fs from 'fs';
import * as path from 'path';

// Load test configuration
const testConfig = JSON.parse(
  fs.readFileSync(path.join(__dirname, '../test-config.json'), 'utf-8')
);

// Iterate over all pages with JSON schemas
testConfig.pages.forEach((pageConfig: any) => {
  if (pageConfig.type !== 'json_schema_form' || !pageConfig.schema) {
    return; // Skip non-schema pages
  }

  // Load the JSON Schema
  const schemaPath = path.join(__dirname, '..', pageConfig.schema);
  const schema = JSON.parse(fs.readFileSync(schemaPath, 'utf-8'));

  // Create test suite for this page
  test.describe(`${pageConfig.name} - Schema Validation (Reflective)`, () => {

    test.beforeEach(async ({ page }) => {
      await page.goto(pageConfig.url);
    });

    // Dynamically generate tests for each property in the schema
    if (schema.properties) {
      Object.entries(schema.properties).forEach(([fieldName, fieldSchema]: [string, any]) => {

        test.describe(`${fieldName} validation`, () => {

          // STRING VALIDATION TESTS
          if (fieldSchema.type === 'string') {

            // minLength test
            if (fieldSchema.minLength !== undefined) {
              test(`rejects ${fieldName} shorter than ${fieldSchema.minLength} chars`, async ({ page }) => {
                const input = page.locator(`[name="${fieldName}"]`);
                await input.fill('a'.repeat(fieldSchema.minLength - 1));
                await input.blur();

                // Should show error
                const errorMsg = page.locator('.error-message:visible, .invalid-feedback:visible');
                await expect(errorMsg).toBeVisible({ timeout: 2000 });
              });

              test(`accepts ${fieldName} with exactly ${fieldSchema.minLength} chars`, async ({ page }) => {
                const input = page.locator(`[name="${fieldName}"]`);
                await input.fill('a'.repeat(fieldSchema.minLength));
                await input.blur();

                // Should NOT show error
                const errorMsg = page.locator('.error-message:visible, .invalid-feedback:visible');
                await expect(errorMsg).toHaveCount(0);
              });
            }

            // maxLength test
            if (fieldSchema.maxLength !== undefined) {
              test(`rejects ${fieldName} longer than ${fieldSchema.maxLength} chars`, async ({ page }) => {
                const input = page.locator(`[name="${fieldName}"]`);
                await input.fill('a'.repeat(fieldSchema.maxLength + 1));
                await input.blur();

                // Should show error
                const errorMsg = page.locator('.error-message:visible, .invalid-feedback:visible');
                await expect(errorMsg).toBeVisible({ timeout: 2000 });
              });
            }

            // pattern test (common patterns)
            if (fieldSchema.pattern) {
              const pattern = fieldSchema.pattern;

              // Lowercase-only pattern
              if (pattern.includes('[a-z]') && !pattern.includes('[A-Z]')) {
                test(`rejects ${fieldName} with uppercase letters`, async ({ page }) => {
                  const input = page.locator(`[name="${fieldName}"]`);
                  await input.fill('ABC');
                  await input.blur();

                  const errorMsg = page.locator('.error-message:visible, .invalid-feedback:visible');
                  await expect(errorMsg).toBeVisible({ timeout: 2000 });
                });
              }
            }

            // format validation
            if (fieldSchema.format === 'email') {
              test(`rejects invalid email for ${fieldName}`, async ({ page }) => {
                const input = page.locator(`[name="${fieldName}"]`);
                await input.fill('not-an-email');
                await input.blur();

                const errorMsg = page.locator('.error-message:visible, .invalid-feedback:visible');
                await expect(errorMsg).toBeVisible({ timeout: 2000 });
              });

              test(`accepts valid email for ${fieldName}`, async ({ page }) => {
                const input = page.locator(`[name="${fieldName}"]`);
                await input.fill('test@example.com');
                await input.blur();

                const errorMsg = page.locator('.error-message:visible, .invalid-feedback:visible');
                await expect(errorMsg).toHaveCount(0);
              });
            }
          }

          // NUMBER VALIDATION TESTS
          if (fieldSchema.type === 'number' || fieldSchema.type === 'integer') {

            if (fieldSchema.minimum !== undefined) {
              test(`rejects ${fieldName} below minimum (${fieldSchema.minimum})`, async ({ page }) => {
                const input = page.locator(`[name="${fieldName}"]`);
                await input.fill(String(fieldSchema.minimum - 1));
                await input.blur();

                const errorMsg = page.locator('.error-message:visible, .invalid-feedback:visible');
                await expect(errorMsg).toBeVisible({ timeout: 2000 });
              });

              test(`accepts ${fieldName} at minimum (${fieldSchema.minimum})`, async ({ page }) => {
                const input = page.locator(`[name="${fieldName}"]`);
                await input.fill(String(fieldSchema.minimum));
                await input.blur();

                const errorMsg = page.locator('.error-message:visible, .invalid-feedback:visible');
                await expect(errorMsg).toHaveCount(0);
              });
            }

            if (fieldSchema.maximum !== undefined) {
              test(`rejects ${fieldName} above maximum (${fieldSchema.maximum})`, async ({ page }) => {
                const input = page.locator(`[name="${fieldName}"]`);
                await input.fill(String(fieldSchema.maximum + 1));
                await input.blur();

                const errorMsg = page.locator('.error-message:visible, .invalid-feedback:visible');
                await expect(errorMsg).toBeVisible({ timeout: 2000 });
              });
            }
          }

          // ENUM VALIDATION TESTS
          if (fieldSchema.enum && Array.isArray(fieldSchema.enum)) {
            test(`accepts valid enum value for ${fieldName}`, async ({ page }) => {
              const select = page.locator(`[name="${fieldName}"]`);
              await select.selectOption(fieldSchema.enum[0]);

              // Should not show error
              const errorMsg = page.locator('.error-message:visible, .invalid-feedback:visible');
              await expect(errorMsg).toHaveCount(0);
            });
          }
        });
      });
    }

    // REQUIRED FIELDS TESTS
    if (schema.required && Array.isArray(schema.required)) {
      test.describe('Required fields', () => {
        schema.required.forEach((fieldName: string) => {
          test(`${fieldName} is required`, async ({ page }) => {
            // Try to submit form without filling required field
            const submitBtn = page.locator('button[type="submit"], button:has-text("Generate"), button:has-text("Create")');
            await submitBtn.first().click();

            // Should show required error for this field
            const input = page.locator(`[name="${fieldName}"]`);
            const errorMsg = page.locator('.error-message:visible, .invalid-feedback:visible');

            // Either the field itself or nearby error message should indicate required
            const hasError = await errorMsg.count() > 0;
            const isInvalid = await input.getAttribute('aria-invalid');

            expect(hasError || isInvalid === 'true').toBeTruthy();
          });
        });
      });
    }
  });
});
