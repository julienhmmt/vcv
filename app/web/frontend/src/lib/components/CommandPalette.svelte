<script lang="ts">
  import Moon from '@lucide/svelte/icons/moon'
  import Sun from '@lucide/svelte/icons/sun'
  import Shield from '@lucide/svelte/icons/shield'
  import Languages from '@lucide/svelte/icons/languages'
  import FileBadge from '@lucide/svelte/icons/file-badge'
  import Settings from '@lucide/svelte/icons/settings'
  import * as Command from '$lib/components/ui/command'
  import { getI18n, LANGUAGES } from '$lib/stores/i18n.svelte'
  import type { Certificate, CertStatus } from '$lib/types'

  interface Props {
    open: boolean
    onOpenChange: (open: boolean) => void
    certs: Certificate[]
    theme: 'light' | 'dark'
    onSelectCert: (cert: Certificate) => void
    onToggleStatus: (status: CertStatus) => void
    onToggleTheme: () => void
    onSetLang: (code: string) => void
  }

  const { open, onOpenChange, certs, theme, onSelectCert, onToggleStatus, onToggleTheme, onSetLang }: Props =
    $props()

  const i18n = getI18n()

  let query = $state('')

  const STATUS_KEYS: CertStatus[] = ['valid', 'warning', 'critical', 'expired', 'revoked']
  const statusLabels = $derived<Record<CertStatus, string>>({
    valid: i18n.t('statusLabelValid', 'Valid'),
    warning: i18n.t('statusLabelWarning', 'Warning'),
    critical: i18n.t('statusLabelCritical', 'Critical'),
    expired: i18n.t('statusLabelExpired', 'Expired'),
    revoked: i18n.t('statusLabelRevoked', 'Revoked'),
  })

  /** Cert matches when the query appears in its CN, serial, or any SAN. Capped for performance. */
  const matchingCerts = $derived.by(() => {
    const q = query.trim().toLowerCase()
    if (!q) return []
    const out: Certificate[] = []
    for (const cert of certs) {
      const haystack = `${cert.commonName} ${cert.serialNumber} ${cert.sans.join(' ')}`.toLowerCase()
      if (haystack.includes(q)) out.push(cert)
      if (out.length >= 25) break
    }
    return out
  })

  function run(action: () => void): void {
    action()
    query = ''
    onOpenChange(false)
  }
</script>

<Command.Dialog
  {open}
  onOpenChange={(value) => {
    if (!value) query = ''
    onOpenChange(value)
  }}
  shouldFilter={false}
  title={i18n.t('commandPaletteTitle', 'Command palette')}
  description={i18n.t('commandPaletteHint', 'Jump to a certificate or run a command')}
>
  <Command.Input bind:value={query} placeholder={i18n.t('commandPalettePlaceholder', 'Search certificates or commands…')} />
  <Command.List>
    <Command.Empty>{i18n.t('commandPaletteEmpty', 'No results found.')}</Command.Empty>

    {#if matchingCerts.length > 0}
      <Command.Group heading={i18n.t('commandPaletteCertsGroup', 'Certificates')}>
        {#each matchingCerts as cert (cert.id)}
          <Command.Item value={cert.id} onSelect={() => run(() => onSelectCert(cert))}>
            <FileBadge class="h-4 w-4" />
            <span>{cert.commonName || cert.serialNumber || '—'}</span>
          </Command.Item>
        {/each}
      </Command.Group>
    {/if}

    <Command.Group heading={i18n.t('commandPaletteFiltersGroup', 'Filter by status')}>
      {#each STATUS_KEYS as status (status)}
        <Command.Item value={`status-${status}`} onSelect={() => run(() => onToggleStatus(status))}>
          <Shield class="h-4 w-4" />
          <span>{statusLabels[status]}</span>
        </Command.Item>
      {/each}
    </Command.Group>

    <Command.Group heading={i18n.t('commandPaletteActionsGroup', 'Actions')}>
      <Command.Item value="toggle-theme" onSelect={() => run(onToggleTheme)}>
        {#if theme === 'dark'}<Sun class="h-4 w-4" />{:else}<Moon class="h-4 w-4" />{/if}
        <span>{i18n.t('buttonToggleTheme', 'Toggle dark mode')}</span>
      </Command.Item>
      <Command.Item value="open-admin" onSelect={() => run(() => (window.location.href = '/admin'))}>
        <Settings class="h-4 w-4" />
        <span>{i18n.t('commandPaletteOpenAdmin', 'Open admin panel')}</span>
      </Command.Item>
    </Command.Group>

    <Command.Group heading={i18n.t('labelLanguage', 'Language')}>
      {#each LANGUAGES as language (language.code)}
        <Command.Item value={`lang-${language.code}`} onSelect={() => run(() => onSetLang(language.code))}>
          <Languages class="h-4 w-4" />
          <span>{language.name}</span>
        </Command.Item>
      {/each}
    </Command.Group>
  </Command.List>
</Command.Dialog>
