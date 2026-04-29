"use client";

/* eslint-disable @next/next/no-img-element */

import { useCallback, useEffect, useMemo, useState } from "react";
import toast from "react-hot-toast";

import { CopyButton } from "@/components/shared/CopyButton";
import { PageIntro, PageSectionHeader } from "@/components/shared/PageLayout";
import { Badge } from "@/components/ui/Badge";
import { Button } from "@/components/ui/Button";
import { Card } from "@/components/ui/Card";
import { Select } from "@/components/ui/Select";
import { useClientToken } from "@/hooks/useClientToken";
import { useUiLocale, useUiTranslations } from "@/hooks/useUiPreferences";
import { publicStorageOptions, uploadImageWithProgress } from "@/lib/api";
import { formatBytes, formatDate } from "@/lib/format";
import { listRecentUploadRecords, saveUploadRecord } from "@/lib/indexeddb/upload-history";
import { useUploadStore } from "@/stores/upload-store";
import type { PublicStorageOption } from "@/types/storage";
import type { UploadHistoryRecord } from "@/types/upload";

import { RecentUploads } from "./RecentUploads";
import { UploadDropzone } from "./UploadDropzone";

export function UploadPageClient() {
  const { token, ready } = useClientToken();
  const locale = useUiLocale();
  const t = useUiTranslations();
  const phase = useUploadStore((state) => state.phase);
  const progress = useUploadStore((state) => state.progress);
  const result = useUploadStore((state) => state.result);
  const error = useUploadStore((state) => state.error);
  const start = useUploadStore((state) => state.start);
  const setProgress = useUploadStore((state) => state.setProgress);
  const succeed = useUploadStore((state) => state.succeed);
  const fail = useUploadStore((state) => state.fail);
  const [isDragging, setIsDragging] = useState(false);
  const [recent, setRecent] = useState<UploadHistoryRecord[]>([]);
  const [storageOptions, setStorageOptions] = useState<PublicStorageOption[]>([]);
  const [selectedStorageKey, setSelectedStorageKey] = useState("");
  const [storageOptionsLoading, setStorageOptionsLoading] = useState(false);
  const [storageOptionsError, setStorageOptionsError] = useState<string | null>(null);
  const hasStorageOptionsError = storageOptionsError !== null;
  const storageErrorText = storageOptionsError || t.upload.storageOptionsFailed;
  const storageHintId = "upload-storage-hint";
  const storageErrorId = "upload-storage-error";
  const storageStatusId = "upload-storage-status";
  const storageDescriptionIds = [
    storageHintId,
    hasStorageOptionsError ? storageErrorId : "",
    storageOptionsLoading ? storageStatusId : ""
  ]
    .filter(Boolean)
    .join(" ");

  const readRecentUploads = useCallback(async () => {
    try {
      return await listRecentUploadRecords(10);
    } catch {
      return null;
    }
  }, []);

  const applyStorageOptions = useCallback((items: PublicStorageOption[]) => {
    setStorageOptions(items);
    setSelectedStorageKey((current) =>
      current && !items.some((item) => item.storage_key === current) ? "" : current
    );
  }, []);

  const loadStorageOptions = useCallback(async (signal?: AbortSignal) => {
    setStorageOptionsLoading(true);
    setStorageOptionsError(null);
    try {
      const result = await publicStorageOptions(signal);
      if (signal?.aborted) {
        return;
      }
      applyStorageOptions(result.items);
    } catch (storageError) {
      if (signal?.aborted) {
        return;
      }
      const message = storageError instanceof Error ? storageError.message : "";
      setStorageOptionsError(message);
    } finally {
      if (!signal?.aborted) {
        setStorageOptionsLoading(false);
      }
    }
  }, [applyStorageOptions]);

  useEffect(() => {
    if (!ready) {
      return;
    }
    let cancelled = false;
    void readRecentUploads().then((items) => {
      if (!cancelled) {
        setRecent(items ?? []);
      }
    });
    return () => {
      cancelled = true;
    };
  }, [readRecentUploads, ready]);

  useEffect(() => {
    const controller = new AbortController();
    void Promise.resolve().then(() => loadStorageOptions(controller.signal));
    return () => {
      controller.abort();
    };
  }, [loadStorageOptions]);

  const refreshStorageOptions = useCallback(() => {
    void loadStorageOptions();
  }, [loadStorageOptions]);

  const handleUpload = useCallback(async (file: File) => {
    if (!token) {
      return;
    }

    start();
    try {
      const uploaded = await uploadImageWithProgress(file, token, setProgress, selectedStorageKey);
      succeed(uploaded);
      const record: UploadHistoryRecord = {
        ...uploaded,
        token,
        original_filename: file.name
      };
      toast.success(uploaded.duplicate ? t.upload.duplicateUploadToast : t.upload.uploadCompleteToast);
      try {
        await saveUploadRecord(record);
        const items = await readRecentUploads();
        if (items) {
          setRecent(items);
        } else {
          toast.error(t.upload.localHistorySaveFailed);
        }
      } catch {
        toast.error(t.upload.localHistorySaveFailed);
      }
    } catch (uploadError) {
      const message = uploadError instanceof Error ? uploadError.message : t.upload.uploadFailed;
      fail(message);
      toast.error(message);
    }
  }, [
    fail,
    selectedStorageKey,
    setProgress,
    start,
    succeed,
    readRecentUploads,
    t.upload.duplicateUploadToast,
    t.upload.localHistorySaveFailed,
    t.upload.uploadCompleteToast,
    t.upload.uploadFailed,
    token
  ]);

  const defaultStorage = useMemo(
    () => storageOptions.find((item) => item.is_default) ?? null,
    [storageOptions]
  );
  const selectedStorage = useMemo(
    () =>
      selectedStorageKey
        ? storageOptions.find((item) => item.storage_key === selectedStorageKey) ?? null
        : defaultStorage,
    [defaultStorage, selectedStorageKey, storageOptions]
  );
  const statusText =
    phase === "uploading"
      ? t.upload.statusUploading(progress)
      : phase === "success"
        ? t.upload.statusSuccess
        : phase === "error"
          ? error ?? t.upload.uploadFailed
          : t.upload.statusIdle;

  return (
    <div className="space-y-6 animate-fade-in lg:space-y-8">
      <PageIntro
        description={t.upload.description}
        eyebrow={t.upload.eyebrow}
        title={<span className="gradient-text">{t.upload.title}</span>}
      />

      <section className="grid gap-6 xl:grid-cols-[minmax(0,1.18fr)_minmax(360px,0.82fr)]">
        <div className="space-y-5">
          <Card className="relative overflow-hidden p-5 sm:p-6 lg:p-7" variant="strong">
            <div className="absolute inset-0 bg-[radial-gradient(circle_at_top_left,rgba(139,92,246,0.16),transparent_30%),radial-gradient(circle_at_84%_22%,rgba(34,211,238,0.12),transparent_30%)]" />
            <div className="relative space-y-6">
              <div className="space-y-2">
                <h2 className="text-2xl font-bold tracking-tight text-slate-950 dark:text-white sm:text-[2rem]">
                  {t.upload.dropTitle}
                </h2>
                <p className="max-w-2xl text-sm leading-7 text-muted">
                  {t.upload.dropDescription}
                </p>
              </div>

              <div className="rounded-[26px] border border-white/45 bg-white/55 p-4 shadow-sm backdrop-blur-xl dark:border-white/10 dark:bg-slate-950/40 sm:p-5">
                <div className="min-w-0 space-y-3">
                  <div className="flex flex-wrap items-center gap-2">
                    <span className="flex h-10 w-10 items-center justify-center rounded-2xl bg-gradient-to-br from-violet-500/20 to-cyan-500/20 text-violet-600 dark:text-violet-200">
                      <StorageIcon />
                    </span>
                    <label className="text-sm font-semibold text-slate-800 dark:text-slate-100" htmlFor="upload-storage">
                      {t.upload.storageTarget}
                    </label>
                    {selectedStorage ? <Badge>{selectedStorage.storage_backend}</Badge> : null}
                  </div>
                  <p className="text-sm text-muted" id={storageHintId}>
                    {selectedStorage
                      ? t.upload.storageSelectionHint(selectedStorage.name, selectedStorage.storage_backend)
                      : t.upload.backendDefaultStorage}
                  </p>
                  <div className="flex flex-col gap-3 sm:flex-row">
                    <Select
                      aria-busy={storageOptionsLoading}
                      aria-describedby={storageDescriptionIds}
                      aria-invalid={hasStorageOptionsError ? true : undefined}
                      disabled={storageOptionsLoading || phase === "uploading"}
                      id="upload-storage"
                      onChange={(event) => setSelectedStorageKey(event.target.value)}
                      value={selectedStorageKey}
                    >
                      <option value="">
                        {defaultStorage
                          ? t.upload.defaultStorageOption(defaultStorage.name, defaultStorage.storage_backend)
                          : t.upload.backendDefaultStorage}
                      </option>
                      {storageOptions.map((option) => (
                        <option key={option.storage_key} value={option.storage_key}>
                          {option.is_default
                            ? t.upload.storageOptionDefault(option.name, option.storage_backend)
                            : t.upload.storageOption(option.name, option.storage_backend)}
                        </option>
                      ))}
                    </Select>
                    <Button
                      aria-controls="upload-storage"
                      className="sm:self-start"
                      disabled={storageOptionsLoading || phase === "uploading"}
                      onClick={refreshStorageOptions}
                      size="icon"
                      variant="secondary"
                    >
                      <RefreshIcon />
                      <span className="sr-only">{storageOptionsLoading ? t.common.loading : t.common.refresh}</span>
                    </Button>
                  </div>
                  {storageOptionsLoading ? (
                    <p className="sr-only" id={storageStatusId} role="status">
                      {t.common.loading}
                    </p>
                  ) : null}
                  {hasStorageOptionsError ? (
                    <p className="text-sm text-danger" id={storageErrorId} role="alert">
                      {storageErrorText}
                    </p>
                  ) : null}
                </div>
              </div>

              <UploadDropzone
                disabled={!ready || phase === "uploading"}
                isDragging={isDragging}
                onDragStateChange={setIsDragging}
                onSelectFile={handleUpload}
              />
            </div>
          </Card>

          <RecentUploads items={recent} title={t.upload.recentUploads} />
        </div>

        <aside className="space-y-5 xl:sticky xl:top-24 xl:self-start">
          <Card className="overflow-hidden p-5" variant="strong">
            <PageSectionHeader
              badge={phase === "success" && result?.duplicate ? <Badge>{t.common.duplicate}</Badge> : null}
              description={phase === "uploading" ? `${progress}%` : statusText}
              title={t.upload.status}
            />
            <div
              aria-label={t.upload.status}
              aria-valuemax={100}
              aria-valuemin={0}
              aria-valuenow={phase === "uploading" ? progress : result ? 100 : 0}
              className="mt-5 h-2.5 overflow-hidden rounded-full bg-slate-200/80 dark:bg-slate-800"
              role="progressbar"
            >
              <div
                className="h-full rounded-full bg-gradient-to-r from-violet-500 to-cyan-400 shadow-[0_0_24px_rgba(139,92,246,0.45)] transition-all duration-300"
                style={{ width: `${phase === "uploading" ? progress : result ? 100 : 0}%` }}
              />
            </div>
            <p className={phase === "error" ? "mt-4 text-sm text-danger" : "mt-4 text-sm text-muted"} role={phase === "error" ? "alert" : "status"}>
              {phase === "uploading" ? `${progress}%` : statusText}
            </p>
          </Card>

          {result ? (
            <Card className="overflow-hidden" variant="strong">
              {result.mime_type.startsWith("image/") ? (
                <div className="relative max-h-72 overflow-hidden bg-slate-200/70 dark:bg-slate-900">
                  <img
                    alt={t.upload.previewAlt}
                    className="h-full max-h-72 w-full object-cover"
                    src={result.url}
                  />
                  <div className="absolute inset-x-0 bottom-0 border-t border-white/20 bg-white/70 px-4 py-3 backdrop-blur-md dark:bg-slate-950/70">
                    <div className="flex items-center justify-between gap-3">
                      <h3 className="text-base font-bold text-slate-900 dark:text-white">{t.upload.latestResult}</h3>
                      {result.duplicate ? <Badge>{t.common.duplicate}</Badge> : null}
                    </div>
                  </div>
                </div>
              ) : null}
              <div className="space-y-4 p-5">
                <dl className="grid gap-3 rounded-[24px] border border-white/40 bg-white/50 p-4 text-sm text-muted backdrop-blur-md dark:border-white/10 dark:bg-slate-950/40">
                  <InfoRow label={t.common.uid} value={result.uid} mono />
                  <InfoRow label={t.common.type} value={result.mime_type} />
                  <InfoRow label={t.common.storage} value={`${result.storage_key} (${result.storage_backend})`} mono />
                  <InfoRow label={t.common.size} value={formatBytes(result.size)} />
                  <InfoRow label={t.common.created} value={formatDate(result.created_at, locale)} />
                </dl>
                <div className="flex flex-wrap gap-2">
                  <CopyButton label={t.common.copyTargets.url} value={result.url} />
                  <CopyButton label={t.common.copyTargets.markdown} value={result.md_url} />
                  <CopyButton label={t.common.copyTargets.bbcode} value={result.bbcode} />
                </div>
              </div>
            </Card>
          ) : (
            <Card className="overflow-hidden p-5" variant="subtle">
              <PageSectionHeader
                description={t.upload.statusIdle}
                icon={<KeyIcon />}
                title={t.upload.latestResult}
              />
              <div className="mt-5 rounded-[26px] border border-white/45 bg-white/55 p-4 backdrop-blur-xl dark:border-white/10 dark:bg-slate-950/40">
                <div className="skeleton-glass h-44" />
                <div className="mt-4 space-y-3">
                  <div className="skeleton-glass h-4 w-2/3" />
                  <div className="skeleton-glass h-4 w-1/2" />
                </div>
              </div>
            </Card>
          )}

          <Card className="p-5" variant="subtle">
            <PageSectionHeader
              description={t.upload.backendDefaultStorage}
              icon={<KeyIcon />}
              title={t.common.clientToken}
            />
            <p className="mt-4 break-all rounded-2xl border border-white/40 bg-white/60 px-4 py-3 font-mono text-xs text-slate-700 backdrop-blur-xl dark:border-white/10 dark:bg-slate-950/50 dark:text-slate-300">
              {ready ? token : t.common.preparingToken}
            </p>
          </Card>
        </aside>
      </section>
    </div>
  );
}

function InfoRow({ label, mono, value }: { label: string; mono?: boolean; value: string }) {
  return (
    <div className="grid gap-1 sm:grid-cols-[96px_1fr]">
      <dt className="font-semibold text-slate-700 dark:text-slate-300">{label}</dt>
      <dd className={mono ? "break-all font-mono text-xs" : "break-all"}>{value}</dd>
    </div>
  );
}

function StorageIcon() {
  return (
    <svg aria-hidden="true" className="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
      <path d="M4 7c0 1.7 3.6 3 8 3s8-1.3 8-3-3.6-3-8-3-8 1.3-8 3Zm0 0v10c0 1.7 3.6 3 8 3s8-1.3 8-3V7M4 12c0 1.7 3.6 3 8 3s8-1.3 8-3" strokeLinecap="round" strokeLinejoin="round" />
    </svg>
  );
}

function RefreshIcon() {
  return (
    <svg aria-hidden="true" className="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
      <path d="M20 11a8.1 8.1 0 0 0-15.5-2M4 5v4h4m-4 4a8.1 8.1 0 0 0 15.5 2M20 19v-4h-4" strokeLinecap="round" strokeLinejoin="round" />
    </svg>
  );
}

function KeyIcon() {
  return (
    <svg aria-hidden="true" className="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
      <path d="M15.75 7.5a5.25 5.25 0 1 0-4 5.1L5 19.35V22h2.65l1.2-1.2V19h1.8l1.4-1.4v-1.8l2.85-2.85a5.22 5.22 0 0 0 .85-5.45ZM16.5 6.75h.01" strokeLinecap="round" strokeLinejoin="round" />
    </svg>
  );
}
