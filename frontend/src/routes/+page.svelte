<script lang="ts">
  import { Link, Loader2 } from 'lucide-svelte';
  import CanvasDropzone from '@/components/studio/CanvasDropzone.svelte';
  import ImageDataRow from '@/components/studio/ImageDataRow.svelte';
  import StorageInspector from '@/components/studio/StorageInspector.svelte';
  import { ApiError, getRuntimeSettings, uploadImageWithProgress } from '@/api';
  import { getClientToken } from '@/preferences';
  import { saveUploadToHistory, getRecentUploads } from '@/indexeddb/upload-history';
  import { t } from '@/i18n';
  import { preferences, setRuntimeSettings, setSelectedStorageKey } from '@/stores/preferences.svelte';
  import { toast } from '@/stores/toast.svelte';
  import type { Language, UploadHistoryRecord, UploadResult } from '@/types';

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
  let counter = 0;

  const activeTasks = $derived(tasks.filter((task) => task.status === 'pending' || task.status === 'uploading' || task.status === 'error'));
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

  $effect(() => {
    loadRuntime();
    loadRecent();
    const pasteHandler = (event: ClipboardEvent) => {
      const target = event.target as HTMLElement;
      if (target.tagName === 'INPUT' || target.tagName === 'TEXTAREA' || target.isContentEditable) return;
      const files = Array.from(event.clipboardData?.items ?? [])
        .filter((item) => item.type.startsWith('image/'))
        .map((item) => item.getAsFile())
        .filter((file): file is File => !!file);
      if (files.length) {
        event.preventDefault();
        handleFiles(files);
      }
    };
    document.addEventListener('paste', pasteHandler);
    return () => document.removeEventListener('paste', pasteHandler);
  });
</script>

<svelte:head><title>OmePic</title></svelte:head>

<div class="grid gap-6 lg:grid-cols-[1fr_320px]">
  <section class="space-y-6">
    {#if runtimeError}
      <div class="studio-panel border-[hsl(var(--danger))] p-4 text-sm font-bold text-[hsl(var(--danger))]">{runtimeError}</div>
    {/if}
    {#if maintenanceMode}
      <div class="studio-panel bg-[hsl(var(--marker-yellow)/0.35)] p-4 text-sm font-bold">{preferences.runtimeSettings?.features.maintenance_message}</div>
    {/if}

    <div role="presentation" ondragenter={() => (dragging = true)} ondragleave={() => (dragging = false)}>
      <CanvasDropzone language={preferences.language} disabled={uploadDisabled} {dragging} onFiles={handleFiles} />
    </div>

    <div class="grid gap-3 border-y-[3px] ink-line py-5 md:grid-cols-[1fr_auto] md:items-end">
      <label class="grid gap-2 text-sm font-black">
        {t(preferences.language, 'upload.urlLabel')}
        <input class="studio-input" bind:value={urlInput} placeholder={t(preferences.language, 'upload.urlPlaceholder')} onkeydown={(event) => event.key === 'Enter' && handleUrlUpload()} />
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
      <h2 class="border-b-2 ink-line pb-2 text-xl font-black">{t(preferences.language, 'upload.recentTitle')}</h2>
      {#if activeTasks.length}
        <div class="mt-3 grid gap-3">
          {#each activeTasks as task (task.id)}
            <div class="border-b-2 border-dashed border-[hsl(var(--ink)/0.32)] pb-3 text-sm">
              <div class="flex justify-between gap-3 font-black"><span class="truncate">{task.file.name}</span><span>{task.progress}%</span></div>
              <div class="mt-2 h-2 border-2 ink-line"><div class="h-full bg-[hsl(var(--marker-green))]" style={`width:${task.progress}%`}></div></div>
              {#if task.error}<p class="mt-1 text-xs text-[hsl(var(--danger))]">{task.error}</p>{/if}
            </div>
          {/each}
        </div>
      {/if}
      {#if !activeTasks.length && !recentUploads.length}
        <p class="mt-3 text-sm text-[hsl(var(--ink-muted))]">{t(preferences.language, 'upload.noRecent')}</p>
      {/if}
      {#each recentUploads.slice(0, 4) as record (record.uid)}
        <div class="mt-3 border-b-2 border-dashed border-[hsl(var(--ink)/0.22)] pb-3 text-sm">
          <button class="font-black hover:marker-highlight" type="button" onclick={() => copy(record.url)}>{record.original_filename || record.uid}</button>
          <p class="truncate text-xs text-[hsl(var(--ink-muted))]">{record.url}</p>
        </div>
      {/each}
    </div>
  </aside>
</div>

{#if recentUploads.length}
  <section class="mt-10">
    <div class="mb-4 flex items-end justify-between border-b-[3px] ink-line pb-3">
      <div>
        <p class="text-xs font-black uppercase text-[hsl(var(--ink-muted))]">File desk</p>
        <h2 class="text-3xl font-black">{t(preferences.language, 'upload.recentTitle')}</h2>
      </div>
    </div>
    <div class="overflow-x-auto">
      <div class="min-w-[720px]">
        {#each recentUploads as record (record.uid)}
          <ImageDataRow language={preferences.language} {record} onCopy={copy} />
        {/each}
      </div>
    </div>
  </section>
{/if}

