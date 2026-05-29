<script lang="ts">
  import { page } from '$app/state';
  import { KeyRound, Save, TriangleAlert } from 'lucide-svelte';
  import { adminChangePassword, adminGetConfig, adminGetSystemSettings, adminPurgeCloudflareImageCache, adminUpdateSystemSettings } from '@/api';
  import AnnouncementManager from '@/components/studio/AnnouncementManager.svelte';
  import PageTitle from '@/components/studio/PageTitle.svelte';
  import StorageInstanceManager from '@/components/studio/StorageInstanceManager.svelte';
  import { t } from '@/i18n';
  import { isAbortError } from '@/utils';
  import { preferences } from '@/stores/preferences.svelte';
  import { toast } from '@/stores/toast.svelte';
  import { runAsyncAction, toastApiError } from '@/ui-errors';
  import { isValidAdminPasswordStrength } from '@/password-policy';
  import type { AdminConfig, AdminSystemSettings } from '@/types';

  let config = $state.raw<AdminConfig | null>(null);
  let system = $state<AdminSystemSettings | null>(null);
  let savingRuntime = $state(false);
  let changingPassword = $state(false);
  let purgingCloudflare = $state(false);
  let cloudflarePurgeUrl = $state('');
  let oldPassword = $state('');
  let newPassword = $state('');

  const activeTab = $derived(page.url.searchParams.get('tab') ?? 'runtime');
  const siteName = $derived(system?.runtime.site_name || preferences.runtimeSettings?.site.name || 'OmePic');
  const cloudflarePurgeConfigured = $derived(system?.readonly.service.cloudflare_purge_configured ?? false);
  const realIpSourceOptions = [
    { value: 'remote-addr', labelKey: 'admin.realIpSourceRemoteAddr' },
    { value: 'x-forwarded-for', labelKey: 'admin.realIpSourceXForwardedFor' },
    { value: 'x-real-ip', labelKey: 'admin.realIpSourceXRealIp' },
    { value: 'cf-connecting-ip', labelKey: 'admin.realIpSourceCfConnectingIp' },
  ] as const;
  const securityWarnings = $derived.by(() => {
    const warnings: string[] = [];
    if (!system) return warnings;
    if (!system.readonly.security.jwt_secret.configured) warnings.push(t(preferences.language, 'admin.runtimeWarningJwtDefault'));
    if (!system.readonly.security.uid_encryption_key.configured) warnings.push(t(preferences.language, 'admin.runtimeWarningUidDefault'));
    if (!system.readonly.security.admin_password.configured) warnings.push(t(preferences.language, 'admin.runtimeWarningAdminPasswordBootstrap'));
    return warnings;
  });

  async function load(signal?: AbortSignal) {
    if (!preferences.adminToken || activeTab === 'announcements') return;
    try {
      [config, system] = await Promise.all([adminGetConfig(preferences.adminToken, signal), adminGetSystemSettings(preferences.adminToken, signal)]);
      if (signal?.aborted) return;
    } catch (err) {
      if (isAbortError(err)) return;
      toastApiError(err, preferences.language);
    }
  }



  async function saveRuntime() {
    const token = preferences.adminToken;
    if (!token || !system) return;
    await runAsyncAction({
      language: preferences.language,
      setBusy: (value) => (savingRuntime = value),
      successMessage: t(preferences.language, 'common.success'),
      action: () => {
        if (!system) throw new Error(t(preferences.language, 'common.error'));
        return adminUpdateSystemSettings(token, system.runtime);
      },
      onSuccess: (nextSystem) => {
        system = nextSystem;
      },
    });
  }

  async function changePassword() {
    const token = preferences.adminToken;
    if (!token) return;
    if (!isValidAdminPasswordStrength(newPassword)) {
      toast.error(t(preferences.language, 'admin.passwordStrengthError'));
      return;
    }
    await runAsyncAction({
      language: preferences.language,
      setBusy: (value) => (changingPassword = value),
      successMessage: t(preferences.language, 'admin.passwordChanged'),
      fallbackErrorKey: 'admin.passwordChangeError',
      action: () => adminChangePassword(token, oldPassword, newPassword),
      onSuccess: async () => {
        oldPassword = '';
        newPassword = '';
        system = await adminGetSystemSettings(token);
      },
    });
  }

  async function purgeCloudflareCache() {
    const token = preferences.adminToken;
    const url = cloudflarePurgeUrl.trim();
    if (!token || !url) return;
    await runAsyncAction({
      language: preferences.language,
      setBusy: (value) => (purgingCloudflare = value),
      successMessage: (result) => t(preferences.language, 'admin.cloudflarePurgeSuccess', { url: result.url }),
      fallbackErrorKey: 'admin.cloudflarePurgeError',
      action: () => adminPurgeCloudflareImageCache(token, url),
      onSuccess: () => {
        cloudflarePurgeUrl = '';
      },
    });
  }


  $effect(() => {
    const controller = new AbortController();
    void load(controller.signal);
    return () => controller.abort();
  });
</script>

<svelte:head><title>{t(preferences.language, 'admin.settingsTitle')} · {siteName}</title></svelte:head>

<div class="space-y-6">
  {#if activeTab === 'storage'}
    {#if config}
      <StorageInstanceManager {config} onChange={(next) => (config = next)} />
    {/if}
  {:else if activeTab === 'runtime'}
    {#if system}
      <PageTitle eyebrow={t(preferences.language, 'admin.submenuRuntime')} title={t(preferences.language, 'admin.runtimeTitle')} subtitle={t(preferences.language, 'admin.runtimeDescription')} tone="yellow" />
      <div class="mt-6 flex justify-end border-b-[3px] ink-line pb-3">
        <button class="studio-button w-fit" data-tone="primary" type="button" disabled={savingRuntime} onclick={saveRuntime}><Save class="size-4" />{t(preferences.language, 'common.save')}</button>
      </div>

      <div class="grid gap-4">
        {#if securityWarnings.length > 0}
          <div class="grid gap-2 rounded-none border-2 ink-line bg-[hsl(var(--marker-yellow))] p-4 text-[hsl(var(--marker-ink))]">
            <div class="flex items-center gap-2 font-black">
              <TriangleAlert class="size-4" />
              {t(preferences.language, 'admin.runtimeSecurityWarnings')}
            </div>
            <ul class="list-disc space-y-1 pl-5 text-sm font-bold">
              {#each securityWarnings as warning (warning)}
                <li>{warning}</li>
              {/each}
            </ul>
          </div>
        {/if}
        <div class="grid gap-4 rounded-none border-2 ink-line bg-[hsl(var(--paper))] p-4 md:grid-cols-2">
            <div class="md:col-span-2">
              <span class="tape-label rotate-[1deg]" style="background:hsl(var(--marker-green))">{t(preferences.language, 'admin.runtimeSiteIdentity')}</span>
            </div>
            <label class="grid gap-2 text-sm font-black">
              {t(preferences.language, 'admin.runtimeSiteName')}
              <input class="studio-input" bind:value={system.runtime.site_name} />
            </label>
            <label class="grid gap-2 text-sm font-black">
              {t(preferences.language, 'admin.runtimeSiteTagline')}
              <input class="studio-input" bind:value={system.runtime.site_tagline} />
            </label>
          </div>

          <div class="grid gap-4 rounded-none border-2 ink-line bg-[hsl(var(--paper))] p-4">
            <label class="flex items-center gap-3 border-y-2 ink-line py-3 font-black">
              <input type="checkbox" bind:checked={system.runtime.allow_storage_selection} />
              {t(preferences.language, 'admin.settingsAllowSelection')}
            </label>
          </div>

          <div class="grid gap-4 rounded-none border-2 ink-line bg-[hsl(var(--paper))] p-4 md:grid-cols-2">
            <div class="md:col-span-2">
              <span class="tape-label rotate-[-1deg]" style="background:hsl(var(--marker-green))">{t(preferences.language, 'admin.runtimePublicAccess')}</span>
            </div>
            <label class="grid gap-2 text-sm font-black md:col-span-2">
              {t(preferences.language, 'admin.runtimePublicUrl')}
              <input class="studio-input" bind:value={system.runtime.public_base_url} />
            </label>
            <label class="grid gap-2 text-sm font-black md:col-span-2">
              {t(preferences.language, 'admin.realIpSource')}
              <select class="studio-input" bind:value={system.runtime.real_ip_source}>
                {#each realIpSourceOptions as option (option.value)}
                  <option value={option.value}>{t(preferences.language, option.labelKey)}</option>
                {/each}
              </select>
              <span class="text-sm font-bold text-[hsl(var(--ink-muted))]">{t(preferences.language, 'admin.realIpSourceHint')}</span>
            </label>
          </div>

          <div class="grid gap-4 rounded-none border-2 ink-line bg-[hsl(var(--paper))] p-4 md:grid-cols-2">
            <div class="md:col-span-2">
              <span class="tape-label rotate-[1deg]" style="background:hsl(var(--marker-blue))">Cloudflare</span>
            </div>
            <label class="flex items-start gap-3 border-y-2 ink-line py-3 font-black md:col-span-2">
              <input class="mt-1" type="checkbox" bind:checked={system.runtime.cloudflare_purge_enabled} />
              <span class="grid gap-1">
                <span>{t(preferences.language, 'admin.cloudflarePurgeEnabled')}</span>
                <span class="text-sm font-bold text-[hsl(var(--ink-muted))]">{t(preferences.language, 'admin.cloudflarePurgeDescription')}</span>
              </span>
            </label>
            <label class="grid gap-2 text-sm font-black">
              {t(preferences.language, 'admin.cloudflareZoneId')}
              <input class="studio-input font-mono text-sm" autocomplete="off" bind:value={system.runtime.cloudflare_zone_id} />
            </label>
            <label class="grid gap-2 text-sm font-black">
              {t(preferences.language, 'admin.cloudflareApiToken')}
              <input class="studio-input font-mono text-sm" type="password" autocomplete="new-password" placeholder={t(preferences.language, 'admin.cloudflareApiTokenPlaceholder')} bind:value={system.runtime.cloudflare_api_token} />
            </label>
            <label class="grid gap-2 text-sm font-black md:col-span-2">
              {t(preferences.language, 'admin.cloudflareApiBaseUrl')}
              <input class="studio-input font-mono text-sm" placeholder="https://api.cloudflare.com/client/v4" bind:value={system.runtime.cloudflare_api_base_url} />
              <span class="text-sm font-bold text-[hsl(var(--ink-muted))]">{t(preferences.language, 'admin.cloudflareApiBaseUrlHint')}</span>
            </label>
            <p class="text-sm font-bold text-[hsl(var(--ink-muted))] md:col-span-2">
              {t(preferences.language, cloudflarePurgeConfigured ? 'admin.cloudflareConfigured' : 'admin.cloudflareNotConfigured')}
            </p>
            {#if system.runtime.cloudflare_purge_enabled}
              <div class="grid gap-3 border-t-2 ink-line pt-3 md:col-span-2">
                <label class="grid gap-2 text-sm font-black">
                  {t(preferences.language, 'admin.cloudflareManualPurgeUrl')}
                  <input class="studio-input" placeholder={t(preferences.language, 'admin.cloudflareManualPurgePlaceholder')} bind:value={cloudflarePurgeUrl} />
                </label>
                <div class="flex flex-col gap-3 md:flex-row md:items-center md:justify-between">
                  <p class="text-sm font-bold text-[hsl(var(--ink-muted))]">{t(preferences.language, 'admin.cloudflareManualPurgeHint')}</p>
                  <button class="studio-button w-full md:w-fit" data-tone="blue" type="button" disabled={purgingCloudflare || !cloudflarePurgeConfigured || !cloudflarePurgeUrl.trim()} onclick={purgeCloudflareCache}>{t(preferences.language, 'admin.cloudflareManualPurge')}</button>
                </div>
              </div>
            {/if}
          </div>

          <div class="grid gap-4 rounded-none border-2 ink-line bg-[hsl(var(--paper))] p-4">
            <div class="flex flex-col gap-3 md:flex-row md:items-center md:justify-between">
              <span class="tape-label rotate-[1deg]" style="background:hsl(var(--marker-pink))">{t(preferences.language, 'admin.runtimeMaintenanceMode')}</span>
              <label class="flex items-center gap-3 border-y-2 ink-line py-3 font-black md:border-y-0 md:py-0">
                <input type="checkbox" bind:checked={system.runtime.maintenance_mode} />
                {t(preferences.language, 'admin.runtimeMaintenance')}
              </label>
            </div>
            <label class="grid gap-2 text-sm font-black">
              {t(preferences.language, 'admin.runtimeMaintenanceMessage')}
              <textarea class="studio-input min-h-24" bind:value={system.runtime.maintenance_message}></textarea>
            </label>
          </div>

          <form class="grid gap-4 rounded-none border-2 ink-line bg-[hsl(var(--paper))] p-4 md:grid-cols-2" onsubmit={(event) => { event.preventDefault(); void changePassword(); }}>
            <div class="md:col-span-2">
              <span class="tape-label rotate-[-1deg]" style="background:hsl(var(--marker-yellow))">{t(preferences.language, 'admin.changePassword')}</span>
            </div>
            <label class="grid gap-2 text-sm font-black">
              {t(preferences.language, 'admin.oldPassword')}
              <input class="studio-input" type="password" autocomplete="current-password" bind:value={oldPassword} />
            </label>
            <label class="grid gap-2 text-sm font-black">
              {t(preferences.language, 'admin.newPassword')}
              <input class="studio-input" type="password" autocomplete="new-password" bind:value={newPassword} />
            </label>
            <div class="md:col-span-2 flex justify-end">
              <button class="studio-button w-fit" data-tone="blue" type="submit" disabled={changingPassword || !oldPassword || !newPassword}><KeyRound class="size-4" />{t(preferences.language, 'admin.changePassword')}</button>
            </div>
          </form>
      </div>
    {/if}
  {:else if activeTab === 'announcements'}
    <AnnouncementManager />
  {/if}
</div>
