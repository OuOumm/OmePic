"use client";

import { useState, useEffect, useCallback, useRef } from "react";
import { Card } from "@/components/ui/Card";
import { Button } from "@/components/ui/Button";
import { Input } from "@/components/ui/Input";
import { UploadHistoryLightbox } from "@/components/shared/UploadHistoryLightbox";
import { ImgStyleImageCard } from "@/components/shared/ImgStyleImageCard";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/Table";
import { useAdminSessionStore } from "@/stores/admin-session-store";
import { useUiPreferencesStore } from "@/stores/ui-preferences-store";
import { adminGetImages, adminDeleteImages, adminCreateIPBan, adminDeleteIPBanImages } from "@/lib/api";
import { t } from "@/lib/i18n";
import { formatBytes, formatDate, getImageUrl } from "@/lib/utils";
import {
  Search,
  LayoutGrid,
  List,
  ChevronLeft,
  ChevronRight,
  Trash2,
  Loader2,
  AlertCircle,
  CheckSquare,
  Square,
  Ban,
  Eye,
  EyeOff,
} from "lucide-react";
import toast from "react-hot-toast";
import type { AdminImage, ViewMode } from "@/types";

export function ImageTable() {
  const token = useAdminSessionStore((state) => state.token);
  const language = useUiPreferencesStore((state) => state.language);

  const [images, setImages] = useState<AdminImage[]>([]);
  const [total, setTotal] = useState(0);
  const [page, setPage] = useState(1);
  const [pageSize] = useState(20);
  const [search, setSearch] = useState("");
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");
  const [viewMode, setViewMode] = useState<ViewMode>("grid");
  const [selectedUids, setSelectedUids] = useState<Set<string>>(new Set());
  const [deleting, setDeleting] = useState(false);
  const [banningUid, setBanningUid] = useState<string | null>(null);
  const [showFullIPs, setShowFullIPs] = useState<Set<string>>(new Set());
  const [previewImage, setPreviewImage] = useState<AdminImage | null>(null);

  const debounceRef = useRef<NodeJS.Timeout | null>(null);
  const [searchInput, setSearchInput] = useState("");

  const loadImages = useCallback(async () => {
    if (!token) return;
    setLoading(true);
    setError("");
    try {
      const result = await adminGetImages(token, page, pageSize, search || undefined);
      setImages(result.items);
      setTotal(result.total);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to load images");
    } finally {
      setLoading(false);
    }
  }, [token, page, pageSize, search]);

  useEffect(() => {
    loadImages();
  }, [loadImages]);

  // Search debounce
  const handleSearchChange = (value: string) => {
    setSearchInput(value);
    if (debounceRef.current) clearTimeout(debounceRef.current);
    debounceRef.current = setTimeout(() => {
      setSearch(value);
      setPage(1);
    }, 250);
  };

  const handleSelectAll = () => {
    const allUids = new Set(images.map((img) => img.uid));
    if (selectedUids.size === images.length) {
      setSelectedUids(new Set());
    } else {
      setSelectedUids(allUids);
    }
  };

  const handleSelect = (uid: string, checked: boolean) => {
    const next = new Set(selectedUids);
    if (checked) {
      next.add(uid);
    } else {
      next.delete(uid);
    }
    setSelectedUids(next);
  };

  const handleBatchDelete = async () => {
    if (selectedUids.size === 0 || !token) return;
    const count = selectedUids.size;
    if (!window.confirm(t(language, "admin.imagesDeleteConfirm", { count }))) return;
    setDeleting(true);
    try {
      await adminDeleteImages(token, Array.from(selectedUids));
      setSelectedUids(new Set());
      toast.success(t(language, "admin.imagesDeleted", { count }));
      loadImages();
    } catch (err) {
      toast.error(t(language, "admin.imagesDeleteError"));
    } finally {
      setDeleting(false);
    }
  };

  const handleDeleteImage = async (image: AdminImage) => {
    if (!token) return;
    if (!window.confirm(t(language, "admin.imagesDeleteConfirm", { count: 1 }))) return;
    setDeleting(true);
    try {
      await adminDeleteImages(token, [image.uid]);
      setPreviewImage(null);
      setSelectedUids((prev) => {
        const next = new Set(prev);
        next.delete(image.uid);
        return next;
      });
      toast.success(t(language, "admin.imagesDeleted", { count: 1 }));
      loadImages();
    } catch (err) {
      toast.error(t(language, "admin.imagesDeleteError"));
    } finally {
      setDeleting(false);
    }
  };

  const handleToggleIPVisibility = (uid: string) => {
    setShowFullIPs((prev) => {
      const next = new Set(prev);
      if (next.has(uid)) {
        next.delete(uid);
      } else {
        next.add(uid);
      }
      return next;
    });
  };

  const displayIP = (image: AdminImage) => {
    if (showFullIPs.has(image.uid)) {
      return image.ip_address || "-";
    }
    return image.ip_address_masked || image.ip_address || "-";
  };

  const handleBanIP = async (image: AdminImage) => {
    if (!token || !image.ip_address) return;
    const durationInput = window.prompt(t(language, "admin.ipBanDurationPrompt"), "24");
    if (durationInput === null) return;
    const durationHours = Number(durationInput.trim());
    if (!Number.isFinite(durationHours) || durationHours < 0) {
      toast.error(t(language, "admin.ipBanInvalidDuration"));
      return;
    }
    if (!window.confirm(t(language, "admin.ipBanConfirm", { ip: image.ip_address_masked || image.ip_address }))) return;
    setBanningUid(image.uid);
    try {
      const result = await adminCreateIPBan(token, {
        uid: image.uid,
        duration_hours: durationHours,
        reason: `Abusive upload from image ${image.uid}`,
      });
      toast.success(t(language, "admin.ipBanCreated"));
      const shouldDelete = result.affected_image_count > 0 && window.confirm(t(language, "admin.ipBanDeleteImagesConfirm", {
        count: result.affected_image_count,
        size: formatBytes(result.affected_total_size),
      }));
      if (shouldDelete) {
        const deleted = await adminDeleteIPBanImages(token, result.ban.id);
        toast.success(t(language, "admin.ipBanImagesDeleted", { count: deleted.deleted_count }));
        setSelectedUids(new Set());
        loadImages();
      }
    } catch (err) {
      toast.error(err instanceof Error ? err.message : t(language, "admin.ipBanCreateError"));
    } finally {
      setBanningUid(null);
    }
  };

  const lang = language;
  const totalPages = Math.ceil(total / pageSize);

  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between flex-wrap gap-2">
        <h1 className="text-xl font-bold">{t(lang, "admin.imagesTitle")}</h1>
        <div className="flex items-center gap-2">
          <Button
            variant={viewMode === "grid" ? "secondary" : "ghost"}
            size="sm"
            onClick={() => setViewMode("grid")}
            className="cursor-pointer"
          >
            <LayoutGrid className="h-4 w-4" />
            <span className="hidden sm:inline">{t(lang, "admin.imagesGridView")}</span>
          </Button>
          <Button
            variant={viewMode === "list" ? "secondary" : "ghost"}
            size="sm"
            onClick={() => setViewMode("list")}
            className="cursor-pointer"
          >
            <List className="h-4 w-4" />
            <span className="hidden sm:inline">{t(lang, "admin.imagesListView")}</span>
          </Button>
        </div>
      </div>

      {/* Toolbar */}
      <div className="flex items-center gap-2 flex-wrap">
        <div className="relative flex-1 min-w-[200px]">
          <Search className="absolute left-2 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
          <Input
            value={searchInput}
            onChange={(e) => handleSearchChange(e.target.value)}
            placeholder={t(lang, "admin.imagesSearch")}
            className="pl-8 h-8 text-sm"
          />
        </div>
        <Button variant="outline" size="sm" onClick={handleSelectAll} className="cursor-pointer">
          {selectedUids.size === images.length && images.length > 0 ? <CheckSquare className="h-3.5 w-3.5" /> : <Square className="h-3.5 w-3.5" />}
          {selectedUids.size === images.length ? t(lang, "admin.imagesDeselectAll") : t(lang, "admin.imagesSelectAll")}
        </Button>
        {selectedUids.size > 0 && (
          <Button variant="destructive" size="sm" onClick={handleBatchDelete} disabled={deleting} className="cursor-pointer">
            {deleting ? <Loader2 className="h-3.5 w-3.5 animate-spin" /> : <Trash2 className="h-3.5 w-3.5" />}
            {t(lang, "admin.imagesDelete")} ({selectedUids.size})
          </Button>
        )}
      </div>

      {/* Info */}
      <div className="flex items-center gap-2 text-xs text-muted-foreground">
        <span>{t(lang, "admin.imagesTotal", { total })}</span>
        {selectedUids.size > 0 && (
          <span>· {t(lang, "admin.imagesSelected", { count: selectedUids.size })}</span>
        )}
      </div>

      {/* Content */}
      {loading ? (
        <div className="flex items-center justify-center py-20">
          <Loader2 className="h-6 w-6 animate-spin text-muted-foreground" />
        </div>
      ) : error ? (
        <div className="flex items-center gap-2 text-destructive" role="alert">
          <AlertCircle className="h-5 w-5" />
          {error}
        </div>
      ) : viewMode === "grid" ? (
        <div className="grid grid-cols-2 sm:grid-cols-3 md:grid-cols-4 lg:grid-cols-5 gap-3">
          {images.map((img) => (
            <div key={img.uid} className="relative group">
              <ImgStyleImageCard
                src={getImageUrl(img.uid)}
                alt={img.uid}
                title={img.uid.slice(0, 8)}
                sizeLabel={formatBytes(img.size)}
                onPreview={() => setPreviewImage(img)}
                previewLabel={t(lang, "admin.imagesViewPreview")}
                selected={selectedUids.has(img.uid)}
                showCheckbox
                onSelect={(checked) => handleSelect(img.uid, checked)}
              />
            </div>
          ))}
        </div>
      ) : (
        <Card>
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead className="w-10">
                  <input
                    type="checkbox"
                    checked={selectedUids.size === images.length && images.length > 0}
                    onChange={handleSelectAll}
                    className="cursor-pointer"
                  />
                </TableHead>
                <TableHead className="w-16">{t(lang, "image.preview")}</TableHead>
                <TableHead>{t(lang, "image.uid")}</TableHead>
                <TableHead>{t(lang, "image.type")}</TableHead>
                <TableHead>{t(lang, "image.size")}</TableHead>
                <TableHead>{t(lang, "image.token")}</TableHead>
                <TableHead>{t(lang, "image.ip")}</TableHead>
                <TableHead>MD5</TableHead>
                <TableHead>{t(lang, "image.storageKey")}</TableHead>
                <TableHead>{t(lang, "image.storageBackend")}</TableHead>
                <TableHead>{t(lang, "image.created")}</TableHead>
                <TableHead>{t(lang, "admin.imagesActions")}</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {images.map((img) => (
                <TableRow
                  key={img.uid}
                  className="cursor-pointer"
                  onClick={() => setPreviewImage(img)}
                >
                  <TableCell onClick={(e) => e.stopPropagation()}>
                    <input
                      type="checkbox"
                      checked={selectedUids.has(img.uid)}
                      onChange={(e) => handleSelect(img.uid, e.target.checked)}
                      className="cursor-pointer"
                    />
                  </TableCell>
                  <TableCell>
                    <button
                      type="button"
                      onClick={(e) => {
                        e.stopPropagation();
                        setPreviewImage(img);
                      }}
                      className="block h-11 w-11 overflow-hidden rounded-md border bg-muted transition-colors hover:border-primary focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2"
                      aria-label={t(lang, "admin.imagesViewPreview")}
                    >
                      <img
                        src={getImageUrl(img.uid)}
                        alt={img.uid}
                        loading="lazy"
                        className="h-full w-full object-cover"
                      />
                    </button>
                  </TableCell>
                  <TableCell className="font-mono text-xs">{img.uid.slice(0, 12)}...</TableCell>
                  <TableCell className="text-xs">{img.mime_type}</TableCell>
                  <TableCell className="text-xs">{formatBytes(img.size)}</TableCell>
                  <TableCell className="font-mono text-xs">{img.token.slice(0, 8)}...</TableCell>
                  <TableCell className="font-mono text-xs whitespace-nowrap" onClick={(e) => e.stopPropagation()}>
                    <button
                      type="button"
                      onClick={() => handleToggleIPVisibility(img.uid)}
                      className="inline-flex items-center gap-1 rounded px-1 py-0.5 transition-colors hover:bg-muted focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring"
                    >
                      <span>{displayIP(img)}</span>
                      {showFullIPs.has(img.uid) ? <EyeOff className="h-3 w-3" /> : <Eye className="h-3 w-3" />}
                    </button>
                  </TableCell>
                  <TableCell className="font-mono text-xs">{img.md5_hash.slice(0, 8)}...</TableCell>
                  <TableCell className="text-xs">{img.storage_key}</TableCell>
                  <TableCell className="text-xs">{img.storage_backend}</TableCell>
                  <TableCell className="text-xs whitespace-nowrap">{formatDate(img.created_at)}</TableCell>
                  <TableCell onClick={(e) => e.stopPropagation()}>
                    <Button
                      variant="outline"
                      size="sm"
                      onClick={() => handleBanIP(img)}
                      disabled={banningUid === img.uid || !img.ip_address}
                      className="h-7 cursor-pointer text-xs"
                    >
                      {banningUid === img.uid ? <Loader2 className="h-3 w-3 animate-spin" /> : <Ban className="h-3 w-3" />}
                      {t(lang, "admin.ipBanAction")}
                    </Button>
                  </TableCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>
        </Card>
      )}

      {/* Pagination */}
      {totalPages > 1 && (
        <div className="flex items-center justify-center gap-2">
          <Button
            variant="outline"
            size="sm"
            disabled={page <= 1}
            onClick={() => setPage((p) => p - 1)}
            className="cursor-pointer"
          >
            <ChevronLeft className="h-4 w-4" />
            {t(lang, "admin.imagesPrev")}
          </Button>
          <span className="text-sm text-muted-foreground">
            {t(lang, "admin.imagesPage", { page })} / {totalPages}
          </span>
          <Button
            variant="outline"
            size="sm"
            disabled={page >= totalPages}
            onClick={() => setPage((p) => p + 1)}
            className="cursor-pointer"
          >
            {t(lang, "admin.imagesNext")}
            <ChevronRight className="h-4 w-4" />
          </Button>
        </div>
      )}

      {/* Preview */}
      <UploadHistoryLightbox
        open={!!previewImage}
        onClose={() => setPreviewImage(null)}
        selectedUid={previewImage?.uid ?? null}
        items={images.map((image) => ({ type: "admin", image }))}
        language={lang}
        metadataLabel={t(lang, "admin.imagesViewPreview")}
        onDeleteAdminImage={handleDeleteImage}
      />
    </div>
  );
}
