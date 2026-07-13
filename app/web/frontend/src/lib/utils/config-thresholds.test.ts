import { describe, it, expect } from 'vitest'
import { DEFAULT_THRESHOLDS } from '$lib/utils/cert-status'
import { thresholdsFromConfig } from '$lib/utils/config-thresholds'

describe('thresholdsFromConfig', () => {
  it('returns fallback when raw is missing', () => {
    expect(thresholdsFromConfig(undefined)).toEqual(DEFAULT_THRESHOLDS)
  })

  it('returns fallback when critical or warning is missing', () => {
    expect(thresholdsFromConfig({ critical: 14 })).toEqual(DEFAULT_THRESHOLDS)
    expect(thresholdsFromConfig({ warning: 60 })).toEqual(DEFAULT_THRESHOLDS)
    expect(thresholdsFromConfig({})).toEqual(DEFAULT_THRESHOLDS)
  })

  it('returns fallback for zero or negative values', () => {
    expect(thresholdsFromConfig({ critical: 0, warning: 30 })).toEqual(DEFAULT_THRESHOLDS)
    expect(thresholdsFromConfig({ critical: 7, warning: 0 })).toEqual(DEFAULT_THRESHOLDS)
    expect(thresholdsFromConfig({ critical: -1, warning: 30 })).toEqual(DEFAULT_THRESHOLDS)
  })

  it('returns fallback for non-finite numbers', () => {
    expect(thresholdsFromConfig({ critical: Number.NaN, warning: 30 })).toEqual(DEFAULT_THRESHOLDS)
    expect(thresholdsFromConfig({ critical: 7, warning: Number.POSITIVE_INFINITY })).toEqual(
      DEFAULT_THRESHOLDS,
    )
  })

  it('returns custom thresholds when both are positive numbers', () => {
    expect(thresholdsFromConfig({ critical: 14, warning: 60 })).toEqual({
      critical: 14,
      warning: 60,
    })
  })

  it('allows warning < critical without inventing product rules', () => {
    expect(thresholdsFromConfig({ critical: 30, warning: 7 })).toEqual({
      critical: 30,
      warning: 7,
    })
  })

  it('uses an explicit fallback when provided', () => {
    const custom = { critical: 1, warning: 2 }
    expect(thresholdsFromConfig(undefined, custom)).toEqual(custom)
  })
})
