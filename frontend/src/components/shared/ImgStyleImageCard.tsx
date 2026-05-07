"use client";

import { useState, useRef, useCallback } from "react";
import { cn } from "@/lib/utils";
import { ZoomIn, Loader2, XCircle, Check } from "lucide-react";
import { Button } from "@/components/ui/Button";

export function ImgStyleImageCard({
  src,
  alt,
  title,
  sizeLabel,
  onPreview,
  previewLabel,
  selected,
  topLeft,
  showCheckbox,
  onSelect,
  uploadStatus,
  uploadProgress = 0,
  filename,
  actionButtons,
}: {
  src?: string;
  alt?: string;
  title?: string;
  sizeLabel?: string;
  onPreview?: () => void;
  previewLabel?: string;
  selected?: boolean;
  topLeft?: React.ReactNode;
  showCheckbox?: boolean;
  onSelect?: (checked: boolean) => void;
  uploadStatus?: "pending" | "uploading" | "success" | "error";
  uploadProgress?: number;
  filename?: string;
  actionButtons?: { label: string; onClick: () => void }[];
}) {
  const [clickedIndex, setClickedIndex] = useState<number | null>(null);
  const timerRef = useRef<ReturnType<typeof setTimeout> | null>(null);

  const handleActionClick = useCallback((idx: number, onClick: () => void) => {
    onClick();
    setClickedIndex(idx);
    if (timerRef.current) clearTimeout(timerRef.current);
    timerRef.current = setTimeout(() => setClickedIndex(null), 1000);
  }, []);

  const isUploading = uploadStatus === "pending" || uploadStatus === "uploading";
  const isError = uploadStatus === "error";
  const showImage = !isUploading && !isError && src;
  const isTriZone = showImage && actionButtons && actionButtons.length > 0;

  // 3-zone flex layout: success/normal cards with action buttons
  if (isTriZone) {
    return (
      <div
        className={cn(
          "flex flex-col rounded-lg border bg-card overflow-hidden cursor-pointer",
          selected && "ring-2 ring-primary"
        )}
      >
        {/* Zone 1: Image (aspect-square, w-full) */}
        <div className="group relative aspect-square w-full">
          <img
            src={src}
            alt={alt ?? ""}
            className="h-full w-full object-cover"
            loading="lazy"
          />

          {/* Preview overlay */}
          {onPreview && (
            <div
              className="absolute inset-0 bg-black/0 group-hover:bg-black/20 transition-colors flex items-center justify-center"
              onClick={onPreview}
              role="button"
              aria-label={previewLabel}
            >
              <ZoomIn className="h-8 w-8 text-white opacity-0 group-hover:opacity-100 transition-opacity" />
            </div>
          )}

          {/* Checkbox (admin grid) */}
          {showCheckbox && onSelect && (
            <div className="absolute top-2 left-2" onClick={(e) => e.stopPropagation()}>
              <input
                type="checkbox"
                checked={selected}
                onChange={(e) => onSelect(e.target.checked)}
                className="h-4 w-4 cursor-pointer"
                aria-label={`Select ${alt}`}
              />
            </div>
          )}

          {topLeft && <div className="absolute top-2 right-2">{topLeft}</div>}
        </div>

        {/* Zone 2: Button Row */}
        <div className="flex items-center justify-center gap-1.5 px-2 py-1.5 bg-muted/30">
          {actionButtons.map((btn, idx) => (
            <Button
              key={btn.label}
              type="button"
              size="sm"
              variant="outline"
              className="h-7 min-w-[36px] text-xs"
              onClick={(e) => {
                e.stopPropagation();
                handleActionClick(idx, btn.onClick);
              }}
            >
              {clickedIndex === idx ? (
                <Check className="h-3 w-3" />
              ) : (
                btn.label
              )}
            </Button>
          ))}
        </div>

        {/* Zone 3: Metadata */}
        <div className="px-3 pb-2 pt-1 space-y-0.5">
          {(filename || title) && (
            <p className="text-xs text-foreground truncate font-medium">
              {title || filename}
            </p>
          )}
          {sizeLabel && (
            <p className="text-xs text-muted-foreground">{sizeLabel}</p>
          )}
        </div>
      </div>
    );
  }

  // Original layout: uploading, error, or normal without actionButtons
  return (
    <div
      className={cn(
        "group relative aspect-square rounded-lg overflow-hidden border bg-muted transition-colors",
        isUploading && "border-primary/30",
        isError && "border-destructive/30 bg-destructive/10",
        !isUploading && !isError && "cursor-pointer",
        selected && "ring-2 ring-primary"
      )}
    >
      {/* Uploading overlay */}
      {isUploading && (
        <div className="absolute inset-0 flex flex-col items-center justify-center">
          <Loader2 className="h-8 w-8 animate-spin text-muted-foreground" />
          <div className="mt-3 w-4/5 h-1.5 bg-muted-foreground/20 rounded-full overflow-hidden">
            <div
              className="h-full bg-primary rounded-full transition-all duration-300"
              style={{ width: `${uploadProgress}%` }}
            />
          </div>
          <p className="text-xs text-foreground mt-1">{uploadProgress}%</p>
        </div>
      )}

      {/* Error overlay */}
      {isError && (
        <div className="absolute inset-0 flex flex-col items-center justify-center">
          <XCircle className="h-8 w-8 text-destructive" />
          <p className="text-xs text-destructive mt-1">Upload failed</p>
        </div>
      )}

      {/* Image */}
      {showImage && (
        <img
          src={src}
          alt={alt ?? ""}
          className="h-full w-full object-cover"
          loading="lazy"
        />
      )}

      {/* Preview overlay */}
      {showImage && onPreview && (
        <div
          className="absolute inset-0 bg-black/0 group-hover:bg-black/20 transition-colors flex items-center justify-center"
          onClick={onPreview}
          role="button"
          aria-label={previewLabel}
        >
          <ZoomIn className="h-8 w-8 text-white opacity-0 group-hover:opacity-100 transition-opacity" />
        </div>
      )}

      {/* Checkbox (admin grid) */}
      {showCheckbox && onSelect && (
        <div className="absolute top-2 left-2" onClick={(e) => e.stopPropagation()}>
          <input
            type="checkbox"
            checked={selected}
            onChange={(e) => onSelect(e.target.checked)}
            className="h-4 w-4 cursor-pointer"
            aria-label={`Select ${alt}`}
          />
        </div>
      )}

      {topLeft && <div className="absolute top-2 right-2">{topLeft}</div>}

      {/* Bottom strip -- uploading */}
      {isUploading && (
        <div className="absolute bottom-0 left-0 right-0 bg-gradient-to-t from-black/60 to-transparent p-2">
          <p className="text-white text-xs truncate">
            {filename} {uploadProgress}%
          </p>
        </div>
      )}

      {/* Bottom strip -- error */}
      {isError && (
        <div className="absolute bottom-0 left-0 right-0 bg-gradient-to-t from-black/60 to-transparent p-2">
          <p className="text-white text-xs truncate">{filename}</p>
        </div>
      )}

      {/* Bottom strip -- success/normal (no actionButtons in this layout) */}
      {!isUploading && !isError && (title || sizeLabel) && (
        <div className="absolute bottom-0 left-0 right-0 bg-gradient-to-t from-black/60 to-transparent p-2">
          {title && (
            <p className="text-white text-xs truncate font-medium">{title}</p>
          )}
          {sizeLabel && (
            <p className="text-white/70 text-[10px]">{sizeLabel}</p>
          )}
        </div>
      )}
    </div>
  );
}
