// @vitest-environment jsdom
import { describe, it, expect, vi } from 'vitest'
import { render, screen, fireEvent, within, waitFor } from '@testing-library/svelte'
import type { SettingsFile, VaultInstance } from '$lib/types'

vi.mock('$lib/stores/i18n.svelte', () => ({
  getI18n: () => ({
    t: (_key: string, fallback?: string, params?: Record<string, string>) =>
      (fallback ?? _key).replace(/\{(\w+)\}/g, (match, name: string) => params?.[name] ?? match),
  }),
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
    onAddVault: vi.fn(async () => null),
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

function vault(id: string, displayName: string): VaultInstance {
  return { id, display_name: displayName, address: `https://${id}.example`, pki_mounts: ['pki'], pki_mount: 'pki', enabled: true }
}

describe('AdminPanel unsaved edits survive vault add and remove', () => {
  it('adding a vault keeps an unsaved threshold edit and lists the new vault', async () => {
    const onAddVault = vi.fn(async () => vault('', 'Staging'))
    render(AdminPanel, { props: { ...baseProps(), onAddVault } })

    await fireEvent.input(screen.getByDisplayValue('30'), { target: { value: '14' } })
    await fireEvent.click(screen.getByRole('button', { name: '+ Add vault' }))

    expect(onAddVault).toHaveBeenCalledTimes(1)
    expect(await screen.findByText('Staging')).toBeInTheDocument()
    expect(screen.getByDisplayValue('14')).toBeInTheDocument()
  })

  it('Cancel in the remove confirmation keeps the vault', async () => {
    render(AdminPanel, { props: baseProps({ vaults: [vault('prod', 'Prod')] }) })

    await fireEvent.click(screen.getByRole('button', { name: 'Remove vault: Prod' }))
    const dialog = await screen.findByRole('dialog')
    expect(within(dialog).getByText('Remove vault Prod? It is deleted when you save.')).toBeInTheDocument()
    await fireEvent.click(within(dialog).getByRole('button', { name: 'Cancel' }))

    await waitFor(() => expect(screen.queryByRole('dialog')).not.toBeInTheDocument())
    expect(screen.getByText('Prod')).toBeInTheDocument()
  })

  it('Remove after confirm drops the vault from the form and Save sends the reduced list', async () => {
    const onSave = vi.fn()
    render(AdminPanel, { props: { ...baseProps({ vaults: [vault('prod', 'Prod'), vault('edge', 'Edge')] }), onSave } })

    await fireEvent.click(screen.getByRole('button', { name: 'Remove vault: Prod' }))
    const dialog = await screen.findByRole('dialog')
    await fireEvent.click(within(dialog).getByRole('button', { name: 'Remove' }))

    await waitFor(() => expect(screen.queryByText('Prod')).not.toBeInTheDocument())
    expect(screen.getByText('Edge')).toBeInTheDocument()

    const form = screen.getByPlaceholderText('Enter a new webhook URL to replace the stored one').closest('form')
    await fireEvent.submit(form as HTMLFormElement)
    const saved = onSave.mock.calls[0][0] as SettingsFile
    expect(saved.vaults.map((v) => v.id)).toEqual(['edge'])
  })
})
