<script lang="ts">
  import { Ban, CheckCircle2, Copy, ExternalLink, Trash2, X } from 'lucide-svelte';
  import { adminCreateIPBan, adminDeleteImages, adminGetAbuseIPDetail } from '@/api';
  import { attachAccessibleDialog } from '@/actions/accessible-dialog';
  import ConfirmDialog from './ConfirmDialog.svelte';
  import ImageSwitchButton from './ImageSwitchButton.svelte';
  import { copyToClipboard } from '@/clipboard';
  import { t } from '@/i18n';
  import { preferences } from '@/stores/preferences.svelte';
  import { formatBytes, formatDate, getImagePath, isAbortError } from '@/utils';
  import { runAsyncAction, toastApiError } from '@/ui-errors';
  import type { AdminAbuseIPDetail, AdminImage } from '@/types';
  import BanIPDialog from './BanIPDialog.svelte';

  type Props = {
    image: AdminImage | null;
    images?: AdminImage[];
    onClose: () => void;
    onDeleted: () => void;
    onNavigate?: (image: AdminImage) => void;
  };

  let { image, images = [], onClose, onDeleted, onNavigate }: Props = $props();
  let busy = $state(false);
  let ipDetail = $state<AdminAbuseIPDetail | null>(null);
  let ipDetailLoading = $state(false);
  let banTarget = $state<{ ip: string; label?: string } | null>(null);
  let deleteOpen = $state(false);
  let imageLoaded = $state(false);
  let loadedIpAddress = $state('');

  const imageUrl = $derived(image ? getImagePath(image.uid) : '');
  const targetIp = $derived(image?.ip_address ?? '');
  const targetIpLabel = $derived(image?.ip_address_masked ?? '');
  const isIpBanned = $derived(Boolean(ipDetail?.is_banned));
  const currentIndex = $derived(image ? images.findIndex((item) => item.uid === image.uid) : -1);
  const hasNavigation = $derived(Boolean(onNavigate) && images.length > 1 && currentIndex >= 0);
  const previousImage = $derived(hasNavigation && currentIndex > 0 ? images[currentIndex - 1] : null);
  const nextImage = $derived(hasNavigation && currentIndex < images.length - 1 ? images[currentIndex + 1] : null);

  function navigateTo(target: AdminImage | null) {
    if (target && onNavigate) onNavigate(target);
  }

  function handleKeydown(event: KeyboardEvent) {
    if (event.key === 'ArrowLeft' && previousImage) {
      event.preventDefault();
      navigateTo(previousImage);
    }
    if (event.key === 'ArrowRight' && nextImage) {
      event.preventDefault();
      navigateTo(nextImage);
    }
  }

  $effect(() => {
    const token = preferences.adminToken;
    const currentIp = targetIp;
    const currentImageUid = image?.uid ?? '';
    imageLoaded = false;
    ipDetail = null;
    banTarget = null;
    loadedIpAddress = currentIp;

    if (!token || !currentIp || !currentImageUid) {
      ipDetailLoading = false;
      loadedIpAddress = '';
      return;
    }

    const controller = new AbortController();
    ipDetailLoading = true;
    adminGetAbuseIPDetail(token, currentIp, controller.signal)
      .then((detail) => {
        if (!controller.signal.aborted && loadedIpAddress === currentIp) ipDetail = detail;
      })
      .catch((err) => {
        if (isAbortError(err)) return;
        if (loadedIpAddress === currentIp) ipDetail = null;
        toastApiError(err, preferences.language, 'admin.ipDetailLoadError');
      })
      .finally(() => {
        if (!controller.signal.aborted && loadedIpAddress === currentIp) ipDetailLoading = false;
      });

    return () => controller.abort();
  });

  function copy(value: string) {
    void copyToClipboard(value, preferences.language);
  }

  async function remove() {
    const token = preferences.adminToken;
    if (!token || !image) return;
    const uid = image.uid;
    await runAsyncAction({
      language: preferences.language,
      setBusy: (value) => (busy = value),
      successMessage: t(preferences.language, 'common.success'),
      action: () => adminDeleteImages(token, [uid]),
      onSuccess: () => {
        const fallback = nextImage ?? previousImage;
        deleteOpen = false;
        onDeleted();
        if (fallback && onNavigate) onNavigate(fallback);
        else onClose();
      },
    });
  }

  async function banIp(input: { ip: string; reason: string; durationHours: number | null }) {
    const token = preferences.adminToken;
    if (!token || !image || !input.ip || isIpBanned) return;
    const uid = image.uid;
    await runAsyncAction({
      language: preferences.language,
      setBusy: (value) => (busy = value),
      successMessage: t(preferences.language, 'common.success'),
      action: () => adminCreateIPBan(token, { uid, ip_address: input.ip, duration_hours: input.durationHours, reason: input.reason }),
      onSuccess: (result) => {
        ipDetail = {
          ip_address: result.ban.ip_address,
          ip_address_masked: result.ban.ip_address_masked,
          upload_count: result.affected_image_count,
          total_size: result.affected_total_size,
          is_banned: true,
          ban: result.ban,
        };
        loadedIpAddress = result.ban.ip_address;
        banTarget = null;
      },
    });
  }
</script>

{#if image}
  <div class="fixed inset-0 z-[70] grid place-items-center p-2 sm:p-4" role="dialog" aria-modal="true" aria-label={t(preferences.language, 'admin.imageDetails')} tabindex="-1" onkeydown={handleKeydown} {@attach attachAccessibleDialog(() => ({ onClose }))}>
    <button class="absolute inset-0 cursor-default bg-[hsl(var(--ink))]/35 backdrop-blur-[2px]" type="button" onclick={onClose} aria-label={t(preferences.language, 'common.close')}></button>
    <div class="studio-panel relative max-h-[calc(100dvh-1rem)] w-full max-w-2xl overflow-y-auto bg-[hsl(var(--paper))] p-3 rotate-[0.25deg] sketch-enter sm:max-h-[calc(100dvh-3rem)] sm:p-5">
      <div class="mb-2 flex items-start justify-between gap-2 sm:mb-4 sm:gap-3">
        <span class="tape-label rotate-[-2deg]">{t(preferences.language, 'admin.imageDetails')}</span>
        <button class="studio-button p-1.5 sm:p-2" type="button" onclick={onClose} aria-label={t(preferences.language, 'common.close')}><X class="size-4" /></button>
      </div>

    <div class="relative mb-3 grid min-h-28 place-items-center overflow-hidden border-[3px] ink-line bg-[hsl(var(--paper-deep))] sm:mb-5 sm:min-h-40">
      {#if !imageLoaded}
        <div class="absolute inset-2 animate-pulse border-2 border-dashed border-[hsl(var(--ink)/0.32)] bg-[hsl(var(--paper)/0.38)] sm:inset-3" aria-hidden="true"></div>
      {/if}
      <img src={imageUrl} alt={image.uid} class="max-h-[28dvh] w-full object-contain transition-opacity duration-200 sm:max-h-80 {imageLoaded ? 'opacity-100' : 'opacity-0'}" loading="eager" decoding="async" width="960" height="540" onload={() => (imageLoaded = true)} />
      {#if hasNavigation}
        <ImageSwitchButton direction="previous" language={preferences.language} disabled={!previousImage} onclick={() => navigateTo(previousImage)} />
        <ImageSwitchButton direction="next" language={preferences.language} disabled={!nextImage} onclick={() => navigateTo(nextImage)} />
      {/if}
    </div>

    <dl class="grid grid-cols-2 gap-x-3 gap-y-1 text-xs sm:text-sm md:grid-cols-2">
      <div class="col-span-2 grid grid-cols-[3.75rem_minmax(0,1fr)_auto] items-center gap-1.5 border-b-2 border-dashed border-[hsl(var(--ink)/0.28)] py-1 md:col-span-2"><dt class="font-black">{t(preferences.language, 'image.uid')}</dt><dd class="min-w-0 truncate font-bold" title={image.uid}>{image.uid}</dd><button class="studio-button px-1.5 py-1 sm:px-2 sm:py-1.5" type="button" onclick={() => copy(image.uid)} aria-label={t(preferences.language, 'image.copyUid')}><Copy class="size-3.5 sm:size-4" /></button></div>
      <div class="col-span-2 grid grid-cols-[3.75rem_minmax(0,1fr)_auto] items-center gap-1.5 border-b-2 border-dashed border-[hsl(var(--ink)/0.28)] py-1 md:col-span-2"><dt class="font-black">{t(preferences.language, 'image.url')}</dt><dd class="min-w-0 truncate" title={imageUrl}>{imageUrl}</dd><button class="studio-button px-1.5 py-1 sm:px-2 sm:py-1.5" type="button" onclick={() => copy(imageUrl)} aria-label={t(preferences.language, 'common.copyUrl')}><Copy class="size-3.5 sm:size-4" /></button></div>
      <div class="col-span-2 grid grid-cols-[3.75rem_minmax(0,1fr)_auto] items-center gap-1.5 border-b-2 border-dashed border-[hsl(var(--ink)/0.28)] py-1 md:col-span-2"><dt class="font-black">{t(preferences.language, 'image.md5')}</dt><dd class="min-w-0 truncate" title={image.md5_hash}>{image.md5_hash}</dd><button class="studio-button px-1.5 py-1 sm:px-2 sm:py-1.5" type="button" onclick={() => copy(image.md5_hash)} aria-label={t(preferences.language, 'image.copyMd5')}><Copy class="size-3.5 sm:size-4" /></button></div>
      <div class="col-span-2 grid grid-cols-[3.75rem_minmax(0,1fr)_auto] items-center gap-1.5 border-b-2 border-dashed border-[hsl(var(--ink)/0.28)] py-1 md:col-span-2"><dt class="font-black">{t(preferences.language, 'image.token')}</dt><dd class="min-w-0 truncate" title={image.token}>{image.token}</dd><button class="studio-button px-1.5 py-1 sm:px-2 sm:py-1.5" type="button" onclick={() => copy(image.token)} aria-label={t(preferences.language, 'image.copyToken')}><Copy class="size-3.5 sm:size-4" /></button></div>
      <div class="grid grid-cols-[3.75rem_minmax(0,1fr)] gap-1.5 border-b-2 border-dashed border-[hsl(var(--ink)/0.28)] py-1"><dt class="font-black">{t(preferences.language, 'image.ip')}</dt><dd class="min-w-0 truncate" title={targetIpLabel}>{targetIpLabel}</dd></div>
      <div class="grid grid-cols-[3.75rem_minmax(0,1fr)] gap-1.5 border-b-2 border-dashed border-[hsl(var(--ink)/0.28)] py-1"><dt class="font-black">{t(preferences.language, 'admin.securityStatus')}</dt><dd class="min-w-0 truncate {isIpBanned ? 'font-black text-[hsl(var(--marker-pink))]' : ''}">{ipDetailLoading ? t(preferences.language, 'common.loading') : isIpBanned ? t(preferences.language, 'admin.securityBanned') : t(preferences.language, 'admin.securityNotBanned')}</dd></div>
      <div class="grid grid-cols-[3.75rem_minmax(0,1fr)] gap-1.5 border-b-2 border-dashed border-[hsl(var(--ink)/0.28)] py-1"><dt class="font-black">{t(preferences.language, 'image.size')}</dt><dd class="tabular-nums">{formatBytes(image.size, preferences.language)}</dd></div>
      <div class="grid grid-cols-[3.75rem_minmax(0,1fr)] gap-1.5 border-b-2 border-dashed border-[hsl(var(--ink)/0.28)] py-1"><dt class="font-black">{t(preferences.language, 'image.type')}</dt><dd class="min-w-0 truncate" title={image.mime_type}>{image.mime_type}</dd></div>
      <div class="grid grid-cols-[3.75rem_minmax(0,1fr)] gap-1.5 border-b-2 border-dashed border-[hsl(var(--ink)/0.28)] py-1"><dt class="font-black">{t(preferences.language, 'image.storage')}</dt><dd class="min-w-0 truncate" title={`${image.storage_key} · ${image.storage_backend}`}>{image.storage_key} · {image.storage_backend}</dd></div>
      <div class="grid grid-cols-[3.75rem_minmax(0,1fr)] gap-1.5 border-b-2 border-dashed border-[hsl(var(--ink)/0.28)] py-1"><dt class="font-black">{t(preferences.language, 'image.created')}</dt><dd class="min-w-0 truncate">{formatDate(image.created_at, preferences.language)}</dd></div>
    </dl>

    <div class="mt-3 flex flex-wrap gap-1.5 sm:mt-4 sm:gap-2">
      <a class="studio-button px-2 py-1.5 text-xs sm:px-3 sm:py-2 sm:text-sm" data-tone="blue" href={imageUrl} target="_blank" rel="noopener noreferrer"><ExternalLink class="size-3.5 sm:size-4" />{t(preferences.language, 'admin.imageOpen')}</a>
      {#if isIpBanned}
        <button class="studio-button px-2 py-1.5 text-xs sm:px-3 sm:py-2 sm:text-sm" type="button" disabled><CheckCircle2 class="size-3.5 sm:size-4" />{t(preferences.language, 'admin.securityBanned')}</button>
      {:else}
        <button class="studio-button px-2 py-1.5 text-xs sm:px-3 sm:py-2 sm:text-sm" type="button" disabled={busy || ipDetailLoading || !targetIp} onclick={() => (banTarget = { ip: targetIp, label: targetIpLabel })}><Ban class="size-3.5 sm:size-4" />{t(preferences.language, 'admin.securityBan')}</button>
      {/if}
      <button class="studio-button px-2 py-1.5 text-xs sm:px-3 sm:py-2 sm:text-sm" data-tone="danger" type="button" disabled={busy} onclick={() => (deleteOpen = true)}><Trash2 class="size-3.5 sm:size-4" />{t(preferences.language, 'common.delete')}</button>
    </div>
    </div>
    <BanIPDialog target={banTarget} busy={busy} onClose={() => (banTarget = null)} onConfirm={banIp} />
    <ConfirmDialog
      open={deleteOpen}
      title={`${t(preferences.language, 'common.delete')} ${image.uid}?`}
      description={image.md5_hash}
      confirmLabel={t(preferences.language, 'common.delete')}
      cancelLabel={t(preferences.language, 'common.cancel')}
      {busy}
      onClose={() => (deleteOpen = false)}
      onConfirm={remove}
    />
  </div>
{/if}
