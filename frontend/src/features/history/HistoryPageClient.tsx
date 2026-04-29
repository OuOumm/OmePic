"use client";

import { useEffect, useState } from "react";
import toast from "react-hot-toast";

import { CopyButton } from "@/components/shared/CopyButton";
import { ImageLightbox } from "@/components/shared/ImageLightbox";
import { ImgGalleryEmptyState, ImgStyleImageCard } from "@/components/shared/ImgStyleImageCard";
import { PageDetailPill, PageIntro } from "@/components/shared/PageLayout";
import { Button } from "@/components/ui/Button";
import { Card } from "@/components/ui/Card";
import { useClientToken } from "@/hooks/useClientToken";
import { useUiLocale, useUiTranslations } from "@/hooks/useUiPreferences";
import { deleteImage } from "@/lib/api";
import { formatBytes, formatDate } from "@/lib/format";
import {
  clearUploadRecords,
  listUploadRecords,
  removeUploadRecord
} from "@/lib/indexeddb/upload-history";
import type { UploadHistoryRecord } from "@/types/upload";

export function HistoryPageClient() {
  const { token, ready } = useClientToken();
  const t = useUiTranslations();
  const locale = useUiLocale();
  const [records, setRecords] = useState<UploadHistoryRecord[]>([]);
  const [previewRecord, setPreviewRecord] = useState<UploadHistoryRecord | null>(null);
  const [message, setMessage] = useState<string | null>(null);
  const [errorMessage, setErrorMessage] = useState<string | null>(null);
  const previewTitle = previewRecord?.original_filename || previewRecord?.uid || "";
  const previewCanDelete = previewRecord ? token === previewRecord.token : false;

  useEffect(() => {
    if (!ready) {
      return;
    }
    let cancelled = false;
    void listUploadRecords().then((items) => {
      if (!cancelled) {
        setRecords(items);
      }
    });
    return () => {
      cancelled = true;
    };
  }, [ready]);

  async function handleDelete(record: UploadHistoryRecord) {
    setErrorMessage(null);
    setMessage(null);
    try {
      await deleteImage(record.uid, token);
      await removeUploadRecord(record.uid);
      setRecords(await listUploadRecords());
      setPreviewRecord((current) => (current?.uid === record.uid ? null : current));
      setMessage(t.historyPage.deleteSuccessToast);
      toast.success(t.historyPage.deleteSuccessToast);
    } catch (error) {
      const nextError = error instanceof Error ? error.message : t.historyPage.deleteFailed;
      setErrorMessage(nextError);
      toast.error(nextError);
    }
  }

  async function handleClear() {
    await clearUploadRecords();
    setRecords([]);
    setPreviewRecord(null);
    setMessage(null);
    setErrorMessage(null);
  }

  return (
    <div className="space-y-6 animate-fade-in">
      <PageIntro
        actions={
          <Button disabled={records.length === 0} onClick={() => void handleClear()} variant="secondary">
            <TrashIcon />
            {t.historyPage.clear}
          </Button>
        }
        aside={
          <>
            <PageDetailPill label={t.common.total} value={t.common.items(records.length)} />
            <PageDetailPill label={t.common.preview} value={previewRecord ? previewTitle : t.historyPage.empty} />
          </>
        }
        description={t.historyPage.description}
        eyebrow={t.historyPage.eyebrow}
        title={t.historyPage.title}
      />

      {message ? (
        <Card className="border-violet-300/40 bg-violet-500/10 p-4 text-sm text-slate-800 dark:text-violet-100" role="status" variant="subtle">
          {message}
        </Card>
      ) : null}

      {errorMessage ? (
        <Card className="border-rose-400/30 bg-rose-500/10 p-4 text-sm text-danger" role="alert" variant="subtle">
          {errorMessage}
        </Card>
      ) : null}

      {records.length === 0 ? (
        <ImgGalleryEmptyState icon={<GalleryIcon />} title={t.historyPage.empty} />
      ) : (
        <ul className="gallery-grid">
          {records.map((record, index) => {
            return (
              <li key={record.uid}>
                <ImgStyleImageCard
                  alt={record.original_filename || t.upload.previewAlt}
                  animationDelay={`${index * 50}ms`}
                  onPreview={() => setPreviewRecord(record)}
                  previewLabel={t.common.openPreview(record.original_filename || record.uid)}
                  sizeLabel={formatBytes(record.size)}
                  src={record.url}
                  title={record.original_filename || record.uid}
                />
              </li>
            );
          })}
        </ul>
      )}

      <ImageLightbox
        actions={
          previewRecord ? (
            <>
              <CopyButton label={t.common.copyTargets.url} value={previewRecord.url} />
              <CopyButton label={t.common.copyTargets.markdown} value={previewRecord.md_url} />
              <CopyButton label={t.common.copyTargets.bbcode} value={previewRecord.bbcode} />
              <Button
                disabled={!previewCanDelete}
                onClick={() => void handleDelete(previewRecord)}
                size="sm"
                title={
                  previewCanDelete
                    ? t.historyPage.deleteEnabledTitle
                    : t.historyPage.deleteDisabledTitle
                }
                variant="danger"
              >
                <TrashIcon />
                {t.historyPage.deleteUid}
              </Button>
            </>
          ) : null
        }
        closeLabel={t.common.closePreview}
        eyebrowLabel={t.common.preview}
        image={
          previewRecord
            ? {
                alt: previewRecord.original_filename || t.upload.previewAlt,
                metadata: [
                  { label: t.common.uid, value: previewRecord.uid, mono: true },
                  {
                    label: t.common.storage,
                    value: `${previewRecord.storage_key} (${previewRecord.storage_backend})`,
                    mono: true
                  },
                  { label: t.common.type, value: previewRecord.mime_type },
                  { label: t.common.size, value: formatBytes(previewRecord.size) },
                  { label: t.common.created, value: formatDate(previewRecord.created_at, locale) }
                ],
                src: previewRecord.url,
                subtitle: previewRecord.uid,
                title: previewTitle
              }
            : null
        }
        metadataLabel={t.common.previewMetadata}
        onClose={() => setPreviewRecord(null)}
      />
    </div>
  );
}

function GalleryIcon() {
  return (
    <svg aria-hidden="true" className="h-12 w-12" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={1.6}>
      <path d="M4 5a2 2 0 0 1 2-2h12a2 2 0 0 1 2 2v14a2 2 0 0 1-2 2H6a2 2 0 0 1-2-2V5Zm3 12 3.5-4 2.5 3 2-2.4 2 3.4M15 8h.01" strokeLinecap="round" strokeLinejoin="round" />
    </svg>
  );
}

function TrashIcon() {
  return (
    <svg aria-hidden="true" className="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
      <path d="m19 7-.8 12.1A2 2 0 0 1 16.2 21H7.8a2 2 0 0 1-2-1.9L5 7m5 4v6m4-6v6M9 7V4h6v3M4 7h16" strokeLinecap="round" strokeLinejoin="round" />
    </svg>
  );
}
