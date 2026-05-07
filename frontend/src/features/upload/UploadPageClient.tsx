"use client";

import { useState, useCallback, useRef, useEffect, useMemo } from "react";
import { UploadDropzone } from "./UploadDropzone";
import { StorageSelector } from "./StorageSelector";
import { ImageLightbox } from "@/components/shared/ImageLightbox";
import { ImgStyleImageCard } from "@/components/shared/ImgStyleImageCard";
import { Card, CardContent } from "@/components/ui/Card";
import { Button } from "@/components/ui/Button";
import { Input } from "@/components/ui/Input";
import { Label } from "@/components/ui/Label";
import { useUploadStore } from "@/stores/upload-store";
import { useUiPreferencesStore } from "@/stores/ui-preferences-store";
import { uploadImageWithProgress } from "@/lib/api";
import { getClientToken } from "@/lib/preferences";
import {
  saveUploadToHistory,
  getRecentUploads,
} from "@/lib/indexeddb/upload-history";
import { t } from "@/lib/i18n";
import { formatBytes, formatDate } from "@/lib/utils";
import { Loader2, Link } from "lucide-react";
import toast from "react-hot-toast";
import type { UploadResult, UploadHistoryRecord } from "@/types";

interface UploadTask {
  id: string;
  file: File;
  progress: number;
  status: "pending" | "uploading" | "success" | "error";
  result?: UploadResult;
  error?: string;
}

let taskIdCounter = 0;

export function UploadPageClient() {
  const language = useUiPreferencesStore((state) => state.language);
  const selectedStorageKey = useUploadStore((state) => state.selectedStorageKey);
  const hasHydrated = useUiPreferencesStore((state) => state.hasHydrated);

  const [tasks, setTasks] = useState<UploadTask[]>([]);
  const [isDragOver, setIsDragOver] = useState(false);
  const [recentUploads, setRecentUploads] = useState<UploadHistoryRecord[]>([]);
  const [previewRecord, setPreviewRecord] = useState<UploadHistoryRecord | null>(null);

  // URL upload state
  const [urlInput, setUrlInput] = useState("");
  const [urlUploading, setUrlUploading] = useState(false);

  const fileInputRef = useRef<HTMLInputElement>(null);
  const processingRef = useRef(false);
  const languageRef = useRef(language);
  const storageKeyRef = useRef(selectedStorageKey);
  languageRef.current = language;
  storageKeyRef.current = selectedStorageKey;

  const loadRecentUploads = useCallback(async () => {
    try {
      const records = await getRecentUploads(10);
      setRecentUploads(records);
    } catch { /* ignore */ }
  }, []);

  useEffect(() => {
    loadRecentUploads();
  }, [loadRecentUploads]);

  const processQueue = useCallback(() => {
    if (processingRef.current) return;

    setTasks((prev) => {
      const pending = prev.filter((t) => t.status === "pending");
      if (pending.length === 0) return prev;

      processingRef.current = true;

      // Mark pending tasks as uploading
      const updated = prev.map((t) =>
        t.status === "pending" ? { ...t, status: "uploading" as const, progress: 0 } : t
      );

      const lang = languageRef.current;
      const storageKey = storageKeyRef.current || undefined;

      // Start all uploads concurrently in the background
      (async () => {
        const token = getClientToken();

        const settled = await Promise.allSettled(
          pending.map((task) =>
            uploadImageWithProgress(
              task.file,
              token,
              (pct) =>
                setTasks((prev) =>
                  prev.map((t) => (t.id === task.id ? { ...t, progress: pct } : t))
                ),
              storageKey
            )
              .then((result) => {
                setTasks((prev) =>
                  prev.map((t) =>
                    t.id === task.id
                      ? { ...t, status: "success" as const, result, progress: 100 }
                      : t
                  )
                );
                return { task, result: result as UploadResult };
              })
              .catch((err) => {
                const msg = err instanceof Error ? err.message : t(lang, "upload.error");
                setTasks((prev) =>
                  prev.map((t) =>
                    t.id === task.id
                      ? { ...t, status: "error" as const, error: msg }
                      : t
                  )
                );
                return { task, result: null };
              })
          )
        );

        // Collect successful results
        const savedResults: { task: UploadTask; result: UploadResult }[] = [];
        for (const s of settled) {
          if (s.status === "fulfilled" && s.value.result) {
            savedResults.push(s.value);
          }
        }

        // Save all successful results to IndexedDB
        const totalCount = pending.length;
        for (const { task: tsk, result: r } of savedResults) {
          const record: UploadHistoryRecord = {
            uid: r.uid,
            url: r.url,
            mime_type: r.mime_type,
            size: r.size,
            created_at: r.created_at,
            is_duplicate: r.is_duplicate,
            storage_key: r.storage_key,
            storage_backend: r.storage_backend,
            markdown: r.markdown,
            bbcode: r.bbcode,
            client_token: token,
            original_filename: tsk.file.name,
            saved_at: new Date().toISOString(),
          };
          saveUploadToHistory(record).catch(() => {});
        }

        // Show toast summary
        const successCount = savedResults.length;
        if (successCount === totalCount && totalCount === 1) {
          const r = savedResults[0].result;
          toast.success(
            r.is_duplicate ? t(lang, "upload.duplicate") : t(lang, "upload.success")
          );
        } else if (successCount === totalCount) {
          toast.success(t(lang, "upload.multiSuccess", { count: successCount }));
        } else if (successCount > 0) {
          toast(t(lang, "upload.multiPartial", { success: successCount, total: totalCount }));
        }
        // If all failed, individual error toasts are handled in the catch above
        // via setTasks, and the user sees the error card

        // Refresh recent uploads
        try {
          const recent = await getRecentUploads(10);
          setRecentUploads(recent);
        } catch { /* ignore */ }

        processingRef.current = false;
        // Check for newly added pending tasks
        processQueue();
      })();

      return updated;
    });
  }, []);

  const handleUploads = useCallback(
    (files: File[]) => {
      const newTasks: UploadTask[] = files.map((file) => ({
        id: `task-${++taskIdCounter}`,
        file,
        progress: 0,
        status: "pending" as const,
      }));

      setTasks((prev) => [...prev, ...newTasks]);
      // Schedule processing after state is committed
      setTimeout(() => processQueue(), 0);
    },
    [processQueue]
  );

  const handleUrlUpload = useCallback(async () => {
    let url = urlInput.trim();
    if (!url) return;
    if (!/^https?:\/\//i.test(url)) {
      toast.error(t(language, "upload.invalidUrl"));
      return;
    }

    setUrlUploading(true);
    try {
      const resp = await fetch(url);
      if (!resp.ok) throw new Error("Download failed");
      const blob = await resp.blob();
      let mimeType = resp.headers.get("Content-Type") || "";
      if (
        !mimeType ||
        mimeType === "application/octet-stream" ||
        mimeType === "binary/octet-stream"
      ) {
        // Infer from extension
        const extMap: Record<string, string> = {
          avif: "image/avif",
          png: "image/png",
          jpg: "image/jpeg",
          jpeg: "image/jpeg",
          gif: "image/gif",
          webp: "image/webp",
          bmp: "image/bmp",
        };
        const ext = url.split(".").pop()?.toLowerCase().split("?")[0] ?? "";
        mimeType = extMap[ext] || "";
      }
      if (!mimeType.startsWith("image/")) {
        toast.error(t(language, "upload.urlNotImage"));
        return;
      }
      const filename = url.split("/").pop()?.split("?")[0] || "image";
      const file = new File([blob], filename, { type: mimeType });
      toast.success(t(language, "upload.urlSuccess"));
      setUrlInput("");
      handleUploads([file]);
    } catch {
      toast.error(t(language, "upload.urlDownloadFail"));
    } finally {
      setUrlUploading(false);
    }
  }, [urlInput, language, handleUploads]);

  // Global paste handler
  useEffect(() => {
    const handler = (e: ClipboardEvent) => {
      const target = e.target as HTMLElement;
      if (
        target.tagName === "INPUT" ||
        target.tagName === "TEXTAREA" ||
        target.isContentEditable
      )
        return;
      // Skip if paste happened inside the dropzone (which handles it via React onPaste)
      if (target.closest("[data-paste-managed]")) return;

      const items = e.clipboardData?.items;
      if (!items) return;
      for (let i = 0; i < items.length; i++) {
        if (items[i].type.startsWith("image/")) {
          e.preventDefault();
          const file = items[i].getAsFile();
          if (file) handleUploads([file]);
          return;
        }
      }
      // No image found in clipboard — notify the user
      toast.error(t(language, "upload.noClipboard"));
    };
    document.addEventListener("paste", handler);
    return () => document.removeEventListener("paste", handler);
  }, [handleUploads, language]);

  // Determine if any upload is in progress
  const isUploading = tasks.some(
    (t) => t.status === "pending" || t.status === "uploading"
  );

  // Unified grid: uploading/pending tasks first, error tasks next, then recent history.
  // Within each group, newest first (task id descending, saved_at descending).
  const mergedItems = useMemo(() => {
    const extractTaskNum = (task: UploadTask) => {
      const parts = task.id.split("-");
      return parseInt(parts[parts.length - 1], 10) || 0;
    };

    // Uploading/pending tasks — newest first
    const activeTasks = tasks
      .filter((t) => t.status === "pending" || t.status === "uploading")
      .sort((a, b) => extractTaskNum(b) - extractTaskNum(a));

    // Error tasks — newest first
    const errorTasks = tasks
      .filter((t) => t.status === "error")
      .sort((a, b) => extractTaskNum(b) - extractTaskNum(a));

    // History records — newest first by saved_at
    const sortedHistory = [...recentUploads].sort(
      (a, b) => new Date(b.saved_at).getTime() - new Date(a.saved_at).getTime()
    );

    const items: Array<
      | { type: "task"; task: UploadTask }
      | { type: "history"; record: UploadHistoryRecord }
    > = [];

    for (const task of activeTasks) {
      items.push({ type: "task", task });
    }
    for (const task of errorTasks) {
      items.push({ type: "task", task });
    }
    for (const rec of sortedHistory) {
      items.push({ type: "history", record: rec });
    }

    return items.slice(0, 12);
  }, [tasks, recentUploads]);

  if (!hasHydrated) {
    return (
      <div className="flex items-center justify-center py-20">
        <Loader2 className="h-6 w-6 animate-spin text-muted-foreground" />
      </div>
    );
  }

  const lang = language;

  return (
    <div className="space-y-8" id="main-content">
      {/* Upload area */}
      <Card>
        <CardContent className="pt-6">
          <StorageSelector />
          <div className="mt-4">
            <UploadDropzone
            disabled={isUploading}
            isDragging={isDragOver}
            onDragStateChange={setIsDragOver}
            onSelectFiles={handleUploads}
            fileInputRef={fileInputRef}
            language={lang}
          />
          </div>
        </CardContent>
      </Card>

      {/* URL upload */}
      <Card>
        <CardContent className="pt-6">
          <div className="flex flex-col sm:flex-row gap-2 items-start sm:items-end">
            <div className="flex-1 w-full">
              <Label htmlFor="url-upload" className="text-xs">
                {t(lang, "upload.urlLabel")}
              </Label>
              <Input
                id="url-upload"
                value={urlInput}
                onChange={(e) => setUrlInput(e.target.value)}
                placeholder={t(lang, "upload.urlPlaceholder")}
                className="h-8 text-sm"
                onKeyDown={(e) => e.key === "Enter" && handleUrlUpload()}
              />
            </div>
            <Button
              variant="outline"
              size="sm"
              onClick={handleUrlUpload}
              disabled={urlUploading || isUploading || !urlInput.trim()}
              className="cursor-pointer h-8"
            >
              {urlUploading ? (
                <Loader2 className="h-3.5 w-3.5 animate-spin" />
              ) : (
                <Link className="h-3.5 w-3.5" />
              )}
              {t(lang, "upload.urlUpload")}
            </Button>
          </div>
        </CardContent>
      </Card>

      {/* Unified image card grid */}
      <div>
        <h2 className="text-lg font-semibold mb-4">
          {t(lang, "upload.recentTitle")}
        </h2>
        {mergedItems.length === 0 ? (
          <p className="text-sm text-muted-foreground">
            {t(lang, "upload.noRecent")}
          </p>
        ) : (
          <div className="grid grid-cols-2 sm:grid-cols-3 md:grid-cols-5 gap-3">
            {mergedItems.map((item) => {
              if (item.type === "task") {
                const task = item.task;
                return (
                  <ImgStyleImageCard
                    key={task.id}
                    alt={task.file.name}
                    uploadStatus={task.status}
                    uploadProgress={task.progress}
                    filename={task.file.name}
                  />
                );
              }

              // History record — success card with hover action buttons
              const rec = item.record;
              return (
                <ImgStyleImageCard
                  key={rec.uid}
                  src={rec.url}
                  alt={rec.original_filename || rec.uid}
                  title={rec.uid}
                  filename={rec.original_filename}
                  sizeLabel={formatBytes(rec.size)}
                  onPreview={() => setPreviewRecord(rec)}
                  previewLabel={t(lang, "common.openPreview", {
                    title: rec.uid,
                  })}
                  actionButtons={[
                    {
                      label: "URL",
                      onClick: () => {
                        navigator.clipboard.writeText(rec.url);
                        toast.success(t(lang, "common.copied"));
                      },
                    },
                    {
                      label: "MD",
                      onClick: () => {
                        navigator.clipboard.writeText(
                          rec.markdown || `![](${rec.url})`
                        );
                        toast.success(t(lang, "common.copied"));
                      },
                    },
                    {
                      label: "BB",
                      onClick: () => {
                        navigator.clipboard.writeText(
                          rec.bbcode || `[img]${rec.url}[/img]`
                        );
                        toast.success(t(lang, "common.copied"));
                      },
                    },
                  ]}
                />
              );
            })}
          </div>
        )}
      </div>

      {/* Preview lightbox */}
      <ImageLightbox
        open={!!previewRecord}
        onClose={() => setPreviewRecord(null)}
        initialIndex={previewRecord ? recentUploads.findIndex((r) => r.uid === previewRecord.uid) : 0}
        images={recentUploads.map((r) => ({
          url: r.url,
          alt: r.original_filename,
          metadata: [
            { label: t(lang, "image.uid"), value: r.uid },
            { label: t(lang, "image.storageKey"), value: r.storage_key },
            { label: t(lang, "image.storageBackend"), value: r.storage_backend },
            { label: t(lang, "image.type"), value: r.mime_type },
            { label: t(lang, "image.size"), value: formatBytes(r.size) },
            { label: t(lang, "image.created"), value: formatDate(r.created_at) },
          ],
        }))}
        getActions={(_, idx) => {
          const r = recentUploads[idx];
          if (!r) return [];
          return [
            {
              label: t(lang, "common.copyUrl"),
              onClick: () => {
                navigator.clipboard.writeText(r.url);
                toast.success(t(lang, "common.copied"));
              },
            },
            {
              label: t(lang, "common.copyMarkdown"),
              onClick: () => {
                navigator.clipboard.writeText(r.markdown || `![](${r.url})`);
                toast.success(t(lang, "common.copied"));
              },
            },
            {
              label: t(lang, "common.copyBBCode"),
              onClick: () => {
                navigator.clipboard.writeText(r.bbcode || `[img]${r.url}[/img]`);
                toast.success(t(lang, "common.copied"));
              },
            },
          ];
        }}
        closeLabel={t(lang, "common.close")}
        metadataLabel={t(lang, "history.viewPreview")}
      />
    </div>
  );
}
