import { describe, it, expect } from 'vitest'
import { expiryTier, shouldNotifyExpiry, type ExpiryTier } from '$lib/utils/expiry-notify'

describe('expiryTier', () => {
  it('prefers critical over warning', () => {
    expect(expiryTier({ critical: 1, warning: 5 })).toBe('critical')
  })

  it('returns warning when only warning counts', () => {
    expect(expiryTier({ critical: 0, warning: 2 })).toBe('warning')
  })

  it('returns none when both zero', () => {
    expect(expiryTier({ critical: 0, warning: 0 })).toBe('none')
  })
})

describe('shouldNotifyExpiry', () => {
  const cases: {
    isInitial: boolean
    tier: ExpiryTier
    last: ExpiryTier
    expect: boolean
  }[] = [
    { isInitial: true, tier: 'warning', last: 'none', expect: true },
    { isInitial: true, tier: 'critical', last: 'none', expect: true },
    { isInitial: true, tier: 'none', last: 'none', expect: false },
    { isInitial: false, tier: 'warning', last: 'warning', expect: false },
    { isInitial: false, tier: 'critical', last: 'warning', expect: true },
    { isInitial: false, tier: 'warning', last: 'critical', expect: false },
    { isInitial: false, tier: 'none', last: 'warning', expect: false },
    { isInitial: false, tier: 'critical', last: 'critical', expect: false },
    { isInitial: false, tier: 'warning', last: 'none', expect: true },
  ]

  for (const row of cases) {
    it(`isInitial=${row.isInitial} tier=${row.tier} last=${row.last} → ${row.expect}`, () => {
      expect(
        shouldNotifyExpiry({
          isInitial: row.isInitial,
          tier: row.tier,
          lastNotifiedTier: row.last,
        }),
      ).toBe(row.expect)
    })
  }
})
