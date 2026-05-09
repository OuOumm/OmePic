<script lang="ts">
  import { ExternalLink, Image as ImageIcon, Trash2 } from 'lucide-svelte';
  import type { Language, UploadHistoryRecord } from '@/types';
  import { t } from '@/i18n';
  import { formatBytes } from '@/utils';

  export let language: Language;
  export let record: UploadHistoryRecord;
  export let canDelete = false;
  export let onCopy: (value: string) => void;
  export let onPreview: (() => void) | undefined = undefined;
  export let onDelete: (() => void) | undefined = undefined;
</script>

<div class="grid grid-cols-[minmax(0,1.5fr)_90px_minmax(90px,120px)_minmax(0,1fr)] items-center gap-3 studio-table-row py-3 text-sm">
  <button class="flex min-w-0 items-center gap-3 text-left" type="button" onclick={onPreview} disabled={!onPreview} aria-label={t(language, 'common.openPreview', { title: record.original_filename || record.uid })}>
    <span class="grid size-12 shrink-0 place-items-center overflow-hidden border-2 ink-line bg-[hsl(var(--paper-deep))]">
      {#if record.url}
        <img src={record.url} alt={record.original_filename || record.uid} class="h-full w-full object-cover" loading="lazy" />
      {:else}
        <ImageIcon class="size-5" />
      {/if}
    </span>
    <span class="min-w-0">
      <span class="block truncate font-black">{record.original_filename || record.uid}</span>
      <span class="block truncate text-xs text-[hsl(var(--ink-muted))]">{record.uid}</span>
    </span>
  </button>
  <div class="font-bold">{formatBytes(record.size)}</div>
  <div class="min-w-0 truncate font-bold">{record.storage_key}</div>
  <div class="flex min-w-0 flex-wrap justify-end gap-1.5 overflow-hidden">
    <button class="studio-button p-2 text-xs" type="button" onclick={() => onCopy(record.url)} aria-label={t(language, 'common.copyUrl')}>URL</button>
    <button class="studio-button p-2 text-xs" type="button" onclick={() => onCopy(record.markdown)} aria-label={t(language, 'common.copyMarkdown')}>MD</button>
    <button class="studio-button p-2 text-xs" type="button" onclick={() => onCopy(record.bbcode)} aria-label={t(language, 'common.copyBBCode')}>BB</button>
    <a class="studio-button p-2" href={record.url} target="_blank" rel="noreferrer" aria-label={t(language, 'common.openPreview', { title: record.uid })}><ExternalLink class="size-4" /></a>
    {#if canDelete && onDelete}
      <button class="studio-button p-2" data-tone="danger" type="button" onclick={onDelete} aria-label={t(language, 'history.delete')}><Trash2 class="size-4" /></button>
    {/if}
  </div>
</div>

