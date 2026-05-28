<script lang="ts">
  import { Button } from '$lib/components/ui/button'
  import { Card, CardContent, CardHeader, CardTitle } from '$lib/components/ui/card'
  import { Input } from '$lib/components/ui/input'
  import { Label } from '$lib/components/ui/label'

  interface Props {
    loading: boolean
    error: string | null
    onSubmit: (username: string, password: string) => void
  }

  const { loading, error, onSubmit }: Props = $props()

  let username = $state('admin')
  let password = $state('')

  function submit(event: SubmitEvent): void {
    event.preventDefault()
    onSubmit(username, password)
  }
</script>

<Card class="mx-auto max-w-sm">
  <CardHeader>
    <CardTitle>Admin Sign In</CardTitle>
  </CardHeader>
  <CardContent>
    <form class="space-y-4" onsubmit={submit}>
      <div class="space-y-2">
        <Label for="username">Username</Label>
        <Input id="username" type="text" bind:value={username} required autocomplete="username" />
      </div>
      <div class="space-y-2">
        <Label for="password">Password</Label>
        <Input id="password" type="password" bind:value={password} required autocomplete="current-password" />
      </div>
      {#if error}
        <p class="text-sm text-destructive">{error}</p>
      {/if}
      <Button type="submit" class="w-full" disabled={loading}>
        {loading ? 'Signing in…' : 'Sign In'}
      </Button>
    </form>
  </CardContent>
</Card>
