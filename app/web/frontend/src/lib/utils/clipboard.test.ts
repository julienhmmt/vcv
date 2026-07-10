// @vitest-environment jsdom
import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import { copyToClipboard } from '$lib/utils/clipboard'

describe('copyToClipboard', () => {
  let writeText: ReturnType<typeof vi.fn>
  let execCommand: ReturnType<typeof vi.fn>

  beforeEach(() => {
    writeText = vi.fn()
    // jsdom does not implement execCommand, so define it rather than spy on it.
    execCommand = vi.fn().mockReturnValue(true)
    Object.defineProperty(document, 'execCommand', { value: execCommand, configurable: true, writable: true })
  })

  afterEach(() => {
    vi.restoreAllMocks()
  })

  it('uses the async Clipboard API when available and returns true on success', async () => {
    writeText.mockResolvedValue(undefined)
    Object.defineProperty(navigator, 'clipboard', { value: { writeText }, configurable: true })
    expect(await copyToClipboard('hello')).toBe(true)
    expect(writeText).toHaveBeenCalledWith('hello')
    expect(execCommand).not.toHaveBeenCalled()
  })

  it('falls back to execCommand when navigator.clipboard is undefined', async () => {
    Object.defineProperty(navigator, 'clipboard', { value: undefined, configurable: true })
    expect(await copyToClipboard('hello')).toBe(true)
    expect(execCommand).toHaveBeenCalledWith('copy')
  })

  it('falls back to execCommand when writeText rejects', async () => {
    writeText.mockRejectedValue(new Error('denied'))
    Object.defineProperty(navigator, 'clipboard', { value: { writeText }, configurable: true })
    expect(await copyToClipboard('hello')).toBe(true)
    expect(execCommand).toHaveBeenCalledWith('copy')
  })

  it('returns false when both paths fail', async () => {
    Object.defineProperty(navigator, 'clipboard', { value: undefined, configurable: true })
    execCommand.mockReturnValue(false)
    expect(await copyToClipboard('hello')).toBe(false)
  })
})
