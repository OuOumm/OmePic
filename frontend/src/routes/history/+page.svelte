<script lang="ts">
  import { Trash2 } from 'lucide-svelte';
  import ConfirmDialog from '@/components/studio/ConfirmDialog.svelte';
  import ImageDataTable from '@/components/studio/ImageDataTable.svelte';
  import ImagePreviewDialog from '@/components/studio/ImagePreviewDialog.svelte';
  import PageTitle from '@/components/studio/PageTitle.svelte';
  import { deleteImageByUid } from '@/api';
  import { clearUploadHistory, deleteUploadFromHistory, getAllUploads, getUploadCount } from '@/indexeddb/upload-history';
  import { getClientToken } from '@/preferences';
  import { t } from '@/i18n';
  import { preferences } from '@/stores/preferences.svelte';
  import { toast } from '@/stores/toast.svelte';
  import type { UploadHistoryRecord } from '@/types';

  let records = $state<UploadHistoryRecord[]>([]);
  let count = $state(0);
  let loading = $state(true);
  let previewRecord = $state<UploadHistoryRecord | null>(null);
  let confirmTarget = $state<UploadHistoryRecord | 'clear' | null>(null);
  let busy = $state(false);

  async function loadData() {
    loading = true;
    try {
      [records, count] = await Promise.all([getAllUploads(), getUploadCount()]);
    } finally {
      loading = false;
    }
  }

  async function clearAll() {
    busy = true;
    await clearUploadHistory();
    records = [];
    count = 0;
    confirmTarget = null;
    busy = false;
    toast.success(t(preferences.language, 'history.cleared'));
  }

  async function remove(record: UploadHistoryRecord) {
    busy = true;
    try {
      await deleteImageByUid(record.uid, record.client_token);
      await deleteUploadFromHistory(record.uid);
      records = records.filter((item) => item.uid !== record.uid);
      if (previewRecord?.uid === record.uid) previewRecord = null;
      confirmTarget = null;
      count -= 1;
      toast.success(t(preferences.language, 'history.deleted'));
    } catch {
      toast.error(t(preferences.language, 'history.deleteError'));
    } finally {
      busy = false;
    }
  }

  function copy(value: string) {
    navigator.clipboard.writeText(value);
    toast.success(t(preferences.language, 'common.copied'));
  }

  $effect(() => { loadData(); });
</script>

<svelte:head><title>{t(preferences.language, 'history.title')} · OmePic</title></svelte:head>

<div class="space-y-7">
  <div class="flex flex-col gap-4 md:flex-row md:items-end md:justify-between">
    <PageTitle eyebrow={t(preferences.language, 'admin.fileDeskEyebrow')} title={t(preferences.language, 'history.title')} subtitle={t(preferences.language, 'history.subtitle')} tone="blue" />
    {#if records.length > 0}
      <button class="studio-button" data-tone="danger" type="button" onclick={() => (confirmTarget = 'clear')}>
        <Trash2 class="size-4" />
        {t(preferences.language, 'history.clear')}
      </button>
    {/if}
  </div>

  <div class="border-y-[3px] ink-line py-3 text-sm font-black">{loading ? t(preferences.language, 'common.loading') : t(preferences.language, 'history.count', { count })}</div>

  {#if !loading && records.length === 0}
    <div class="min-h-72 border-[3px] border-dashed ink-line grid place-items-center text-center">
      <div>
        <p class="text-4xl font-black">{t(preferences.language, 'history.empty')}</p>
        <a class="studio-button mt-5" href="/" data-tone="primary">{t(preferences.language, 'nav.upload')}</a>
      </div>
    </div>
  {:else}
    <ImageDataTable language={preferences.language} {records} canDelete={(record) => record.client_token === getClientToken()} onCopy={copy} onPreview={(record) => (previewRecord = record)} onDelete={(record) => (confirmTarget = record)} />
  {/if}

  <ImagePreviewDialog language={preferences.language} record={previewRecord} records={records} canDelete={previewRecord?.client_token === getClientToken()} onCopy={copy} onDelete={() => previewRecord && (confirmTarget = previewRecord)} onNavigate={(record) => (previewRecord = record)} onClose={() => (previewRecord = null)} />
  <ConfirmDialog
    open={confirmTarget !== null}
    title={confirmTarget === 'clear' ? t(preferences.language, 'history.clearConfirm') : t(preferences.language, 'history.deleteConfirm')}
    description={confirmTarget === 'clear' ? t(preferences.language, 'history.count', { count }) : confirmTarget?.original_filename || confirmTarget?.uid || ''}
    confirmLabel={confirmTarget === 'clear' ? t(preferences.language, 'history.clear') : t(preferences.language, 'common.delete')}
    cancelLabel={t(preferences.language, 'common.cancel')}
    {busy}
    onClose={() => (confirmTarget = null)}
    onConfirm={() => confirmTarget === 'clear' ? clearAll() : confirmTarget && remove(confirmTarget)}
  />
</div>

