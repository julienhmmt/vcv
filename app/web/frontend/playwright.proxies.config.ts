import { defineConfig, devices } from '@playwright/test'

// vcv behind the dev reverse proxies (docker-compose.dev.yml).
// Bring the stack up first: `make proxy-up` (or `make proxy-e2e` for the full
// cycle). Every proxy terminates HTTPS with the dev CA from
// dev/proxies/gen-cert.sh, so TLS errors are ignored here.
const proxies = [
  { name: 'nginx', url: 'https://localhost:53001' },
  { name: 'caddy', url: 'https://localhost:53002' },
  { name: 'traefik', url: 'https://localhost:53003' },
  // Host rewritten to app:52000, so only X-Forwarded-Host keeps the CSRF
  // origin right (ingress / Gateway API setups).
  { name: 'traefik-host-rewrite', url: 'https://localhost:53004' },
  { name: 'haproxy', url: 'https://localhost:53005' },
  // Its proxy host is created through its API by dev/proxies/npm-setup.sh.
  { name: 'nginx-proxy-manager', url: 'https://vcv.localhost' },
]

export default defineConfig({
  testDir: './e2e-proxies',
  fullyParallel: false,
  workers: 1,
  reporter: process.env.CI ? 'github' : 'list',
  timeout: 60_000,
  expect: {
    timeout: 15_000,
  },
  use: { trace: 'retain-on-failure', ignoreHTTPSErrors: true },
  projects: proxies.map(({ name, url }) => ({
    name,
    use: { ...devices['Desktop Chrome'], baseURL: url },
  })),
})
