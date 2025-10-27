import { defineConfig, devices } from '@playwright/test';

/**
 * Playwright configuration for wellknown GCP setup testing
 *
 * See https://playwright.dev/docs/test-configuration
 */
export default defineConfig({
  testDir: './e2e',

  /* Maximum time one test can run for */
  timeout: 120 * 1000, // 2 minutes per test (GCP operations can be slow)

  expect: {
    timeout: 10000 // 10 seconds for assertions
  },

  /* Run tests in files in parallel */
  fullyParallel: false, // Sequential for GCP setup flow

  /* Fail the build on CI if you accidentally left test.only in the source code */
  forbidOnly: !!process.env.CI,

  /* Retry on CI only */
  retries: process.env.CI ? 2 : 0,

  /* Opt out of parallel tests on CI */
  workers: process.env.CI ? 1 : 1, // Sequential execution

  /* Reporter to use */
  reporter: [
    ['html'],
    ['list'],
    ['json', { outputFile: 'test-results/results.json' }]
  ],

  /* Shared settings for all the projects below */
  use: {
    /* Base URL for local server */
    baseURL: 'http://localhost:8080',

    /* Collect trace when retrying the failed test */
    trace: 'on-first-retry',

    /* Screenshot on failure */
    screenshot: 'only-on-failure',

    /* Video on failure */
    video: 'retain-on-failure',

    /* Slow down operations (useful for debugging) */
    // launchOptions: {
    //   slowMo: 500
    // }
  },

  /* Configure projects for major browsers */
  projects: [
    {
      name: 'webkit',
      use: {
        ...devices['Desktop Safari'],
        // Keep browser open to maintain GCP login
        headless: false,
      },
    },

    // Chromium for comparison (uncomment if needed)
    // {
    //   name: 'chromium',
    //   use: {
    //     ...devices['Desktop Chrome'],
    //     headless: false,
    //   },
    // },

    // {
    //   name: 'firefox',
    //   use: { ...devices['Desktop Firefox'] },
    // },
  ],

  /* Run your local dev server before starting the tests */
  webServer: {
    command: 'make dev',
    url: 'http://localhost:8080',
    reuseExistingServer: !process.env.CI,
    timeout: 120 * 1000, // 2 minutes to start server
  },
});
