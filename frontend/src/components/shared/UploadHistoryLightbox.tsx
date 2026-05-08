"use client";

import toast from "react-hot-toast";
import { ImageLightbox } from "@/components/shared/ImageLightbox";
import { t } from "@/lib/i18n";
import { formatBytes, formatDate, getAbsoluteImageUrl, getAbsoluteUrl, getImageUrl } from "@/lib/utils";
import type { AdminImage, Language, UploadHistoryRecord } from "@/types";

type UploadHistoryLightboxItem =
  | { type: "upload"; record: UploadHistoryRecord }
  | { type: "admin"; image: AdminImage };

type UploadHistoryLightboxProps = {
  open: boolean;
  items: UploadHistoryLightboxItem[];
  selectedUid: string | null;
  language: Language;
  onClose: () => void;
  metadataLabel?: string;
  canDelete?: (record: UploadHistoryRecord) => boolean;
  onDelete?: (record: UploadHistoryRecord) => void;
  onDeleteAdminImage?: (image: AdminImage) => void;
};

function getItemUid(item: UploadHistoryLightboxItem) {
  return item.type === "upload" ? item.record.uid : item.image.uid;
}

function getItemUrl(item: UploadHistoryLightboxItem) {
  return item.type === "upload" ? item.record.url : getImageUrl(item.image.uid);
}

function getItemCopyUrl(item: UploadHistoryLightboxItem) {
  return item.type === "upload" ? getAbsoluteUrl(item.record.url) : getAbsoluteImageUrl(item.image.uid);
}

function getItemAlt(item: UploadHistoryLightboxItem) {
  if (item.type === "upload") return item.record.original_filename || item.record.uid;
  return item.image.uid;
}

function getItemMetadata(item: UploadHistoryLightboxItem, language: Language) {
  if (item.type === "upload") {
    const record = item.record;
    return [
      { label: t(language, "image.uid"), value: record.uid },
      { label: t(language, "image.storageKey"), value: record.storage_key },
      { label: t(language, "image.storageBackend"), value: record.storage_backend },
      { label: t(language, "image.type"), value: record.mime_type },
      { label: t(language, "image.size"), value: formatBytes(record.size) },
      { label: t(language, "image.created"), value: formatDate(record.created_at) },
    ];
  }

  const image = item.image;
  return [
    { label: t(language, "image.uid"), value: image.uid },
    { label: t(language, "image.type"), value: image.mime_type },
    { label: t(language, "image.size"), value: formatBytes(image.size) },
    { label: t(language, "image.token"), value: image.token },
    { label: t(language, "image.md5"), value: image.md5_hash },
    { label: t(language, "image.storageKey"), value: image.storage_key },
    { label: t(language, "image.storageBackend"), value: image.storage_backend },
    { label: t(language, "image.created"), value: formatDate(image.created_at) },
  ];
}

function copyText(text: string, language: Language) {
  navigator.clipboard.writeText(text);
  toast.success(t(language, "common.copied"));
}

export function UploadHistoryLightbox({
  open,
  items,
  selectedUid,
  language,
  onClose,
  metadataLabel,
  canDelete,
  onDelete,
  onDeleteAdminImage,
}: UploadHistoryLightboxProps) {
  const initialIndex = selectedUid
    ? Math.max(0, items.findIndex((item) => getItemUid(item) === selectedUid))
    : 0;

  return (
    <ImageLightbox
      open={open}
      onClose={onClose}
      initialIndex={initialIndex}
      images={items.map((item) => ({
        url: getItemUrl(item),
        alt: getItemAlt(item),
        metadata: getItemMetadata(item, language),
      }))}
      getActions={(_, index) => {
        const item = items[index];
        if (!item) return [];

        const url = getItemCopyUrl(item);
        const markdown = item.type === "upload" ? item.record.markdown || `![](${url})` : `![](${url})`;
        const bbcode = item.type === "upload" ? item.record.bbcode || `[img]${url}[/img]` : `[img]${url}[/img]`;

        return [
          ...(item.type === "upload" && canDelete?.(item.record) && onDelete
            ? [{ label: t(language, "history.delete"), onClick: () => onDelete(item.record), variant: "destructive" as const }]
            : []),
          ...(item.type === "admin" && onDeleteAdminImage
            ? [{ label: t(language, "history.delete"), onClick: () => onDeleteAdminImage(item.image), variant: "destructive" as const }]
            : []),
          {
            label: t(language, "common.copyUrl"),
            onClick: () => copyText(url, language),
          },
          {
            label: t(language, "common.copyMarkdown"),
            onClick: () => copyText(markdown, language),
          },
          {
            label: t(language, "common.copyBBCode"),
            onClick: () => copyText(bbcode, language),
          },
        ];
      }}
      closeLabel={t(language, "common.close")}
      metadataLabel={metadataLabel ?? t(language, "history.viewPreview")}
    />
  );
}
