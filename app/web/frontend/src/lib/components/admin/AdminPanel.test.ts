// @vitest-environment jsdom
import { describe, it, expect, vi } from 'vitest'
import { render, screen, fireEvent, within } from '@testing-library/svelte'
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
})

describe('AdminPanel routed webhooks', () => {
  function notificationsSection(container: HTMLElement): HTMLElement {
    const section = container.querySelector('#notifications')
    if (!(section instanceof HTMLElement)) throw new Error('notifications section missing')
    return section
  }

  it('renders one row per configured target with its levels checked', () => {
    const { container } = render(
      AdminPanel,
      {
        props: baseProps({
          notifications: {
            webhook_url: '',
            webhooks: [
              { url: '', levels: ['critical'] },
              { url: '', levels: [] },
            ],
          },
        }),
      },
    )

    const section = within(notificationsSection(container))
    expect(section.getByText('Routed webhooks')).toBeInTheDocument()
    const checkboxes = section.getAllByRole('checkbox') as HTMLInputElement[]
    expect(checkboxes).toHaveLength(4)
    // First row: warning off, critical on. Second row (all tiers): both on.
    expect(checkboxes.map((c) => c.checked)).toEqual([false, true, true, true])
  })

  it('adds a row defaulting to all tiers and removes rows', async () => {
    const { container } = render(AdminPanel, { props: baseProps() })
    const section = () => within(notificationsSection(container))

    await fireEvent.click(section().getByRole('button', { name: '+ Add webhook' }))
    const checkboxes = section().getAllByRole('checkbox') as HTMLInputElement[]
    expect(checkboxes).toHaveLength(2)
    expect(checkboxes.every((c) => c.checked)).toBe(true)

    await fireEvent.click(section().getByRole('button', { name: '+ Add webhook' }))
    expect(section().getAllByRole('checkbox')).toHaveLength(4)

    await fireEvent.click(section().getAllByRole('button', { name: 'Remove' })[0])
    expect(section().getAllByRole('checkbox')).toHaveLength(2)
  })

  it('submits levels as a subset and omits them when all tiers apply', async () => {
    const onSave = vi.fn()
    const { container } = render(AdminPanel, { props: { ...baseProps(), onSave } })
    const section = within(notificationsSection(container))

    await fireEvent.click(section.getByRole('button', { name: '+ Add webhook' }))
    const checkboxes = section.getAllByRole('checkbox') as HTMLInputElement[]
    // Uncheck warning on the new row: critical-only.
    await fireEvent.click(checkboxes[0])

    const form = section.getByPlaceholderText('Enter a new webhook URL to replace the stored one').closest('form')
    await fireEvent.submit(form as HTMLFormElement)

    const saved = onSave.mock.calls[0][0] as SettingsFile
    expect(saved.notifications?.webhooks).toEqual([{ url: '', levels: ['critical'] }])
  })
})

describe('AdminPanel masked round-trip', () => {
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
