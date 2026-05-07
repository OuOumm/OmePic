"use client";

import { useState, useEffect, useCallback } from "react";
import { Card, CardContent } from "@/components/ui/Card";
import { Button } from "@/components/ui/Button";
import { ImageLightbox } from "@/components/shared/ImageLightbox";
import { ImgStyleImageCard } from "@/components/shared/ImgStyleImageCard";
import { Separator } from "@/components/ui/Separator";
import { useUiPreferencesStore } from "@/stores/ui-preferences-store";
import {
  getAllUploads,
  deleteUploadFromHistory,
  clearUploadHistory,
  getUploadCount,
} from "@/lib/indexeddb/upload-history";
import { deleteImageByUid } from "@/lib/api";
import { getClientToken } from "@/lib/preferences";
import { t } from "@/lib/i18n";
import { formatBytes, formatDate } from "@/lib/utils";
import { Trash2, AlertTriangle, Loader2, Inbox } from "lucide-react";
import toast from "react-hot-toast";
import type { UploadHistoryRecord } from "@/types";

export function HistoryPageClient() {
  const language = useUiPreferencesStore((state) => state.language);
  const hasHydrated = useUiPreferencesStore((state) => state.hasHydrated);

  const [records, setRecords] = useState<UploadHistoryRecord[]>([]);
  const [count, setCount] = useState(0);
  const [loading, setLoading] = useState(true);
  const [previewRecord, setPreviewRecord] = useState<UploadHistoryRecord | null>(null);
  const [deleting, setDeleting] = useState<string | null>(null);

  const loadData = useCallback(async () => {
    try {
      const [all, cnt] = await Promise.all([getAllUploads(), getUploadCount()]);
      setRecords(all);
      setCount(cnt);
    } catch { /* ignore */ }
    setLoading(false);
  }, []);

  useEffect(() => {
    loadData();
  }, [loadData]);

  const handleClear = useCallback(async () => {
    if (!window.confirm(t(language, "history.clearConfirm"))) return;
    try {
      await clearUploadHistory();
      setRecords([]);
      setCount(0);
      toast.success(t(language, "history.cleared"));
    } catch {
      toast.error(t(language, "common.error"));
    }
  }, [language]);

  const handleDelete = useCallback(
    async (record: UploadHistoryRecord) => {
      if (!window.confirm(t(language, "history.deleteConfirm"))) return;
      const currentToken = getClientToken();
      if (record.client_token !== currentToken) {
        toast.error("Only the original uploader can delete this image");
        return;
      }
      setDeleting(record.uid);
      try {
        await deleteImageByUid(record.uid, currentToken);
        await deleteUploadFromHistory(record.uid);
        setRecords((prev) => prev.filter((r) => r.uid !== record.uid));
        setCount((c) => c - 1);
        if (previewRecord?.uid === record.uid) setPreviewRecord(null);
        toast.success(t(language, "history.deleted"));
      } catch {
        toast.error(t(language, "history.deleteError"));
      } finally {
        setDeleting(null);
      }
    },
    [language, previewRecord]
  );

  if (!hasHydrated || loading) {
    return (
      <div className="flex items-center justify-center py-20">
        <Loader2 className="h-6 w-6 animate-spin text-muted-foreground" />
      </div>
    );
  }

  const lang = language;

  return (
    <div className="space-y-6" id="main-content">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-xl font-bold">{t(lang, "history.title")}</h1>
          <p className="text-sm text-muted-foreground">{t(lang, "history.count", { count })}</p>
        </div>
        {records.length > 0 && (
          <Button
            variant="outline"
            size="sm"
            onClick={handleClear}
            className="cursor-pointer gap-1 text-destructive hover:text-destructive"
          >
            <Trash2 className="h-3.5 w-3.5" />
            {t(lang, "history.clear")}
          </Button>
        )}
      </div>

      {records.length === 0 ? (
        <Card>
          <CardContent className="pt-6 flex flex-col items-center py-12 gap-3">
            <Inbox className="h-10 w-10 text-muted-foreground" />
            <p className="text-muted-foreground">{t(lang, "history.empty")}</p>
          </CardContent>
        </Card>
      ) : (
        <div className="grid grid-cols-2 sm:grid-cols-3 md:grid-cols-4 lg:grid-cols-5 gap-3">
          {records.map((rec) => (
            <ImgStyleImageCard
              key={rec.uid}
              src={rec.url}
              alt={rec.original_filename || rec.uid}
              title={rec.uid}
              filename={rec.original_filename}
              sizeLabel={formatBytes(rec.size)}
              onPreview={() => setPreviewRecord(rec)}
              previewLabel={t(lang, "common.openPreview", { title: rec.uid })}
              topLeft={
                rec.client_token === getClientToken() && (
                  <Button
                    variant="destructive"
                    size="icon"
                    className="h-6 w-6 cursor-pointer"
                    onClick={(e) => { e.stopPropagation(); handleDelete(rec); }}
                    disabled={deleting === rec.uid}
                    aria-label={t(lang, "history.delete")}
                  >
                    {deleting === rec.uid ? <Loader2 className="h-3 w-3 animate-spin" /> : <Trash2 className="h-3 w-3" />}
                  </Button>
                )
              }
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
          ))}
        </div>
      )}

      {/* Preview lightbox */}
      <ImageLightbox
        open={!!previewRecord}
        onClose={() => setPreviewRecord(null)}
        initialIndex={previewRecord ? records.findIndex((r) => r.uid === previewRecord.uid) : 0}
        images={records.map((r) => ({
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
          const r = records[idx];
          if (!r) return [];
          const token = getClientToken();
          return [
            ...(r.client_token === token
              ? [{ label: t(lang, "history.delete"), onClick: () => handleDelete(r) }]
              : []),
            {
              label: t(lang, "common.copyUrl"),
              onClick: () => { navigator.clipboard.writeText(r.url); toast.success(t(lang, "common.copied")); },
            },
            {
              label: "MD",
              onClick: () => { navigator.clipboard.writeText(r.markdown || `![](${r.url})`); toast.success(t(lang, "common.copied")); },
            },
            {
              label: "BB",
              onClick: () => { navigator.clipboard.writeText(r.bbcode || `[img]${r.url}[/img]`); toast.success(t(lang, "common.copied")); },
            },
          ];
        }}
        closeLabel={t(lang, "common.close")}
        metadataLabel={t(lang, "history.viewPreview")}
      />
    </div>
  );
}
