import { test, expect } from '@playwright/test'

// One project per proxy in docker-compose.dev.yml. The app runs with
// app.trust_proxy: true, so these pass only when the proxy overwrites
// X-Forwarded-Host/-Proto.
test('serves the SPA and the certificate API over HTTPS', async ({ page }) => {
  const response = await page.goto('/')
  expect(response?.status()).toBe(200)
  // SecurityHeaders emits HSTS only when the request looks like HTTPS.
  expect(response?.headers()['strict-transport-security']).toBeTruthy()

  // Seeded e2e stack: the table renders real Vault/OpenBao data.
  await expect(page.getByRole('table')).toBeVisible()
  await expect(page.getByRole('button', { name: /: Details/ }).first()).toBeVisible({ timeout: 45_000 })
})

test('accepts a same-origin admin POST', async ({ request, baseURL }) => {
  // CSRF compares Origin with the app's target origin, built from
  // X-Forwarded-Proto/-Host. Without them the app sees http while the browser
  // says https, and every unsafe request is a 403. No session: 204, not 401.
  const response = await request.post('/api/admin/logout', { headers: { Origin: baseURL! } })
  expect(response.status()).toBe(204)
})
