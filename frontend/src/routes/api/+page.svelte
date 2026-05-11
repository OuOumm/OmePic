<script lang="ts">
  import { Copy, KeyRound, Terminal } from 'lucide-svelte';
  import PageTitle from '@/components/studio/PageTitle.svelte';
  import { getClientToken } from '@/client-token';
  import { t } from '@/i18n';
  import { preferences } from '@/stores/preferences.svelte';
  import { toast } from '@/stores/toast.svelte';

  const token = typeof window === 'undefined' ? '' : getClientToken();
  const examples = $derived([
    {
      title: t(preferences.language, 'api.exampleUpload'),
      code: `curl -X POST "$ORIGIN/v1/image" \\\n  -H "X-Token: ${token || '<token>'}" \\\n  -F "file=@image.png"`,
    },
    {
      title: t(preferences.language, 'api.exampleDelete'),
      code: `curl -X DELETE "$ORIGIN/i/<uid>.avif" \\\n  -H "X-Token: ${token || '<token>'}"`,
    },
    {
      title: t(preferences.language, 'api.exampleResponse'),
      code: `{
  "success": true,
  "data": {
    "uid": "abc123",
    "url": "https://example.com/i/abc123.avif",
    "markdown": "![](https://example.com/i/abc123.avif)"
  }
}`,
    },
  ]);

  function copy(value: string) {
    navigator.clipboard.writeText(value);
    toast.success(t(preferences.language, 'common.copied'));
  }
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
        <section class="min-w-0 border-b-[3px] ink-line pb-6">
          <div class="mb-3 flex min-w-0 flex-col items-stretch gap-3 sm:flex-row sm:items-center sm:justify-between">
            <h2 class="flex min-w-0 items-center gap-2 overflow-wrap-anywhere text-xl font-black sm:text-2xl"><Terminal class="size-5 shrink-0" />{example.title}</h2>
            <button class="studio-button w-full shrink-0 text-xs sm:w-auto" type="button" onclick={() => copy(example.code)}><Copy class="size-4" />{t(preferences.language, 'common.copy')}</button>
          </div>
          <div class="max-w-full min-w-0 overflow-hidden border-2 ink-line bg-[hsl(var(--ink))]">
            <pre class="max-w-full overflow-x-auto p-3 text-xs text-[hsl(var(--paper))] sm:p-4 sm:text-sm"><code class="block min-w-0 whitespace-pre-wrap break-words">{example.code}</code></pre>
          </div>
        </section>
      {/each}
    </div>
  </section>
</div>

