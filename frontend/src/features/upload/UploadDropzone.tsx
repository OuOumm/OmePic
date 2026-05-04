"use client";

import { useEffect, useRef } from "react";
import { UploadCloud } from "lucide-react";

import { Button } from "@/components/ui/Button";
import { useUiTranslations } from "@/hooks/useUiPreferences";
import { cn } from "@/lib/utils";

type UploadDropzoneProps = {
  disabled?: boolean;
  isDragging: boolean;
  onDragStateChange: (value: boolean) => void;
  onSelectFile: (file: File) => void;
  onPasteImage?: React.ClipboardEventHandler<HTMLElement>;
};

export function UploadDropzone({
  disabled,
  isDragging,
  onDragStateChange,
  onPasteImage,
  onSelectFile
}: UploadDropzoneProps) {
  const inputRef = useRef<HTMLInputElement | null>(null);
  const dragDepthRef = useRef(0);
  const t = useUiTranslations();
  const titleId = "upload-dropzone-title";
  const descriptionId = "upload-dropzone-description";

  useEffect(() => {
    if (!disabled) {
      return;
    }
    dragDepthRef.current = 0;
    onDragStateChange(false);
  }, [disabled, onDragStateChange]);

  function handleFiles(files: FileList | null) {
    if (disabled) {
      return;
    }
    const file = files?.[0];
    if (!file) {
      return;
    }
    onSelectFile(file);
  }

  return (
    <section
      aria-describedby={descriptionId}
      aria-disabled={disabled}
      aria-labelledby={titleId}
      className={cn(
        "group relative flex min-h-[min(42rem,calc(100vh-10rem))] overflow-hidden rounded-lg border border-dashed p-8 text-center shadow-sm transition-colors sm:p-10 lg:p-12",
        disabled
          ? "cursor-not-allowed border-border bg-muted/30 opacity-70"
          : "cursor-pointer border-border bg-card hover:border-primary hover:bg-muted/30",
        isDragging &&
          "border-primary bg-muted"
      )}
      role="group"
      tabIndex={disabled ? undefined : 0}
      onClick={() => {
        if (!disabled) {
          inputRef.current?.click();
        }
      }}
      onDragEnter={(event) => {
        event.preventDefault();
        if (!disabled) {
          dragDepthRef.current += 1;
          onDragStateChange(true);
        }
      }}
      onDragLeave={(event) => {
        event.preventDefault();
        if (disabled) {
          return;
        }
        dragDepthRef.current = Math.max(0, dragDepthRef.current - 1);
        if (dragDepthRef.current === 0) {
          onDragStateChange(false);
        }
      }}
      onDragOver={(event) => event.preventDefault()}
      onDrop={(event) => {
        event.preventDefault();
        dragDepthRef.current = 0;
        onDragStateChange(false);
        handleFiles(event.dataTransfer.files);
      }}
      onPaste={onPasteImage}
    >
      <div className="pointer-events-none absolute inset-0 bg-[radial-gradient(circle_at_top,rgba(99,102,241,0.16),transparent_42%),radial-gradient(circle_at_bottom_right,rgba(6,182,212,0.14),transparent_38%)]" aria-hidden="true" />
      <div className="pointer-events-none absolute inset-x-8 top-8 h-px bg-gradient-to-r from-transparent via-primary/35 to-transparent" aria-hidden="true" />
      <div className="relative mx-auto flex max-w-3xl flex-1 flex-col items-center justify-center gap-6">
        <div className="flex h-20 w-20 items-center justify-center rounded-lg border border-border bg-background text-primary shadow-sm">
          <UploadCloud aria-hidden="true" className="h-10 w-10" />
        </div>
        <div className="space-y-2">
          <h1 className="text-3xl font-semibold tracking-tight text-foreground sm:text-4xl" id={titleId}>
            {t.upload.dropTitle}
          </h1>
          <p className="mx-auto max-w-2xl text-base leading-7 text-muted-foreground" id={descriptionId}>
            {t.upload.dropDescription}
          </p>
        </div>
        <Button
          aria-describedby={descriptionId}
          disabled={disabled}
          onClick={(event) => {
            event.stopPropagation();
            inputRef.current?.click();
          }}
        >
          <UploadCloud aria-hidden="true" className="h-4 w-4" />
          {t.upload.chooseFile}
        </Button>
        <div className="flex flex-wrap items-center justify-center gap-2 text-[11px] font-medium uppercase tracking-wider text-muted-foreground">
          {["AVIF", "PNG", "JPG", "GIF", "WEBP", "BMP"].map((format) => (
            <span
              className="rounded-md border border-border bg-muted px-2.5 py-1"
              key={format}
            >
              {format}
            </span>
          ))}
        </div>
      </div>
      <input
        ref={inputRef}
        accept=".avif,.png,.jpg,.jpeg,.gif,.webp,.bmp,image/avif,image/png,image/jpeg,image/gif,image/webp,image/bmp"
        aria-describedby={descriptionId}
        aria-label={t.upload.fileInputLabel}
        className="sr-only"
        disabled={disabled}
        onChange={(event) => handleFiles(event.target.files)}
        type="file"
      />
    </section>
  );
}
