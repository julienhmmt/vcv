import { describe, it, expect } from 'vitest'
import { parseUrlState, serializeUrlState, type UrlState } from '$lib/utils/url-state'

const defaults: UrlState = {
  search: '',
  statusFilters: [],
  certTypeFilter: 'all',
  mountFilter: null,
  sortKey: 'expiresAt',
  sortDir: 'asc',
  pageSize: 25,
  pageIndex: 0,
}

describe('parseUrlState', () => {
  it('returns defaults for an empty query string', () => {
    expect(parseUrlState(defaults, '')).toEqual(defaults)
  })

  it('parses each supported param', () => {
    const result = parseUrlState(
      defaults,
      '?q=acme&status=critical,expired&type=machine&mounts=pki%2Fa,pki%2Fb&sort=commonName&dir=desc&size=50&page=3',
    )
    expect(result).toEqual({
      search: 'acme',
      statusFilters: ['critical', 'expired'],
      certTypeFilter: 'machine',
      mountFilter: ['pki/a', 'pki/b'],
      sortKey: 'commonName',
      sortDir: 'desc',
      pageSize: 50,
      pageIndex: 2, // page is 1-based in the URL
    })
  })

  it('ignores invalid enum values and falls back to defaults', () => {
    const result = parseUrlState(defaults, '?status=bogus&type=nope&sort=nope&dir=nope&size=999')
    expect(result.statusFilters).toEqual([])
    expect(result.certTypeFilter).toBe('all')
    expect(result.sortKey).toBe('expiresAt')
    expect(result.sortDir).toBe('asc')
    expect(result.pageSize).toBe(25)
  })

  it('parses size=all and an empty mounts param as an empty selection', () => {
    const result = parseUrlState(defaults, '?size=all&mounts=')
    expect(result.pageSize).toBe('all')
    expect(result.mountFilter).toEqual([])
  })

  it('clamps a non-positive page to the default index', () => {
    expect(parseUrlState(defaults, '?page=0').pageIndex).toBe(0)
    expect(parseUrlState(defaults, '?page=-2').pageIndex).toBe(0)
  })
})

describe('serializeUrlState', () => {
  it('omits params left at their defaults', () => {
    expect(serializeUrlState(defaults, defaults)).toBe('')
  })

  it('serializes only the changed params', () => {
    const state: UrlState = {
      ...defaults,
      search: 'web',
      statusFilters: ['warning'],
      pageIndex: 1,
    }
    const params = new URLSearchParams(serializeUrlState(state, defaults))
    expect(params.get('q')).toBe('web')
    expect(params.get('status')).toBe('warning')
    expect(params.get('page')).toBe('2')
    expect(params.get('type')).toBeNull()
  })

  it('round-trips through parse', () => {
    const state: UrlState = {
      search: 'x',
      statusFilters: ['valid', 'revoked'],
      certTypeFilter: 'user',
      mountFilter: ['pki/a'],
      sortKey: 'vault',
      sortDir: 'desc',
      pageSize: 'all',
      pageIndex: 4,
    }
    const query = serializeUrlState(state, defaults)
    expect(parseUrlState(defaults, `?${query}`)).toEqual(state)
  })
})
