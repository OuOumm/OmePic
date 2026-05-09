<script lang="ts">
  import { Search, Trash2 } from 'lucide-svelte';
  import ImageDetailDrawer from '@/components/studio/ImageDetailDrawer.svelte';
  import PageTitle from '@/components/studio/PageTitle.svelte';
  import { adminDeleteImages, adminGetImages } from '@/api';
  import { t } from '@/i18n';
  import { preferences } from '@/stores/preferences.svelte';
  import { toast } from '@/stores/toast.svelte';
  import { formatBytes } from '@/utils';
  import type { AdminImage } from '@/types';

  let images = $state<AdminImage[]>([]);
  let total = $state(0);
  let page = $state(1);
  let search = $state('');
  let selected = $state<Set<string>>(new Set());
  let activeImage = $state<AdminImage | null>(null);
  const pageSize = 30;

  async function load() {
    if (!preferences.adminToken) return;
    const data = await adminGetImages(preferences.adminToken, page, pageSize, search || undefined);
    images = data.items;
    total = data.total;
    selected = new Set();
  }

  async function removeSelected() {
    if (!preferences.adminToken || selected.size === 0) return;
    if (!confirm(t(preferences.language, 'admin.imagesDeleteConfirm', { count: selected.size }))) return;
    await adminDeleteImages(preferences.adminToken, Array.from(selected));
    toast.success(t(preferences.language, 'admin.imagesDeleted', { count: selected.size }));
    await load();
  }

  function toggle(uid: string) {
    selected = new Set(selected.has(uid) ? Array.from(selected).filter((item) => item !== uid) : [...selected, uid]);
  }

  $effect(() => { load(); });
</script>

<svelte:head><title>{t(preferences.language, 'admin.imagesTitle')} · OmePic</title></svelte:head>

<div class="space-y-6">
  <PageTitle eyebrow="File desk" title={t(preferences.language, 'admin.imagesTitle')} subtitle={t(preferences.language, 'admin.imagesDescription')} />
  <div class="grid gap-3 border-y-[3px] ink-line py-4 md:grid-cols-[1fr_auto]">
    <label class="relative">
      <Search class="absolute left-3 top-1/2 size-4 -translate-y-1/2" />
      <input class="studio-input pl-10" bind:value={search} placeholder={t(preferences.language, 'admin.imagesSearch')} onkeydown={(event) => event.key === 'Enter' && load()} />
    </label>
    <button class="studio-button" data-tone="danger" type="button" disabled={selected.size === 0} onclick={removeSelected}><Trash2 class="size-4" />{t(preferences.language, 'admin.imagesDelete')} ({selected.size})</button>
  </div>
  <p class="font-black">{t(preferences.language, 'admin.imagesTotal', { total })}</p>
  <div class="overflow-x-auto">
    <div class="min-w-[990px]">
      {#each images as image (image.uid)}
        <div class="grid grid-cols-[36px_72px_1fr_120px_110px_120px_170px] items-center gap-4 studio-table-row py-3 text-sm">
          <input type="checkbox" checked={selected.has(image.uid)} onchange={() => toggle(image.uid)} />
          <button class="grid size-14 place-items-center overflow-hidden border-2 ink-line bg-[hsl(var(--paper-deep))]" type="button" onclick={() => (activeImage = image)} aria-label={t(preferences.language, 'common.openPreview', { title: image.uid })}>
            <img src={`/i/${image.uid}.avif`} alt={image.uid} class="h-full w-full object-cover" loading="lazy" />
          </button>
          <button class="min-w-0 text-left" type="button" onclick={() => (activeImage = image)}><p class="truncate font-black hover:marker-highlight">{image.uid}</p><p class="truncate text-xs text-[hsl(var(--ink-muted))]">{image.md5_hash}</p></button>
          <div>{formatBytes(image.size)}</div>
          <div>{image.storage_key}</div>
          <div>{image.ip_address_masked}</div>
          <div class="flex gap-2"><button class="studio-button p-2 text-xs" type="button" onclick={() => (activeImage = image)}>Details</button><a class="studio-button p-2 text-xs" href={`/i/${image.uid}.avif`} target="_blank">Open</a></div>
        </div>
      {/each}
    </div>
  </div>
  <div class="flex justify-between border-t-[3px] ink-line pt-4">
    <button class="studio-button" disabled={page <= 1} onclick={() => { page -= 1; load(); }}>{t(preferences.language, 'admin.imagesPrev')}</button>
    <span class="font-black">{t(preferences.language, 'admin.imagesPage', { page })}</span>
    <button class="studio-button" disabled={page * pageSize >= total} onclick={() => { page += 1; load(); }}>{t(preferences.language, 'admin.imagesNext')}</button>
  </div>
  <ImageDetailDrawer image={activeImage} onClose={() => (activeImage = null)} onDeleted={() => { activeImage = null; load(); }} />
</div>
