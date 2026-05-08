<script lang="ts">
  import { Copy, KeyRound, Terminal } from 'lucide-svelte';
  import PageTitle from '@/components/studio/PageTitle.svelte';
  import { getClientToken } from '@/preferences';
  import { t } from '@/i18n';
  import { preferences } from '@/stores/preferences.svelte';
  import { toast } from '@/stores/toast.svelte';

  const token = typeof window === 'undefined' ? '' : getClientToken();
  const examples = $derived([
    {
      title: 'Upload',
      code: `curl -X POST "$ORIGIN/v1/image" \\\n  -H "X-Token: ${token || '<token>'}" \\\n  -F "file=@image.png"`,
    },
    {
      title: 'Delete',
      code: `curl -X DELETE "$ORIGIN/i/<uid>.avif" \\\n  -H "X-Token: ${token || '<token>'}"`,
    },
    {
      title: 'Response',
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
  <PageTitle eyebrow="Developer notes" title={t(preferences.language, 'api.title')} subtitle={t(preferences.language, 'api.subtitle')} />

  <section class="grid gap-6 lg:grid-cols-[260px_1fr]">
    <aside class="studio-panel h-fit p-5 rotate-[-0.35deg]">
      <KeyRound class="mb-3 size-8" />
      <p class="text-xs font-black uppercase text-[hsl(var(--ink-muted))]">{t(preferences.language, 'common.token')}</p>
      <p class="mt-2 break-all font-mono text-sm">{token}</p>
      <button class="studio-button mt-4 text-xs" type="button" onclick={() => copy(token)}>
        <Copy class="size-4" />
        {t(preferences.language, 'common.copyToken')}
      </button>
    </aside>

    <div class="space-y-6">
      {#each examples as example (example.title)}
        <section class="border-b-[3px] ink-line pb-6">
          <div class="mb-3 flex items-center justify-between gap-3">
            <h2 class="flex items-center gap-2 text-2xl font-black"><Terminal class="size-5" />{example.title}</h2>
            <button class="studio-button text-xs" type="button" onclick={() => copy(example.code)}><Copy class="size-4" />Copy</button>
          </div>
          <pre class="overflow-x-auto border-2 ink-line bg-[hsl(var(--ink))] p-4 text-sm text-[hsl(var(--paper))]"><code>{example.code}</code></pre>
        </section>
      {/each}
    </div>
  </section>
</div>

