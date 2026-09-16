import '@testing-library/jest-dom/vitest'
import { afterEach } from 'vitest'

class ResizeObserverMock {
  constructor(_callback: ResizeObserverCallback) {}

  disconnect(): void {}

  observe(_target: Element): void {}

  unobserve(_target: Element): void {}
}

Object.defineProperty(globalThis, 'ResizeObserver', { configurable: true, value: ResizeObserverMock })

// Node ≥23 defines a non-functional global `localStorage` (it requires
// --localstorage-file). Its presence makes vitest's jsdom environment skip
// copying jsdom's working localStorage onto the test global, leaving
// window.localStorage undefined. Re-expose jsdom's implementation.
if (typeof window !== 'undefined' && typeof window.localStorage === 'undefined') {
  const jsdom = (globalThis as { jsdom?: { window?: { localStorage?: Storage } } }).jsdom
  const storage = jsdom?.window?.localStorage
  if (storage) {
    Object.defineProperty(globalThis, 'localStorage', { value: storage, configurable: true })
  }
}

// jsdom doesn't implement scrollIntoView; bits-ui's Command component calls it
// when the highlighted item changes (e.g. CommandPalette filtering as you type).
// Most test files run in the default 'node' environment, where Element is undefined.
if (typeof Element !== 'undefined' && !Element.prototype.scrollIntoView) {
  Element.prototype.scrollIntoView = () => {}
}

// jsdom doesn't implement matchMedia; the theme store reads it to detect the
// OS color scheme on mount. Most test files run in the default 'node'
// environment, where window is undefined.
if (typeof window !== 'undefined' && !window.matchMedia) {
  window.matchMedia = (query: string) =>
    ({
      matches: false,
      media: query,
      onchange: null,
      addListener: () => {},
      removeListener: () => {},
      addEventListener: () => {},
      removeEventListener: () => {},
      dispatchEvent: () => false,
    }) as MediaQueryList
}

// UI primitives (bits-ui Dialog scroll-lock restore) and components schedule
// real timers that can fire after a test file's jsdom globals are torn down,
// failing the run with an unhandled "document is not defined". Track real
// timers and drop leftovers after each test so nothing outlives its
// environment. Fake-timer tests are unaffected: vi.useFakeTimers swaps these
// globals out, and the per-file restore hooks run before this cleanup.
type TimerID = ReturnType<typeof setTimeout>
const pendingTimers = new Set<TimerID>()
const realSetTimeout = globalThis.setTimeout.bind(globalThis)
const realClearTimeout = globalThis.clearTimeout.bind(globalThis)
const realSetInterval = globalThis.setInterval.bind(globalThis)
const realClearInterval = globalThis.clearInterval.bind(globalThis)

globalThis.setTimeout = ((
  handler: TimerHandler,
  timeout?: number,
  ...args: unknown[]
): TimerID => {
  const id: TimerID = realSetTimeout(
    () => {
      pendingTimers.delete(id)
      if (typeof handler === 'function') {
        ;(handler as (...callArgs: unknown[]) => void)(...args)
      }
    },
    timeout,
  )
  pendingTimers.add(id)
  return id
}) as typeof setTimeout

globalThis.clearTimeout = ((id?: TimerID): void => {
  if (id !== undefined) pendingTimers.delete(id)
  realClearTimeout(id)
}) as typeof clearTimeout

globalThis.setInterval = ((
  handler: TimerHandler,
  timeout?: number,
  ...args: unknown[]
): TimerID => {
  const id: TimerID = realSetInterval(
    () => {
      if (typeof handler === 'function') {
        ;(handler as (...callArgs: unknown[]) => void)(...args)
      }
    },
    timeout,
  )
  pendingTimers.add(id)
  return id
}) as typeof setInterval

globalThis.clearInterval = ((id?: TimerID): void => {
  if (id !== undefined) pendingTimers.delete(id)
  realClearInterval(id)
}) as typeof clearInterval

afterEach(() => {
  for (const id of pendingTimers) {
    realClearTimeout(id)
    realClearInterval(id)
  }
  pendingTimers.clear()
})
