"use client";

import { useState } from "react";
import { Grid3X3, ImageOff, LoaderCircle } from "lucide-react";

import { CopyButton } from "@/components/shared/CopyButton";
import { ImageLightbox } from "@/components/shared/ImageLightbox";
import { ImgGalleryEmptyState, ImgStyleImageCard } from "@/components/shared/ImgStyleImageCard";
import { PageSectionHeader } from "@/components/shared/PageLayout";
import { Card } from "@/components/ui/Card";
import { useUiLocale, useUiTranslations } from "@/hooks/useUiPreferences";
import { formatBytes, formatDate } from "@/lib/format";
import type { UploadHistoryRecord } from "@/types/upload";

type RecentUploadsProps = {
  activeUpload?: {
    error?: string;
    progress?: number;
    status: string;
  } | null;
  title: string;
  items: UploadHistoryRecord[];
  latestResultUid?: string | null;
};

export function RecentUploads({
  activeUpload,
  latestResultUid = null,
  title,
  items
}: RecentUploadsProps) {
  const t = useUiTranslations();
  const locale = useUiLocale();
  const [previewRecord, setPreviewRecord] = useState<UploadHistoryRecord | null>(null);
  const previewTitle = previewRecord?.original_filename || previewRecord?.uid || "";
  const showEmptyState = items.length === 0 && !activeUpload;

  return (
    <Card className="space-y-5 p-5 sm:p-6" variant="strong">
      <PageSectionHeader
        badge={
          <span className="rounded-md bg-muted px-2.5 py-1 text-xs font-medium text-muted-foreground">
            {t.common.items(items.length)}
          </span>
        }
        description={items.length === 0 ? t.upload.emptyRecent : undefined}
        icon={<Grid3X3 aria-hidden="true" className="h-4 w-4" />}
        title={title}
      />

      {showEmptyState ? (
        <ImgGalleryEmptyState icon={<ImageOff aria-hidden="true" className="h-10 w-10" />} title={t.upload.emptyRecent} />
      ) : (
        <ul className="grid grid-cols-2 gap-3 sm:gap-4 md:grid-cols-4 lg:grid-cols-5">
          {activeUpload ? (
            <li>
              <UploadProgressCard upload={activeUpload} />
            </li>
          ) : null}
          {items.map((item, index) => (
            <li key={item.uid}>
              <ImgStyleImageCard
                alt={item.original_filename || t.upload.previewAlt}
                animationDelay={`${index * 50}ms`}
                onPreview={() => setPreviewRecord(item)}
                previewLabel={t.common.openPreview(item.original_filename || item.uid)}
                sizeLabel={formatBytes(item.size)}
                src={item.url}
                title={item.original_filename || item.uid}
              />
              {latestResultUid === item.uid ? (
                <div className="mt-2 grid grid-cols-3 gap-2 rounded-lg border border-border bg-muted/40 p-2">
                  <CopyButton className="min-w-0 px-2 text-xs" label={t.common.copyTargets.url} value={item.url} />
                  <CopyButton className="min-w-0 px-2 text-xs" label={t.common.copyTargets.markdownShort} value={item.md_url} />
                  <CopyButton className="min-w-0 px-2 text-xs" label={t.common.copyTargets.bbcodeShort} value={item.bbcode} />
                </div>
              ) : null}
            </li>
          ))}
        </ul>
      )}

      <ImageLightbox
        actions={
          previewRecord ? (
            <>
              <CopyButton label={t.common.copyTargets.url} value={previewRecord.url} />
              <CopyButton label={t.common.copyTargets.markdown} value={previewRecord.md_url} />
              <CopyButton label={t.common.copyTargets.bbcode} value={previewRecord.bbcode} />
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
    </Card>
  );
}

function UploadProgressCard({
  upload
}: {
  upload: {
    error?: string;
    progress?: number;
    status: string;
  };
}) {
  const t = useUiTranslations();
  const progress = upload.progress ?? 0;
  const hasError = Boolean(upload.error);

  return (
    <div className="relative animate-fade-in overflow-hidden rounded-lg border border-border bg-card shadow-sm">
      <div className="flex aspect-square flex-col items-center justify-center gap-4 bg-muted/40 p-5 text-center">
        <span className="flex h-12 w-12 items-center justify-center rounded-md border border-border bg-background text-primary shadow-sm">
          <LoaderCircle
            aria-hidden="true"
            className={hasError ? "h-5 w-5" : "h-5 w-5 animate-spin"}
          />
        </span>
        <div className="w-full max-w-44 space-y-3">
          <div
            aria-label={t.upload.status}
            aria-valuemax={100}
            aria-valuemin={0}
            aria-valuenow={progress}
            className="h-2 overflow-hidden rounded-full bg-background"
            role="progressbar"
          >
            <div
              className={hasError ? "h-full rounded-full bg-danger transition-all duration-300" : "h-full rounded-full bg-primary transition-all duration-300"}
              style={{ width: `${progress}%` }}
            />
          </div>
          <p
            className={hasError ? "text-xs font-medium text-danger" : "text-xs font-medium text-muted-foreground"}
            role={hasError ? "alert" : "status"}
          >
            {upload.status}
          </p>
        </div>
      </div>
      <div className="absolute bottom-0 left-0 right-0 flex items-center justify-between border-t border-border bg-background/95 px-3 py-2 text-xs">
        <span className="truncate font-medium text-foreground">{t.upload.uploadingCardTitle}</span>
        <span className="flex-shrink-0 text-muted-foreground">{progress}%</span>
      </div>
    </div>
  );
}
