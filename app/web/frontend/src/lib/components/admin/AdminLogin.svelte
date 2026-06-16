<script lang="ts">
  import { Button } from '$lib/components/ui/button'
  import { Input } from '$lib/components/ui/input'
  import { Label } from '$lib/components/ui/label'
  import { getI18n } from '$lib/stores/i18n.svelte'

  interface Props {
    loading: boolean
    error: string | null
    onSubmit: (username: string, password: string) => void
  }

  const { loading, error, onSubmit }: Props = $props()
  const i18n = getI18n()

  let username = $state('')
  let password = $state('')

  function submit(event: SubmitEvent): void {
    event.preventDefault()
    onSubmit(username, password)
    password = ''
  }
</script>

<div class="admin-login-root">
  <div class="admin-login-brand">
    <div class="admin-login-icon">
      <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">
        <rect x="3" y="11" width="18" height="11" rx="2" ry="2"/>
        <path d="M7 11V7a5 5 0 0 1 10 0v4"/>
      </svg>
    </div>
    <div class="admin-login-wordmark">VCV Admin</div>
    <div class="admin-login-path">/admin</div>
  </div>

  <form class="admin-login-form" onsubmit={submit} autocomplete="on">
    <div class="admin-login-field">
      <Label for="username" class="admin-login-label">{i18n.t('adminUsername', 'Username')}</Label>
      <Input
        id="username"
        type="text"
        bind:value={username}
        required
        autocomplete="username"
        class="admin-login-input"
      />
    </div>

    <div class="admin-login-field">
      <Label for="password" class="admin-login-label">{i18n.t('adminPassword', 'Password')}</Label>
      <Input
        id="password"
        type="password"
        bind:value={password}
        required
        autocomplete="current-password"
        class="admin-login-input"
      />
    </div>

    {#if error}
      <p class="admin-login-error" role="alert">{error}</p>
    {/if}

    <Button type="submit" class="admin-login-submit" disabled={loading}>
      {loading ? i18n.t('adminSigningIn', 'Signing in…') : i18n.t('adminLogin', 'Sign In')}
    </Button>

    <a href="/" class="admin-login-back">{i18n.t('adminBackToVCV', 'Back to VCV')}</a>
  </form>
</div>

<style>
  .admin-login-root {
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    min-height: 100svh;
    padding: 2rem 1rem;
    background: var(--vcv-color-bg);
  }

  .admin-login-brand {
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 0.5rem;
    margin-bottom: 2.5rem;
  }

  .admin-login-icon {
    width: 2.5rem;
    height: 2.5rem;
    color: var(--vcv-color-primary);
    margin-bottom: 0.25rem;
  }

  .admin-login-icon svg {
    width: 100%;
    height: 100%;
  }

  .admin-login-wordmark {
    font-size: 1rem;
    font-weight: 600;
    letter-spacing: 0.02em;
    color: var(--vcv-color-text-strong);
  }

  .admin-login-path {
    font-family: ui-monospace, SFMono-Regular, Menlo, monospace;
    font-size: 0.7rem;
    color: var(--vcv-color-muted);
    letter-spacing: 0.04em;
  }

  .admin-login-form {
    width: 100%;
    max-width: 22rem;
    display: flex;
    flex-direction: column;
    gap: 1rem;
  }

  .admin-login-field {
    display: flex;
    flex-direction: column;
    gap: 0.375rem;
  }

  :global(.admin-login-label) {
    font-size: 0.75rem;
    font-weight: 500;
    letter-spacing: 0.04em;
    text-transform: uppercase;
    color: var(--vcv-color-muted);
  }

  :global(.admin-login-input) {
    background: var(--vcv-color-surface);
    border-color: var(--vcv-color-border-strong);
    font-size: 0.875rem;
  }

  .admin-login-error {
    font-size: 0.8rem;
    color: var(--vcv-color-danger);
    padding: 0.5rem 0.75rem;
    background: var(--vcv-color-danger-surface);
    border: 1px solid var(--vcv-color-danger-border);
    border-radius: var(--vcv-radius-md);
    margin: 0;
  }

  :global(.admin-login-submit) {
    width: 100%;
    margin-top: 0.25rem;
  }

  .admin-login-back {
    display: block;
    text-align: center;
    font-size: 0.75rem;
    color: var(--vcv-color-muted);
    text-decoration: none;
    margin-top: 0.5rem;
    transition: color 0.12s;
  }

  .admin-login-back:hover {
    color: var(--vcv-color-text);
  }
</style>
