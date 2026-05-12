<script lang="ts">
  import { Bell, Link, Loader2 } from 'lucide-svelte';
  import AnnouncementDialog from '@/components/studio/AnnouncementDialog.svelte';
  import CanvasDropzone from '@/components/studio/CanvasDropzone.svelte';
  import ConfirmDialog from '@/components/studio/ConfirmDialog.svelte';
  import ImageDataTable from '@/components/studio/ImageDataTable.svelte';
  import ImagePreviewDialog from '@/components/studio/ImagePreviewDialog.svelte';
  import StorageInspector from '@/components/studio/StorageInspector.svelte';
  import { ApiError, deleteImageByUid, getAnnouncements, getRuntimeSettings, uploadImageWithProgress } from '@/api';
  import { copyToClipboard } from '@/clipboard';
  import { getClientToken } from '@/client-token';
  import { saveUploadToHistory, deleteUploadFromHistory, getRecentUploads } from '@/indexeddb/upload-history';
  import { t } from '@/i18n';
  import { preferences, setRuntimeSettings, setSelectedStorageKey } from '@/stores/preferences.svelte';
  import { toast } from '@/stores/toast.svelte';
  import type { Announcement, Language, UploadHistoryRecord, UploadResult } from '@/types';
  import { imageAcceptFromMimeTypes, imageUrlAllowedOrigins, isAbortError, isAllowedImageMimeType, isBlockedImageMimeType, normalizeDownloadFilename, normalizedImageMimeType } from '@/utils';
  import { errorMessage } from '@/ui-errors';
  import { createProgressReporter, runWithConcurrency } from '@/upload-queue';

  type UploadTask = {
    id: string;
    file: File;
    progress: number;
    status: 'pending' | 'uploading' | 'success' | 'error';
    result?: UploadResult;
    error?: string;
  };

  let tasks = $state<UploadTask[]>([]);
  let recentUploads = $state.raw<UploadHistoryRecord[]>([]);
  let runtimeLoading = $state(true);
  let runtimeError = $state<string | null>(null);
  let dragging = $state(false);
  let urlInput = $state('');
  let urlUploading = $state(false);
  let announcements = $state.raw<Announcement[]>([]);
  let announcementDialogOpen = $state(false);
  let announcementDialogMode = $state<'detail' | 'history'>('detail');
  let previewRecord = $state<UploadHistoryRecord | null>(null);
  let deleteTarget = $state<UploadHistoryRecord | null>(null);
  let deleting = $state(false);
  let counter = 0;

  const activeTasks = $derived(tasks.filter((task) => task.status === 'pending' || task.status === 'uploading' || task.status === 'error'));
  const siteName = $derived(preferences.runtimeSettings?.site.name || 'OmePic');
  const siteTitle = $derived(preferences.runtimeSettings?.site.tagline ? `${siteName} - ${preferences.runtimeSettings.site.tagline}` : siteName);
  const allowedMimeTypes = $derived(preferences.runtimeSettings?.upload.effective_allowed_mime_types ?? []);
  const allowedMimeTypesText = $derived(allowedMimeTypes.join(', '));
  const uploadAccept = $derived(imageAcceptFromMimeTypes(allowedMimeTypes));
  const publicImageAllowedOrigins = $derived(imageUrlAllowedOrigins(preferences.runtimeSettings?.access.public_base_url));
  const maintenanceMode = $derived(preferences.runtimeSettings?.features.maintenance_mode ?? false);
  const uploadDisabled = $derived(runtimeLoading || maintenanceMode || activeTasks.some((task) => task.status === 'pending' || task.status === 'uploading'));
  const uploadConcurrency = 3;

  function uploadErrorMessage(lang: Language, err: unknown): string {
    if (err instanceof ApiError && err.code === 'rate_limited') {
      return typeof err.retryAfter === 'number'
        ? t(lang, 'upload.rateLimitedWithRetry', { seconds: err.retryAfter })
        : t(lang, 'upload.rateLimited');
    }
    if (err instanceof ApiError && err.code === 'network_error') return t(lang, 'upload.networkError');
    return err instanceof Error ? err.message : t(lang, 'upload.error');
  }

  async function loadRuntime(showLoading = true, signal?: AbortSignal) {
    if (showLoading) runtimeLoading = true;
    runtimeError = null;
    try {
      const settings = await getRuntimeSettings(signal);
      if (signal?.aborted) return;
      setRuntimeSettings(settings);
      if (preferences.selectedStorageKey && !settings.storage.options.some((option) => option.storage_key === preferences.selectedStorageKey)) {
        setSelectedStorageKey('');
      }
    } catch (err) {
      if (isAbortError(err)) return;
      runtimeError = errorMessage(err, preferences.language);
      setRuntimeSettings(null);
    } finally {
      if (!signal?.aborted) runtimeLoading = false;
    }
  }

  async function loadRecent() {
    try {
      recentUploads = await getRecentUploads(10);
    } catch {
      recentUploads = [];
    }
  }

  async function loadAnnouncements(signal?: AbortSignal) {
    try {
      const items = await getAnnouncements(signal);
      if (signal?.aborted) return;
      announcements = items;
      const latestStamp = items[0]?.updated_at || items[0]?.created_at || '';
      const seenStamp = localStorage.getItem('omepic:announcement:lastSeen') ?? '';
      if (latestStamp && latestStamp !== seenStamp) {
        announcementDialogMode = 'detail';
        announcementDialogOpen = true;
      }
    } catch (err) {
      if (isAbortError(err)) return;
      announcements = [];
    }
  }

  function closeAnnouncementDialog() {
    announcementDialogOpen = false;
  }

  function acknowledgeAnnouncementDialog() {
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
      if (allowedTypes.length > 0 && !isAllowedImageMimeType(file.type, allowedTypes)) {
        toast.error(`${file.name}: ${t(preferences.language, 'upload.error')}`);
        return false;
      }
      return true;
    });
  }

  function updateTask(id: string, values: Partial<Omit<UploadTask, 'id' | 'file'>>) {
    const task = tasks.find((item) => item.id === id);
    if (task) Object.assign(task, values);
  }

  async function uploadTask(task: UploadTask) {
    updateTask(task.id, { status: 'uploading', progress: 0 });
    const token = getClientToken();
    const reportProgress = createProgressReporter((progress) => {
      updateTask(task.id, { progress });
    });
    try {
      const result = await uploadImageWithProgress(
        task.file,
        token,
        reportProgress,
        preferences.selectedStorageKey || undefined,
      );
      updateTask(task.id, { status: 'success', progress: 100, result });
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
      return result;
    } catch (err) {
      const message = uploadErrorMessage(preferences.language, err);
      updateTask(task.id, { status: 'error', error: message });
      toast.error(`${task.file.name}: ${message}`);
      return null;
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
    const results = await runWithConcurrency(next.map((task) => () => uploadTask(task)), uploadConcurrency);
    if (results.some(Boolean)) await loadRecent();
  }

  async function handleUrlUpload() {
    const rawUrl = urlInput.trim();
    if (!rawUrl) return;

    let parsed: URL;
    try {
      parsed = new URL(rawUrl);
    } catch {
      toast.error(t(preferences.language, 'upload.invalidUrl'));
      return;
    }
    if (parsed.protocol !== 'http:' && parsed.protocol !== 'https:') {
      toast.error(t(preferences.language, 'upload.invalidUrl'));
      return;
    }

    urlUploading = true;
    try {
      const response = await fetch(parsed.toString());
      if (!response.ok) throw new Error('Download failed');
      const redirected = new URL(response.url || parsed.toString());
      if (redirected.protocol !== 'http:' && redirected.protocol !== 'https:') {
        toast.error(t(preferences.language, 'upload.invalidUrl'));
        return;
      }
      const maxBytes = preferences.runtimeSettings?.upload.max_upload_size_mb ? preferences.runtimeSettings.upload.max_upload_size_mb * 1024 * 1024 : 0;
      const contentLength = Number(response.headers.get('Content-Length') ?? 0);
      if (maxBytes > 0 && Number.isFinite(contentLength) && contentLength > maxBytes) {
        toast.error(t(preferences.language, 'upload.error'));
        return;
      }
      const blob = await response.blob();
      if (maxBytes > 0 && blob.size > maxBytes) {
        toast.error(t(preferences.language, 'upload.error'));
        return;
      }
      const mimeType = response.headers.get('Content-Type') || blob.type;
      const allowedTypes = preferences.runtimeSettings?.upload.effective_allowed_mime_types ?? [];
      if (!mimeType.startsWith('image/') || (allowedTypes.length > 0 && !isAllowedImageMimeType(mimeType, allowedTypes))) {
        toast.error(t(preferences.language, 'upload.urlNotImage'));
        return;
      }
      const filename = normalizeDownloadFilename(decodeURIComponent(redirected.pathname.split('/').pop() ?? ''), 'image');
      toast.success(t(preferences.language, 'upload.urlSuccess'));
      urlInput = '';
      await handleFiles([new File([blob], filename, { type: mimeType.split(';', 1)[0].trim().toLowerCase() })]);
    } catch {
      toast.error(t(preferences.language, 'upload.urlDownloadFail'));
    } finally {
      urlUploading = false;
    }
  }

  function copy(value: string) {
    void copyToClipboard(value, preferences.language);
  }

  async function removeRecent(record: UploadHistoryRecord) {
    deleting = true;
    try {
      await deleteImageByUid(record.uid, record.client_token);
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

  function handlePaste(event: ClipboardEvent) {
    const target = event.target as HTMLElement;
    if (target.tagName === 'INPUT' || target.tagName === 'TEXTAREA' || target.isContentEditable) return;
    const files = Array.from(event.clipboardData?.items ?? [])
      .filter((item) => normalizedImageMimeType(item.type).startsWith('image/') && !isBlockedImageMimeType(item.type))
      .map((item) => item.getAsFile())
      .filter((file): file is File => Boolean(file));
    if (files.length) {
      event.preventDefault();
      handleFiles(files);
    }
  }

  $effect(() => {
    const controller = new AbortController();
    loadRuntime(true, controller.signal);
    loadRecent();
    loadAnnouncements(controller.signal);
    return () => controller.abort();
  });
</script>

<svelte:window onpaste={handlePaste} />
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
      <CanvasDropzone language={preferences.language} disabled={uploadDisabled} {dragging} allowedTypes={allowedMimeTypesText} accept={uploadAccept} onFiles={handleFiles} onDragStateChange={(value) => (dragging = value)} />
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
    <ImageDataTable language={preferences.language} records={recentUploads} allowedImageOrigins={publicImageAllowedOrigins} canDelete={(record) => record.client_token === getClientToken()} onCopy={copy} onPreview={(record) => (previewRecord = record)} onDelete={(record) => (deleteTarget = record)} />
  </section>
{/if}

<ImagePreviewDialog language={preferences.language} record={previewRecord} records={recentUploads} allowedImageOrigins={publicImageAllowedOrigins} canDelete={previewRecord?.client_token === getClientToken()} onCopy={copy} onDelete={() => previewRecord && (deleteTarget = previewRecord)} onNavigate={(record) => (previewRecord = record)} onClose={() => (previewRecord = null)} />
<AnnouncementDialog language={preferences.language} announcements={announcements} open={announcementDialogOpen} initialMode={announcementDialogMode} onClose={closeAnnouncementDialog} onAcknowledge={acknowledgeAnnouncementDialog} />
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
