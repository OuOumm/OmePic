"use client";

import { useCallback, useEffect, useRef, useState } from "react";
import { Check, Grid3X3, Images, List, Search, SquareCheckBig, Trash2 } from "lucide-react";
import toast from "react-hot-toast";

import { CopyButton } from "@/components/shared/CopyButton";
import { ImageLightbox } from "@/components/shared/ImageLightbox";
import { ImgGalleryEmptyState, ImgStyleImageCard } from "@/components/shared/ImgStyleImageCard";
import { PageDetailPill, PageIntro, PageSectionHeader } from "@/components/shared/PageLayout";
import { Button } from "@/components/ui/Button";
import { Card } from "@/components/ui/Card";
import { Input } from "@/components/ui/Input";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow
} from "@/components/ui/Table";
import { useUiLocale, useUiTranslations } from "@/hooks/useUiPreferences";
import { adminDeleteImages, adminImages, apiUrl } from "@/lib/api";
import { formatBytes, formatDate } from "@/lib/format";
import { cn } from "@/lib/utils";
import { useAdminSessionStore } from "@/stores/admin-session-store";
import type { AdminImageItem } from "@/types/admin";

type ViewMode = "grid" | "list";

export function ImageTable() {
  const token = useAdminSessionStore((state) => state.token);
  const locale = useUiLocale();
  const t = useUiTranslations();
  const [items, setItems] = useState<AdminImageItem[]>([]);
  const [selected, setSelected] = useState<string[]>([]);
  const [previewItem, setPreviewItem] = useState<AdminImageItem | null>(null);
  const [searchInput, setSearchInput] = useState("");
  const [search, setSearch] = useState("");
  const [viewMode, setViewMode] = useState<ViewMode>("grid");
  const [page, setPage] = useState(1);
  const [total, setTotal] = useState(0);
  const [loading, setLoading] = useState(true);
  const [errorMessage, setErrorMessage] = useState<string | null>(null);
  const loadImagesFailedRef = useRef(t.admin.loadImagesFailed);

  useEffect(() => {
    loadImagesFailedRef.current = t.admin.loadImagesFailed;
  }, [t.admin.loadImagesFailed]);

  const loadImages = useCallback(async (signal?: AbortSignal) => {
    setLoading(true);
    setErrorMessage(null);
    try {
      const result = await adminImages(token, search, page, signal);
      if (signal?.aborted) {
        return;
      }
      setItems(result.items);
      setTotal(result.total);
      const resultUids = new Set(result.items.map((item) => item.uid));
      setSelected((current) => current.filter((uid) => resultUids.has(uid)));
      setPreviewItem((current) => (current && resultUids.has(current.uid) ? current : null));
    } catch (error) {
      if (signal?.aborted) {
        return;
      }
      const nextError = error instanceof Error ? error.message : loadImagesFailedRef.current;
      setErrorMessage(nextError);
      toast.error(nextError);
    } finally {
      if (!signal?.aborted) {
        setLoading(false);
      }
    }
  }, [page, search, token]);

  useEffect(() => {
    const handle = window.setTimeout(() => {
      setPage(1);
      setSearch(searchInput.trim());
    }, 250);

    return () => {
      window.clearTimeout(handle);
    };
  }, [searchInput]);

  useEffect(() => {
    const controller = new AbortController();

    void Promise.resolve().then(() => loadImages(controller.signal));

    return () => {
      controller.abort();
    };
  }, [loadImages]);

  async function handleBatchDelete() {
    if (selected.length === 0) {
      return;
    }
    const deleted = new Set(selected);
    try {
      await adminDeleteImages(token, selected);
      toast.success(t.admin.deleteSelectedSuccessToast);
      setSelected([]);
      setPreviewItem((current) => (current && deleted.has(current.uid) ? null : current));
      await loadImages();
    } catch (error) {
      setLoading(false);
      const nextError = error instanceof Error ? error.message : t.admin.batchDeleteFailed;
      setErrorMessage(nextError);
      toast.error(nextError);
    }
  }

  const visibleSelected = items.length > 0 && items.every((item) => selected.includes(item.uid));

  function toggleSelected(uid: string, checked: boolean) {
    setSelected((current) => (checked ? [...current, uid] : current.filter((item) => item !== uid)));
  }

  function toggleVisible() {
    if (visibleSelected) {
      const visible = new Set(items.map((item) => item.uid));
      setSelected((current) => current.filter((uid) => !visible.has(uid)));
      return;
    }
    setSelected((current) => Array.from(new Set([...current, ...items.map((item) => item.uid)])));
  }

  return (
    <div className="space-y-5 animate-fade-in">
      <PageIntro
        aside={
          <>
            <PageDetailPill label={t.common.total} value={String(total)} />
            <PageDetailPill label={t.common.select} value={t.common.items(selected.length)} />
          </>
        }
        description={t.admin.imageManagementDescription}
        eyebrow={t.admin.nav.images}
        title={t.admin.imageManagementTitle}
      />

      <Card aria-busy={loading} className="space-y-5 p-4 sm:p-5" variant="strong">
        <PageSectionHeader
          actions={
            <div className="flex flex-wrap items-center gap-2">
              <div className="flex rounded-md border border-border bg-muted p-1">
                <ViewButton active={viewMode === "grid"} label={t.admin.gridView} onClick={() => setViewMode("grid")}>
                  <Grid3X3 aria-hidden="true" className="h-4 w-4" />
                </ViewButton>
                <ViewButton active={viewMode === "list"} label={t.admin.listView} onClick={() => setViewMode("list")}>
                  <List aria-hidden="true" className="h-4 w-4" />
                </ViewButton>
              </div>
              <Button
                disabled={items.length === 0}
                onClick={toggleVisible}
                variant="secondary"
              >
                <SquareCheckBig aria-hidden="true" className="h-4 w-4" />
                {visibleSelected ? t.admin.deselectVisible : t.admin.selectVisible}
              </Button>
              <Button
                disabled={selected.length === 0}
                onClick={() => void handleBatchDelete()}
                title={t.admin.deleteSelectedTitle}
                variant="danger"
              >
                <Trash2 aria-hidden="true" className="h-4 w-4" />
                {t.admin.deleteSelected}
              </Button>
            </div>
          }
          description={t.admin.imageManagementDescription}
          title={t.admin.imageManagementTitle}
        />

        <div className="flex flex-col gap-3 lg:flex-row lg:items-center lg:justify-between">
          <div className="relative flex-1" role="search">
            <Search aria-hidden="true" className="pointer-events-none absolute left-4 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
            <Input
              aria-label={t.admin.searchInputLabel}
              aria-describedby={errorMessage ? "admin-images-error" : undefined}
              className="pl-11"
              onChange={(event) => setSearchInput(event.target.value)}
              placeholder={t.admin.searchPlaceholder}
              value={searchInput}
            />
          </div>
        </div>

        {errorMessage ? (
          <p className="rounded-md border border-rose-400/30 bg-rose-500/10 p-3 text-sm text-danger" id="admin-images-error" role="alert">
            {errorMessage}
          </p>
        ) : null}

        {loading ? (
          <LoadingGrid />
        ) : items.length === 0 ? (
          <ImgGalleryEmptyState icon={<Images aria-hidden="true" className="h-10 w-10" />} title={t.admin.noImagesFound} />
        ) : viewMode === "grid" ? (
          <div className="gallery-grid">
            {items.map((item, index) => (
              <ImageGridCard
                checked={selected.includes(item.uid)}
                index={index}
                item={item}
                key={item.uid}
                onCheckedChange={toggleSelected}
                onPreview={() => setPreviewItem(item)}
                t={t}
              />
            ))}
          </div>
        ) : (
          <div className="table-surface">
            <Table>
              <TableHeader className="bg-muted/50">
                <TableRow>
                  <TableHead scope="col">{t.admin.table.select}</TableHead>
                  <TableHead scope="col">{t.admin.table.uid}</TableHead>
                  <TableHead scope="col">{t.admin.table.type}</TableHead>
                  <TableHead scope="col">{t.admin.table.size}</TableHead>
                  <TableHead scope="col">{t.admin.table.token}</TableHead>
                  <TableHead scope="col">{t.admin.table.md5}</TableHead>
                  <TableHead scope="col">{t.admin.table.storageKey}</TableHead>
                  <TableHead scope="col">{t.admin.table.backend}</TableHead>
                  <TableHead scope="col">{t.admin.table.created}</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {items.map((item) => {
                  const checked = selected.includes(item.uid);
                  return (
                    <TableRow
                      className={cn(
                        "hover:bg-muted/60",
                        checked ? "bg-muted" : ""
                      )}
                      key={item.uid}
                    >
                      <TableCell>
                        <ImageCheckbox
                          checked={checked}
                          label={t.admin.selectImage(item.uid)}
                          onChange={(value) => toggleSelected(item.uid, value)}
                        />
                      </TableCell>
                      <TableCell className="font-mono text-xs">{item.uid}</TableCell>
                      <TableCell>{item.mime_type}</TableCell>
                      <TableCell className="tabular-nums">{formatBytes(item.size)}</TableCell>
                      <TableCell className="font-mono text-xs">{item.token}</TableCell>
                      <TableCell className="font-mono text-xs">{item.md5_hash}</TableCell>
                      <TableCell className="font-mono text-xs">{item.storage_key}</TableCell>
                      <TableCell>{item.storage_backend}</TableCell>
                      <TableCell className="whitespace-nowrap">{formatDate(item.created_at, locale)}</TableCell>
                    </TableRow>
                  );
                })}
              </TableBody>
            </Table>
          </div>
        )}

        <div className="flex flex-col gap-3 border-t border-border pt-4 text-sm text-muted-foreground sm:flex-row sm:items-center sm:justify-between">
          <p>{t.common.totalResults(total)}</p>
          <div className="flex gap-2">
            <Button
              disabled={page <= 1}
              onClick={() => {
                setLoading(true);
                setPage((value) => value - 1);
              }}
              variant="secondary"
            >
              {t.common.previous}
            </Button>
            <Button
              disabled={page * 20 >= total}
              onClick={() => {
                setLoading(true);
                setPage((value) => value + 1);
              }}
              variant="secondary"
            >
              {t.common.next}
            </Button>
          </div>
        </div>
      </Card>

      <div
        className={cn(
          "fixed inset-x-4 bottom-5 z-40 mx-auto flex max-w-xl items-center justify-between gap-3 rounded-lg border border-border bg-popover p-3 shadow-lg transition-all duration-300",
          selected.length > 0 ? "translate-y-0 opacity-100" : "pointer-events-none translate-y-6 opacity-0"
        )}
      >
        <p className="text-sm font-medium text-foreground">
          {t.common.items(selected.length)}
        </p>
        <div className="flex gap-2">
          <Button onClick={toggleVisible} size="sm" variant="secondary">
            <SquareCheckBig aria-hidden="true" className="h-4 w-4" />
            {visibleSelected ? t.admin.deselectVisible : t.admin.selectVisible}
          </Button>
          <Button onClick={() => void handleBatchDelete()} size="sm" variant="danger">
            <Trash2 aria-hidden="true" className="h-4 w-4" />
            {t.admin.deleteSelected}
          </Button>
        </div>
      </div>

      <ImageLightbox
        actions={
          previewItem ? (
            <CopyButton
              label={t.common.copyTargets.url}
              value={apiUrl(`/i/${previewItem.uid}.avif`)}
            />
          ) : null
        }
        closeLabel={t.common.closePreview}
        eyebrowLabel={t.common.preview}
        image={
          previewItem
            ? {
                alt: t.admin.imagePreviewAlt(previewItem.uid),
                metadata: [
                  { label: t.common.uid, value: previewItem.uid, mono: true },
                  { label: t.common.type, value: previewItem.mime_type },
                  { label: t.common.size, value: formatBytes(previewItem.size) },
                  { label: t.common.token, value: previewItem.token, mono: true },
                  { label: t.common.md5, value: previewItem.md5_hash, mono: true },
                  {
                    label: t.common.storage,
                    value: `${previewItem.storage_key} (${previewItem.storage_backend})`,
                    mono: true
                  },
                  { label: t.common.created, value: formatDate(previewItem.created_at, locale) }
                ],
                src: apiUrl(`/i/${previewItem.uid}.avif`),
                subtitle: previewItem.uid,
                title: previewItem.uid
              }
            : null
        }
        metadataLabel={t.common.previewMetadata}
        onClose={() => setPreviewItem(null)}
      />
    </div>
  );
}

function ImageGridCard({
  checked,
  index,
  item,
  onCheckedChange,
  onPreview,
  t
}: {
  checked: boolean;
  index: number;
  item: AdminImageItem;
  onCheckedChange: (uid: string, checked: boolean) => void;
  onPreview: () => void;
  t: ReturnType<typeof useUiTranslations>;
}) {
  return (
    <ImgStyleImageCard
      alt={t.admin.imagePreviewAlt(item.uid)}
      animationDelay={`${index * 50}ms`}
      onPreview={onPreview}
      previewLabel={t.common.openPreview(item.uid)}
      selected={checked}
      sizeLabel={formatBytes(item.size)}
      src={apiUrl(`/i/${item.uid}.avif`)}
      title={item.uid}
      topLeft={
        <ImageCheckbox
          checked={checked}
          label={t.admin.selectImage(item.uid)}
          onChange={(value) => onCheckedChange(item.uid, value)}
        />
      }
    />
  );
}

function ImageCheckbox({
  checked,
  label,
  onChange
}: {
  checked: boolean;
  label: string;
  onChange: (checked: boolean) => void;
}) {
  return (
    <label className="inline-flex">
      <span className="sr-only">{label}</span>
      <input
        checked={checked}
        className="peer sr-only"
        onChange={(event) => onChange(event.target.checked)}
        type="checkbox"
      />
      <span className="flex h-5 w-5 items-center justify-center rounded-sm border border-input bg-background text-primary transition-colors peer-checked:border-primary peer-focus-visible:ring-2 peer-focus-visible:ring-ring">
        {checked ? <Check aria-hidden="true" className="h-4 w-4" /> : null}
      </span>
    </label>
  );
}

function ViewButton({
  active,
  children,
  label,
  onClick
}: {
  active: boolean;
  children: React.ReactNode;
  label: string;
  onClick: () => void;
}) {
  return (
    <button
      aria-label={label}
      aria-pressed={active}
      className={cn(
        "flex h-9 w-9 items-center justify-center rounded-sm transition-colors focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring",
        active
          ? "bg-background text-foreground shadow-sm"
          : "text-muted-foreground hover:bg-background/70 hover:text-foreground"
      )}
      onClick={onClick}
      type="button"
    >
      {children}
    </button>
  );
}

function LoadingGrid() {
  return (
    <div className="gallery-grid" role="status">
      {[0, 1, 2, 3, 4, 5].map((item) => (
        <div
          className="overflow-hidden rounded-lg border border-border bg-card shadow-sm"
          key={item}
        >
          <div className="skeleton-glass aspect-square rounded-none" />
        </div>
      ))}
    </div>
  );
}
