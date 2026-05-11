<script lang="ts">
  import { Bell, Link, Loader2 } from 'lucide-svelte';
  import AnnouncementDialog from '@/components/studio/AnnouncementDialog.svelte';
  import CanvasDropzone from '@/components/studio/CanvasDropzone.svelte';
  import ConfirmDialog from '@/components/studio/ConfirmDialog.svelte';
  import ImageDataTable from '@/components/studio/ImageDataTable.svelte';
  import ImagePreviewDialog from '@/components/studio/ImagePreviewDialog.svelte';
  import StorageInspector from '@/components/studio/StorageInspector.svelte';
  import { ApiError, deleteImageByUid, getAnnouncements, getRuntimeSettings, uploadImageWithProgress } from '@/api';
  import { getClientToken } from '@/preferences';
  import { saveUploadToHistory, deleteUploadFromHistory, getRecentUploads } from '@/indexeddb/upload-history';
  import { t } from '@/i18n';
  import { preferences, setRuntimeSettings, setSelectedStorageKey } from '@/stores/preferences.svelte';
  import { toast } from '@/stores/toast.svelte';
  import type { Announcement, Language, UploadHistoryRecord, UploadResult } from '@/types';

  type UploadTask = {
    id: string;
    file: File;
    progress: number;
    status: 'pending' | 'uploading' | 'success' | 'error';
    result?: UploadResult;
    error?: string;
  };

  let tasks = $state<UploadTask[]>([]);
  let recentUploads = $state<UploadHistoryRecord[]>([]);
  let runtimeLoading = $state(true);
  let runtimeError = $state<string | null>(null);
  let dragging = $state(false);
  let urlInput = $state('');
  let urlUploading = $state(false);
  let announcements = $state<Announcement[]>([]);
  let announcementDialogOpen = $state(false);
  let announcementDialogMode = $state<'detail' | 'history'>('detail');
  let previewRecord = $state<UploadHistoryRecord | null>(null);
  let deleteTarget = $state<UploadHistoryRecord | null>(null);
  let deleting = $state(false);
  let counter = 0;

  const activeTasks = $derived(tasks.filter((task) => task.status === 'pending' || task.status === 'uploading' || task.status === 'error'));
  const siteName = $derived(preferences.runtimeSettings?.site.name || 'OmePic');
  const siteTitle = $derived(preferences.runtimeSettings?.site.tagline ? `${siteName} - ${preferences.runtimeSettings.site.tagline}` : siteName);
  const allowedMimeTypesText = $derived(preferences.runtimeSettings?.upload.effective_allowed_mime_types?.join(', ') ?? '');
  const maintenanceMode = $derived(preferences.runtimeSettings?.features.maintenance_mode ?? false);
  const uploadDisabled = $derived(runtimeLoading || maintenanceMode || activeTasks.some((task) => task.status === 'pending' || task.status === 'uploading'));

  function uploadErrorMessage(lang: Language, err: unknown): string {
    if (err instanceof ApiError && err.code === 'rate_limited') {
      return typeof err.retryAfter === 'number'
        ? t(lang, 'upload.rateLimitedWithRetry', { seconds: err.retryAfter })
        : t(lang, 'upload.rateLimited');
    }
    if (err instanceof ApiError && err.code === 'network_error') return t(lang, 'upload.networkError');
    return err instanceof Error ? err.message : t(lang, 'upload.error');
  }

  async function loadRuntime(showLoading = true) {
    if (showLoading) runtimeLoading = true;
    runtimeError = null;
    try {
      const settings = await getRuntimeSettings();
      setRuntimeSettings(settings);
      if (preferences.selectedStorageKey && !settings.storage.options.some((option) => option.storage_key === preferences.selectedStorageKey)) {
        setSelectedStorageKey('');
      }
    } catch (err) {
      runtimeError = err instanceof Error ? err.message : t(preferences.language, 'common.error');
      setRuntimeSettings(null);
    } finally {
      runtimeLoading = false;
    }
  }

  async function loadRecent() {
    try {
      recentUploads = await getRecentUploads(10);
    } catch {
      recentUploads = [];
    }
  }

  async function loadAnnouncements() {
    try {
      const items = await getAnnouncements();
      announcements = items;
      const latestStamp = items[0]?.updated_at || items[0]?.created_at || '';
      const seenStamp = localStorage.getItem('omepic:announcement:lastSeen') ?? '';
      if (latestStamp && latestStamp !== seenStamp) {
        announcementDialogMode = 'detail';
        announcementDialogOpen = true;
      }
    } catch {
      announcements = [];
    }
  }

  function closeAnnouncementDialog() {
    const latestStamp = announcements[0]?.updated_at || announcements[0]?.created_at || '';
    if (latestStamp) localStorage.setItem('omepic:announcement:lastSeen', latestStamp);
    announcementDialogOpen = false;
  }

  function openAnnouncementDialog(mode: 'detail' | 'history' = 'detail') {
    announcementDialogMode = mode;
    announcementDialogOpen = true;
  }

  function validateFiles(files: File[]) {
    const settings = preferences.runtimeSettings;
    if (!settings) return files;
    const maxBytes = settings.upload.max_upload_size_mb > 0 ? settings.upload.max_upload_size_mb * 1024 * 1024 : 0;
    const allowedTypes = settings.upload.effective_allowed_mime_types;
    return files.filter((file) => {
      if (maxBytes > 0 && file.size > maxBytes) {
        toast.error(`${file.name}: ${t(preferences.language, 'upload.error')}`);
        return false;
      }
      if (allowedTypes.length > 0 && !allowedTypes.includes(file.type.toLowerCase())) {
        toast.error(`${file.name}: ${t(preferences.language, 'upload.error')}`);
        return false;
      }
      return true;
    });
  }

  async function uploadTask(task: UploadTask) {
    tasks = tasks.map((item) => (item.id === task.id ? { ...item, status: 'uploading', progress: 0 } : item));
    const token = getClientToken();
    try {
      const result = await uploadImageWithProgress(
        task.file,
        token,
        (progress) => {
          tasks = tasks.map((item) => (item.id === task.id ? { ...item, progress } : item));
        },
        preferences.selectedStorageKey || undefined,
      );
      tasks = tasks.map((item) => (item.id === task.id ? { ...item, status: 'success', progress: 100, result } : item));
      await saveUploadToHistory({
        uid: result.uid,
        url: result.url,
        mime_type: result.mime_type,
        size: result.size,
        created_at: result.created_at,
        is_duplicate: result.is_duplicate,
        storage_key: result.storage_key,
        storage_backend: result.storage_backend,
        markdown: result.markdown,
        bbcode: result.bbcode,
        client_token: token,
        original_filename: task.file.name,
        saved_at: new Date().toISOString(),
      });
      toast.success(result.is_duplicate ? t(preferences.language, 'upload.duplicate') : t(preferences.language, 'upload.success'));
      await loadRecent();
    } catch (err) {
      const message = uploadErrorMessage(preferences.language, err);
      tasks = tasks.map((item) => (item.id === task.id ? { ...item, status: 'error', error: message } : item));
      toast.error(`${task.file.name}: ${message}`);
    }
  }

  async function handleFiles(files: File[]) {
    if (maintenanceMode) {
      toast.error(preferences.runtimeSettings?.features.maintenance_message ?? t(preferences.language, 'common.error'));
      return;
    }
    const accepted = validateFiles(files);
    const next = accepted.map((file) => ({ id: `task-${++counter}`, file, progress: 0, status: 'pending' as const }));
    tasks = [...next, ...tasks];
    await Promise.all(next.map(uploadTask));
  }

  async function handleUrlUpload() {
    const url = urlInput.trim();
    if (!url) return;
    if (!/^https?:\/\//i.test(url)) {
      toast.error(t(preferences.language, 'upload.invalidUrl'));
      return;
    }
    urlUploading = true;
    try {
      const response = await fetch(url);
      if (!response.ok) throw new Error('Download failed');
      const blob = await response.blob();
      const mimeType = response.headers.get('Content-Type') || blob.type;
      if (!mimeType.startsWith('image/')) {
        toast.error(t(preferences.language, 'upload.urlNotImage'));
        return;
      }
      const filename = url.split('/').pop()?.split('?')[0] || 'image';
      toast.success(t(preferences.language, 'upload.urlSuccess'));
      urlInput = '';
      await handleFiles([new File([blob], filename, { type: mimeType })]);
    } catch {
      toast.error(t(preferences.language, 'upload.urlDownloadFail'));
    } finally {
      urlUploading = false;
    }
  }

  function copy(value: string) {
    navigator.clipboard.writeText(value);
    toast.success(t(preferences.language, 'common.copied'));
  }

  async function removeRecent(record: UploadHistoryRecord) {
    deleting = true;
    try {
      await deleteImageByUid(record.uid, getClientToken());
      await deleteUploadFromHistory(record.uid);
      recentUploads = recentUploads.filter((item) => item.uid !== record.uid);
      if (previewRecord?.uid === record.uid) previewRecord = null;
      deleteTarget = null;
      toast.success(t(preferences.language, 'history.deleted'));
    } catch {
      toast.error(t(preferences.language, 'history.deleteError'));
    } finally {
      deleting = false;
    }
  }

  $effect(() => {
    loadRuntime();
    loadRecent();
    loadAnnouncements();
    const pasteHandler = (event: ClipboardEvent) => {
      const target = event.target as HTMLElement;
      if (target.tagName === 'INPUT' || target.tagName === 'TEXTAREA' || target.isContentEditable) return;
      const files = Array.from(event.clipboardData?.items ?? [])
        .filter((item) => item.type.startsWith('image/'))
        .map((item) => item.getAsFile())
        .filter((file): file is File => Boolean(file));
      if (files.length) {
        event.preventDefault();
        handleFiles(files);
      }
    };
    window.addEventListener('paste', pasteHandler);
    return () => window.removeEventListener('paste', pasteHandler);
  });
</script>

<svelte:head><title>{siteTitle}</title></svelte:head>

<div class="grid gap-6 lg:grid-cols-[1fr_320px]">
  <section class="space-y-6">
    {#if runtimeError}
      <div class="studio-panel border-[hsl(var(--danger))] p-4 text-sm font-bold text-[hsl(var(--danger))]" role="alert">{runtimeError}</div>
    {/if}
    {#if maintenanceMode}
      <div class="studio-panel bg-[hsl(var(--marker-yellow)/0.35)] p-4 text-sm font-bold">{preferences.runtimeSettings?.features.maintenance_message}</div>
    {/if}

    <div class="relative" role="presentation" ondragenter={() => (dragging = true)} ondragleave={() => (dragging = false)}>
      {#if announcements.length}
        <button class="studio-button absolute right-3 top-3 z-20 p-2 text-xs shadow-[4px_4px_0_hsl(var(--ink))] rotate-[1deg] sm:p-2.5 sm:text-sm" type="button" onclick={() => openAnnouncementDialog('history')}>
          <Bell class="size-4" />
          {t(preferences.language, 'announcement.entry', { count: announcements.length })}
        </button>
      {/if}
      <CanvasDropzone language={preferences.language} disabled={uploadDisabled} {dragging} allowedTypes={allowedMimeTypesText} onFiles={handleFiles} />
    </div>

    <div class="grid gap-3 border-y-[3px] ink-line py-5 md:grid-cols-[1fr_auto] md:items-end">
      <label class="grid gap-2 text-sm font-black">
        {t(preferences.language, 'upload.urlLabel')}
        <input class="studio-input" bind:value={urlInput} type="url" name="image-url" autocomplete="url" inputmode="url" placeholder={t(preferences.language, 'upload.urlPlaceholder')} onkeydown={(event) => event.key === 'Enter' && handleUrlUpload()} />
      </label>
      <button class="studio-button" type="button" onclick={handleUrlUpload} disabled={urlUploading || uploadDisabled || !urlInput.trim()} data-tone="primary">
        {#if urlUploading}<Loader2 class="size-4 animate-spin" />{:else}<Link class="size-4" />{/if}
        {t(preferences.language, 'upload.urlUpload')}
      </button>
    </div>
  </section>

  <aside class="space-y-5">
    <StorageInspector
      language={preferences.language}
      settings={preferences.runtimeSettings}
      selected={preferences.selectedStorageKey}
      refreshing={runtimeLoading}
      onSelect={setSelectedStorageKey}
      onRefresh={() => loadRuntime(false)}
    />
    <div class="studio-panel p-4 rotate-[-0.35deg]">
      <h2 class="border-b-2 ink-line pb-2 text-xl font-black">{t(preferences.language, 'upload.queueTitle')}</h2>
      {#if activeTasks.length}
        <ul class="mt-3 grid gap-3">
          {#each activeTasks as task (task.id)}
            <li class="min-w-0 overflow-hidden border-b-2 border-dashed border-[hsl(var(--ink)/0.32)] pb-3 text-sm">
              <div class="grid min-w-0 grid-cols-[minmax(0,1fr)_auto] items-center gap-3 font-black overflow-hidden"><span class="block min-w-0 truncate" title={task.file.name}>{task.file.name}</span><span class="text-right whitespace-nowrap">{task.progress}%</span></div>
              <div class="mt-2 h-2 border-2 ink-line" role="progressbar" aria-label={task.file.name} aria-valuemin="0" aria-valuemax="100" aria-valuenow={task.progress}><div class="h-full bg-[hsl(var(--marker-green))]" style={`width:${task.progress}%`}></div></div>
              {#if task.error}<p class="mt-1 text-xs text-[hsl(var(--danger))]">{task.error}</p>{/if}
            </li>
          {/each}
        </ul>
      {:else}
        <p class="mt-3 text-sm text-[hsl(var(--ink-muted))]">{t(preferences.language, 'upload.queueEmpty')}</p>
      {/if}
    </div>
  </aside>
</div>

{#if recentUploads.length}
  <section class="mt-10">
    <div class="mb-4 flex items-end justify-between border-b-[3px] ink-line pb-3">
      <div>
        <p class="text-xs font-black uppercase text-[hsl(var(--ink-muted))]">{t(preferences.language, 'admin.fileDeskEyebrow')}</p>
        <h2 class="text-3xl font-black">{t(preferences.language, 'upload.recentTitle')}</h2>
      </div>
    </div>
    <ImageDataTable language={preferences.language} records={recentUploads} canDelete={(record) => record.client_token === getClientToken()} onCopy={copy} onPreview={(record) => (previewRecord = record)} onDelete={(record) => (deleteTarget = record)} />
  </section>
{/if}

<ImagePreviewDialog language={preferences.language} record={previewRecord} records={recentUploads} canDelete={previewRecord?.client_token === getClientToken()} onCopy={copy} onDelete={() => previewRecord && (deleteTarget = previewRecord)} onNavigate={(record) => (previewRecord = record)} onClose={() => (previewRecord = null)} />
<AnnouncementDialog language={preferences.language} announcements={announcements} open={announcementDialogOpen} initialMode={announcementDialogMode} onClose={closeAnnouncementDialog} />
<ConfirmDialog
  open={deleteTarget !== null}
  title={t(preferences.language, 'history.deleteConfirm')}
  description={deleteTarget?.original_filename || deleteTarget?.uid || ''}
  confirmLabel={t(preferences.language, 'common.delete')}
  cancelLabel={t(preferences.language, 'common.cancel')}
  busy={deleting}
  onClose={() => (deleteTarget = null)}
  onConfirm={() => deleteTarget && removeRecent(deleteTarget)}
/>
