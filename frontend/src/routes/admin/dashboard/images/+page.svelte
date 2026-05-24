<script lang="ts">
  import { Ban, Search, Trash2 } from 'lucide-svelte';
  import BanIPDialog from '@/components/studio/BanIPDialog.svelte';
  import ConfirmDialog from '@/components/studio/ConfirmDialog.svelte';
  import ImageDetailDrawer from '@/components/studio/ImageDetailDrawer.svelte';
  import PageTitle from '@/components/studio/PageTitle.svelte';
  import { adminCreateIPBan, adminDeleteImages, adminGetImages } from '@/api';
  import { t } from '@/i18n';
  import { preferences } from '@/stores/preferences.svelte';
  import { isAbortError, formatBytes, getImagePath } from '@/utils';
  import { runAsyncAction, toastApiError } from '@/ui-errors';
  import type { AdminImage } from '@/types';

  let images = $state.raw<AdminImage[]>([]);
  let total = $state(0);
  let page = $state(1);
  let search = $state('');
  let selected = $state<Set<string>>(new Set());
  let activeImage = $state<AdminImage | null>(null);
  let banTargetImage = $state<AdminImage | null>(null);
  let deleteTarget = $state<AdminImage | 'selected' | null>(null);
  let busy = $state(false);
  let debouncedSearch = $state('');
  let pageSize = $state(30);
  const pageSizeOptions = [10, 30, 50, 100];
  const searchDebounceMs = 300;
  const totalPages = $derived(Math.max(1, Math.ceil(total / pageSize)));

  async function load(currentPage = page, currentSearch = debouncedSearch, currentPageSize = pageSize, signal?: AbortSignal) {
    if (!preferences.adminToken) return;
    try {
      const data = await adminGetImages(preferences.adminToken, currentPage, currentPageSize, currentSearch || undefined, signal);
      if (signal?.aborted) return;
      images = data.items;
      total = data.total;
      selected = new Set();
    } catch (err) {
      if (isAbortError(err)) return;
      toastApiError(err, preferences.language);
    }
  }

  async function removeSelected() {
    const token = preferences.adminToken;
    if (!token || selected.size === 0) return;
    const uids = Array.from(selected);
    await runAsyncAction({
      language: preferences.language,
      setBusy: (value) => (busy = value),
      successMessage: t(preferences.language, 'admin.imagesDeleted', { count: uids.length }),
      fallbackErrorKey: 'admin.imagesDeleteError',
      action: () => adminDeleteImages(token, uids),
      onSuccess: async () => {
        deleteTarget = null;
        await load();
      },
    });
  }

  async function removeOne(image: AdminImage) {
    const token = preferences.adminToken;
    if (!token) return;
    await runAsyncAction({
      language: preferences.language,
      setBusy: (value) => (busy = value),
      successMessage: t(preferences.language, 'admin.imagesDeleted', { count: 1 }),
      fallbackErrorKey: 'admin.imagesDeleteError',
      action: () => adminDeleteImages(token, [image.uid]),
      onSuccess: async () => {
        deleteTarget = null;
        await load();
      },
    });
  }

  async function banOne(input: { ip: string; reason: string; durationHours: number | null }) {
    const token = preferences.adminToken;
    if (!token || !banTargetImage) return;
    const uid = banTargetImage.uid;
    await runAsyncAction({
      language: preferences.language,
      setBusy: (value) => (busy = value),
      successMessage: t(preferences.language, 'common.success'),
      action: () => adminCreateIPBan(token, { uid, ip_address: input.ip, duration_hours: input.durationHours, reason: input.reason }),
      onSuccess: () => {
        banTargetImage = null;
      },
    });
  }

  function loadCurrent() {
    void load();
  }

  function toggle(uid: string) {
    selected = new Set(selected.has(uid) ? Array.from(selected).filter((item) => item !== uid) : [...selected, uid]);
  }

  function setPageSize(value: string) {
    const nextPageSize = Number(value);
    pageSize = pageSizeOptions.includes(nextPageSize) ? nextPageSize : 30;
    page = 1;
  }

  $effect(() => {
    const nextSearch = search.trim();
    const timer = window.setTimeout(() => {
      if (debouncedSearch !== nextSearch) {
        debouncedSearch = nextSearch;
        page = 1;
      }
    }, searchDebounceMs);
    return () => window.clearTimeout(timer);
  });

  $effect(() => {
    const controller = new AbortController();
    void load(page, debouncedSearch, pageSize, controller.signal);
    return () => controller.abort();
  });
</script>

<svelte:head><title>{t(preferences.language, 'admin.imagesTitle')} · OmePic</title></svelte:head>

<div class="space-y-6">
  <PageTitle eyebrow={t(preferences.language, 'admin.fileDeskEyebrow')} title={t(preferences.language, 'admin.imagesTitle')} subtitle={t(preferences.language, 'admin.imagesDescription')} />
  <div class="grid grid-cols-[minmax(0,1fr)_auto] items-center gap-2 border-b-[3px] ink-line pb-4 sm:gap-3 md:grid-cols-[minmax(0,1fr)_auto_auto]">
    <label class="relative min-w-0">
      <span class="sr-only">{t(preferences.language, 'common.search')}</span>
      <Search class="absolute left-3 top-1/2 size-4 -translate-y-1/2" />
      <input class="studio-input w-full pl-10" bind:value={search} name="admin-image-search" autocomplete="off" placeholder={t(preferences.language, 'admin.imagesSearch')} />
    </label>
    <label class="flex items-center gap-1 text-sm font-black sm:gap-2">
      <span class="whitespace-nowrap">{t(preferences.language, 'admin.imagesPageSize')}</span>
      <select class="studio-input w-16 py-2 sm:w-24" value={pageSize} aria-label={t(preferences.language, 'admin.imagesPageSize')} onchange={(event) => setPageSize(event.currentTarget.value)}>
        {#each pageSizeOptions as option (option)}
          <option value={option}>{option}</option>
        {/each}
      </select>
    </label>
    <button class="studio-button col-span-2 w-full justify-center md:col-span-1 md:w-auto" data-tone="danger" type="button" disabled={selected.size === 0} onclick={() => (deleteTarget = 'selected')}><Trash2 class="size-4" />{t(preferences.language, 'admin.imagesDelete')} ({selected.size})</button>
  </div>
  <p class="font-black">{t(preferences.language, 'admin.imagesTotal', { total })}</p>
  <div class="w-full min-w-0 max-w-full overflow-x-auto xl:overflow-x-visible">
    <table class="w-full min-w-[760px] table-fixed border-collapse text-sm">
      <thead>
        <tr class="border-b-[3px] ink-line text-left text-xs font-black uppercase tracking-[0.08em] text-[hsl(var(--ink-muted))]">
          <th class="w-7 py-2 pr-2" scope="col">
            <input type="checkbox" checked={images.length > 0 && selected.size === images.length} aria-label={t(preferences.language, 'admin.imagesSelectAll')} onchange={(event) => selected = event.currentTarget.checked ? new Set(images.map((image) => image.uid)) : new Set()} />
          </th>
          <th class="w-16 px-1 py-2" scope="col">{t(preferences.language, 'admin.imagesTablePreview')}</th>
          <th class="w-[30%] px-2 py-2" scope="col">{t(preferences.language, 'image.uid')}</th>
          <th class="w-[9%] px-2 py-2" scope="col">{t(preferences.language, 'image.size')}</th>
          <th class="w-[15%] px-2 py-2" scope="col">{t(preferences.language, 'image.storageKey')}</th>
          <th class="w-[8%] px-2 py-2" scope="col">{t(preferences.language, 'image.ip')}</th>
          <th class="w-52 px-2 py-2 text-right" scope="col">{t(preferences.language, 'admin.imagesTableActions')}</th>
        </tr>
      </thead>
      <tbody>
        {#each images as image (image.uid)}
          <tr class="studio-table-row align-middle">
            <td class="py-2 pr-2"><input type="checkbox" checked={selected.has(image.uid)} onchange={() => toggle(image.uid)} aria-label={t(preferences.language, 'admin.imagesSelect', { uid: image.uid })} /></td>
            <td class="px-1 py-2">
              <button class="grid size-12 place-items-center overflow-hidden border-2 ink-line bg-[hsl(var(--paper-deep))]" type="button" onclick={() => (activeImage = image)} aria-label={t(preferences.language, 'common.openPreview', { title: image.uid })}>
                <img src={getImagePath(image.uid)} alt={image.uid} class="h-full w-full object-cover" loading="lazy" decoding="async" width="48" height="48" />
              </button>
            </td>
            <th class="min-w-0 px-2 py-2 text-left font-normal" scope="row"><button class="block min-w-0 max-w-full text-left" type="button" onclick={() => (activeImage = image)}><span class="block truncate font-black hover:marker-highlight">{image.uid}</span><span class="block truncate text-xs text-[hsl(var(--ink-muted))]">{image.md5_hash}</span></button></th>
            <td class="px-2 py-2 font-bold tabular-nums">{formatBytes(image.size, preferences.language)}</td>
            <td class="px-2 py-2 font-bold" title={image.storage_key}>{image.storage_key}</td>
            <td class="truncate px-2 py-2" title={image.ip_address}>{image.ip_address}</td>
            <td class="px-2 py-2">
              <div class="flex justify-end gap-1.5"><button class="studio-button px-2 py-1.5 text-xs" type="button" onclick={() => (activeImage = image)}>{t(preferences.language, 'admin.imageDetails')}</button><a class="studio-button px-2 py-1.5 text-xs" href={getImagePath(image.uid)} target="_blank" rel="noopener noreferrer">{t(preferences.language, 'admin.imageOpen')}</a><button class="studio-button px-2 py-1.5 text-xs" data-tone="danger" type="button" onclick={() => (banTargetImage = image)} aria-label={`${t(preferences.language, 'admin.securityBan')} ${image.ip_address}`}><Ban class="size-4" /></button><button class="studio-button px-2 py-1.5 text-xs" data-tone="danger" type="button" onclick={() => (deleteTarget = image)} aria-label={`${t(preferences.language, 'common.delete')} ${image.uid}`}><Trash2 class="size-4" /></button></div>
            </td>
          </tr>
        {/each}
      </tbody>
    </table>
  </div>
  <div class="grid grid-cols-[auto_1fr_auto] items-center gap-2 border-t-[3px] ink-line pt-3">
    <button class="studio-button px-3 py-1.5 text-sm" disabled={page <= 1} onclick={() => { page -= 1; }}>{t(preferences.language, 'admin.imagesPrev')}</button>
    <span class="min-w-0 justify-center text-center text-sm font-black">{t(preferences.language, 'admin.imagesPageStatus', { page, totalPages })}</span>
    <button class="studio-button px-3 py-1.5 text-sm" disabled={page >= totalPages} onclick={() => { page += 1; }}>{t(preferences.language, 'admin.imagesNext')}</button>
  </div>
  <ImageDetailDrawer image={activeImage} images={images} onNavigate={(image) => (activeImage = image)} onClose={() => (activeImage = null)} onDeleted={loadCurrent} />
  <BanIPDialog target={banTargetImage ? { ip: banTargetImage.ip_address, label: banTargetImage.ip_address } : null} busy={busy} onClose={() => (banTargetImage = null)} onConfirm={banOne} />
  <ConfirmDialog
    open={deleteTarget !== null}
    title={deleteTarget === 'selected' ? t(preferences.language, 'admin.imagesDeleteConfirm', { count: selected.size }) : `${t(preferences.language, 'common.delete')} ${deleteTarget?.uid ?? ''}?`}
    description={deleteTarget === 'selected' ? t(preferences.language, 'admin.imagesSelected', { count: selected.size }) : deleteTarget?.md5_hash ?? ''}
    confirmLabel={t(preferences.language, 'common.delete')}
    cancelLabel={t(preferences.language, 'common.cancel')}
    {busy}
    onClose={() => (deleteTarget = null)}
    onConfirm={() => deleteTarget === 'selected' ? removeSelected() : deleteTarget && removeOne(deleteTarget)}
  />
</div>
