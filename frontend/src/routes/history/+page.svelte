<script lang="ts">
  import { Search, Trash2 } from 'lucide-svelte';
  import ConfirmDialog from '@/components/studio/ConfirmDialog.svelte';
  import ImageDataTable from '@/components/studio/ImageDataTable.svelte';
  import ImagePreviewDialog from '@/components/studio/ImagePreviewDialog.svelte';
  import PageTitle from '@/components/studio/PageTitle.svelte';
  import { deleteImageByUid } from '@/api';
  import { copyToClipboard } from '@/clipboard';
  import { buildUploadHistoryPage, clearUploadHistory, deleteUploadFromHistory, getAllUploads } from '@/indexeddb/upload-history';
  import { getClientToken } from '@/client-token';
  import { t } from '@/i18n';
  import { preferences } from '@/stores/preferences.svelte';
  import { toast } from '@/stores/toast.svelte';
  import type { UploadHistoryRecord } from '@/types';

  let allRecords = $state<UploadHistoryRecord[]>([]);
  let records = $state<UploadHistoryRecord[]>([]);
  let filteredRecords = $state<UploadHistoryRecord[]>([]);
  let count = $state(0);
  let filteredCount = $state(0);
  let page = $state(1);
  let pageSize = $state(20);
  let totalPages = $state(1);
  let search = $state('');
  let loading = $state(true);
  let previewRecord = $state<UploadHistoryRecord | null>(null);
  let selectedUids = $state<string[]>([]);
  let confirmTarget = $state<UploadHistoryRecord | 'clear' | 'selected' | null>(null);
  let busy = $state(false);

  const pageSizeOptions = [10, 20, 50];
  const hasSearch = $derived(search.trim().length > 0);
  const selectedCount = $derived(selectedUids.length);
  const showingEmptyState = $derived(!loading && count === 0);
  const showingNoMatches = $derived(!loading && count > 0 && filteredCount === 0);

  function applyHistoryPage(currentPage = page) {
    const pageData = buildUploadHistoryPage(allRecords, { query: search, page: currentPage, pageSize });
    const navigationData = buildUploadHistoryPage(allRecords, { query: search, page: 1, pageSize: Math.max(pageData.filteredTotal, 1) });
    records = pageData.records;
    filteredRecords = navigationData.records;
    selectedUids = selectedUids.filter((uid) => allRecords.some((record) => record.uid === uid));
    count = pageData.total;
    filteredCount = pageData.filteredTotal;
    page = pageData.page;
    totalPages = pageData.totalPages;
  }

  async function loadData(currentPage = page) {
    loading = true;
    try {
      allRecords = await getAllUploads();
      applyHistoryPage(currentPage);
    } finally {
      loading = false;
    }
  }

  async function clearAll() {
    busy = true;
    await clearUploadHistory();
    allRecords = [];
    records = [];
    filteredRecords = [];
    count = 0;
    filteredCount = 0;
    page = 1;
    totalPages = 1;
    selectedUids = [];
    confirmTarget = null;
    busy = false;
    toast.success(t(preferences.language, 'history.cleared'));
  }

  async function remove(record: UploadHistoryRecord) {
    busy = true;
    try {
      await deleteImageByUid(record.uid, record.client_token);
      await deleteUploadFromHistory(record.uid);
      allRecords = allRecords.filter((item) => item.uid !== record.uid);
      selectedUids = selectedUids.filter((uid) => uid !== record.uid);
      applyHistoryPage(page);
      if (previewRecord?.uid === record.uid) previewRecord = null;
      confirmTarget = null;
      toast.success(t(preferences.language, 'history.deleted'));
    } catch {
      toast.error(t(preferences.language, 'history.deleteError'));
    } finally {
      busy = false;
    }
  }

  async function removeSelected() {
    const targets = allRecords.filter((record) => selectedUids.includes(record.uid));
    if (targets.length === 0) return;

    busy = true;
    try {
      await Promise.all(targets.map((record) => deleteImageByUid(record.uid, record.client_token)));
      await Promise.all(targets.map((record) => deleteUploadFromHistory(record.uid)));
      const targetUids = targets.map((record) => record.uid);
      allRecords = allRecords.filter((record) => !targetUids.includes(record.uid));
      selectedUids = [];
      if (previewRecord && targetUids.includes(previewRecord.uid)) previewRecord = null;
      applyHistoryPage(page);
      confirmTarget = null;
      toast.success(t(preferences.language, 'history.bulkDeleted', { count: targets.length }));
    } catch {
      toast.error(t(preferences.language, 'history.deleteError'));
    } finally {
      busy = false;
    }
  }

  function copy(value: string) {
    void copyToClipboard(value, preferences.language);
  }

  function toggleSelected(record: UploadHistoryRecord) {
    selectedUids = selectedUids.includes(record.uid)
      ? selectedUids.filter((uid) => uid !== record.uid)
      : [...selectedUids, record.uid];
  }

  function toggleSelectPage(checked: boolean) {
    if (checked) {
      selectedUids = [...selectedUids, ...records.map((record) => record.uid).filter((uid) => !selectedUids.includes(uid))];
      return;
    }

    const pageUids = records.map((record) => record.uid);
    selectedUids = selectedUids.filter((uid) => !pageUids.includes(uid));
  }

  function clearSelected() {
    selectedUids = [];
  }

  function setSearch(value: string) {
    search = value;
    applyHistoryPage(1);
  }

  function setPageSize(value: string) {
    const nextPageSize = Number(value);
    pageSize = pageSizeOptions.includes(nextPageSize) ? nextPageSize : 20;
    applyHistoryPage(1);
  }

  function goToPage(nextPage: number) {
    applyHistoryPage(nextPage);
  }

  $effect(() => { loadData(1); });
</script>

<svelte:head><title>{t(preferences.language, 'history.title')} · OmePic</title></svelte:head>

<div class="space-y-7">
  <div class="flex flex-col gap-4 md:flex-row md:items-end md:justify-between">
    <PageTitle eyebrow={t(preferences.language, 'admin.fileDeskEyebrow')} title={t(preferences.language, 'history.title')} subtitle={t(preferences.language, 'history.subtitle')} tone="blue" />
    {#if count > 0}
      <button class="studio-button" data-tone="danger" type="button" onclick={() => (confirmTarget = 'clear')}>
        <Trash2 class="size-4" />
        {t(preferences.language, 'history.clear')}
      </button>
    {/if}
  </div>

  <div class="grid grid-cols-[minmax(0,1fr)_auto_auto] items-center gap-2 border-y-[3px] ink-line py-3 sm:gap-3">
    <label class="relative min-w-0">
      <span class="sr-only">{t(preferences.language, 'history.searchLabel')}</span>
      <Search class="absolute left-3 top-1/2 size-4 -translate-y-1/2" aria-hidden="true" />
      <input class="studio-input w-full pl-10" value={search} name="history-search" autocomplete="off" placeholder={t(preferences.language, 'history.searchPlaceholder')} oninput={(event) => setSearch(event.currentTarget.value)} />
    </label>
    <label class="flex items-center gap-1 text-sm font-black sm:gap-2">
      <span class="whitespace-nowrap">{t(preferences.language, 'history.pageSize')}</span>
      <select class="studio-input w-16 py-2 sm:w-24" value={pageSize} aria-label={t(preferences.language, 'history.pageSize')} onchange={(event) => setPageSize(event.currentTarget.value)}>
        {#each pageSizeOptions as option (option)}
          <option value={option}>{option}</option>
        {/each}
      </select>
    </label>
    <div class="whitespace-nowrap text-right text-sm font-black">
      {#if loading}
        {t(preferences.language, 'common.loading')}
      {:else}
        {t(preferences.language, 'history.count', { count: hasSearch ? filteredCount : count })}
      {/if}
    </div>
    {#if count > 0}
      <div class="col-span-3 grid gap-2 border-t-2 border-dashed border-[hsl(var(--ink)/0.28)] pt-3 md:flex md:justify-end">
        <div class="grid grid-cols-2 gap-2 md:flex md:justify-end">
          <button class="studio-button justify-center px-3 py-2 text-sm" type="button" disabled={selectedCount === 0} onclick={clearSelected}>{t(preferences.language, 'history.clearSelection')}</button>
          <button class="studio-button justify-center px-3 py-2 text-sm" data-tone="danger" type="button" disabled={selectedCount === 0} onclick={() => (confirmTarget = 'selected')}><Trash2 class="size-4" />{t(preferences.language, 'history.deleteSelected', { count: selectedCount })}</button>
        </div>
      </div>
    {/if}
  </div>

  {#if showingEmptyState}
    <div class="min-h-72 border-[3px] border-dashed ink-line grid place-items-center text-center">
      <div>
        <p class="text-4xl font-black">{t(preferences.language, 'history.empty')}</p>
        <a class="studio-button mt-5" href="/" data-tone="primary">{t(preferences.language, 'nav.upload')}</a>
      </div>
    </div>
  {:else if showingNoMatches}
    <div class="min-h-64 border-[3px] border-dashed ink-line grid place-items-center text-center">
      <div>
        <p class="text-3xl font-black">{t(preferences.language, 'history.noMatches')}</p>
        <p class="mt-2 text-sm font-bold text-[hsl(var(--ink-muted))]">{t(preferences.language, 'history.noMatchesHint')}</p>
      </div>
    </div>
  {:else}
    <ImageDataTable
      language={preferences.language}
      {records}
      selectable
      selectedUids={selectedUids}
      canDelete={(record) => record.client_token === getClientToken()}
      onToggleSelect={toggleSelected}
      onToggleSelectAll={toggleSelectPage}
      onCopy={copy}
      onPreview={(record) => (previewRecord = record)}
      onDelete={(record) => (confirmTarget = record)}
    />
    <div class="grid grid-cols-[auto_1fr_auto] items-center gap-2 border-t-[3px] ink-line pt-3">
      <button class="studio-button px-3 py-1.5 text-sm" type="button" disabled={page <= 1} onclick={() => goToPage(page - 1)}>{t(preferences.language, 'admin.imagesPrev')}</button>
      <span class="min-w-0 justify-center text-center text-sm font-black">{t(preferences.language, 'history.pageStatus', { page, totalPages })}</span>
      <button class="studio-button px-3 py-1.5 text-sm" type="button" disabled={page >= totalPages} onclick={() => goToPage(page + 1)}>{t(preferences.language, 'admin.imagesNext')}</button>
    </div>
  {/if}

  <ImagePreviewDialog language={preferences.language} record={previewRecord} records={filteredRecords} canDelete={previewRecord?.client_token === getClientToken()} onCopy={copy} onDelete={() => previewRecord && (confirmTarget = previewRecord)} onNavigate={(record) => (previewRecord = record)} onClose={() => (previewRecord = null)} />
  <ConfirmDialog
    open={confirmTarget !== null}
    title={confirmTarget === 'clear' ? t(preferences.language, 'history.clearConfirm') : confirmTarget === 'selected' ? t(preferences.language, 'history.deleteSelectedConfirm', { count: selectedCount }) : t(preferences.language, 'history.deleteConfirm')}
    description={confirmTarget === 'clear' ? t(preferences.language, 'history.count', { count }) : confirmTarget === 'selected' ? t(preferences.language, 'history.selectedCount', { count: selectedCount }) : confirmTarget?.original_filename || confirmTarget?.uid || ''}
    confirmLabel={confirmTarget === 'clear' ? t(preferences.language, 'history.clear') : t(preferences.language, 'common.delete')}
    cancelLabel={t(preferences.language, 'common.cancel')}
    {busy}
    onClose={() => (confirmTarget = null)}
    onConfirm={() => confirmTarget === 'clear' ? clearAll() : confirmTarget === 'selected' ? removeSelected() : confirmTarget && remove(confirmTarget)}
  />
</div>
