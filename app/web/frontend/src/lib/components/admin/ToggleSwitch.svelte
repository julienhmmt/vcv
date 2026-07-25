<script lang="ts">
  interface Props {
    checked: boolean
    onCheckedChange: (checked: boolean) => void
    name?: string
    disabled?: boolean
  }

  const { checked, onCheckedChange, name, disabled = false }: Props = $props()
</script>

<span class="tgl-switch">
  <input
    type="checkbox"
    class="tgl-input"
    {name}
    {checked}
    {disabled}
    onchange={(event) => onCheckedChange((event.target as HTMLInputElement).checked)}
  />
  <span class="tgl-track" aria-hidden="true">
    <span class="tgl-thumb"></span>
  </span>
</span>

<style>
  .tgl-switch {
    position: relative;
    display: inline-flex;
    align-items: center;
    width: 2.25rem;
    height: 1.25rem;
    flex-shrink: 0;
  }

  .tgl-input {
    position: absolute;
    inset: 0;
    margin: 0;
    opacity: 0;
    cursor: pointer;
    z-index: 1;
  }

  .tgl-track {
    display: block;
    width: 100%;
    height: 100%;
    background: var(--vcv-color-border-strong);
    border-radius: var(--vcv-radius-full);
    transition: background-color 0.15s;
  }

  .tgl-thumb {
    display: block;
    width: 1rem;
    height: 1rem;
    margin: 0.125rem;
    background: var(--vcv-color-surface);
    border-radius: 50%;
    box-shadow: var(--vcv-shadow-button);
    transition: transform 0.15s;
  }

  .tgl-input:checked + .tgl-track {
    background: var(--vcv-color-primary);
  }

  .tgl-input:checked + .tgl-track .tgl-thumb {
    transform: translateX(1rem);
  }

  .tgl-input:focus-visible + .tgl-track {
    outline: none;
    box-shadow: 0 0 0 3px var(--vcv-color-focus-ring);
  }

  .tgl-input:disabled {
    cursor: not-allowed;
  }

  .tgl-input:disabled + .tgl-track {
    opacity: 0.6;
  }
</style>
