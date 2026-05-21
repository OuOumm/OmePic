import { ApiError, uploadImageWithProgress } from '@/api';
import { getClientToken } from '@/client-token';
import { t } from '@/i18n';
import { saveUploadToHistory } from '@/indexeddb/upload-history';
import { preferences } from '@/stores/preferences.svelte';
import { toast } from '@/stores/toast.svelte';
import type { Language, UploadResult } from '@/types';
import { bbcodeForImageUrl, isAllowedImageMimeType, markdownForImageUrl, uidFromImageUrl } from '@/utils';
import { createProgressReporter, runWithConcurrency } from '@/upload-queue';

export type UploadTask = {
  id: string;
  file: File;
  progress: number;
  status: 'pending' | 'uploading' | 'success' | 'error';
  result?: UploadResult;
  error?: string;
};

let tasks = $state<UploadTask[]>([]);
let counter = 0;

/** Reactive getter — use in a $derived context in the consuming component. */
export function getActiveTasks(): UploadTask[] {
  return tasks.filter((task) => task.status === 'pending' || task.status === 'uploading' || task.status === 'error');
}

export function uploadErrorMessageWithT(lang: Language, err: unknown): string {
  if (err instanceof ApiError && err.code === 'rate_limited') {
    return typeof err.retryAfter === 'number'
      ? t(lang, 'upload.rateLimitedWithRetry', { seconds: err.retryAfter })
      : t(lang, 'upload.rateLimited');
  }
  if (err instanceof ApiError && err.code === 'network_error') return t(lang, 'upload.networkError');
  return err instanceof Error ? err.message : t(lang, 'upload.error');
}

function validateFiles(files: File[]): File[] {
  const settings = preferences.runtimeSettings;
  if (!settings) return files;
  const maxBytes = settings.upload.max_upload_size_mb > 0 ? settings.upload.max_upload_size_mb * 1024 * 1024 : 0;
  const allowedTypes = settings.upload.allowed_mime_types;
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

async function uploadOneTask(task: UploadTask): Promise<UploadResult | null> {
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
    const uid = uidFromImageUrl(result.url);
    if (!uid) throw new Error(t(preferences.language, 'upload.error'));
    const createdAt = new Date().toISOString();
    const selectedKey = preferences.selectedStorageKey.trim();
    const selectedOption = preferences.runtimeSettings?.storage.options.find((option) => option.storage_key === selectedKey)
      ?? preferences.runtimeSettings?.storage.options.find((option) => option.is_default)
      ?? preferences.runtimeSettings?.storage.options[0];
    await saveUploadToHistory({
      uid,
      url: result.url,
      mime_type: 'image/avif',
      size: task.file.size,
      created_at: createdAt,
      is_duplicate: result.duplicate,
      storage_key: selectedOption?.storage_key ?? selectedKey,
      storage_backend: selectedOption?.storage_backend ?? 'local',
      markdown: markdownForImageUrl(result.url, task.file.name),
      bbcode: bbcodeForImageUrl(result.url),
      client_token: token,
      original_filename: task.file.name,
      saved_at: createdAt,
    });
    toast.success(result.duplicate ? t(preferences.language, 'upload.duplicate') : t(preferences.language, 'upload.success'));
    return result;
  } catch (err) {
    const message = uploadErrorMessageWithT(preferences.language, err);
    updateTask(task.id, { status: 'error', error: message });
    toast.error(`${task.file.name}: ${message}`);
    return null;
  }
}

export async function enqueueFiles(files: File[], maintenanceMode: boolean): Promise<boolean> {
  if (maintenanceMode) {
    toast.error(preferences.runtimeSettings?.features.maintenance_message ?? t(preferences.language, 'common.error'));
    return false;
  }
  const accepted = validateFiles(files);
  if (accepted.length === 0) return false;
  const next = accepted.map((file) => ({ id: `task-${++counter}`, file, progress: 0, status: 'pending' as const }));
  tasks = [...next, ...tasks];
  const results = await runWithConcurrency(next.map((task) => () => uploadOneTask(task)), 3);
  return results.some(Boolean);
}
