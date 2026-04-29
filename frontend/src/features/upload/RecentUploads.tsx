"use client";

import { useState } from "react";

import { CopyButton } from "@/components/shared/CopyButton";
import { ImageLightbox } from "@/components/shared/ImageLightbox";
import { ImgGalleryEmptyState, ImgStyleImageCard } from "@/components/shared/ImgStyleImageCard";
import { PageSectionHeader } from "@/components/shared/PageLayout";
import { Card } from "@/components/ui/Card";
import { useUiLocale, useUiTranslations } from "@/hooks/useUiPreferences";
import { formatBytes, formatDate } from "@/lib/format";
import type { UploadHistoryRecord } from "@/types/upload";

type RecentUploadsProps = {
  title: string;
  items: UploadHistoryRecord[];
};

export function RecentUploads({ title, items }: RecentUploadsProps) {
  const t = useUiTranslations();
  const locale = useUiLocale();
  const [previewRecord, setPreviewRecord] = useState<UploadHistoryRecord | null>(null);
  const previewTitle = previewRecord?.original_filename || previewRecord?.uid || "";

  return (
    <Card className="space-y-5 p-5 sm:p-6" variant="strong">
      <PageSectionHeader
        badge={
          <span className="rounded-full bg-white/60 px-3 py-1 text-xs font-semibold text-muted backdrop-blur-xl dark:bg-slate-800/60">
            {t.common.items(items.length)}
          </span>
        }
        description={items.length === 0 ? t.upload.emptyRecent : undefined}
        icon={<GridIcon />}
        title={title}
      />

      {items.length === 0 ? (
        <ImgGalleryEmptyState icon={<EmptyImageIcon />} title={t.upload.emptyRecent} />
      ) : (
        <ul className="gallery-grid">
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

function GridIcon() {
  return (
    <svg aria-hidden="true" className="h-5 w-5 text-violet-500 dark:text-violet-300" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
      <path d="M4 5a1 1 0 0 1 1-1h5v5H4V5Zm10-1h5a1 1 0 0 1 1 1v5h-6V4ZM4 14h6v6H5a1 1 0 0 1-1-1v-5Zm10 0h6v5a1 1 0 0 1-1 1h-5v-6Z" strokeLinecap="round" strokeLinejoin="round" />
    </svg>
  );
}

function EmptyImageIcon() {
  return (
    <svg aria-hidden="true" className="h-10 w-10" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={1.6}>
      <path d="m4 16 4.6-4.6a2 2 0 0 1 2.8 0L16 16m-2-2 1.6-1.6a2 2 0 0 1 2.8 0L20 14M14 8h.01M6 20h12a2 2 0 0 0 2-2V6a2 2 0 0 0-2-2H6a2 2 0 0 0-2 2v12a2 2 0 0 0 2 2Z" strokeLinecap="round" strokeLinejoin="round" />
    </svg>
  );
}
