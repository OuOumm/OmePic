<script lang="ts">
  import { Ban, ShieldCheck, Trash2, X } from 'lucide-svelte';
  import { adminDeleteIPBan, adminDeleteIPBanImages, adminGetAbuseIPDetail } from '@/api';
  import { t } from '@/i18n';
  import { preferences } from '@/stores/preferences.svelte';
  import { toast } from '@/stores/toast.svelte';
  import { formatBytes } from '@/utils';
  import type { AdminAbuseIPDetail } from '@/types';

  type Props = {
    ip: string | null;
    onClose: () => void;
    onChanged: () => void;
    onBan: (target: { ip: string; label?: string }) => void;
  };

  let { ip, onClose, onChanged, onBan }: Props = $props();
  let detail = $state<AdminAbuseIPDetail | null>(null);
  let error = $state('');
  let loading = $state(false);
  let busy = $state(false);

  async function load() {
    if (!preferences.adminToken || !ip) {
      detail = null;
      return;
    }
    loading = true;
    error = '';
    try {
      detail = await adminGetAbuseIPDetail(preferences.adminToken, ip);
    } catch (err) {
      error = err instanceof Error ? err.message : t(preferences.language, 'common.error');
      toast.error(error);
    } finally {
      loading = false;
    }
  }

  function requestBan() {
    if (!ip) return;
    onBan({ ip, label: detail?.ip_address_masked ?? ip });
  }

  async function unban() {
    if (!preferences.adminToken || !detail?.ban) return;
    busy = true;
    try {
      await adminDeleteIPBan(preferences.adminToken, detail.ban.id);
      toast.success(t(preferences.language, 'common.success'));
      await load();
      onChanged();
    } catch (err) {
      toast.error(err instanceof Error ? err.message : t(preferences.language, 'common.error'));
    } finally {
      busy = false;
    }
  }

  async function purgeImages() {
    if (!preferences.adminToken || !detail?.ban || !confirm('Delete images uploaded from this IP?')) return;
    busy = true;
    try {
      const result = await adminDeleteIPBanImages(preferences.adminToken, detail.ban.id);
      toast.success(`Deleted ${result.deleted_count}`);
      await load();
      onChanged();
    } catch (err) {
      toast.error(err instanceof Error ? err.message : t(preferences.language, 'common.error'));
    } finally {
      busy = false;
    }
  }

  $effect(() => { load(); });
</script>

{#if ip}
  <div class="fixed inset-0 z-[70] bg-[hsl(var(--ink)/0.38)] backdrop-blur-sm" role="presentation" onclick={onClose}></div>
  <aside class="fixed right-0 top-0 z-[80] h-dvh w-full max-w-lg overflow-y-auto border-l-[3px] ink-line bg-[hsl(var(--paper))] p-5 shadow-[-8px_0_0_hsl(var(--ink))] sketch-enter">
    <div class="mb-5 flex items-start justify-between gap-3 border-b-[3px] ink-line pb-3">
      <div>
        <span class="tape-label rotate-[-2deg]">ip detail</span>
        <h2 class="mt-3 break-all text-3xl font-black">{detail?.ip_address_masked ?? ip}</h2>
      </div>
      <button class="studio-button p-2" type="button" onclick={onClose} aria-label="close"><X class="size-4" /></button>
    </div>

    {#if loading}
      <p class="font-black">{t(preferences.language, 'common.loading')}</p>
    {:else if error}
      <div class="grid gap-4">
        <div class="border-[3px] border-dashed ink-line p-4">
          <h3 class="text-xl font-black">Unable to load IP detail</h3>
          <p class="mt-2 text-sm font-semibold text-[hsl(var(--ink-muted))]">{error}</p>
        </div>
        <button class="studio-button" data-tone="danger" type="button" disabled={busy} onclick={requestBan}><Ban class="size-4" />Ban IP anyway</button>
      </div>
    {:else if detail}
      <div class="grid gap-4">
        <div class="grid gap-3 sm:grid-cols-2">
          <div class="border-y-[3px] ink-line py-4"><p class="text-xs font-black uppercase">Uploads</p><p class="text-3xl font-black">{detail.upload_count}</p></div>
          <div class="border-y-[3px] ink-line py-4"><p class="text-xs font-black uppercase">Size</p><p class="text-3xl font-black">{formatBytes(detail.total_size)}</p></div>
        </div>
        <section class="blueprint-grid border-[3px] ink-line p-4">
          <h3 class="flex items-center gap-2 text-xl font-black">{detail.is_banned ? 'Banned' : 'Not banned'} {#if detail.is_banned}<Ban class="size-5" />{:else}<ShieldCheck class="size-5" />{/if}</h3>
          {#if detail.ban}
            <p class="mt-2 text-sm font-semibold text-[hsl(var(--ink-muted))]">{detail.ban.reason}</p>
            <p class="mt-1 text-xs font-bold text-[hsl(var(--ink-muted))]">{detail.ban.expires_at ?? 'never expires'}</p>
          {/if}
        </section>
        <div class="flex flex-wrap gap-3">
          {#if detail.ban}
            <button class="studio-button" type="button" disabled={busy} onclick={unban}><Ban class="size-4" />Unban</button>
            <button class="studio-button" data-tone="danger" type="button" disabled={busy} onclick={purgeImages}><Trash2 class="size-4" />Purge images</button>
          {:else}
            <button class="studio-button" data-tone="danger" type="button" disabled={busy} onclick={requestBan}><Ban class="size-4" />Ban IP</button>
          {/if}
        </div>
      </div>
    {/if}
  </aside>
{/if}
