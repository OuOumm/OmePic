"use client";

/* eslint-disable @next/next/no-img-element */

import { useCallback, useEffect, useMemo, useState } from "react";
import { Database, KeyRound, RefreshCw } from "lucide-react";
import toast from "react-hot-toast";

import { CopyButton } from "@/components/shared/CopyButton";
import { PageIntro, PageSectionHeader } from "@/components/shared/PageLayout";
import { Badge } from "@/components/ui/Badge";
import { Button } from "@/components/ui/Button";
import { Card } from "@/components/ui/Card";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/Select";
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

const defaultStorageSelectValue = "__default__";

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
        title={t.upload.title}
      />

      <section className="grid gap-6 xl:grid-cols-[minmax(0,1.18fr)_minmax(360px,0.82fr)]">
        <div className="space-y-5">
          <Card className="p-5 sm:p-6 lg:p-7" variant="strong">
            <div className="space-y-6">
              <div className="space-y-2">
                <h2 className="text-2xl font-semibold tracking-tight text-foreground sm:text-[2rem]">
                  {t.upload.dropTitle}
                </h2>
                <p className="max-w-2xl text-sm leading-6 text-muted-foreground">
                  {t.upload.dropDescription}
                </p>
              </div>

              <div className="rounded-lg border border-border bg-muted/30 p-4 sm:p-5">
                <div className="min-w-0 space-y-3">
                  <div className="flex flex-wrap items-center gap-2">
                    <span className="flex h-9 w-9 items-center justify-center rounded-md border border-border bg-card text-muted-foreground">
                      <Database aria-hidden="true" className="h-4 w-4" />
                    </span>
                    <label className="text-sm font-medium text-foreground" htmlFor="upload-storage">
                      {t.upload.storageTarget}
                    </label>
                    {selectedStorage ? <Badge>{selectedStorage.storage_backend}</Badge> : null}
                  </div>
                  <p className="text-sm text-muted-foreground" id={storageHintId}>
                    {selectedStorage
                      ? t.upload.storageSelectionHint(selectedStorage.name, selectedStorage.storage_backend)
                      : t.upload.backendDefaultStorage}
                  </p>
                  <div className="flex flex-col gap-3 sm:flex-row">
                    <Select
                      disabled={storageOptionsLoading || phase === "uploading"}
                      onValueChange={(value) =>
                        setSelectedStorageKey(value === defaultStorageSelectValue ? "" : value)
                      }
                      value={selectedStorageKey || defaultStorageSelectValue}
                    >
                      <SelectTrigger
                        aria-busy={storageOptionsLoading}
                        aria-describedby={storageDescriptionIds}
                        aria-invalid={hasStorageOptionsError ? true : undefined}
                        id="upload-storage"
                      >
                        <SelectValue />
                      </SelectTrigger>
                      <SelectContent>
                        <SelectItem value={defaultStorageSelectValue}>
                          {defaultStorage
                            ? t.upload.defaultStorageOption(defaultStorage.name, defaultStorage.storage_backend)
                            : t.upload.backendDefaultStorage}
                        </SelectItem>
                        {storageOptions.map((option) => (
                          <SelectItem key={option.storage_key} value={option.storage_key}>
                            {option.is_default
                              ? t.upload.storageOptionDefault(option.name, option.storage_backend)
                              : t.upload.storageOption(option.name, option.storage_backend)}
                          </SelectItem>
                        ))}
                      </SelectContent>
                    </Select>
                    <Button
                      aria-controls="upload-storage"
                      className="sm:self-start"
                      disabled={storageOptionsLoading || phase === "uploading"}
                      onClick={refreshStorageOptions}
                      size="icon"
                      variant="secondary"
                    >
                      <RefreshCw aria-hidden="true" className="h-4 w-4" />
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
              className="mt-5 h-2 overflow-hidden rounded-full bg-muted"
              role="progressbar"
            >
              <div
                className="h-full rounded-full bg-primary transition-all duration-300"
                style={{ width: `${phase === "uploading" ? progress : result ? 100 : 0}%` }}
              />
            </div>
            <p className={phase === "error" ? "mt-4 text-sm text-danger" : "mt-4 text-sm text-muted-foreground"} role={phase === "error" ? "alert" : "status"}>
              {phase === "uploading" ? `${progress}%` : statusText}
            </p>
          </Card>

          {result ? (
            <Card className="overflow-hidden" variant="strong">
              {result.mime_type.startsWith("image/") ? (
                  <div className="relative max-h-72 overflow-hidden bg-muted">
                  <img
                    alt={t.upload.previewAlt}
                    className="h-full max-h-72 w-full object-cover"
                    src={result.url}
                  />
                  <div className="absolute inset-x-0 bottom-0 border-t border-border bg-background/95 px-4 py-3">
                    <div className="flex items-center justify-between gap-3">
                      <h3 className="text-base font-semibold text-foreground">{t.upload.latestResult}</h3>
                      {result.duplicate ? <Badge>{t.common.duplicate}</Badge> : null}
                    </div>
                  </div>
                </div>
              ) : null}
              <div className="space-y-4 p-5">
                <dl className="grid gap-3 rounded-lg border border-border bg-muted/30 p-4 text-sm text-muted-foreground">
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
                icon={<KeyRound aria-hidden="true" className="h-4 w-4" />}
                title={t.upload.latestResult}
              />
              <div className="mt-5 rounded-lg border border-border bg-muted/30 p-4">
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
              icon={<KeyRound aria-hidden="true" className="h-4 w-4" />}
              title={t.common.clientToken}
            />
            <p className="mt-4 break-all rounded-md border border-border bg-muted px-4 py-3 font-mono text-xs text-muted-foreground">
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
      <dt className="font-medium text-foreground">{label}</dt>
      <dd className={mono ? "break-all font-mono text-xs" : "break-all"}>{value}</dd>
    </div>
  );
}
