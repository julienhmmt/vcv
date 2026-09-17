import { test, expect } from '@playwright/test'

// Smoke coverage against the seeded e2e stack (see docker-compose.e2e.yml):
// vault-main/pki is seeded with deterministic CNs like
// web.vault-main-pki.local, machine-web.*, expiring-*, expired-*, revoked-*.
// The default sort is expires-ascending, so expired certs land on page 1.
const PAGE_ONE_CN = 'expired-1.vault-main-pki.local'

test.beforeEach(async ({ page }) => {
  await page.goto('/')
  // Inventory fetch from real Vault/OpenBao backends can take a few seconds
  // on a cold cache; wait for the table to render its first page.
  await expect(page.getByRole('table')).toBeVisible()
  await expect(page.getByRole('button', { name: /: Details/ }).first()).toBeVisible({ timeout: 45_000 })
  // exact: "0 certificates" is a substring of counts like "120 certificates".
  await expect(page.getByText('0 certificates', { exact: true })).toHaveCount(0)
})

test('lists seeded certificates and filters by search', async ({ page }) => {
  await expect(page.getByText(PAGE_ONE_CN).first()).toBeVisible()

  // Full CN is unique; bare "user-alice" is seeded once per mount (6 hits).
  await page.getByLabel('Search certificates').fill('user-alice.vault-main-pki.local')
  await expect(page.getByText('1 certificates', { exact: true })).toBeVisible()
  await expect(page.getByText('user-alice.vault-main-pki.local').first()).toBeVisible()
  await expect(page.getByText('user-bob.vault-main-pki.local')).toHaveCount(0)
})

test('sorts by expiration date from the column header', async ({ page }) => {
  const expiresHead = page.getByRole('columnheader', { name: /expires/i })
  const before = await expiresHead.getAttribute('aria-sort')

  await expiresHead.getByRole('button').click()
  await expect
    .poll(async () => expiresHead.getAttribute('aria-sort'), { timeout: 10_000 })
    .not.toBe(before)
})

test('opens the certificate detail modal with issuer info', async ({ page }) => {
  await page.getByRole('button', { name: `${PAGE_ONE_CN}: Details` }).first().click()

  // Scope to the dialog heading: the CN also stays in the table row behind
  // the modal and repeats inside the dialog (subject, technical details).
  const dialog = page.getByRole('dialog')
  await expect(dialog.getByRole('heading', { name: PAGE_ONE_CN })).toBeVisible()
  await expect(dialog.getByText('Issuer', { exact: true }).first()).toBeVisible()

  await page.keyboard.press('Escape')
  await expect(page.getByRole('dialog')).toHaveCount(0)
})

test('mount selector lists mounts and toggles scope', async ({ page }) => {
  const countText = await page.getByText(/certificates$/).first().textContent()

  // The toolbar button is labelled "Sources: n/n".
  await page.getByRole('button', { name: /sources/i }).click()
  const dialog = page.getByRole('dialog')
  await expect(dialog).toContainText('Certificate sources')
  await expect(dialog).toContainText('pki')

  // Deselect everything: the table drains to zero. (exact: "Select all" is a
  // substring of "Deselect all" under Playwright's default name matching.)
  await dialog.getByRole('button', { name: 'Deselect all', exact: true }).click()
  await expect(page.getByText('0 certificates', { exact: true })).toBeVisible()

  // Restore full scope.
  await dialog.getByRole('button', { name: 'Select all', exact: true }).click()
  await expect(page.getByText(/certificates$/).first()).toHaveText(countText ?? '')

  await dialog.getByRole('button', { name: 'Done', exact: true }).click()
  await expect(page.getByRole('dialog')).toHaveCount(0)
})

test('admin login opens the settings panel', async ({ page }) => {
  await page.goto('/admin')

  await page.getByLabel('Username').fill('admin')
  await page.getByLabel('Password').fill('e2e-admin-password')
  // en bundle renders adminLogin as "Login" (not the "Sign In" fallback).
  await page.getByRole('button', { name: 'Login' }).click()

  await expect(page.getByText('Expiration thresholds')).toBeVisible({ timeout: 20_000 })
})
