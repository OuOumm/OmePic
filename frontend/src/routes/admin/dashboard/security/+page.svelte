<script lang="ts">
  import { page } from '$app/state';
  import { Gauge, Save, ShieldCheck, Trash2, Unlock } from 'lucide-svelte';
  import BanIPDialog from '@/components/studio/BanIPDialog.svelte';
  import ConfirmDialog from '@/components/studio/ConfirmDialog.svelte';
  import MetricStrip from '@/components/studio/MetricStrip.svelte';
  import PageTitle from '@/components/studio/PageTitle.svelte';
  import { adminCreateIPBan, adminDeleteIPBan, adminDeleteIPBanImages, adminGetAbuseOverview, adminGetIPBans, adminGetSystemSettings, adminUpdateSystemSettings } from '@/api';
  import { t } from '@/i18n';
  import { preferences } from '@/stores/preferences.svelte';
  import { formatBytes, isAbortError } from '@/utils';
  import { runAsyncAction, toastApiError } from '@/ui-errors';
  import type { AdminAbuseOverview, AdminIPBan, AdminSystemSettings } from '@/types';

  let overview = $state.raw<AdminAbuseOverview | null>(null);
  let bans = $state.raw<AdminIPBan[]>([]);
  let system = $state<AdminSystemSettings | null>(null);
  let banTarget = $state<{ ip: string; label?: string } | null>(null);
  let confirmTarget = $state<{ action: 'unban' | 'purge'; ban: AdminIPBan } | null>(null);
  let banning = $state(false);
  let confirmBusy = $state(false);
  let savingRateLimit = $state(false);

  const activeTab = $derived(page.url.searchParams.get('tab') ?? 'abuse');
  const topIps = $derived(Array.isArray(overview?.top_ips) ? overview.top_ips : []);
  const safeBans = $derived(Array.isArray(bans) ? bans : []);
  const siteName = $derived(system?.runtime.site_name || preferences.runtimeSettings?.site.name || 'OmePic');

  async function loadAbuse(signal?: AbortSignal) {
    if (!preferences.adminToken) return;
    try {
      const [nextOverview, nextBans] = await Promise.all([adminGetAbuseOverview(preferences.adminToken, undefined, undefined, signal), adminGetIPBans(preferences.adminToken, signal)]);
      if (signal?.aborted) return;
      overview = nextOverview ? { ...nextOverview, top_ips: Array.isArray(nextOverview.top_ips) ? nextOverview.top_ips : [] } : null;
      bans = Array.isArray(nextBans) ? nextBans : [];
    } catch (err) {
      if (isAbortError(err)) return;
      toastApiError(err, preferences.language);
    }
  }

  async function loadRateLimit(signal?: AbortSignal) {
    if (!preferences.adminToken) return;
    try {
      const nextSystem = await adminGetSystemSettings(preferences.adminToken, signal);
      if (!signal?.aborted) system = nextSystem;
    } catch (err) {
      if (isAbortError(err)) return;
      toastApiError(err, preferences.language);
    }
  }

  async function load(signal?: AbortSignal) {
    if (activeTab === 'rate-limit') {
      await loadRateLimit(signal);
      return;
    }
    await loadAbuse(signal);
  }

  function normalizeLimit(value: number) {
    return Math.max(0, Math.trunc(Number.isFinite(value) ? value : 0));
  }

  function limitSummary(windowMinutes: number, maxRequests: number) {
    if (windowMinutes === 0 || maxRequests === 0) return t(preferences.language, 'admin.rateLimitDisabled');
    return t(preferences.language, 'admin.rateLimitSummary', { count: maxRequests, minutes: windowMinutes });
  }

  async function saveRateLimit() {
    const token = preferences.adminToken;
    if (!token || !system) return;
    await runAsyncAction({
      language: preferences.language,
      setBusy: (value) => (savingRateLimit = value),
      successMessage: t(preferences.language, 'common.success'),
      action: () => {
        if (!system) throw new Error(t(preferences.language, 'common.error'));
        system.runtime.rate_limit_window_minutes = normalizeLimit(system.runtime.rate_limit_window_minutes);
        system.runtime.rate_limit_max_requests = normalizeLimit(system.runtime.rate_limit_max_requests);
        system.runtime.upload_rate_limit_window_minutes = normalizeLimit(system.runtime.upload_rate_limit_window_minutes);
        system.runtime.upload_rate_limit_max_requests = normalizeLimit(system.runtime.upload_rate_limit_max_requests);
        return adminUpdateSystemSettings(token, system.runtime);
      },
      onSuccess: (nextSystem) => {
        system = nextSystem;
      },
    });
  }

  async function banIp(input: { ip: string; reason: string; durationHours: number | null }) {
    const token = preferences.adminToken;
    if (!token) return;
    await runAsyncAction({
      language: preferences.language,
      setBusy: (value) => (banning = value),
      successMessage: t(preferences.language, 'common.success'),
      action: () => adminCreateIPBan(token, { ip_address: input.ip, duration_hours: input.durationHours, reason: input.reason }),
      onSuccess: async () => {
        banTarget = null;
        await loadAbuse();
      },
    });
  }

  async function unban(id: number) {
    const token = preferences.adminToken;
    if (!token) return;
    await runAsyncAction({
      language: preferences.language,
      setBusy: (value) => (confirmBusy = value),
      successMessage: t(preferences.language, 'common.success'),
      action: () => adminDeleteIPBan(token, id),
      onSuccess: async () => {
        confirmTarget = null;
        await loadAbuse();
      },
    });
  }

  async function purgeImages(id: number) {
    const token = preferences.adminToken;
    if (!token) return;
    await runAsyncAction({
      language: preferences.language,
      setBusy: (value) => (confirmBusy = value),
      successMessage: (result) => t(preferences.language, 'admin.securityDeletedImages', { count: result.deleted_count }),
      action: () => adminDeleteIPBanImages(token, id),
      onSuccess: async () => {
        confirmTarget = null;
        await loadAbuse();
      },
    });
  }

  $effect(() => {
    const controller = new AbortController();
    void load(controller.signal);
    return () => controller.abort();
  });
</script>

<svelte:head><title>{t(preferences.language, 'admin.abuseTitle')} · {siteName}</title></svelte:head>

<div class="min-w-0 space-y-6 overflow-hidden">
  {#if activeTab === 'rate-limit'}
    <PageTitle eyebrow={t(preferences.language, 'admin.submenuRateLimit')} title={t(preferences.language, 'admin.rateLimitTitle')} subtitle={t(preferences.language, 'admin.rateLimitDescription')} tone="blue" />
    {#if system}
      <div class="mt-6 flex justify-end border-b-[3px] ink-line pb-3">
        <button class="studio-button w-full sm:w-fit" data-tone="primary" type="button" disabled={savingRateLimit} onclick={saveRateLimit}><Save class="size-4" />{t(preferences.language, 'common.save')}</button>
      </div>

      <div class="grid gap-4 lg:grid-cols-2">
        <section class="grid gap-4 rounded-none border-2 ink-line bg-[hsl(var(--paper))] p-4">
          <div class="flex min-w-0 items-start gap-3">
            <span class="grid size-10 shrink-0 place-items-center border-2 ink-line bg-[hsl(var(--marker-yellow))] text-[hsl(var(--marker-ink))]"><Gauge class="size-5" /></span>
            <div class="min-w-0">
              <span class="tape-label rotate-[-1deg]" style="background:hsl(var(--marker-yellow))">{t(preferences.language, 'admin.rateLimitApi')}</span>
              <p class="mt-3 overflow-wrap-anywhere text-sm font-bold text-[hsl(var(--ink-muted))]">{limitSummary(system.runtime.rate_limit_window_minutes, system.runtime.rate_limit_max_requests)}</p>
            </div>
          </div>
          <label class="grid gap-2 text-sm font-black">
            {t(preferences.language, 'admin.rateLimitWindowMinutes')}
            <input class="studio-input" type="number" min="0" inputmode="numeric" bind:value={system.runtime.rate_limit_window_minutes} />
          </label>
          <label class="grid gap-2 text-sm font-black">
            {t(preferences.language, 'admin.rateLimitMaxRequests')}
            <input class="studio-input" type="number" min="0" inputmode="numeric" bind:value={system.runtime.rate_limit_max_requests} />
          </label>
        </section>

        <section class="grid gap-4 rounded-none border-2 ink-line bg-[hsl(var(--paper))] p-4">
          <div class="flex min-w-0 items-start gap-3">
            <span class="grid size-10 shrink-0 place-items-center border-2 ink-line bg-[hsl(var(--marker-pink))] text-[hsl(var(--marker-ink))]"><Gauge class="size-5" /></span>
            <div class="min-w-0">
              <span class="tape-label rotate-[1deg]" style="background:hsl(var(--marker-pink))">{t(preferences.language, 'admin.rateLimitUpload')}</span>
              <p class="mt-3 overflow-wrap-anywhere text-sm font-bold text-[hsl(var(--ink-muted))]">{limitSummary(system.runtime.upload_rate_limit_window_minutes, system.runtime.upload_rate_limit_max_requests)}</p>
            </div>
          </div>
          <label class="grid gap-2 text-sm font-black">
            {t(preferences.language, 'admin.rateLimitWindowMinutes')}
            <input class="studio-input" type="number" min="0" inputmode="numeric" bind:value={system.runtime.upload_rate_limit_window_minutes} />
          </label>
          <label class="grid gap-2 text-sm font-black">
            {t(preferences.language, 'admin.rateLimitMaxRequests')}
            <input class="studio-input" type="number" min="0" inputmode="numeric" bind:value={system.runtime.upload_rate_limit_max_requests} />
          </label>
        </section>
      </div>

      <p class="border-l-[6px] ink-line bg-[hsl(var(--paper))] p-4 text-sm font-bold text-[hsl(var(--ink-muted))]">{t(preferences.language, 'admin.rateLimitZeroHint')}</p>
    {/if}
  {:else}
    <PageTitle eyebrow={t(preferences.language, 'admin.securityEyebrow')} title={t(preferences.language, 'admin.abuseTitle')} subtitle={t(preferences.language, 'admin.abuseDescription')} tone="pink" />
    {#if overview}
      <div class="grid gap-4 md:grid-cols-3">
        <MetricStrip label={t(preferences.language, 'admin.securityUploads')} value={overview.upload_count} tone="yellow" />
        <MetricStrip label={t(preferences.language, 'admin.securitySize')} value={formatBytes(overview.upload_size, preferences.language)} tone="blue" />
        <MetricStrip label={t(preferences.language, 'admin.securityBans')} value={overview.active_ip_ban_count} tone="pink" />
      </div>
      <div class="grid min-w-0 gap-8 lg:grid-cols-[minmax(0,1fr)_minmax(0,1fr)]">
        <section class="min-w-0 overflow-hidden">
          <div class="mb-3 flex items-center justify-between border-b-[3px] ink-line pb-2">
            <h2 class="text-2xl font-black">{t(preferences.language, 'admin.securityTopIps')}</h2>
          </div>
          <div class="w-full min-w-0 max-w-full touch-pan-x overflow-x-auto overscroll-x-contain [-webkit-overflow-scrolling:touch]">
            <table class="w-full min-w-[420px] border-collapse text-sm">
              <thead>
                <tr class="border-b-[3px] ink-line text-left text-xs font-black uppercase text-[hsl(var(--ink-muted))]">
                  <th class="px-3 py-2" scope="col">{t(preferences.language, 'admin.securityTableIp')}</th>
                  <th class="w-20 px-3 py-2" scope="col">{t(preferences.language, 'admin.securityUploads')}</th>
                  <th class="w-[110px] px-3 py-2" scope="col">{t(preferences.language, 'admin.securitySize')}</th>
                  <th class="w-[110px] px-3 py-2 text-center" scope="col">{t(preferences.language, 'admin.securityTableActions')}</th>
                </tr>
              </thead>
              <tbody>
                {#each topIps as item (item.ip_address)}
                  <tr class="studio-table-row align-middle">
                    <th class="px-3 py-3 text-left font-black" scope="row">{item.ip_address}</th>
                    <td class="px-3 py-3">{item.upload_count}</td>
                    <td class="px-3 py-3">{formatBytes(item.total_size, preferences.language)}</td>
                    <td class="px-3 py-3 text-center">
                      <button class="studio-button p-2 text-xs" data-tone={item.is_banned ? 'green' : 'danger'} type="button" disabled={item.is_banned} onclick={() => (banTarget = { ip: item.ip_address, label: item.ip_address })}>
                        {item.is_banned ? t(preferences.language, 'admin.securityBanned') : t(preferences.language, 'admin.securityBan')}
                      </button>
                    </td>
                  </tr>
                {/each}
              </tbody>
            </table>
          </div>
        </section>
        <section class="min-w-0 overflow-hidden">
          <div class="mb-3 flex items-center justify-between border-b-[3px] ink-line pb-2">
            <h2 class="text-2xl font-black">{t(preferences.language, 'admin.securityBannedIps')}</h2>
          </div>
          {#if safeBans.length === 0}
            <div class="grid min-h-32 min-w-0 max-w-full place-items-center border-[3px] border-dashed ink-line px-3 text-center"><p class="flex min-w-0 items-center gap-2 font-black"><ShieldCheck class="size-5 shrink-0" />{t(preferences.language, 'admin.securityNoActiveBans')}</p></div>
          {:else}
            <div class="w-full min-w-0 max-w-full touch-pan-x overflow-x-auto overscroll-x-contain [-webkit-overflow-scrolling:touch]">
              <table class="w-full min-w-[420px] table-fixed border-collapse text-sm">
                <thead>
                  <tr class="border-b-[3px] ink-line text-left text-xs font-black uppercase text-[hsl(var(--ink-muted))]">
                    <th class="w-14 px-2 py-2" scope="col">{t(preferences.language, 'admin.securityTableIp')}</th>
                    <th class="w-[38%] px-2 py-2" scope="col">{t(preferences.language, 'admin.securityBanReason')}</th>
                    <th class="w-24 px-2 py-2" scope="col">{t(preferences.language, 'admin.securityBanDuration')}</th>
                    <th class="w-20 px-2 py-2 text-center" scope="col">{t(preferences.language, 'admin.securityTableActions')}</th>
                  </tr>
                </thead>
                <tbody>
                  {#each safeBans as ban (ban.id)}
                    <tr class="studio-table-row align-middle">
                      <th class="px-2 py-3 text-left font-black" scope="row">{ban.ip_address}</th>
                      <td class="min-w-0 px-2 py-3"><p class="truncate text-xs text-[hsl(var(--ink-muted))]" title={ban.reason}>{ban.reason}</p></td>
                      <td class="break-words px-2 py-3 text-xs text-[hsl(var(--ink-muted))]">{ban.expires_at ?? t(preferences.language, 'admin.securityNeverExpires')}</td>
                      <td class="px-2 py-3">
                        <div class="flex justify-center gap-2">
                          <button class="studio-button p-2" data-tone="green" type="button" onclick={() => (confirmTarget = { action: 'unban', ban })} aria-label={t(preferences.language, 'admin.securityUnban')}><Unlock class="size-4" /></button>
                          <button class="studio-button p-2" data-tone="danger" type="button" onclick={() => (confirmTarget = { action: 'purge', ban })} aria-label={t(preferences.language, 'admin.securityPurge')}><Trash2 class="size-4" /></button>
                        </div>
                      </td>
                    </tr>
                  {/each}
                </tbody>
              </table>
            </div>
          {/if}
        </section>
      </div>
    {/if}
  {/if}
  <BanIPDialog target={banTarget} busy={banning} onClose={() => (banTarget = null)} onConfirm={banIp} />
  <ConfirmDialog
    open={confirmTarget !== null}
    title={confirmTarget?.action === 'unban' ? t(preferences.language, 'admin.securityUnbanConfirm') : t(preferences.language, 'admin.securityDeleteBanImagesConfirm')}
    description={confirmTarget?.ban.ip_address ?? ''}
    confirmLabel={confirmTarget?.action === 'unban' ? t(preferences.language, 'admin.securityUnban') : t(preferences.language, 'admin.securityPurge')}
    cancelLabel={t(preferences.language, 'common.cancel')}
    tone={confirmTarget?.action === 'unban' ? 'primary' : 'danger'}
    busy={confirmBusy}
    onClose={() => (confirmTarget = null)}
    onConfirm={() => confirmTarget?.action === 'unban' ? unban(confirmTarget.ban.id) : confirmTarget && purgeImages(confirmTarget.ban.id)}
  />
</div>
