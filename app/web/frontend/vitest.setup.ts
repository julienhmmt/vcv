import '@testing-library/jest-dom/vitest'

class ResizeObserverMock {
  constructor(_callback: ResizeObserverCallback) {}

  disconnect(): void {}

  observe(_target: Element): void {}

  unobserve(_target: Element): void {}
}

Object.defineProperty(globalThis, 'ResizeObserver', { configurable: true, value: ResizeObserverMock })
