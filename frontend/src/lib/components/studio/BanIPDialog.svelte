<script lang="ts">
  import { Ban, X } from 'lucide-svelte';
  import { t } from '@/i18n';
  import { preferences } from '@/stores/preferences.svelte';

  type BanDialogTarget = {
    ip: string;
    label?: string;
  };

  type BanDialogValue = {
    ip: string;
    reason: string;
    durationHours: number | null;
  };

  type Props = {
    target: BanDialogTarget | null;
    busy?: boolean;
    onClose: () => void;
    onConfirm: (value: BanDialogValue) => void;
  };

  let { target, busy = false, onClose, onConfirm }: Props = $props();
  let reason = $state('manual review');
  let selectedPreset = $state('24');
  let customHours = $state(24);

  const presets = [
    { label: '1 hour', value: '1' },
    { label: '24 hours', value: '24' },
    { label: '7 days', value: '168' },
    { label: '30 days', value: '720' },
    { label: 'Permanent', value: 'permanent' },
    { label: 'Custom', value: 'custom' },
  ];

  const durationHours = $derived(
    selectedPreset === 'permanent' ? null : selectedPreset === 'custom' ? Math.max(1, Number(customHours) || 1) : Number(selectedPreset),
  );

  function confirmBan() {
    if (!target || !reason.trim()) return;
    onConfirm({ ip: target.ip, reason: reason.trim(), durationHours });
  }

  $effect(() => {
    if (target) {
      reason = 'manual review';
      selectedPreset = '24';
      customHours = 24;
    }
  });
</script>

{#if target}
  <div class="fixed inset-0 z-[90] grid min-h-dvh place-items-center overflow-y-auto bg-[hsl(var(--ink)/0.42)] p-4 backdrop-blur-sm sm:p-6" role="presentation" onclick={(event) => event.target === event.currentTarget && onClose()}>
    <div class="w-full max-w-lg max-h-[calc(100dvh-2rem)] overflow-y-auto border-[3px] ink-line bg-[hsl(var(--paper))] p-4 shadow-[8px_8px_0_hsl(var(--ink))] sketch-enter sm:max-h-[calc(100dvh-3rem)] sm:p-5" role="dialog" tabindex="-1" aria-modal="true" aria-labelledby="ban-dialog-title">
      <div class="mb-5 flex items-start justify-between gap-3 border-b-[3px] ink-line pb-3">
        <div>
          <span class="tape-label rotate-[-2deg]" style="background:hsl(var(--marker-pink))">moderation</span>
          <h2 id="ban-dialog-title" class="mt-3 text-3xl font-black">Ban IP</h2>
          <p class="mt-1 break-all text-sm font-bold text-[hsl(var(--ink-muted))]">{target.label ?? target.ip}</p>
        </div>
        <button class="studio-button p-2" type="button" onclick={onClose} aria-label="close"><X class="size-4" /></button>
      </div>

      <label class="grid gap-2 text-sm font-black">
        Ban reason
        <textarea class="studio-input min-h-24" bind:value={reason}></textarea>
      </label>

      <div class="mt-5 grid gap-3">
        <p class="text-sm font-black">Ban duration</p>
        <div class="grid grid-cols-2 gap-2 sm:grid-cols-3">
          {#each presets as preset (preset.value)}
            <button class="studio-button p-2 text-xs" data-tone={selectedPreset === preset.value ? 'danger' : 'blue'} type="button" onclick={() => (selectedPreset = preset.value)}>{preset.label}</button>
          {/each}
        </div>
        {#if selectedPreset === 'custom'}
          <label class="grid gap-2 text-sm font-black">
            Custom hours
            <input class="studio-input" type="number" min="1" bind:value={customHours} />
          </label>
        {/if}
      </div>

      <div class="mt-6 flex flex-wrap justify-end gap-3 border-t-[3px] ink-line pt-4">
        <button class="studio-button" type="button" onclick={onClose}>Cancel</button>
        <button class="studio-button" data-tone="danger" type="button" disabled={busy || !reason.trim()} onclick={confirmBan}><Ban class="size-4" />{t(preferences.language, 'common.confirm')}</button>
      </div>
    </div>
  </div>
{/if}
