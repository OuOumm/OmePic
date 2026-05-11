<script lang="ts">
  import { AlertTriangle, X } from 'lucide-svelte';
  import { accessibleDialog } from '@/actions/accessible-dialog';

  type Props = {
    open: boolean;
    title: string;
    description?: string;
    confirmLabel: string;
    cancelLabel: string;
    busy?: boolean;
    tone?: 'danger' | 'primary';
    onConfirm: () => void;
    onClose: () => void;
  };

  let { open, title, description = '', confirmLabel, cancelLabel, busy = false, tone = 'danger', onConfirm, onClose }: Props = $props();

  function submitConfirm() {
    if (busy) return;
    onConfirm();
  }
</script>

{#if open}
  <div class="fixed inset-0 z-[110] grid min-h-dvh place-items-center overflow-x-hidden overflow-y-auto bg-[hsl(var(--ink)/0.48)] px-3 py-4 backdrop-blur-sm sm:p-6" role="presentation" onclick={(event) => event.target === event.currentTarget && !busy && onClose()}>
    <div class="w-[min(100%,calc(100vw-1.5rem))] max-w-md overflow-hidden border-[3px] ink-line bg-[hsl(var(--paper))] p-3 shadow-[5px_5px_0_hsl(var(--ink))] sketch-enter sm:p-5 sm:shadow-[8px_8px_0_hsl(var(--ink))]" role="dialog" tabindex="-1" aria-modal="true" aria-labelledby="confirm-dialog-title" aria-describedby={description ? 'confirm-dialog-description' : undefined} use:accessibleDialog={{ onClose: busy ? undefined : onClose }}>
      <div class="mb-4 flex min-w-0 items-start justify-between gap-2 border-b-[3px] ink-line pb-3 sm:mb-5 sm:gap-3">
        <div class="min-w-0 flex-1 overflow-hidden">
          <span class="tape-label max-w-full rotate-[-2deg] truncate" style={tone === 'danger' ? 'background:hsl(var(--marker-pink))' : 'background:hsl(var(--marker-blue))'}><AlertTriangle class="inline size-4" aria-hidden="true" /> {confirmLabel}</span>
          <h2 id="confirm-dialog-title" class="mt-3 min-w-0 overflow-wrap-anywhere text-xl font-black leading-tight sm:text-3xl">{title}</h2>
        </div>
        <button class="studio-button shrink-0 p-2" type="button" disabled={busy} onclick={onClose} aria-label={cancelLabel}><X class="size-4" aria-hidden="true" /></button>
      </div>

      {#if description}
        <p id="confirm-dialog-description" class="max-h-40 overflow-y-auto overflow-wrap-anywhere text-xs font-bold leading-relaxed text-[hsl(var(--ink-muted))] sm:text-sm">{description}</p>
      {/if}

      <div class="mt-5 grid grid-cols-1 gap-3 border-t-[3px] ink-line pt-4 sm:mt-6 sm:flex sm:flex-wrap sm:justify-end">
        <button class="studio-button w-full sm:w-auto" type="button" disabled={busy} onclick={onClose}>{cancelLabel}</button>
        <button class="studio-button w-full sm:w-auto" data-tone={tone} type="button" disabled={busy} onclick={submitConfirm}>{confirmLabel}</button>
      </div>
    </div>
  </div>
{/if}
