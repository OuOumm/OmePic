<script lang="ts">
  import { ExternalLink, Image as ImageIcon, Trash2 } from 'lucide-svelte';
  import type { Language, UploadHistoryRecord } from '@/types';
  import { t } from '@/i18n';
  import { formatBytes, safeImageUrl } from '@/utils';

  export let language: Language;
  export let records: UploadHistoryRecord[] = [];
  export let canDelete: (record: UploadHistoryRecord) => boolean = () => false;
  export let selectable = false;
  export let selectedUids: ReadonlySet<string> | readonly string[] = [];
  export let canSelect: (record: UploadHistoryRecord) => boolean = () => true;
  export let onToggleSelect: (record: UploadHistoryRecord) => void = () => {};
  export let onToggleSelectAll: (checked: boolean) => void = () => {};
  export let onCopy: (value: string) => void;
  export let onPreview: (record: UploadHistoryRecord) => void;
  export let onDelete: (record: UploadHistoryRecord) => void;

  $: selectedUidList = 'has' in selectedUids ? Array.from(selectedUids) : [...selectedUids];
  $: selectableRecords = selectable ? records.filter(canSelect) : [];
  $: allSelectableSelected = selectableRecords.length > 0 && selectableRecords.every((record) => selectedUidList.includes(record.uid));
</script>

<div class="overflow-x-auto">
  <table class="w-full min-w-[900px] border-collapse text-sm">
    <thead>
      <tr class="border-b-[3px] ink-line text-left text-xs font-black uppercase tracking-[0.12em] text-[hsl(var(--ink-muted))]">
        {#if selectable}
          <th class="w-9 px-2 py-2" scope="col">
            <input type="checkbox" checked={allSelectableSelected} disabled={selectableRecords.length === 0} aria-label={t(language, 'history.selectAll')} onchange={(event) => onToggleSelectAll(event.currentTarget.checked)} />
          </th>
        {/if}
        <th class="px-2 py-2" scope="col">{t(language, 'image.filename')}</th>
        <th class="w-[90px] px-2 py-2" scope="col">{t(language, 'image.size')}</th>
        <th class="w-[120px] px-2 py-2" scope="col">{t(language, 'image.storage')}</th>
        <th class="w-[180px] px-2 py-2 text-right" scope="col">{t(language, 'admin.imagesTableActions')}</th>
      </tr>
    </thead>
    <tbody>
      {#each records as record (record.uid)}
        {@const imageUrl = safeImageUrl(record.url)}
        <tr class="studio-table-row align-middle">
          {#if selectable}
            <td class="px-2 py-3 align-middle">
              <input type="checkbox" checked={selectedUidList.includes(record.uid)} disabled={!canSelect(record)} aria-label={t(language, 'history.selectRecord', { title: record.original_filename || record.uid })} onchange={() => onToggleSelect(record)} />
            </td>
          {/if}
          <th class="min-w-0 px-2 py-3 text-left font-normal" scope="row">
            <button class="flex min-w-0 items-center gap-3 text-left" type="button" onclick={() => onPreview(record)} aria-label={t(language, 'common.openPreview', { title: record.original_filename || record.uid })}>
              <span class="grid size-12 shrink-0 place-items-center overflow-hidden border-2 ink-line bg-[hsl(var(--paper-deep))]">
                {#if imageUrl}
                  <img src={imageUrl} alt={record.original_filename || record.uid} class="h-full w-full object-cover" loading="lazy" decoding="async" width="48" height="48" />
                {:else}
                  <ImageIcon class="size-5" aria-hidden="true" />
                {/if}
              </span>
              <span class="min-w-0">
                <span class="block truncate font-black">{record.original_filename || record.uid}</span>
                <span class="block truncate text-xs text-[hsl(var(--ink-muted))]">{record.uid}</span>
              </span>
            </button>
          </th>
          <td class="px-2 py-3 font-bold">{formatBytes(record.size, language)}</td>
          <td class="min-w-0 px-2 py-3 font-bold"><span class="block truncate">{record.storage_key}</span></td>
          <td class="px-2 py-3">
            <div class="flex min-w-[170px] flex-nowrap justify-end gap-1.5 overflow-visible whitespace-nowrap">
              <button class="studio-button p-2 text-xs" type="button" onclick={() => onCopy(record.url)} aria-label={t(language, 'common.copyUrl')}>URL</button>
              <button class="studio-button p-2 text-xs" type="button" onclick={() => onCopy(record.markdown)} aria-label={t(language, 'common.copyMarkdown')}>MD</button>
              <button class="studio-button p-2 text-xs" type="button" onclick={() => onCopy(record.bbcode)} aria-label={t(language, 'common.copyBBCode')}>BB</button>
              {#if imageUrl}
                <a class="studio-button p-2" href={imageUrl} target="_blank" rel="noopener noreferrer" aria-label={t(language, 'common.openPreview', { title: record.uid })}><ExternalLink class="size-4" aria-hidden="true" /></a>
              {/if}
              {#if canDelete(record)}
                <button class="studio-button p-2" data-tone="danger" type="button" onclick={() => onDelete(record)} aria-label={t(language, 'history.delete')}><Trash2 class="size-4" aria-hidden="true" /></button>
              {/if}
            </div>
          </td>
        </tr>
      {/each}
    </tbody>
  </table>
</div>
