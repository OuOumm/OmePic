<script lang="ts">
  import { RefreshCw } from 'lucide-svelte';
  import type { Language, PublicRuntimeSettings } from '@/types';
  import { t } from '@/i18n';

  export let language: Language;
  export let settings: PublicRuntimeSettings | null;
  export let selected = '';
  export let refreshing = false;
  export let onSelect: (key: string) => void;
  export let onRefresh: () => void;
</script>

<div class="studio-panel p-4 rotate-[0.25deg]">
  <div class="mb-3 flex items-center justify-between gap-3 border-b-2 ink-line pb-2">
    <div>
      <p class="text-xs font-black uppercase text-[hsl(var(--ink-muted))]">Inspector</p>
      <h2 class="text-xl font-black">{t(language, 'upload.storage')}</h2>
    </div>
    <button type="button" class="studio-button p-2" onclick={onRefresh} disabled={refreshing} aria-label="refresh">
      <RefreshCw class="size-4 {refreshing ? 'animate-spin' : ''}" />
    </button>
  </div>

  {#if settings}
    <label class="grid gap-2 text-sm font-bold">
      {t(language, 'upload.storage')}
      <select class="studio-input" value={selected} onchange={(event) => onSelect((event.currentTarget as HTMLSelectElement).value)} disabled={!settings.features.allow_storage_selection}>
        <option value="">{settings.storage.default_storage_key} · {t(language, 'common.default')}</option>
        {#each settings.storage.options as option (option.storage_key)}
          <option value={option.storage_key}>{option.name} · {option.storage_backend}</option>
        {/each}
      </select>
    </label>
    <dl class="mt-4 grid gap-2 text-sm">
      <div class="flex justify-between gap-3"><dt>{t(language, 'admin.settingsMaxUpload')}</dt><dd class="font-black">{settings.upload.max_upload_size_mb} MB</dd></div>
      <div class="flex justify-between gap-3"><dt>{t(language, 'admin.settingsAllowSelection')}</dt><dd class="font-black">{settings.features.allow_storage_selection ? t(language, 'common.enabled') : t(language, 'common.disabled')}</dd></div>
      <div class="flex justify-between gap-3"><dt>{t(language, 'admin.systemDefaultStorage')}</dt><dd class="font-black">{settings.storage.default_storage_key}</dd></div>
    </dl>
  {:else}
    <p class="text-sm text-[hsl(var(--ink-muted))]">{t(language, 'common.loading')}</p>
  {/if}
</div>

