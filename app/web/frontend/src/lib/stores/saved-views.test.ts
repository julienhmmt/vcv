// @vitest-environment jsdom
import { describe, it, expect, beforeEach } from 'vitest'
import { createSavedViewsStore, type ViewSnapshot } from '$lib/stores/saved-views.svelte'

const snapshot: ViewSnapshot = {
  search: '',
  statusFilters: ['critical'],
  certTypeFilter: 'all',
  mountFilter: null,
  sortKey: 'expiresAt',
  sortDir: 'asc',
  pageSize: 25,
  minDays: 0,
  maxDays: 7,
}

beforeEach(() => {
  window.localStorage.clear()
})

describe('createSavedViewsStore', () => {
  it('persists a view and reloads it in a new store instance', () => {
    createSavedViewsStore().save('Critical', snapshot)

    const reloaded = createSavedViewsStore()
    expect(reloaded.views).toHaveLength(1)
    expect(reloaded.views[0].name).toBe('Critical')
    expect(reloaded.views[0].snapshot).toEqual(snapshot)
  })

  it('trims the name and ignores a blank one', () => {
    const store = createSavedViewsStore()
    store.save('  Critical  ', snapshot)
    store.save('   ', snapshot)

    expect(store.views).toHaveLength(1)
    expect(store.views[0].name).toBe('Critical')
  })

  it('replaces the snapshot when the name already exists', () => {
    const store = createSavedViewsStore()
    store.save('Critical', snapshot)
    store.save('Critical', { ...snapshot, search: 'web' })

    expect(store.views).toHaveLength(1)
    expect(store.views[0].snapshot.search).toBe('web')
  })

  it('caps the list at ten views, dropping the oldest', () => {
    const store = createSavedViewsStore()
    for (let i = 0; i < 12; i++) store.save(`view-${i}`, snapshot)

    expect(store.views).toHaveLength(10)
    expect(store.views[0].name).toBe('view-2')
  })

  it('removes a view by id', () => {
    const store = createSavedViewsStore()
    store.save('Critical', snapshot)

    store.remove(store.views[0].id)

    expect(store.views).toHaveLength(0)
  })

  it('ignores corrupt storage', () => {
    window.localStorage.setItem('vcv-saved-views', '{not json')

    expect(createSavedViewsStore().views).toEqual([])
  })

  it('survives unreadable storage', () => {
    const original = Storage.prototype.getItem
    Storage.prototype.getItem = () => {
      throw new Error('SecurityError: storage disabled')
    }
    try {
      expect(createSavedViewsStore().views).toEqual([])
    } finally {
      Storage.prototype.getItem = original
    }
  })
})
