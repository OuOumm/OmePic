<script lang="ts">
  import { Copy, Download, Trash2, X } from 'lucide-svelte';
  import { t } from '@/i18n';
  import { accessibleDialog } from '@/actions/accessible-dialog';
  import ImageSwitchButton from './ImageSwitchButton.svelte';
  import type { Language, UploadHistoryRecord } from '@/types';

  type Props = {
    language: Language;
    record: UploadHistoryRecord | null;
    canDelete?: boolean;
    records?: UploadHistoryRecord[];
    onCopy: (value: string) => void;
    onDelete?: () => void;
    onNavigate?: (record: UploadHistoryRecord) => void;
    onClose: () => void;
  };

  let { language, record, canDelete = false, records = [], onCopy, onDelete, onNavigate, onClose }: Props = $props();
  let imageLoaded = $state(false);

  const filename = $derived(record?.original_filename || record?.uid || t(language, 'common.fallbackImage'));
  const currentIndex = $derived(record ? records.findIndex((item) => item.uid === record.uid) : -1);
  const hasNavigation = $derived(Boolean(onNavigate) && records.length > 1 && currentIndex >= 0);
  const previousRecord = $derived(hasNavigation && currentIndex > 0 ? records[currentIndex - 1] : null);
  const nextRecord = $derived(hasNavigation && currentIndex < records.length - 1 ? records[currentIndex + 1] : null);

  function navigateTo(target: UploadHistoryRecord | null) {
    if (target && onNavigate) onNavigate(target);
  }

  function handleKeydown(event: KeyboardEvent) {
    if (event.key === 'ArrowLeft' && previousRecord) {
      event.preventDefault();
      navigateTo(previousRecord);
    }
    if (event.key === 'ArrowRight' && nextRecord) {
      event.preventDefault();
      navigateTo(nextRecord);
    }
  }

  $effect(() => {
    if (record?.url) imageLoaded = false;
  });
</script>

{#if record}
  <div class="fixed inset-0 z-[90] grid min-h-dvh place-items-center overflow-y-auto bg-[hsl(var(--ink)/0.52)] p-2 backdrop-blur-sm sm:p-6" role="presentation" onclick={(event) => event.target === event.currentTarget && onClose()}>
    <div class="grid max-h-[calc(100dvh-1rem)] w-full max-w-5xl grid-rows-[auto_minmax(0,1fr)_auto] overflow-hidden border-[3px] ink-line bg-[hsl(var(--paper))] shadow-[5px_5px_0_hsl(var(--ink))] sketch-enter sm:max-h-[calc(100dvh-3rem)] sm:shadow-[8px_8px_0_hsl(var(--ink))]" role="dialog" tabindex="-1" aria-modal="true" aria-labelledby="image-preview-title" onkeydown={handleKeydown} use:accessibleDialog={{ onClose }}>
      <header class="grid min-w-0 grid-cols-[minmax(0,1fr)_auto] items-start gap-2 border-b-[3px] ink-line p-3 sm:gap-3 sm:p-4">
        <div class="min-w-0 overflow-hidden">
          <span class="tape-label rotate-[-2deg]" style="background:hsl(var(--marker-blue))">{t(language, 'studio.preview')}</span>
          <h2 id="image-preview-title" class="mt-3 block max-w-full truncate text-lg font-black sm:text-2xl" title={filename}>{filename}</h2>
          <p class="mt-1 block max-w-full truncate text-xs font-bold text-[hsl(var(--ink-muted))]" title={record.uid}>{record.uid}</p>
        </div>
        <button class="studio-button shrink-0 p-2" type="button" onclick={onClose} aria-label={t(language, 'common.close')}><X class="size-4" /></button>
      </header>

      <div class="relative grid min-h-0 place-items-center overflow-auto bg-[hsl(var(--paper-deep))] p-4 sm:p-6">
        {#if !imageLoaded}
          <div class="absolute inset-6 animate-pulse border-2 border-dashed border-[hsl(var(--ink)/0.32)] bg-[hsl(var(--paper)/0.38)]" aria-hidden="true"></div>
        {/if}
        <img src={record.url} alt={filename} class="max-h-[62dvh] max-w-full object-contain transition-opacity duration-200 {imageLoaded ? 'opacity-100' : 'opacity-0'}" loading="eager" decoding="async" width="1280" height="720" onload={() => (imageLoaded = true)} />
        {#if hasNavigation}
          <ImageSwitchButton direction="previous" {language} disabled={!previousRecord} onclick={() => navigateTo(previousRecord)} />
          <ImageSwitchButton direction="next" {language} disabled={!nextRecord} onclick={() => navigateTo(nextRecord)} />
        {/if}
      </div>

      <footer class="grid min-w-0 gap-3 border-t-[3px] ink-line p-3 sm:flex sm:flex-wrap sm:items-center sm:justify-between sm:p-4">
        <div class="min-w-0 overflow-hidden text-xs font-bold text-[hsl(var(--ink-muted))]">
          <p class="truncate" title={record.url}>{record.url}</p>
        </div>
        <div class="flex min-w-0 flex-wrap justify-end gap-2">
          <button class="studio-button" type="button" onclick={() => onCopy(record.url)}>
            <Copy class="size-4" />
            {t(language, 'common.copyUrl')}
          </button>
          <a class="studio-button" href={record.url} download={filename} target="_blank" rel="noreferrer">
            <Download class="size-4" />
            {t(language, 'common.download')}
          </a>
          {#if canDelete && onDelete}
            <button class="studio-button" data-tone="danger" type="button" onclick={onDelete}>
              <Trash2 class="size-4" />
              {t(language, 'history.delete')}
            </button>
          {/if}
        </div>
      </footer>
    </div>
  </div>
{/if}
