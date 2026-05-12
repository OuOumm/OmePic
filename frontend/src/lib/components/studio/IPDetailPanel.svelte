<script lang="ts">
  import { Ban, ShieldCheck, Trash2, X } from 'lucide-svelte';
  import { adminDeleteIPBan, adminDeleteIPBanImages, adminGetAbuseIPDetail } from '@/api';
  import { attachAccessibleDialog } from '@/actions/accessible-dialog';
  import ConfirmDialog from './ConfirmDialog.svelte';
  import { t } from '@/i18n';
  import { preferences } from '@/stores/preferences.svelte';
  import { toast } from '@/stores/toast.svelte';
  import { formatBytes, isAbortError } from '@/utils';
  import { errorMessage, runAsyncAction } from '@/ui-errors';
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
  let purgeOpen = $state(false);

  async function load(signal?: AbortSignal) {
    if (!preferences.adminToken || !ip) {
      detail = null;
      return;
    }
    loading = true;
    error = '';
    try {
      const nextDetail = await adminGetAbuseIPDetail(preferences.adminToken, ip, signal);
      if (!signal?.aborted) detail = nextDetail;
    } catch (err) {
      if (isAbortError(err)) return;
      error = errorMessage(err, preferences.language);
      toast.error(error);
    } finally {
      if (!signal?.aborted) loading = false;
    }
  }

  function requestBan() {
    if (!ip) return;
    onBan({ ip, label: detail?.ip_address_masked ?? ip });
  }

  async function unban() {
    const token = preferences.adminToken;
    const banId = detail?.ban?.id;
    if (!token || !banId) return;
    await runAsyncAction({
      language: preferences.language,
      setBusy: (value) => (busy = value),
      successMessage: t(preferences.language, 'common.success'),
      action: () => adminDeleteIPBan(token, banId),
      onSuccess: async () => {
        await load();
        onChanged();
      },
    });
  }

  async function purgeImages() {
    const token = preferences.adminToken;
    const banId = detail?.ban?.id;
    if (!token || !banId) return;
    await runAsyncAction({
      language: preferences.language,
      setBusy: (value) => (busy = value),
      successMessage: (result) => t(preferences.language, 'admin.securityDeletedImages', { count: result.deleted_count }),
      action: () => adminDeleteIPBanImages(token, banId),
      onSuccess: async () => {
        purgeOpen = false;
        await load();
        onChanged();
      },
    });
  }

  $effect(() => {
    const controller = new AbortController();
    void load(controller.signal);
    return () => controller.abort();
  });
</script>

{#if ip}
  <div class="fixed inset-0 z-[70] grid place-items-center p-4" role="dialog" aria-modal="true" aria-labelledby="ip-detail-title" tabindex="-1" {@attach attachAccessibleDialog(() => ({ onClose }))}>
    <button class="absolute inset-0 cursor-default bg-[hsl(var(--ink))]/35 backdrop-blur-[2px]" type="button" onclick={onClose} aria-label={t(preferences.language, 'common.close')}></button>
    <div class="studio-panel relative max-h-[calc(100dvh-3rem)] w-full max-w-2xl overflow-y-auto p-5 rotate-[0.25deg] sketch-enter">
      <div class="mb-5 flex items-start justify-between gap-3 border-b-[3px] ink-line pb-3">
        <div>
          <span class="tape-label rotate-[-2deg]">{t(preferences.language, 'admin.ipDetail')}</span>
          <h2 id="ip-detail-title" class="mt-3 break-all text-3xl font-black">{detail?.ip_address_masked ?? ip}</h2>
        </div>
        <button class="studio-button p-2" type="button" onclick={onClose} aria-label={t(preferences.language, 'common.close')}><X class="size-4" /></button>
      </div>

    {#if loading}
      <p class="font-black">{t(preferences.language, 'common.loading')}</p>
    {:else if error}
      <div class="grid gap-4">
        <div class="border-[3px] border-dashed ink-line p-4">
          <h3 class="text-xl font-black">{t(preferences.language, 'admin.ipDetailLoadError')}</h3>
          <p class="mt-2 text-sm font-semibold text-[hsl(var(--ink-muted))]">{error}</p>
        </div>
        <button class="studio-button" data-tone="danger" type="button" disabled={busy} onclick={requestBan}><Ban class="size-4" />{t(preferences.language, 'admin.ipBanAnyway')}</button>
      </div>
    {:else if detail}
      <div class="grid gap-4">
        <div class="grid gap-3 sm:grid-cols-2">
          <div class="border-y-[3px] ink-line py-4"><p class="text-xs font-black uppercase">{t(preferences.language, 'admin.securityUploads')}</p><p class="text-3xl font-black">{detail.upload_count}</p></div>
          <div class="border-y-[3px] ink-line py-4"><p class="text-xs font-black uppercase">{t(preferences.language, 'admin.securitySize')}</p><p class="text-3xl font-black">{formatBytes(detail.total_size, preferences.language)}</p></div>
        </div>
        <section class="blueprint-grid border-[3px] ink-line p-4">
          <h3 class="flex items-center gap-2 text-xl font-black">{detail.is_banned ? t(preferences.language, 'admin.securityBanned') : t(preferences.language, 'admin.securityNotBanned')} {#if detail.is_banned}<Ban class="size-5" />{:else}<ShieldCheck class="size-5" />{/if}</h3>
          {#if detail.ban}
            <p class="mt-2 text-sm font-semibold text-[hsl(var(--ink-muted))]">{detail.ban.reason}</p>
            <p class="mt-1 text-xs font-bold text-[hsl(var(--ink-muted))]">{detail.ban.expires_at ?? t(preferences.language, 'admin.securityNeverExpires')}</p>
          {/if}
        </section>
        <div class="flex flex-wrap gap-3">
          {#if detail.ban}
            <button class="studio-button" type="button" disabled={busy} onclick={unban}><Ban class="size-4" />{t(preferences.language, 'admin.securityUnban')}</button>
            <button class="studio-button" data-tone="danger" type="button" disabled={busy} onclick={() => (purgeOpen = true)}><Trash2 class="size-4" />{t(preferences.language, 'admin.securityPurge')}</button>
          {:else}
            <button class="studio-button" data-tone="danger" type="button" disabled={busy} onclick={requestBan}><Ban class="size-4" />{t(preferences.language, 'admin.securityBanIp')}</button>
          {/if}
        </div>
      </div>
    {/if}
    </div>
    <ConfirmDialog
      open={purgeOpen}
      title={t(preferences.language, 'admin.securityDeleteIpImagesConfirm')}
      description={detail?.ip_address_masked ?? ip}
      confirmLabel={t(preferences.language, 'admin.securityPurge')}
      cancelLabel={t(preferences.language, 'common.cancel')}
      {busy}
      onClose={() => (purgeOpen = false)}
      onConfirm={purgeImages}
    />
  </div>
{/if}
