<script lang="ts">
  type Props = {
    message: string
    variant?: 'error' | 'warning'
    onDismiss?: () => void
  }

  const { message, variant = 'error', onDismiss }: Props = $props()
</script>

<div class="vcv-error-banner vcv-error-banner-{variant}" role="alert" aria-live="assertive">
  <span class="vcv-error-banner-text">{message}</span>
  {#if onDismiss}
    <button type="button" class="vcv-error-banner-close" aria-label="Dismiss" onclick={onDismiss}>×</button>
  {/if}
</div>
