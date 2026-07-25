// @vitest-environment jsdom
import { describe, it, expect, vi } from 'vitest'
import { render, screen, fireEvent } from '@testing-library/svelte'
import type { SettingsFile } from '$lib/types'

vi.mock('$lib/stores/i18n.svelte', () => ({
  getI18n: () => ({ t: (_key: string, fallback?: string) => fallback ?? _key }),
}))

import AdminPanel from '$lib/components/admin/AdminPanel.svelte'

function settings(overrides: Partial<SettingsFile> = {}): SettingsFile {
  return {
    app: { env: 'dev', port: 52000 },
    certificates: { expiration_thresholds: { critical: 7, warning: 30 } },
    metrics: {},
    cors: {},
    notifications: { webhook_url: '' },
    vaults: [],
    ...overrides,
  }
}

function baseProps(settingsOverrides: Partial<SettingsFile> = {}) {
  return {
    settings: settings(settingsOverrides),
    statuses: [],
    loading: false,
    error: null,
    successMessage: null,
    onSave: vi.fn(),
    onAddVault: vi.fn(),
    onRemoveVault: vi.fn(),
    onInvalidateCache: vi.fn(),
    onLogout: vi.fn(),
  }
}

describe('AdminPanel webhook URL field', () => {
  it('starts empty when the server has no webhook configured', () => {
    render(AdminPanel, { props: baseProps() })

    const input = screen.getByPlaceholderText('Enter a new webhook URL to replace the stored one') as HTMLInputElement
    expect(input.value).toBe('')
  })

  it('submits the entered webhook URL on save', async () => {
    const onSave = vi.fn()
    render(AdminPanel, { props: { ...baseProps(), onSave } })

    const input = screen.getByPlaceholderText('Enter a new webhook URL to replace the stored one')
    await fireEvent.input(input, { target: { value: 'https://hooks.example.com/new' } })

    const form = input.closest('form')
    expect(form).not.toBeNull()
    await fireEvent.submit(form as HTMLFormElement)

    expect(onSave).toHaveBeenCalledTimes(1)
    const saved = onSave.mock.calls[0][0] as SettingsFile
    expect(saved.notifications?.webhook_url).toBe('https://hooks.example.com/new')
  })

  it('resets the input to empty when new (masked) settings arrive after save', async () => {
    const { rerender } = render(AdminPanel, { props: baseProps() })

    const input = screen.getByPlaceholderText('Enter a new webhook URL to replace the stored one') as HTMLInputElement
    await fireEvent.input(input, { target: { value: 'https://hooks.example.com/typed' } })
    expect(input.value).toBe('https://hooks.example.com/typed')

    // Server always returns a masked (empty) webhook_url, same as vault tokens.
    await rerender({ ...baseProps(), settings: settings({ notifications: { webhook_url: '' } }) })

    expect((screen.getByPlaceholderText('Enter a new webhook URL to replace the stored one') as HTMLInputElement).value).toBe('')
  })
})
