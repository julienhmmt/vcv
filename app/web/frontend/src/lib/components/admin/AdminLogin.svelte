<script lang="ts">
  import { Button } from '$lib/components/ui/button'
  import { Card, CardContent, CardHeader, CardTitle } from '$lib/components/ui/card'
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

  let username = $state('admin')
  let password = $state('')

  function submit(event: SubmitEvent): void {
    event.preventDefault()
    onSubmit(username, password)
  }
</script>

<Card class="mx-auto max-w-sm">
  <CardHeader>
    <CardTitle>{i18n.t('adminLogin', 'Admin Sign In')}</CardTitle>
  </CardHeader>
  <CardContent>
    <form class="space-y-4" onsubmit={submit}>
      <div class="space-y-2">
        <Label for="username">{i18n.t('adminUsername', 'Username')}</Label>
        <Input id="username" type="text" bind:value={username} required autocomplete="username" />
      </div>
      <div class="space-y-2">
        <Label for="password">{i18n.t('adminPassword', 'Password')}</Label>
        <Input id="password" type="password" bind:value={password} required autocomplete="current-password" />
      </div>
      {#if error}
        <p class="text-sm text-destructive">{error}</p>
      {/if}
      <p class="text-xs text-muted-foreground">{i18n.t('adminLoginHint', '')}</p>
      <Button type="submit" class="w-full" disabled={loading}>
        {loading ? i18n.t('adminSigningIn', 'Signing in…') : i18n.t('adminLogin', 'Sign In')}
      </Button>
    </form>
  </CardContent>
</Card>
