import '@testing-library/jest-dom/vitest'

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
