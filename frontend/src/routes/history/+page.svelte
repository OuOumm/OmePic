<script lang="ts">
  import { Trash2 } from 'lucide-svelte';
  import ImageDataRow from '@/components/studio/ImageDataRow.svelte';
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

  async function loadData() {
    loading = true;
    try {
      [records, count] = await Promise.all([getAllUploads(), getUploadCount()]);
    } finally {
      loading = false;
    }
  }

  async function clearAll() {
    if (!confirm(t(preferences.language, 'history.clearConfirm'))) return;
    await clearUploadHistory();
    records = [];
    count = 0;
    toast.success(t(preferences.language, 'history.cleared'));
  }

  async function remove(record: UploadHistoryRecord) {
    if (!confirm(t(preferences.language, 'history.deleteConfirm'))) return;
    try {
      await deleteImageByUid(record.uid, getClientToken());
      await deleteUploadFromHistory(record.uid);
      records = records.filter((item) => item.uid !== record.uid);
      if (previewRecord?.uid === record.uid) previewRecord = null;
      count -= 1;
      toast.success(t(preferences.language, 'history.deleted'));
    } catch {
      toast.error(t(preferences.language, 'history.deleteError'));
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
    <PageTitle eyebrow="File desk" title={t(preferences.language, 'history.title')} subtitle={t(preferences.language, 'history.subtitle')} tone="blue" />
    {#if records.length > 0}
      <button class="studio-button" data-tone="danger" type="button" onclick={clearAll}>
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
    <div class="overflow-x-auto">
      <div class="min-w-[760px]">
        {#each records as record (record.uid)}
          <ImageDataRow language={preferences.language} {record} canDelete={record.client_token === getClientToken()} onCopy={copy} onPreview={() => (previewRecord = record)} onDelete={() => remove(record)} />
        {/each}
      </div>
    </div>
  {/if}

  <ImagePreviewDialog language={preferences.language} record={previewRecord} canDelete={previewRecord?.client_token === getClientToken()} onCopy={copy} onDelete={() => previewRecord && remove(previewRecord)} onClose={() => (previewRecord = null)} />
</div>

