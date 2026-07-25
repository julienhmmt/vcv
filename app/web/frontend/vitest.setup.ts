import '@testing-library/jest-dom/vitest'

class ResizeObserverMock {
  constructor(_callback: ResizeObserverCallback) {}

  disconnect(): void {}

  observe(_target: Element): void {}

  unobserve(_target: Element): void {}
}

Object.defineProperty(globalThis, 'ResizeObserver', { configurable: true, value: ResizeObserverMock })

// jsdom doesn't implement scrollIntoView; bits-ui's Command component calls it
// when the highlighted item changes (e.g. CommandPalette filtering as you type).
// Most test files run in the default 'node' environment, where Element is undefined.
if (typeof Element !== 'undefined' && !Element.prototype.scrollIntoView) {
  Element.prototype.scrollIntoView = () => {}
}
