<script lang="ts">
  import PageTitle from '@/components/studio/PageTitle.svelte';
  import { adminGetConfig, adminGetSystemSettings } from '@/api';
  import { t } from '@/i18n';
  import { preferences } from '@/stores/preferences.svelte';
  import type { AdminConfig, AdminSystemSettings } from '@/types';

  let config = $state<AdminConfig | null>(null);
  let system = $state<AdminSystemSettings | null>(null);

  async function load() {
    if (!preferences.adminToken) return;
    [config, system] = await Promise.all([adminGetConfig(preferences.adminToken), adminGetSystemSettings(preferences.adminToken)]);
  }

  $effect(() => { load(); });
</script>

<svelte:head><title>{t(preferences.language, 'admin.settingsTitle')} · OmePic</title></svelte:head>

<div class="space-y-6">
  <PageTitle eyebrow="Settings" title={t(preferences.language, 'admin.settingsTitle')} subtitle={t(preferences.language, 'admin.settingsDescription')} />
  {#if config}
    <section>
      <h2 class="border-b-[3px] ink-line pb-2 text-2xl font-black">Storage instances</h2>
      {#each config.storage_configs as item (item.storage_key)}
        <div class="studio-table-row grid gap-3 py-4 md:grid-cols-[1fr_150px_150px]">
          <div><p class="font-black">{item.name}</p><p class="text-sm text-[hsl(var(--ink-muted))]">{item.storage_key}</p></div>
          <div>{item.storage_backend}</div>
          <div>{item.is_default ? t(preferences.language, 'common.default') : ''}</div>
        </div>
      {/each}
    </section>
  {/if}
  {#if system}
    <section class="blueprint-grid border-[3px] ink-line p-5">
      <h2 class="text-2xl font-black">Runtime</h2>
      <div class="mt-4 grid gap-3 md:grid-cols-2">
        <p><b>{t(preferences.language, 'admin.settingsMaxUpload')}</b> · {system.runtime.max_upload_size_mb} MB</p>
        <p><b>{t(preferences.language, 'admin.settingsAllowSelection')}</b> · {system.runtime.allow_storage_selection ? t(preferences.language, 'common.enabled') : t(preferences.language, 'common.disabled')}</p>
        <p><b>Maintenance</b> · {system.runtime.maintenance_mode ? t(preferences.language, 'common.enabled') : t(preferences.language, 'common.disabled')}</p>
        <p><b>Public URL</b> · {system.runtime.public_base_url || '-'}</p>
      </div>
    </section>
  {/if}
</div>

