<script lang="ts">
  import { getI18n } from '$lib/stores/i18n.svelte'

  type Props = {
    message: string
    variant?: 'error' | 'warning'
    onDismiss?: () => void
  }

  const { message, variant = 'error', onDismiss }: Props = $props()
  const i18n = getI18n()
</script>

<div class="vcv-error-banner vcv-error-banner-{variant}" role="alert" aria-live="assertive">
  <span class="vcv-error-banner-text">{message}</span>
  {#if onDismiss}
    <button type="button" class="vcv-error-banner-close" aria-label={i18n.t('buttonClose', 'Dismiss')} onclick={onDismiss}>×</button>
  {/if}
</div>
