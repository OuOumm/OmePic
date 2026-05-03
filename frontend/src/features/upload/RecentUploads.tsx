"use client";

import { useState } from "react";
import { Grid3X3, ImageOff } from "lucide-react";

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
          <span className="rounded-md bg-muted px-2.5 py-1 text-xs font-medium text-muted-foreground">
            {t.common.items(items.length)}
          </span>
        }
        description={items.length === 0 ? t.upload.emptyRecent : undefined}
        icon={<Grid3X3 aria-hidden="true" className="h-4 w-4" />}
        title={title}
      />

      {items.length === 0 ? (
        <ImgGalleryEmptyState icon={<ImageOff aria-hidden="true" className="h-10 w-10" />} title={t.upload.emptyRecent} />
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
