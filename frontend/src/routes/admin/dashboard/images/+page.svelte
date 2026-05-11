<script lang="ts">
  import { Ban, Search, Trash2 } from 'lucide-svelte';
  import BanIPDialog from '@/components/studio/BanIPDialog.svelte';
  import ConfirmDialog from '@/components/studio/ConfirmDialog.svelte';
  import ImageDetailDrawer from '@/components/studio/ImageDetailDrawer.svelte';
  import PageTitle from '@/components/studio/PageTitle.svelte';
  import { adminCreateIPBan, adminDeleteImages, adminGetImages } from '@/api';
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
  let banTargetImage = $state<AdminImage | null>(null);
  let deleteTarget = $state<AdminImage | 'selected' | null>(null);
  let busy = $state(false);
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
    busy = true;
    try {
      await adminDeleteImages(preferences.adminToken, Array.from(selected));
      toast.success(t(preferences.language, 'admin.imagesDeleted', { count: selected.size }));
      deleteTarget = null;
      await load();
    } catch (err) {
      toast.error(err instanceof Error ? err.message : t(preferences.language, 'admin.imagesDeleteError'));
    } finally {
      busy = false;
    }
  }

  async function removeOne(image: AdminImage) {
    if (!preferences.adminToken) return;
    busy = true;
    try {
      await adminDeleteImages(preferences.adminToken, [image.uid]);
      toast.success(t(preferences.language, 'admin.imagesDeleted', { count: 1 }));
      deleteTarget = null;
      await load();
    } catch (err) {
      toast.error(err instanceof Error ? err.message : t(preferences.language, 'admin.imagesDeleteError'));
    } finally {
      busy = false;
    }
  }

  async function banOne(input: { ip: string; reason: string; durationHours: number | null }) {
    if (!preferences.adminToken || !banTargetImage) return;
    busy = true;
    try {
      await adminCreateIPBan(preferences.adminToken, { uid: banTargetImage.uid, ip_address: input.ip, duration_hours: input.durationHours, reason: input.reason });
      banTargetImage = null;
      toast.success(t(preferences.language, 'common.success'));
    } catch (err) {
      toast.error(err instanceof Error ? err.message : t(preferences.language, 'common.error'));
    } finally {
      busy = false;
    }
  }

  function toggle(uid: string) {
    selected = new Set(selected.has(uid) ? Array.from(selected).filter((item) => item !== uid) : [...selected, uid]);
  }

  $effect(() => { load(); });
</script>

<svelte:head><title>{t(preferences.language, 'admin.imagesTitle')} · OmePic</title></svelte:head>

<div class="space-y-6">
  <PageTitle eyebrow={t(preferences.language, 'admin.fileDeskEyebrow')} title={t(preferences.language, 'admin.imagesTitle')} subtitle={t(preferences.language, 'admin.imagesDescription')} />
  <div class="grid gap-3 border-b-[3px] ink-line pb-4 md:grid-cols-[1fr_auto]">
    <label class="relative min-w-0">
      <span class="sr-only">{t(preferences.language, 'common.search')}</span>
      <Search class="absolute left-3 top-1/2 size-4 -translate-y-1/2" />
      <input class="studio-input w-full pl-10" bind:value={search} name="admin-image-search" autocomplete="off" placeholder={t(preferences.language, 'admin.imagesSearch')} onkeydown={(event) => event.key === 'Enter' && load()} />
    </label>
    <button class="studio-button w-full justify-center md:w-auto" data-tone="danger" type="button" disabled={selected.size === 0} onclick={() => (deleteTarget = 'selected')}><Trash2 class="size-4" />{t(preferences.language, 'admin.imagesDelete')} ({selected.size})</button>
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
                <img src={`/i/${image.uid}.avif`} alt={image.uid} class="h-full w-full object-cover" loading="lazy" decoding="async" width="48" height="48" />
              </button>
            </td>
            <th class="min-w-0 px-2 py-2 text-left font-normal" scope="row"><button class="block min-w-0 max-w-full text-left" type="button" onclick={() => (activeImage = image)}><span class="block truncate font-black hover:marker-highlight">{image.uid}</span><span class="block truncate text-xs text-[hsl(var(--ink-muted))]">{image.md5_hash}</span></button></th>
            <td class="px-2 py-2 font-bold tabular-nums">{formatBytes(image.size, preferences.language)}</td>
            <td class="px-2 py-2 font-bold" title={image.storage_key}>{image.storage_key}</td>
            <td class="truncate px-2 py-2" title={image.ip_address_masked}>{image.ip_address_masked}</td>
            <td class="px-2 py-2">
              <div class="flex justify-end gap-1.5"><button class="studio-button px-2 py-1.5 text-xs" type="button" onclick={() => (activeImage = image)}>{t(preferences.language, 'admin.imageDetails')}</button><a class="studio-button px-2 py-1.5 text-xs" href={`/i/${image.uid}.avif`} target="_blank" rel="noopener noreferrer">{t(preferences.language, 'admin.imageOpen')}</a><button class="studio-button px-2 py-1.5 text-xs" data-tone="danger" type="button" onclick={() => (banTargetImage = image)} aria-label={`${t(preferences.language, 'admin.securityBan')} ${image.ip_address_masked}`}><Ban class="size-4" /></button><button class="studio-button px-2 py-1.5 text-xs" data-tone="danger" type="button" onclick={() => (deleteTarget = image)} aria-label={`${t(preferences.language, 'common.delete')} ${image.uid}`}><Trash2 class="size-4" /></button></div>
            </td>
          </tr>
        {/each}
      </tbody>
    </table>
  </div>
  <div class="grid grid-cols-[auto_1fr_auto] items-center gap-2 border-t-[3px] ink-line pt-3">
    <button class="studio-button px-3 py-1.5 text-sm" disabled={page <= 1} onclick={() => { page -= 1; load(); }}>{t(preferences.language, 'admin.imagesPrev')}</button>
    <span class="min-w-0 justify-center text-center text-sm font-black">{t(preferences.language, 'admin.imagesPage', { page })}</span>
    <button class="studio-button px-3 py-1.5 text-sm" disabled={page * pageSize >= total} onclick={() => { page += 1; load(); }}>{t(preferences.language, 'admin.imagesNext')}</button>
  </div>
  <ImageDetailDrawer image={activeImage} images={images} onNavigate={(image) => (activeImage = image)} onClose={() => (activeImage = null)} onDeleted={() => { load(); }} />
  <BanIPDialog target={banTargetImage ? { ip: banTargetImage.ip_address, label: banTargetImage.ip_address_masked } : null} busy={busy} onClose={() => (banTargetImage = null)} onConfirm={banOne} />
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
