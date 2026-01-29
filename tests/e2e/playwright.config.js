// Playwright configuration for E2E tests
// See https://playwright.dev/docs/test-configuration

const { defineConfig, devices } = require('@playwright/test');

module.exports = defineConfig({
  testDir: './',

  // Maximum time one test can run
  timeout: 30 * 1000,

  // Test execution settings
  fullyParallel: false, // Run tests sequentially to avoid conflicts
  forbidOnly: !!process.env.CI,
  retries: process.env.CI ? 2 : 0,
  workers: 1, // Single worker to avoid race conditions

  // Reporter to use
  reporter: [
    ['html', { outputFolder: 'test-results/html' }],
    ['list'],
  ],

  // Shared settings for all projects
  use: {
    // Base URL for tests
    baseURL: 'http://localhost:7999',

    // Collect trace on failure
    trace: 'on-first-retry',

    // Screenshot on failure
    screenshot: 'only-on-failure',

    // Video on failure
    video: 'retain-on-failure',
  },

  // Configure projects for different test scenarios
  projects: [
    {
      name: 'api-tests',
      testMatch: /.*_test\.js/,
    },
  ],

  // Run backend server before starting tests (optional)
  // webServer: {
  //   command: 'cd ../../backend && go run cmd/server/main.go',
  //   url: 'http://localhost:7999/health',
  //   timeout: 120 * 1000,
  //   reuseExistingServer: !process.env.CI,
  // },
});
