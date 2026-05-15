<script lang="ts">
  import { Copy, KeyRound, Terminal } from 'lucide-svelte';
  import PageTitle from '@/components/studio/PageTitle.svelte';
  import { copyToClipboard } from '@/clipboard';
  import { getClientToken } from '@/client-token';
  import { t } from '@/i18n';
  import { preferences, setRuntimeSettings } from '@/stores/preferences.svelte';
  import { getRuntimeSettings } from '@/api';
  import { getApiExampleBaseUrl, isAbortError } from '@/utils';

  const token = typeof window === 'undefined' ? '' : getClientToken();
  const apiBaseUrl = $derived(getApiExampleBaseUrl(preferences.runtimeSettings?.access.public_base_url));
  const examples = $derived([
    {
      title: t(preferences.language, 'api.exampleUpload'),
      code: `curl -X POST "${apiBaseUrl}/v1/image" \\\n  -H "X-Token: ${token || '<token>'}" \\\n  -F "file=@image.png"`,
    },
    {
      title: t(preferences.language, 'api.exampleDelete'),
      code: `curl -X DELETE "${apiBaseUrl}/i/<uid>.avif" \\\n  -H "X-Token: ${token || '<token>'}"`,
    },
    {
      title: t(preferences.language, 'api.exampleStorage'),
      code: `curl -X GET "${apiBaseUrl}/v1/runtime-settings"`,
    },
    {
      title: t(preferences.language, 'api.exampleStorageResponse'),
      code: `{
  "success": true,
  "data": {
    "features": {
      "allow_storage_selection": true
    },
    "storage": {
      "options": [
        {
          "storage_key": "local-primary",
          "name": "Local Primary",
          "storage_backend": "local",
          "is_default": true
        },
        {
          "storage_key": "s3-archive",
          "name": "Archive Bucket",
          "storage_backend": "s3",
          "is_default": false
        }
      ]
    }
  }
}`,
    },
    {
      title: t(preferences.language, 'api.exampleResponse'),
      code: `{
  "success": true,
  "data": {
    "url": "${apiBaseUrl}/i/abc123.avif",
    "duplicate": false
  }
}`,
    },
  ]);

  async function loadRuntimeSettings(signal?: AbortSignal) {
    try {
      const settings = await getRuntimeSettings(signal);
      if (!signal?.aborted) setRuntimeSettings(settings);
    } catch (err) {
      if (!isAbortError(err)) setRuntimeSettings(null);
    }
  }

  function copy(value: string) {
    void copyToClipboard(value, preferences.language);
  }

  $effect(() => {
    const controller = new AbortController();
    void loadRuntimeSettings(controller.signal);
    return () => controller.abort();
  });
</script>

<svelte:head><title>{t(preferences.language, 'api.title')} · OmePic</title></svelte:head>

<div class="space-y-8">
  <PageTitle eyebrow={t(preferences.language, 'api.eyebrow')} title={t(preferences.language, 'api.title')} subtitle={t(preferences.language, 'api.subtitle')} tone="green" />

  <section class="grid min-w-0 gap-6 lg:grid-cols-[260px_minmax(0,1fr)]">
    <aside class="studio-panel h-fit min-w-0 rotate-[-0.35deg] p-4 sm:p-5">
      <KeyRound class="mb-3 size-8" />
      <p class="text-xs font-black uppercase text-[hsl(var(--ink-muted))]">{t(preferences.language, 'common.token')}</p>
      <p class="mt-2 overflow-wrap-anywhere font-mono text-sm">{token}</p>
      <button class="studio-button mt-4 w-full text-xs sm:w-auto" type="button" onclick={() => copy(token)}>
        <Copy class="size-4" />
        {t(preferences.language, 'common.copyToken')}
      </button>
    </aside>

    <div class="min-w-0 space-y-6">
      {#each examples as example (example.title)}
        <section class="studio-panel rotate-[-0.1deg] p-4 sm:p-5">
          <div class="flex items-start justify-between gap-4">
            <div>
              <div class="flex items-center gap-2">
                <Terminal class="size-4" />
                <h2 class="text-base font-semibold">{example.title}</h2>
              </div>
            </div>
            <button class="studio-button text-xs" type="button" onclick={() => copy(example.code)}>
              <Copy class="size-4" />
              {t(preferences.language, 'common.copy')}
            </button>
          </div>
          <pre class="mt-4 overflow-x-auto rounded-lg bg-[hsl(var(--ink-muted))]/10 p-3 text-xs font-mono leading-snug">{example.code}</pre>
        </section>
      {/each}
    </div>
  </section>
</div>
