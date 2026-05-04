"use client";

import { useCallback, useEffect, useMemo, useState, type ClipboardEvent } from "react";
import Link from "next/link";
import { Clipboard, Link2, Sparkles } from "lucide-react";
import toast from "react-hot-toast";

import { Badge } from "@/components/ui/Badge";
import { Button } from "@/components/ui/Button";
import { Card } from "@/components/ui/Card";
import { Input } from "@/components/ui/Input";
import { useClientToken } from "@/hooks/useClientToken";
import { useUiTranslations } from "@/hooks/useUiPreferences";
import { uploadImageWithProgress } from "@/lib/api";
import { listRecentUploadRecords, saveUploadRecord } from "@/lib/indexeddb/upload-history";
import { useUploadStore } from "@/stores/upload-store";
import type { UploadHistoryRecord } from "@/types/upload";

import { RecentUploads } from "./RecentUploads";
import { UploadDropzone } from "./UploadDropzone";

const pasteInputSelector = "input, textarea, [contenteditable='true'], [role='textbox']";
const supportedSourceImageTypes = new Set([
  "image/avif",
  "image/png",
  "image/jpeg",
  "image/gif",
  "image/webp",
  "image/bmp"
]);
const binaryResponseTypes = new Set([
  "",
  "application/octet-stream",
  "binary/octet-stream"
]);

export function UploadPageClient() {
  const { token, ready } = useClientToken();
  const t = useUiTranslations();
  const phase = useUploadStore((state) => state.phase);
  const progress = useUploadStore((state) => state.progress);
  const result = useUploadStore((state) => state.result);
  const error = useUploadStore((state) => state.error);
  const selectedStorageKey = useUploadStore((state) => state.selectedStorageKey);
  const start = useUploadStore((state) => state.start);
  const setProgress = useUploadStore((state) => state.setProgress);
  const succeed = useUploadStore((state) => state.succeed);
  const fail = useUploadStore((state) => state.fail);
  const [isDragging, setIsDragging] = useState(false);
  const [recent, setRecent] = useState<UploadHistoryRecord[]>([]);
  const [imageUrl, setImageUrl] = useState("");
  const uploadDisabled = !ready || phase === "uploading";
  const urlInputId = "upload-image-url";
  const uploadSourceHelpId = "upload-source-help";

  const readRecentUploads = useCallback(async () => {
    try {
      return await listRecentUploadRecords(10);
    } catch {
      return null;
    }
  }, []);

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

  const handleUpload = useCallback(async (file: File) => {
    if (!token) {
      return false;
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
      setRecent((current) => [
        record,
        ...current.filter((item) => item.uid !== record.uid)
      ].slice(0, 10));
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
      return true;
    } catch (uploadError) {
      const message = uploadError instanceof Error ? uploadError.message : t.upload.uploadFailed;
      fail(message);
      toast.error(message);
      return false;
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

  const reportSourceError = useCallback((message: string) => {
    fail(message);
    toast.error(message);
  }, [fail]);

  const handlePaste = useCallback((event: ClipboardEvent<HTMLElement>) => {
    if (uploadDisabled) {
      return;
    }

    if (event.target instanceof Element && event.target.closest(pasteInputSelector)) {
      return;
    }

    event.stopPropagation();
    const file = fileFromClipboard(event.clipboardData);
    if (!file) {
      reportSourceError(t.upload.clipboardNoImage);
      return;
    }

    event.preventDefault();
    void handleUpload(file);
  }, [handleUpload, reportSourceError, t.upload.clipboardNoImage, uploadDisabled]);

  const handleUrlUpload = useCallback(async () => {
    if (uploadDisabled) {
      return;
    }

    let sourceUrl: URL;
    try {
      sourceUrl = parseImageUrl(imageUrl);
    } catch {
      reportSourceError(t.upload.urlInvalid);
      return;
    }

    try {
      const file = await downloadImageUrl(sourceUrl, t.upload.urlFilenameFallback);
      if (await handleUpload(file)) {
        setImageUrl("");
      }
    } catch (downloadError) {
      if (downloadError instanceof UrlImageDownloadError) {
        reportSourceError(t.upload[downloadError.reason]);
        return;
      }
      reportSourceError(t.upload.urlDownloadFailed);
    }
  }, [
    handleUpload,
    imageUrl,
    reportSourceError,
    t.upload,
    uploadDisabled
  ]);

  useEffect(() => {
    function handleWindowPaste(event: globalThis.ClipboardEvent) {
      if (uploadDisabled) {
        return;
      }
      if (event.target instanceof Element && event.target.closest(pasteInputSelector)) {
        return;
      }

      const file = event.clipboardData ? fileFromClipboard(event.clipboardData) : null;
      if (!file) {
        reportSourceError(t.upload.clipboardNoImage);
        return;
      }

      event.preventDefault();
      void handleUpload(file);
    }

    window.addEventListener("paste", handleWindowPaste);
    return () => {
      window.removeEventListener("paste", handleWindowPaste);
    };
  }, [
    handleUpload,
    reportSourceError,
    t.upload.clipboardNoImage,
    uploadDisabled
  ]);

  const activeUpload = useMemo(
    () =>
      phase === "uploading"
        ? { progress, status: t.upload.statusUploading(progress) }
        : null,
    [phase, progress, t.upload]
  );

  return (
    <div className="mx-auto flex w-full max-w-6xl animate-fade-in flex-col gap-6 lg:gap-8" onPaste={handlePaste}>
      <section className="flex min-h-[calc(100vh-8rem)] flex-col justify-center gap-4">
        <UploadDropzone
          disabled={uploadDisabled}
          isDragging={isDragging}
          onDragStateChange={setIsDragging}
          onPasteImage={handlePaste}
          onSelectFile={handleUpload}
        />

        <Card className="grid gap-4 px-4 py-4 lg:grid-cols-[1fr_minmax(280px,420px)] lg:items-center" variant="default">
          <div className="flex flex-wrap items-center gap-3">
            <Link className="text-sm font-medium text-primary underline-offset-4 hover:underline" href="/history">
              {t.upload.quickHistory}
            </Link>
            <span className="hidden h-4 w-px bg-border sm:block" aria-hidden="true" />
            <Link className="text-sm font-medium text-primary underline-offset-4 hover:underline" href="/api">
              {t.upload.quickApi}
            </Link>
            <span className="hidden h-4 w-px bg-border sm:block" aria-hidden="true" />
            <span className="inline-flex items-center gap-2 text-sm text-muted-foreground">
              <Clipboard aria-hidden="true" className="h-4 w-4" />
              {t.upload.quickPasteHint}
            </span>
            <Badge className="w-fit" variant="secondary">
              <Sparkles aria-hidden="true" className="h-3 w-3" />
              {t.upload.quickSourceBadge}
            </Badge>
          </div>

          <form
            aria-describedby={uploadSourceHelpId}
            className="grid gap-2 sm:grid-cols-[minmax(0,1fr)_auto]"
            onSubmit={(event) => {
              event.preventDefault();
              void handleUrlUpload();
            }}
          >
            <label className="sr-only" htmlFor={urlInputId}>
              {t.upload.urlInputLabel}
            </label>
            <Input
              aria-describedby={uploadSourceHelpId}
              disabled={uploadDisabled}
              id={urlInputId}
              inputMode="url"
              onChange={(event) => setImageUrl(event.target.value)}
              placeholder={t.upload.urlInputPlaceholder}
              type="url"
              value={imageUrl}
            />
            <Button disabled={uploadDisabled || imageUrl.trim().length === 0} type="submit">
              <Link2 aria-hidden="true" className="h-4 w-4" />
              {t.upload.urlUploadAction}
            </Button>
            <p className="text-xs text-muted-foreground sm:col-span-2" id={uploadSourceHelpId}>
              {t.upload.sourceHelp}
            </p>
          </form>
        </Card>

        {phase === "error" && error ? (
          <Card className="border-danger/40 bg-danger/5 p-4 text-sm font-medium text-danger" role="alert">
            {error}
          </Card>
        ) : null}
      </section>

      <RecentUploads
        activeUpload={activeUpload}
        items={recent}
        latestResultUid={phase === "success" && result ? result.uid : null}
        title={t.upload.recentUploads}
      />
    </div>
  );
}

function fileFromClipboard(data: DataTransfer) {
  for (const item of Array.from(data.items)) {
    if (item.kind !== "file" || !isSupportedSourceImageType(item.type)) {
      continue;
    }
    const blob = item.getAsFile();
    if (!blob) {
      continue;
    }
    return new File([blob], clipboardFilename(item.type), {
      type: blob.type || item.type,
      lastModified: Date.now()
    });
  }
  return null;
}

function parseImageUrl(value: string) {
  const url = new URL(value.trim());
  if (url.protocol !== "http:" && url.protocol !== "https:") {
    throw new Error("Unsupported URL protocol");
  }
  return url;
}

type UrlImageDownloadErrorReason = "urlDownloadFailed" | "urlNotImage";

class UrlImageDownloadError extends Error {
  constructor(readonly reason: UrlImageDownloadErrorReason) {
    super(reason);
  }
}

async function downloadImageUrl(url: URL, fallbackName: string) {
  let response: Response;
  try {
    response = await fetch(url.toString());
  } catch {
    throw new UrlImageDownloadError("urlDownloadFailed");
  }

  if (!response.ok) {
    throw new UrlImageDownloadError("urlDownloadFailed");
  }

  const responseType = normalizeMimeType(response.headers.get("content-type") ?? "");
  if (!isSupportedSourceImageType(responseType) && !canInferImageTypeFromUrl(url, responseType)) {
    throw new UrlImageDownloadError("urlNotImage");
  }

  let blob: Blob;
  try {
    blob = await response.blob();
  } catch {
    throw new UrlImageDownloadError("urlDownloadFailed");
  }
  const blobType = normalizeMimeType(blob.type || responseType);
  const inferredType = isSupportedSourceImageType(blobType) ? blobType : imageTypeFromUrl(url);
  if (!inferredType) {
    throw new UrlImageDownloadError("urlNotImage");
  }

  return new File([blob], filenameFromUrl(url, inferredType, fallbackName), {
    type: inferredType,
    lastModified: Date.now()
  });
}

function clipboardFilename(type: string) {
  return `clipboard-image.${extensionFromMimeType(type)}`;
}

function isSupportedSourceImageType(type: string) {
  return supportedSourceImageTypes.has(normalizeMimeType(type));
}

function filenameFromUrl(url: URL, type: string, fallbackName: string) {
  const pathnameName = decodeURIComponent(url.pathname.split("/").filter(Boolean).pop() ?? "");
  const safeName = pathnameName.replace(/[\\/:*?"<>|]+/g, "-").trim();
  if (safeName) {
    return safeName;
  }
  return `${fallbackName}.${extensionFromMimeType(type)}`;
}

function canInferImageTypeFromUrl(url: URL, type: string) {
  return binaryResponseTypes.has(normalizeMimeType(type)) && imageTypeFromUrl(url) !== "";
}

function imageTypeFromUrl(url: URL) {
  return mimeTypeFromExtension(decodeURIComponent(url.pathname.split("/").filter(Boolean).pop() ?? ""));
}

function normalizeMimeType(type: string) {
  return type.split(";")[0]?.trim().toLowerCase() ?? "";
}

function mimeTypeFromExtension(value: string) {
  const extension = value.match(/\.(avif|png|jpe?g|gif|webp|bmp)$/i)?.[1]?.toLowerCase();
  switch (extension) {
    case "avif":
      return "image/avif";
    case "jpg":
    case "jpeg":
      return "image/jpeg";
    case "png":
      return "image/png";
    case "gif":
      return "image/gif";
    case "webp":
      return "image/webp";
    case "bmp":
      return "image/bmp";
    default:
      return "";
  }
}

function extensionFromMimeType(type: string) {
  switch (normalizeMimeType(type)) {
    case "image/avif":
      return "avif";
    case "image/jpeg":
      return "jpg";
    case "image/png":
      return "png";
    case "image/gif":
      return "gif";
    case "image/webp":
      return "webp";
    case "image/bmp":
      return "bmp";
    default:
      return "image";
  }
}
