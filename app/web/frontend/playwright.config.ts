import { defineConfig } from '@playwright/test'

// E2E against the lean docker-compose.e2e.yml stack (app on :52001).
// Bring the stack up first: `make e2e-up` (or `make e2e` for the full cycle).
export default defineConfig({
  testDir: './e2e',
  fullyParallel: false,
  retries: process.env.CI ? 1 : 0,
  reporter: process.env.CI ? 'github' : 'list',
  use: {
    baseURL: process.env.VCV_E2E_URL ?? 'http://localhost:52001',
    trace: 'retain-on-failure',
  },
  timeout: 60_000,
  expect: {
    timeout: 15_000,
  },
})
