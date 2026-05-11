<script lang="ts">
  import { page } from '$app/state';
  import { CircleAlert, Save } from 'lucide-svelte';
  import { adminGetConfig, adminGetSystemSettings, adminUpdateSystemSettings } from '@/api';
  import AnnouncementManager from '@/components/studio/AnnouncementManager.svelte';
  import PageTitle from '@/components/studio/PageTitle.svelte';
  import StorageInstanceManager from '@/components/studio/StorageInstanceManager.svelte';
  import { t } from '@/i18n';
  import { formatMegabytes, isAbortError } from '@/utils';
  import { preferences } from '@/stores/preferences.svelte';
  import { toast } from '@/stores/toast.svelte';
  import type { AdminConfig, AdminSystemSettings } from '@/types';

  let config = $state<AdminConfig | null>(null);
  let system = $state<AdminSystemSettings | null>(null);
  let savingRuntime = $state(false);
  let mimeTypesText = $state('');

  const activeTab = $derived(page.url.searchParams.get('tab') ?? 'runtime');
  const siteName = $derived(system?.runtime.site_name || preferences.runtimeSettings?.site.name || 'OmePic');

  function runtimeMimeTypesText(settings: AdminSystemSettings | null) {
    const runtimeTypes = settings?.runtime.allowed_mime_types;
    return Array.isArray(runtimeTypes) ? runtimeTypes.join(', ') : '';
  }

  async function load(signal?: AbortSignal) {
    if (!preferences.adminToken || activeTab === 'announcements') return;
    try {
      [config, system] = await Promise.all([adminGetConfig(preferences.adminToken, signal), adminGetSystemSettings(preferences.adminToken, signal)]);
      if (signal?.aborted) return;
      mimeTypesText = runtimeMimeTypesText(system);
    } catch (err) {
      if (isAbortError(err)) return;
      toast.error(err instanceof Error ? err.message : t(preferences.language, 'common.error'));
    }
  }

  function parseMimeTypes(value: string) {
    return value
      .split(/[\r\n,]+/)
      .map((item) => item.trim())
      .filter(Boolean);
  }

  async function saveRuntime() {
    if (!preferences.adminToken || !system) return;
    savingRuntime = true;
    try {
      system.runtime.allowed_mime_types = parseMimeTypes(mimeTypesText);
      system = await adminUpdateSystemSettings(preferences.adminToken, system.runtime);
      mimeTypesText = runtimeMimeTypesText(system);
      toast.success(t(preferences.language, 'common.success'));
    } catch (err) {
      toast.error(err instanceof Error ? err.message : t(preferences.language, 'common.error'));
    } finally {
      savingRuntime = false;
    }
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

          <div class="grid gap-4 rounded-none border-2 ink-line bg-[hsl(var(--paper))] p-4 md:grid-cols-2">
            <div class="md:col-span-2">
              <span class="tape-label rotate-[1deg]" style="background:hsl(var(--marker-blue))">{t(preferences.language, 'admin.runtimeUploadPolicy')}</span>
            </div>
            <label class="grid gap-2 text-sm font-black">
              {t(preferences.language, 'admin.settingsMaxUpload')} ({formatMegabytes(system.runtime.max_upload_size_mb, preferences.language)})
              <input class="studio-input" type="number" min="0" bind:value={system.runtime.max_upload_size_mb} />
            </label>
            <label class="grid min-w-0 gap-2 text-sm font-black">
              <span class="flex items-center gap-1">
                {t(preferences.language, 'admin.runtimeAllowedMimeTypes')}
                <span class="inline-grid size-4 place-items-center rounded-full border-2 ink-line bg-[hsl(var(--marker-yellow))]" title={t(preferences.language, 'admin.runtimeAllowedMimeTypesHint')} aria-label={t(preferences.language, 'admin.runtimeAllowedMimeTypesHint')} role="img">
                  <CircleAlert class="size-3" />
                </span>
              </span>
              <input class="studio-input min-w-0 font-mono text-sm" bind:value={mimeTypesText} />
            </label>
            <label class="flex items-center gap-3 border-y-2 ink-line py-3 font-black md:col-span-2">
              <input type="checkbox" bind:checked={system.runtime.allow_storage_selection} />
              {t(preferences.language, 'admin.settingsAllowSelection')}
            </label>
          </div>

          <div class="grid gap-4 rounded-none border-2 ink-line bg-[hsl(var(--paper))] p-4">
            <div>
              <span class="tape-label rotate-[-1deg]" style="background:hsl(var(--marker-green))">{t(preferences.language, 'admin.runtimePublicAccess')}</span>
            </div>
            <label class="grid gap-2 text-sm font-black">
              {t(preferences.language, 'admin.runtimePublicUrl')}
              <input class="studio-input" bind:value={system.runtime.public_base_url} />
            </label>
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
      </div>
    {/if}
  {:else if activeTab === 'announcements'}
    <AnnouncementManager />
  {/if}
</div>
