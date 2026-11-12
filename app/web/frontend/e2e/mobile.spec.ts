import { test, expect } from '@playwright/test'

// Phone layout regressions. These are pure geometry checks so they do not
// depend on the active UI language or on the seeded certificate inventory.
test.use({ viewport: { width: 390, height: 780 } })

test.beforeEach(async ({ page }) => {
  await page.goto('/')
  await expect(page.locator('.vcv-search-wrapper')).toBeVisible()
})

test('page does not scroll horizontally on a phone', async ({ page }) => {
  const overflow = await page
    .locator('html')
    .evaluate((el) => el.scrollWidth - el.clientWidth)
  expect(overflow).toBe(0)
})

test('search field keeps its natural height in the stacked filter bar', async ({ page }) => {
  const height = await page
    .locator('.vcv-search-wrapper')
    .evaluate((el) => el.getBoundingClientRect().height)

  // Regression: flex-basis 18rem became a height once the bar went column.
  expect(height).toBeLessThan(60)
})

test('status overview labels stay inside their buttons', async ({ page }) => {
  const overflowing = await page.locator('.vcv-stat').evaluateAll((stats) =>
    stats
      .map((stat) => {
        const label = stat.querySelector('.vcv-stat-label')
        if (!label) return ''
        const statBox = stat.getBoundingClientRect()
        return label.getBoundingClientRect().right > statBox.right
          ? (label.textContent ?? '').trim()
          : ''
      })
      .filter(Boolean)
  )
  expect(overflowing).toEqual([])
})
