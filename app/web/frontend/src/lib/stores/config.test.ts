import { describe, it, expect, vi, beforeEach } from 'vitest'
import type { PublicConfigResponse } from '$lib/types'

const { config } = vi.hoisted(() => ({ config: vi.fn() }))

vi.mock('$lib/api', () => ({
  api: { config },
  ApiError: class extends Error {},
}))

import { createConfigStore } from '$lib/stores/config.svelte'

const base: PublicConfigResponse = { expirationThresholds: { critical: 7, warning: 30 } }

describe('config store mountKeys', () => {
  beforeEach(() => {
    config.mockReset()
  })

  it('is empty before the first refresh', () => {
    expect(createConfigStore().mountKeys).toEqual([])
  })

  it('flattens vault pki mounts into sorted vault|mount keys', async () => {
    config.mockResolvedValue({
      ...base,
      vaults: [
        { id: 'b', displayName: 'B', pkiMounts: ['pki'] },
        { id: 'a', displayName: 'A', pkiMounts: ['pki-int', 'pki'] },
      ],
    })
    const store = createConfigStore()
    await store.refresh()
    expect(store.mountKeys).toEqual(['a|pki', 'a|pki-int', 'b|pki'])
  })

  it('returns no keys when the response has no vaults', async () => {
    config.mockResolvedValue(base)
    const store = createConfigStore()
    await store.refresh()
    expect(store.mountKeys).toEqual([])
  })
})
