"use client";

import { useCallback, useEffect, useRef, useState } from "react";
import toast from "react-hot-toast";

import { CopyButton } from "@/components/shared/CopyButton";
import { ImageLightbox } from "@/components/shared/ImageLightbox";
import { ImgGalleryEmptyState, ImgStyleImageCard } from "@/components/shared/ImgStyleImageCard";
import { PageDetailPill, PageIntro, PageSectionHeader } from "@/components/shared/PageLayout";
import { Button } from "@/components/ui/Button";
import { Card } from "@/components/ui/Card";
import { Input } from "@/components/ui/Input";
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
              <div className="flex rounded-2xl border border-white/50 bg-white/60 p-1 shadow-sm backdrop-blur-xl dark:border-white/10 dark:bg-slate-950/50">
                <ViewButton active={viewMode === "grid"} label={t.admin.gridView} onClick={() => setViewMode("grid")}>
                  <GridIcon />
                </ViewButton>
                <ViewButton active={viewMode === "list"} label={t.admin.listView} onClick={() => setViewMode("list")}>
                  <ListIcon />
                </ViewButton>
              </div>
              <Button
                disabled={items.length === 0}
                onClick={toggleVisible}
                variant="secondary"
              >
                <SelectIcon />
                {visibleSelected ? t.admin.deselectVisible : t.admin.selectVisible}
              </Button>
              <Button
                disabled={selected.length === 0}
                onClick={() => void handleBatchDelete()}
                title={t.admin.deleteSelectedTitle}
                variant="danger"
              >
                <TrashIcon />
                {t.admin.deleteSelected}
              </Button>
            </div>
          }
          description={t.admin.imageManagementDescription}
          title={t.admin.imageManagementTitle}
        />

        <div className="flex flex-col gap-3 lg:flex-row lg:items-center lg:justify-between">
          <div className="relative flex-1" role="search">
            <SearchIcon />
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
          <p className="rounded-xl border border-rose-400/30 bg-rose-500/10 p-3 text-sm text-danger" id="admin-images-error" role="alert">
            {errorMessage}
          </p>
        ) : null}

        {loading ? (
          <LoadingGrid />
        ) : items.length === 0 ? (
          <ImgGalleryEmptyState icon={<GalleryIcon />} title={t.admin.noImagesFound} />
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
            <table className="min-w-full text-left text-sm">
              <thead className="bg-white/60 text-xs uppercase tracking-wide text-muted backdrop-blur-xl dark:bg-slate-950/60">
                <tr>
                  <th className="px-3 py-3" scope="col">{t.admin.table.select}</th>
                  <th className="px-3 py-3" scope="col">{t.admin.table.uid}</th>
                  <th className="px-3 py-3" scope="col">{t.admin.table.type}</th>
                  <th className="px-3 py-3" scope="col">{t.admin.table.size}</th>
                  <th className="px-3 py-3" scope="col">{t.admin.table.token}</th>
                  <th className="px-3 py-3" scope="col">{t.admin.table.md5}</th>
                  <th className="px-3 py-3" scope="col">{t.admin.table.storageKey}</th>
                  <th className="px-3 py-3" scope="col">{t.admin.table.backend}</th>
                  <th className="px-3 py-3" scope="col">{t.admin.table.created}</th>
                </tr>
              </thead>
              <tbody>
                {items.map((item) => {
                  const checked = selected.includes(item.uid);
                  return (
                    <tr
                      className={cn(
                        "border-t border-white/50 transition-colors duration-200 hover:bg-violet-500/10 dark:border-white/10",
                        checked ? "bg-violet-500/10" : "bg-white/25 dark:bg-slate-950/20"
                      )}
                      key={item.uid}
                    >
                      <td className="px-3 py-3">
                        <ImageCheckbox
                          checked={checked}
                          label={t.admin.selectImage(item.uid)}
                          onChange={(value) => toggleSelected(item.uid, value)}
                        />
                      </td>
                      <td className="px-3 py-3 font-mono text-xs">{item.uid}</td>
                      <td className="px-3 py-3">{item.mime_type}</td>
                      <td className="px-3 py-3 tabular-nums">{formatBytes(item.size)}</td>
                      <td className="px-3 py-3 font-mono text-xs">{item.token}</td>
                      <td className="px-3 py-3 font-mono text-xs">{item.md5_hash}</td>
                      <td className="px-3 py-3 font-mono text-xs">{item.storage_key}</td>
                      <td className="px-3 py-3">{item.storage_backend}</td>
                      <td className="whitespace-nowrap px-3 py-3">{formatDate(item.created_at, locale)}</td>
                    </tr>
                  );
                })}
              </tbody>
            </table>
          </div>
        )}

        <div className="flex flex-col gap-3 border-t border-white/50 pt-4 text-sm text-muted sm:flex-row sm:items-center sm:justify-between dark:border-white/10">
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
          "fixed inset-x-4 bottom-5 z-40 mx-auto flex max-w-xl items-center justify-between gap-3 rounded-[24px] border border-white/50 bg-white/80 p-3 shadow-glow backdrop-blur-xl transition-all duration-300 dark:border-white/10 dark:bg-slate-900/90",
          selected.length > 0 ? "translate-y-0 opacity-100" : "pointer-events-none translate-y-6 opacity-0"
        )}
      >
        <p className="text-sm font-semibold text-slate-800 dark:text-slate-100">
          {t.common.items(selected.length)}
        </p>
        <div className="flex gap-2">
          <Button onClick={toggleVisible} size="sm" variant="secondary">
            <SelectIcon />
            {visibleSelected ? t.admin.deselectVisible : t.admin.selectVisible}
          </Button>
          <Button onClick={() => void handleBatchDelete()} size="sm" variant="danger">
            <TrashIcon />
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
      <span className="flex h-6 w-6 items-center justify-center rounded-full border border-white/70 bg-white/40 text-white shadow-sm backdrop-blur-md transition-all duration-200 peer-checked:border-violet-300 peer-checked:bg-gradient-to-br peer-checked:from-violet-500 peer-checked:to-cyan-500 peer-focus-visible:ring-2 peer-focus-visible:ring-violet-400">
        {checked ? <CheckIcon /> : null}
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
        "flex h-9 w-9 items-center justify-center rounded-lg transition-all duration-200 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-violet-400",
        active
          ? "bg-gradient-to-r from-violet-600 to-cyan-600 text-white shadow-md shadow-violet-500/20"
          : "text-muted hover:bg-white/70 hover:text-violet-700 dark:hover:bg-white/10 dark:hover:text-violet-200"
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
          className="overflow-hidden rounded-2xl border border-slate-200/60 bg-white shadow-sm dark:border-slate-700/40 dark:bg-slate-800/60"
          key={item}
        >
          <div className="skeleton-glass aspect-square rounded-none" />
        </div>
      ))}
    </div>
  );
}

function SearchIcon() {
  return (
    <svg aria-hidden="true" className="pointer-events-none absolute left-4 top-1/2 h-4 w-4 -translate-y-1/2 text-muted" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
      <path d="m21 21-5.8-5.8M17 10a7 7 0 1 1-14 0 7 7 0 0 1 14 0Z" strokeLinecap="round" strokeLinejoin="round" />
    </svg>
  );
}

function GridIcon() {
  return (
    <svg aria-hidden="true" className="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
      <path d="M4 5a1 1 0 0 1 1-1h5v5H4V5Zm10-1h5a1 1 0 0 1 1 1v5h-6V4ZM4 14h6v6H5a1 1 0 0 1-1-1v-5Zm10 0h6v5a1 1 0 0 1-1 1h-5v-6Z" strokeLinecap="round" strokeLinejoin="round" />
    </svg>
  );
}

function ListIcon() {
  return (
    <svg aria-hidden="true" className="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
      <path d="M8 6h12M8 12h12M8 18h12M4 6h.01M4 12h.01M4 18h.01" strokeLinecap="round" strokeLinejoin="round" />
    </svg>
  );
}

function SelectIcon() {
  return (
    <svg aria-hidden="true" className="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
      <path d="M9 11 12 14 22 4M21 12v7a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h11" strokeLinecap="round" strokeLinejoin="round" />
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

function GalleryIcon() {
  return (
    <svg aria-hidden="true" className="h-10 w-10" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={1.6}>
      <path d="M4 5a2 2 0 0 1 2-2h12a2 2 0 0 1 2 2v14a2 2 0 0 1-2 2H6a2 2 0 0 1-2-2V5Zm3 12 3.5-4 2.5 3 2-2.4 2 3.4M15 8h.01" strokeLinecap="round" strokeLinejoin="round" />
    </svg>
  );
}

function CheckIcon() {
  return (
    <svg aria-hidden="true" className="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2.5}>
      <path d="m5 13 4 4L19 7" strokeLinecap="round" strokeLinejoin="round" />
    </svg>
  );
}
