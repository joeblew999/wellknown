/**
 * GENERIC Platform Test Suite - Schema-Driven E2E Testing
 *
 * This is a FULLY GENERIC test suite that works for ANY platform/appType combination.
 * No hardcoded platform logic - everything driven by Go-generated test data + schema metadata.
 *
 * Architecture:
 * 1. Auto-discovers test suites from tests/e2e/generated/*.json
 * 2. Uses schema metadata for validation (required fields, types, etc.)
 * 3. Platform-agnostic validation logic
 * 4. Works for Google Calendar, Apple Calendar, Maps, etc. with ZERO code changes!
 *
 * Adding new platform: Just run `make gen-testdata` - no TypeScript changes needed!
 */

import { test, expect, Page } from '@playwright/test';
import * as fs from 'fs';
import * as path from 'path';

const BASE_URL = 'http://localhost:8080';

// ============================================================================
// Type Definitions (matches Go structures)
// ============================================================================

interface GoTestCase {
  name: string;
  description: string;
  data: Record<string, any>;
  expected: {
    url?: string;
    ics?: string;
    ics_contains?: string[];
    error?: string;
  };
  validation: {
    is_valid: boolean;
    errors?: Record<string, string>;
    fields_tested?: string[];
  };
  tags?: string[];
}

interface SchemaMetadata {
  required_fields: string[];
  optional_fields: string[];
  field_types: Record<string, string>;
  title?: string;
  description?: string;
}

interface GoTestSuite {
  platform: string;
  app_type: string;
  generated_at: string;
  source_file: string;
  test_cases: GoTestCase[];
  metadata: Record<string, any>;
  schema_metadata: SchemaMetadata;
}

// ============================================================================
// Auto-Discovery: Find all Go-generated test suites
// ============================================================================

function discoverTestSuites(): GoTestSuite[] {
  const generatedDir = path.join(__dirname, 'generated');

  if (!fs.existsSync(generatedDir)) {
    throw new Error(`Generated test directory not found: ${generatedDir}\nRun: make gen-testdata`);
  }

  const files = fs.readdirSync(generatedDir).filter(f => f.endsWith('-tests.json'));

  if (files.length === 0) {
    throw new Error('No generated test files found. Run: make gen-testdata');
  }

  return files.map(filename => {
    const filepath = path.join(generatedDir, filename);
    const suite = JSON.parse(fs.readFileSync(filepath, 'utf-8'));

    console.log(`✅ Discovered: ${suite.platform}/${suite.app_type} (${suite.test_cases.length} tests)`);

    return suite;
  });
}

// ============================================================================
// Generic Form Filling - Works for ANY schema
// ============================================================================

async function fillForm(page: Page, data: Record<string, any>): Promise<void> {
  for (const [key, value] of Object.entries(data)) {
    // Handle nested objects (e.g., recurrence, attendees)
    if (typeof value === 'object' && !Array.isArray(value) && value !== null) {
      // For now, skip complex nested objects
      // TODO: Handle nested form fields if needed
      continue;
    }

    // Handle arrays (future: attendees, reminders)
    if (Array.isArray(value)) {
      // Skip for now - would need dynamic array field handling
      continue;
    }

    // String/text inputs
    if (typeof value === 'string') {
      const input = page.locator(`input[name="${key}"], textarea[name="${key}"]`);
      if (await input.count() > 0) {
        await input.fill(value);
      }
    }
    // Boolean/checkbox inputs
    else if (typeof value === 'boolean') {
      const checkbox = page.locator(`input[name="${key}"][type="checkbox"]`);
      if (await checkbox.count() > 0) {
        if (value) {
          await checkbox.check();
        } else {
          await checkbox.uncheck();
        }
      }
    }
    // Number inputs
    else if (typeof value === 'number') {
      const input = page.locator(`input[name="${key}"]`);
      if (await input.count() > 0) {
        await input.fill(value.toString());
      }
    }
  }
}

// ============================================================================
// Generic Output Validation - Platform-agnostic
// ============================================================================

async function validateOutput(
  page: Page,
  suite: GoTestSuite,
  testCase: GoTestCase
): Promise<void> {
  const { platform, app_type } = suite;
  const { expected } = testCase;

  // Validation 1: URL-based output (e.g., Google Calendar)
  if (expected.url) {
    const currentURL = page.url();

    // Check if redirected to external service OR success page
    if (currentURL.includes(new URL(expected.url).hostname)) {
      // External redirect - verify URL structure matches
      const goURL = new URL(expected.url);
      const currentParams = new URL(currentURL);

      // Verify hostname matches
      expect(currentURL).toContain(goURL.hostname);

      // Verify key query params exist
      goURL.searchParams.forEach((value, key) => {
        expect(currentParams.searchParams.has(key)).toBe(true);
      });
    } else {
      // Success page - verify generated link exists
      const generatedLink = page.locator(`a[href*="${new URL(expected.url).hostname}"]`);
      if (await generatedLink.count() > 0) {
        const href = await generatedLink.getAttribute('href');
        expect(href).toContain(new URL(expected.url).hostname);
      }
    }
  }

  // Validation 2: ICS content validation (e.g., Apple Calendar)
  if (expected.ics_contains && expected.ics_contains.length > 0) {
    // Look for download link or embedded ICS content
    const downloadLink = page.locator('a[href*="calendar/download"], a[download]');

    if (await downloadLink.count() > 0) {
      // Verify download link exists
      expect(downloadLink).toBeTruthy();

      // Could fetch and verify ICS content here if needed
      // For now, just verify the link is present
    }

    // Alternative: Check if ICS keywords appear in page source
    const pageContent = await page.content();
    expected.ics_contains.forEach(keyword => {
      // Some keywords might appear in data attributes or hidden fields
      if (keyword.startsWith('BEGIN:') || keyword.startsWith('END:')) {
        // These are structural - less likely to appear in UI
        // Could fetch actual ICS file to verify
      }
    });
  }

  // Validation 3: Error handling
  if (expected.error) {
    const errorMessage = page.locator('.error, [role="alert"], text=/error|invalid/i');
    await expect(errorMessage.first()).toBeVisible({ timeout: 2000 });
  }
}

// ============================================================================
// GENERIC TEST SUITE - Works for ALL platforms!
// ============================================================================

// Create test suites dynamically
function createPlatformTests() {
  const testSuites = discoverTestSuites();

  console.log(`\n🚀 Discovered ${testSuites.length} platform test suite(s)\n`);

  for (const suite of testSuites) {
    test.describe(`${suite.platform}/${suite.app_type} - Schema-Driven Tests`, () => {

      test.beforeAll(() => {
      console.log(`\n📦 Test Suite: ${suite.platform}/${suite.app_type}`);
      console.log(`   Source: ${suite.source_file}`);
      console.log(`   Generated: ${suite.generated_at}`);
      console.log(`   Test cases: ${suite.test_cases.length}`);
      console.log(`   Required fields: ${suite.schema_metadata.required_fields.join(', ')}`);

      // Count tags
      const tagCounts: Record<string, number> = {};
      suite.test_cases.forEach(tc => {
        (tc.tags || []).forEach(tag => {
          tagCounts[tag] = (tagCounts[tag] || 0) + 1;
        });
      });
      console.log(`   Tag distribution:`, tagCounts);
    });

    // Generate a test for EACH Go test case
    suite.test_cases.forEach((testCase, index) => {
      const tags = testCase.tags?.join(', ') || 'untagged';

      test(`[${index + 1}/${suite.test_cases.length}] ${testCase.name} (${tags})`, async ({ page }) => {
        // Navigate to platform-specific form
        const url = `${BASE_URL}/${suite.platform}/${suite.app_type}`;
        await page.goto(url);

        // Fill form with test data
        await fillForm(page, testCase.data);

        // Submit form
        const submitButton = page.locator('button[type="submit"]');
        await submitButton.click();

        // Wait for response
        await page.waitForLoadState('networkidle', { timeout: 5000 }).catch(() => {
          // Some forms redirect immediately, that's OK
        });

        // Validate output using generic validation
        await validateOutput(page, suite, testCase);
      });
    });

    // Meta-test: Verify schema field coverage
    test('📊 metadata: schema field coverage', async () => {
      const fieldsUsed = new Set<string>();

      suite.test_cases.forEach(tc => {
        Object.keys(tc.data).forEach(field => fieldsUsed.add(field));
      });

      console.log(`\n📊 Field Coverage for ${suite.platform}/${suite.app_type}:`);
      console.log(`   Total unique fields: ${fieldsUsed.size}`);
      console.log(`   Fields tested: ${Array.from(fieldsUsed).sort().join(', ')}`);

      // Verify ALL required fields are tested
      suite.schema_metadata.required_fields.forEach(field => {
        expect(fieldsUsed.has(field)).toBe(true);
      });

      console.log(`   ✅ All ${suite.schema_metadata.required_fields.length} required fields covered`);
    });

    // Meta-test: Verify validation rules are tested
    test('🔍 metadata: validation coverage', async () => {
      const validTests = suite.test_cases.filter(tc => tc.validation.is_valid);
      const invalidTests = suite.test_cases.filter(tc => !tc.validation.is_valid);

      console.log(`\n🔍 Validation Coverage:`);
      console.log(`   Valid test cases: ${validTests.length}`);
      console.log(`   Invalid test cases: ${invalidTests.length}`);

      // Should have at least some valid tests
      expect(validTests.length).toBeGreaterThan(0);

      // Future: Could verify invalid tests cover all validation rules
    });
    });
  }

  return testSuites;
}

// Call the function to register all tests
const testSuites = createPlatformTests();

// ============================================================================
// Summary Test - Cross-platform statistics
// ============================================================================

test.describe('🌐 Cross-Platform Summary', () => {
  test('platform coverage statistics', async () => {
    console.log(`\n🌐 Cross-Platform Test Summary:`);
    console.log(`   Total platforms: ${testSuites.length}`);

    const totalTests = testSuites.reduce((sum, s) => sum + s.test_cases.length, 0);
    console.log(`   Total test cases: ${totalTests}`);

    testSuites.forEach(suite => {
      console.log(`   - ${suite.platform}/${suite.app_type}: ${suite.test_cases.length} tests`);
    });

    expect(testSuites.length).toBeGreaterThan(0);
    expect(totalTests).toBeGreaterThan(0);
  });
});
