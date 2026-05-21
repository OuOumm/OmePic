<script lang="ts">
  import { Copy, KeyRound, Terminal } from 'lucide-svelte';
  import PageTitle from '@/components/studio/PageTitle.svelte';
  import { copyToClipboard } from '@/clipboard';
  import { getClientToken } from '@/client-token';
  import { t } from '@/i18n';
  import { getRuntimeSettings } from '@/api';
  import { preferences, setRuntimeSettings } from '@/stores/preferences.svelte';
  import { getApiExampleBaseUrl, isAbortError } from '@/utils';

  type ExampleBlock = {
    title: string;
    code: string;
  };

  type EndpointExample = {
    title: string;
    method: 'POST' | 'DELETE' | 'GET';
    path: string;
    description: string;
    notes: string[];
    blocks: ExampleBlock[];
  };

  const token = typeof window === 'undefined' ? '' : getClientToken();
  const apiBaseUrl = $derived(getApiExampleBaseUrl(preferences.runtimeSettings?.access.public_base_url));
  const exampleUid = 'abc123def456';

  const endpointExamples = $derived([
    {
      title: t(preferences.language, 'api.exampleUpload'),
      method: 'POST',
      path: '/v1/image',
      description: t(preferences.language, 'api.uploadDescription'),
      notes: [
        t(preferences.language, 'api.uploadNoteToken'),
        t(preferences.language, 'api.uploadNoteStorage'),
        t(preferences.language, 'api.uploadNoteResult'),
      ],
      blocks: [
        {
          title: t(preferences.language, 'api.requestExample'),
          code: `curl -X POST "${apiBaseUrl}/v1/image" \\\n  -H "X-Token: ${token || '<client-token>'}" \\\n  -F "file=@image.png" \\\n  -F "storage_key=local-default"`,
        },
        {
          title: t(preferences.language, 'api.successResponse'),
          code: `{
  "success": true,
  "data": {
    "url": "${apiBaseUrl}/i/${exampleUid}.avif",
    "duplicate": false
  }
}`,
        },
      ],
    },
    {
      title: t(preferences.language, 'api.exampleDelete'),
      method: 'DELETE',
      path: '/i/:uid.avif',
      description: t(preferences.language, 'api.deleteDescription'),
      notes: [
        t(preferences.language, 'api.deleteNoteToken'),
        t(preferences.language, 'api.deleteNoteLogical'),
        t(preferences.language, 'api.deleteNoteCloudflare'),
      ],
      blocks: [
        {
          title: t(preferences.language, 'api.requestExample'),
          code: `curl -X DELETE "${apiBaseUrl}/i/${exampleUid}.avif" \\\n  -H "X-Token: ${token || '<client-token>'}"`,
        },
        {
          title: t(preferences.language, 'api.successResponse'),
          code: `{
  "success": true,
  "data": {}
}`,
        },
      ],
    },
    {
      title: t(preferences.language, 'api.exampleStorage'),
      method: 'GET',
      path: '/v1/runtime-settings',
      description: t(preferences.language, 'api.runtimeDescription'),
      notes: [
        t(preferences.language, 'api.runtimeNotePublic'),
        t(preferences.language, 'api.runtimeNoteOptions'),
        t(preferences.language, 'api.runtimeNoteFeature'),
      ],
      blocks: [
        {
          title: t(preferences.language, 'api.requestExample'),
          code: `curl -X GET "${apiBaseUrl}/v1/runtime-settings"`,
        },
        {
          title: t(preferences.language, 'api.readStorageOptions'),
          code: `const response = await fetch("${apiBaseUrl}/v1/runtime-settings");
const result = await response.json();

if (result.success) {
  const storageOptions = result.data.storage.options;
  console.log(storageOptions);
}`,
        },
        {
          title: t(preferences.language, 'api.successResponse'),
          code: `{
  "success": true,
  "data": {
    "site": {
      "name": "OmePic",
      "tagline": "上传、分享和管理图片"
    },
    "access": {
      "public_base_url": "${apiBaseUrl}"
    },
    "upload": {
      "max_upload_size_mb": 20,
      "allowed_mime_types": [
        "image/jpeg",
        "image/png",
        "image/gif",
        "image/webp",
        "image/avif"
      ]
    },
    "features": {
      "allow_storage_selection": true,
      "maintenance_mode": false,
      "maintenance_message": ""
    },
    "storage": {
      "options": [
        {
          "storage_key": "local-default",
          "name": "Default Local Storage",
          "storage_backend": "local",
          "is_default": true
        }
      ]
    }
  }
}`,
        },
      ],
    },
  ] satisfies EndpointExample[]);

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

  function methodMarker(method: EndpointExample['method']): string {
    if (method === 'POST') return 'hsl(var(--marker-green))';
    if (method === 'DELETE') return 'hsl(var(--marker-pink))';
    return 'hsl(var(--marker-blue))';
  }

  $effect(() => {
    const controller = new AbortController();
    void loadRuntimeSettings(controller.signal);
    return () => controller.abort();
  });
</script>

<svelte:head><title>{t(preferences.language, 'api.title')} · OmePic</title></svelte:head>

<div class="space-y-8">
  <PageTitle
    eyebrow={t(preferences.language, 'api.eyebrow')}
    title={t(preferences.language, 'api.title')}
    subtitle={t(preferences.language, 'api.subtitle')}
    tone="green"
  />

  <section class="grid min-w-0 gap-6 lg:grid-cols-[260px_minmax(0,1fr)]">
    <aside class="studio-panel h-fit min-w-0 rotate-[-0.35deg] p-4 sm:p-5">
      <KeyRound class="mb-3 size-8" />
      <p class="text-xs font-black uppercase text-[hsl(var(--ink-muted))]">{t(preferences.language, 'common.token')}</p>
      <p class="mt-2 overflow-wrap-anywhere font-mono text-sm">{token}</p>
      <p class="mt-3 text-sm text-[hsl(var(--ink-muted))]">{t(preferences.language, 'api.tokenHint')}</p>
      <button class="studio-button mt-4 w-full text-xs sm:w-auto" type="button" onclick={() => copy(token)}>
        <Copy class="size-4" />
        {t(preferences.language, 'common.copyToken')}
      </button>
    </aside>

    <div class="min-w-0 space-y-6">
      {#each endpointExamples as endpoint (endpoint.path)}
        <section class="studio-panel rotate-[-0.1deg] p-4 sm:p-5">
          <div class="flex flex-wrap items-start justify-between gap-4">
            <div class="min-w-0">
              <div class="flex items-center gap-2">
                <Terminal class="size-4" />
                <h2 class="text-base font-semibold">{endpoint.title}</h2>
              </div>
              <div class="mt-3 flex flex-wrap items-center gap-2">
                <span class="tape-label rotate-[-1deg]" style={`background:${methodMarker(endpoint.method)}`}>
                  {endpoint.method}
                </span>
                <code class="overflow-wrap-anywhere rounded-md bg-[hsl(var(--ink-muted))]/10 px-2 py-1 font-mono text-xs sm:text-sm">
                  {endpoint.path}
                </code>
              </div>
              <p class="mt-3 max-w-3xl text-sm text-[hsl(var(--ink-muted))]">{endpoint.description}</p>
            </div>
          </div>

          <div class="mt-5 grid min-w-0 gap-4 xl:grid-cols-2">
            {#each endpoint.blocks as block (block.title)}
              <article class={`min-w-0 rounded-xl border-2 border-[hsl(var(--ink-muted))]/20 bg-[hsl(var(--paper))] p-3 ${endpoint.blocks.length % 2 === 1 && block.title === t(preferences.language, 'api.successResponse') ? 'xl:col-span-2' : ''}`}>
                <div class="flex items-center justify-between gap-3">
                  <h3 class="text-xs font-black uppercase tracking-[0.18em] text-[hsl(var(--ink-muted))]">{block.title}</h3>
                  <button class="studio-button px-2 py-1.5 text-xs" type="button" onclick={() => copy(block.code)}>
                    <Copy class="size-4" />
                    {t(preferences.language, 'common.copy')}
                  </button>
                </div>
                <pre class="mt-3 overflow-x-auto rounded-lg bg-[hsl(var(--ink-muted))]/10 p-3 text-xs font-mono leading-snug">{block.code}</pre>
              </article>
            {/each}
          </div>

          <div class="mt-5 space-y-2">
            <p class="text-xs font-black uppercase tracking-[0.18em] text-[hsl(var(--ink-muted))]">{t(preferences.language, 'api.notes')}</p>
            <ul class="space-y-2 pl-5 text-sm text-[hsl(var(--ink-muted))]">
              {#each endpoint.notes as note (note)}
                <li class="list-disc">{note}</li>
              {/each}
            </ul>
          </div>
        </section>
      {/each}
    </div>
  </section>
</div>
