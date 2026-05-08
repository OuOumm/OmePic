<script lang="ts">
  import { Ban, Copy, ExternalLink, Trash2, X } from 'lucide-svelte';
  import { adminCreateIPBan, adminDeleteImages } from '@/api';
  import { t } from '@/i18n';
  import { preferences } from '@/stores/preferences.svelte';
  import { toast } from '@/stores/toast.svelte';
  import { formatBytes, formatDate } from '@/utils';
  import type { AdminImage } from '@/types';

  type Props = {
    image: AdminImage | null;
    onClose: () => void;
    onDeleted: () => void;
  };

  let { image, onClose, onDeleted }: Props = $props();
  let busy = $state(false);

  const imageUrl = $derived(image ? `/i/${image.uid}.avif` : '');

  function copy(value: string) {
    navigator.clipboard.writeText(value);
    toast.success(t(preferences.language, 'common.copied'));
  }

  async function remove() {
    if (!preferences.adminToken || !image || !confirm(`${t(preferences.language, 'common.delete')} ${image.uid}?`)) return;
    busy = true;
    try {
      await adminDeleteImages(preferences.adminToken, [image.uid]);
      toast.success(t(preferences.language, 'common.success'));
      onDeleted();
    } catch (err) {
      toast.error(err instanceof Error ? err.message : t(preferences.language, 'common.error'));
    } finally {
      busy = false;
    }
  }

  async function banIp() {
    if (!preferences.adminToken || !image) return;
    busy = true;
    try {
      await adminCreateIPBan(preferences.adminToken, { uid: image.uid, duration_hours: 24, reason: `Image ${image.uid}` });
      toast.success(t(preferences.language, 'common.success'));
    } catch (err) {
      toast.error(err instanceof Error ? err.message : t(preferences.language, 'common.error'));
    } finally {
      busy = false;
    }
  }
</script>

{#if image}
  <div class="fixed inset-0 z-[70] bg-[hsl(var(--ink)/0.38)] backdrop-blur-sm" role="presentation" onclick={onClose}></div>
  <aside class="fixed right-0 top-0 z-[80] h-dvh w-full max-w-xl overflow-y-auto border-l-[3px] ink-line bg-[hsl(var(--paper))] p-5 shadow-[-8px_0_0_hsl(var(--ink))] sketch-enter">
    <div class="mb-5 flex items-start justify-between gap-3 border-b-[3px] ink-line pb-3">
      <div>
        <span class="tape-label rotate-[-2deg]">image detail</span>
        <h2 class="mt-3 break-all text-3xl font-black">{image.uid}</h2>
      </div>
      <button class="studio-button p-2" type="button" onclick={onClose} aria-label="close"><X class="size-4" /></button>
    </div>

    <div class="mb-5 overflow-hidden border-[3px] ink-line bg-[hsl(var(--paper-deep))]">
      <img src={imageUrl} alt={image.uid} class="max-h-80 w-full object-contain" loading="lazy" />
    </div>

    <div class="grid gap-3">
      <div class="grid grid-cols-[120px_1fr_auto] items-center gap-3 studio-table-row py-2 text-sm"><b>URL</b><span class="truncate">{imageUrl}</span><button class="studio-button p-2" onclick={() => copy(imageUrl)}><Copy class="size-4" /></button></div>
      <div class="grid grid-cols-[120px_1fr_auto] items-center gap-3 studio-table-row py-2 text-sm"><b>MD5</b><span class="truncate">{image.md5_hash}</span><button class="studio-button p-2" onclick={() => copy(image.md5_hash)}><Copy class="size-4" /></button></div>
      <div class="grid grid-cols-[120px_1fr_auto] items-center gap-3 studio-table-row py-2 text-sm"><b>Token</b><span class="truncate">{image.token}</span><button class="studio-button p-2" onclick={() => copy(image.token)}><Copy class="size-4" /></button></div>
      <div class="grid grid-cols-[120px_1fr] gap-3 studio-table-row py-2 text-sm"><b>IP</b><span>{image.ip_address_masked}</span></div>
      <div class="grid grid-cols-[120px_1fr] gap-3 studio-table-row py-2 text-sm"><b>Size</b><span>{formatBytes(image.size)}</span></div>
      <div class="grid grid-cols-[120px_1fr] gap-3 studio-table-row py-2 text-sm"><b>MIME</b><span>{image.mime_type}</span></div>
      <div class="grid grid-cols-[120px_1fr] gap-3 studio-table-row py-2 text-sm"><b>Storage</b><span>{image.storage_key} · {image.storage_backend}</span></div>
      <div class="grid grid-cols-[120px_1fr] gap-3 studio-table-row py-2 text-sm"><b>Created</b><span>{formatDate(image.created_at)}</span></div>
    </div>

    <div class="mt-6 flex flex-wrap gap-3">
      <a class="studio-button" data-tone="blue" href={imageUrl} target="_blank" rel="noreferrer"><ExternalLink class="size-4" />Open</a>
      <button class="studio-button" type="button" disabled={busy} onclick={banIp}><Ban class="size-4" />Ban IP</button>
      <button class="studio-button" data-tone="danger" type="button" disabled={busy} onclick={remove}><Trash2 class="size-4" />{t(preferences.language, 'common.delete')}</button>
    </div>
  </aside>
{/if}
