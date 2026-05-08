<script lang="ts">
  import { Save } from 'lucide-svelte';
  import { adminGetConfig, adminGetSystemSettings, adminSetDefaultStorage, adminUpdateSystemSettings } from '@/api';
  import AnnouncementManager from '@/components/studio/AnnouncementManager.svelte';
  import PageTitle from '@/components/studio/PageTitle.svelte';
  import { t } from '@/i18n';
  import { preferences } from '@/stores/preferences.svelte';
  import { toast } from '@/stores/toast.svelte';
  import type { AdminConfig, AdminSystemSettings } from '@/types';

  let config = $state<AdminConfig | null>(null);
  let system = $state<AdminSystemSettings | null>(null);
  let savingRuntime = $state(false);
  let savingDefault = $state('');

  async function load() {
    if (!preferences.adminToken) return;
    [config, system] = await Promise.all([adminGetConfig(preferences.adminToken), adminGetSystemSettings(preferences.adminToken)]);
  }

  async function saveRuntime() {
    if (!preferences.adminToken || !system) return;
    savingRuntime = true;
    try {
      system = await adminUpdateSystemSettings(preferences.adminToken, system.runtime);
      toast.success(t(preferences.language, 'common.success'));
    } catch (err) {
      toast.error(err instanceof Error ? err.message : t(preferences.language, 'common.error'));
    } finally {
      savingRuntime = false;
    }
  }

  async function setDefault(storageKey: string) {
    if (!preferences.adminToken) return;
    savingDefault = storageKey;
    try {
      config = await adminSetDefaultStorage(preferences.adminToken, storageKey);
      toast.success(t(preferences.language, 'common.success'));
    } catch (err) {
      toast.error(err instanceof Error ? err.message : t(preferences.language, 'common.error'));
    } finally {
      savingDefault = '';
    }
  }

  $effect(() => { load(); });
</script>

<svelte:head><title>{t(preferences.language, 'admin.settingsTitle')} · OmePic</title></svelte:head>

<div class="space-y-10">
  <PageTitle eyebrow="Settings" title={t(preferences.language, 'admin.settingsTitle')} subtitle={t(preferences.language, 'admin.settingsDescription')} tone="green" />
  {#if config}
    <section>
      <h2 class="border-b-[3px] ink-line pb-2 text-2xl font-black">Storage instances</h2>
      {#each config.storage_configs as item (item.storage_key)}
        <div class="studio-table-row grid gap-3 py-4 md:grid-cols-[1fr_150px_170px] md:items-center">
          <div><p class="font-black">{item.name}</p><p class="text-sm text-[hsl(var(--ink-muted))]">{item.storage_key}</p></div>
          <div class="font-bold">{item.storage_backend}</div>
          <button class="studio-button justify-self-start text-xs" data-tone={item.is_default ? 'green' : 'blue'} type="button" disabled={item.is_default || savingDefault === item.storage_key} onclick={() => setDefault(item.storage_key)}>
            {item.is_default ? t(preferences.language, 'common.default') : 'Set default'}
          </button>
        </div>
      {/each}
    </section>
  {/if}
  {#if system}
    <section class="blueprint-grid border-[3px] ink-line p-5">
      <div class="mb-5 flex items-end justify-between gap-3 border-b-2 ink-line pb-3">
        <div>
          <span class="tape-label rotate-[-2deg]">Runtime</span>
          <h2 class="mt-3 text-2xl font-black">Runtime controls</h2>
        </div>
        <button class="studio-button" data-tone="primary" type="button" disabled={savingRuntime} onclick={saveRuntime}><Save class="size-4" />{t(preferences.language, 'common.save')}</button>
      </div>
      <div class="grid gap-4 md:grid-cols-2">
        <label class="grid gap-2 text-sm font-black">
          {t(preferences.language, 'admin.settingsMaxUpload')}
          <input class="studio-input" type="number" min="0" bind:value={system.runtime.max_upload_size_mb} />
        </label>
        <label class="grid gap-2 text-sm font-black">
          Public URL
          <input class="studio-input" bind:value={system.runtime.public_base_url} />
        </label>
        <label class="flex items-center gap-3 border-y-2 ink-line py-3 font-black">
          <input type="checkbox" bind:checked={system.runtime.allow_storage_selection} />
          {t(preferences.language, 'admin.settingsAllowSelection')}
        </label>
        <label class="flex items-center gap-3 border-y-2 ink-line py-3 font-black">
          <input type="checkbox" bind:checked={system.runtime.maintenance_mode} />
          Maintenance
        </label>
        <label class="grid gap-2 text-sm font-black md:col-span-2">
          Maintenance message
          <textarea class="studio-input min-h-24" bind:value={system.runtime.maintenance_message}></textarea>
        </label>
      </div>
    </section>
  {/if}
  <AnnouncementManager />
</div>
