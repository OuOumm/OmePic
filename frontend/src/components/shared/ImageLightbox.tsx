"use client";

import { useEffect, useCallback, useState, useRef } from "react";
import { Dialog, DialogContent } from "@/components/ui/Dialog";
import { Button } from "@/components/ui/Button";
import { X, ChevronLeft, ChevronRight } from "lucide-react";
import { cn } from "@/lib/utils";

interface ImageLightboxAction {
  label: string;
  onClick: () => void;
}

interface ImageLightboxImage {
  url: string;
  alt?: string;
  metadata?: { label: string; value: string }[];
}

export function ImageLightbox({
  images,
  initialIndex = 0,
  open,
  onClose,
  getActions,
  closeLabel,
  metadataLabel,
}: {
  images: ImageLightboxImage[];
  initialIndex?: number;
  open: boolean;
  onClose: () => void;
  getActions?: (image: ImageLightboxImage, index: number) => ImageLightboxAction[];
  closeLabel: string;
  metadataLabel?: string;
}) {
  const [index, setIndex] = useState(initialIndex);
  const stripRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    if (open) setIndex(initialIndex);
  }, [open, initialIndex]);

  const total = images.length;
  const hasMultiple = total > 1;

  const goPrev = useCallback(() => {
    setIndex((i) => (i > 0 ? i - 1 : total - 1));
  }, [total]);

  const goNext = useCallback(() => {
    setIndex((i) => (i < total - 1 ? i + 1 : 0));
  }, [total]);

  // Auto-scroll filmstrip
  useEffect(() => {
    const strip = stripRef.current;
    if (!strip || !hasMultiple) return;
    const thumb = strip.children[index] as HTMLElement | undefined;
    if (thumb) {
      thumb.scrollIntoView({ behavior: "smooth", block: "nearest", inline: "center" });
    }
  }, [index, hasMultiple]);

  // Keyboard
  useEffect(() => {
    if (!open) return;
    const handler = (e: KeyboardEvent) => {
      if (e.key === "Escape") { onClose(); return; }
      if (e.key === "ArrowLeft") { e.preventDefault(); goPrev(); return; }
      if (e.key === "ArrowRight") { e.preventDefault(); goNext(); return; }
    };
    document.addEventListener("keydown", handler);
    return () => document.removeEventListener("keydown", handler);
  }, [open, onClose, goPrev, goNext]);

  const current = images[index];
  if (!current) return null;

  const actions = getActions ? getActions(current, index) : [];

  return (
    <Dialog open={open} onOpenChange={(o) => { if (!o) onClose(); }}>
      <DialogContent
        className="max-w-[95vw] w-[95vw] h-[90vh] max-h-[95vh] p-0 gap-0 flex flex-col overflow-hidden"
        role="dialog"
        aria-modal="true"
        aria-label={metadataLabel ?? "Preview"}
      >
        {/* ====== Top bar ====== */}
        <div className="shrink-0 flex items-center justify-between px-4 py-3 border-b border-white/10">
          <h3 className="font-semibold text-sm select-none">
            {metadataLabel ?? "Preview"}
            {hasMultiple && (
              <span className="text-muted-foreground font-normal ml-2">
                {index + 1} / {total}
              </span>
            )}
          </h3>
          <Button
            variant="ghost"
            size="icon"
            onClick={onClose}
            aria-label={closeLabel}
            className="cursor-pointer"
          >
            <X className="h-4 w-4" />
          </Button>
        </div>

        {/* ====== Image area ====== */}
        <div className="flex-1 min-h-[40vh] relative bg-black/20 dark:bg-white/[0.02]">
          {/* Invisible click zones */}
          {hasMultiple && (
            <>
              <button type="button" onClick={goPrev} aria-label="Previous image"
                className="absolute inset-y-0 left-0 w-1/2 z-20 cursor-pointer" />
              <button type="button" onClick={goNext} aria-label="Next image"
                className="absolute inset-y-0 right-0 w-1/2 z-20 cursor-pointer" />
            </>
          )}

          {/* Visible arrow overlays */}
          {hasMultiple && (
            <>
              <button type="button" onClick={goPrev} aria-hidden="true" tabIndex={-1}
                className="absolute left-3 top-1/2 -translate-y-1/2 z-30 w-10 h-10 rounded-full flex items-center justify-center bg-black/40 hover:bg-black/60 backdrop-blur transition-all duration-200 shadow-lg">
                <ChevronLeft className="h-5 w-5 text-white" />
              </button>
              <button type="button" onClick={goNext} aria-hidden="true" tabIndex={-1}
                className="absolute right-3 top-1/2 -translate-y-1/2 z-30 w-10 h-10 rounded-full flex items-center justify-center bg-black/40 hover:bg-black/60 backdrop-blur transition-all duration-200 shadow-lg">
                <ChevronRight className="h-5 w-5 text-white" />
              </button>
            </>
          )}

          {/* Image */}
          <div className="absolute inset-0 flex items-center justify-center p-6 pointer-events-none">
            <img
              src={current.url}
              alt={current.alt ?? ""}
              className="max-h-full max-w-full object-contain rounded select-none"
              draggable={false}
            />
          </div>
        </div>

        {/* ====== Thumbnail filmstrip ====== */}
        {hasMultiple && (
          <div className="shrink-0 border-t border-white/10 bg-black/10 dark:bg-white/[0.02] px-2 py-2 flex items-center gap-1">
            <button type="button"
              onClick={() => { const s = stripRef.current; if (s) s.scrollBy({ left: -120, behavior: "smooth" }); }}
              className="shrink-0 w-7 h-12 flex items-center justify-center rounded cursor-pointer hover:bg-white/10 transition-colors"
              aria-label="Scroll thumbnails left">
              <ChevronLeft className="h-4 w-4 text-muted-foreground" />
            </button>

            <div ref={stripRef}
              className="flex-1 flex gap-2 overflow-x-auto"
              style={{ scrollbarWidth: "none" }}>
              {images.map((img, i) => (
                <button key={i} type="button" onClick={() => setIndex(i)}
                  className={cn(
                    "shrink-0 h-12 w-12 rounded-md overflow-hidden border-2 cursor-pointer transition-all duration-200",
                    i === index
                      ? "border-primary ring-2 ring-primary/30 scale-105"
                      : "border-transparent opacity-60 hover:opacity-100 hover:border-white/30"
                  )}
                  aria-label={`Image ${i + 1}`}
                  aria-current={i === index ? "true" : undefined}>
                  <img src={img.url} alt="" className="h-full w-full object-cover" loading="lazy" draggable={false} />
                </button>
              ))}
            </div>

            <button type="button"
              onClick={() => { const s = stripRef.current; if (s) s.scrollBy({ left: 120, behavior: "smooth" }); }}
              className="shrink-0 w-7 h-12 flex items-center justify-center rounded cursor-pointer hover:bg-white/10 transition-colors"
              aria-label="Scroll thumbnails right">
              <ChevronRight className="h-4 w-4 text-muted-foreground" />
            </button>
          </div>
        )}

        {/* ====== Info & actions ====== */}
        {(current.metadata && current.metadata.length > 0) || actions.length > 0 ? (
          <div className="shrink-0 border-t border-white/10 px-4 py-3 space-y-2">
            {current.metadata && current.metadata.length > 0 && (
              <div className="flex flex-wrap gap-x-5 gap-y-1 text-sm">
                {current.metadata.map((m) => (
                  <span key={m.label}>
                    <span className="text-muted-foreground">{m.label}: </span>
                    <span className="font-mono text-xs break-all">{m.value}</span>
                  </span>
                ))}
              </div>
            )}
            {actions.length > 0 && (
              <div className="flex gap-2 flex-wrap">
                {actions.map((a) => (
                  <Button key={a.label} variant="outline" size="sm" onClick={a.onClick} className="cursor-pointer">
                    {a.label}
                  </Button>
                ))}
              </div>
            )}
          </div>
        ) : null}
      </DialogContent>
    </Dialog>
  );
}
