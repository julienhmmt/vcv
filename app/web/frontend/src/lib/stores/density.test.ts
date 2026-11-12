// @vitest-environment jsdom
import { describe, it, expect, beforeEach } from 'vitest'
import { createDensityStore } from '$lib/stores/density.svelte'

beforeEach(() => {
  window.localStorage.clear()
})

describe('createDensityStore', () => {
  it('defaults to comfortable rows', () => {
    expect(createDensityStore().compact).toBe(false)
  })

  it('persists the toggled density', () => {
    const store = createDensityStore()
    store.toggle()

    expect(store.compact).toBe(true)
    expect(window.localStorage.getItem('vcv-density')).toBe('compact')
    expect(createDensityStore().compact).toBe(true)
  })

  it('falls back to comfortable when storage is unreadable', () => {
    const original = Storage.prototype.getItem
    Storage.prototype.getItem = () => {
      throw new Error('SecurityError: storage disabled')
    }
    try {
      expect(createDensityStore().compact).toBe(false)
    } finally {
      Storage.prototype.getItem = original
    }
  })
})
