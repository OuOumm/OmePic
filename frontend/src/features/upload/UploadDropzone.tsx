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
};

export function UploadDropzone({
  disabled,
  isDragging,
  onDragStateChange,
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
        "group relative overflow-hidden rounded-lg border border-dashed p-8 text-center transition-colors sm:p-10 lg:p-12",
        disabled
          ? "cursor-not-allowed border-border bg-muted/30 opacity-70"
          : "cursor-pointer border-border bg-background hover:border-primary hover:bg-muted/40",
        isDragging &&
          "border-primary bg-muted"
      )}
      role="group"
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
    >
      <div className="relative mx-auto flex max-w-2xl flex-col items-center gap-5">
        <div className="flex h-16 w-16 items-center justify-center rounded-lg border border-border bg-muted text-muted-foreground">
          <UploadCloud aria-hidden="true" className="h-8 w-8" />
        </div>
        <div className="space-y-2">
          <h2 className="text-2xl font-semibold tracking-tight text-foreground" id={titleId}>
            {t.upload.dropTitle}
          </h2>
          <p className="mx-auto max-w-2xl text-sm leading-6 text-muted-foreground" id={descriptionId}>
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
          {["PNG", "JPG", "GIF", "WEBP", "BMP"].map((format) => (
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
        accept=".png,.jpg,.jpeg,.gif,.webp,.bmp,image/png,image/jpeg,image/gif,image/webp,image/bmp"
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
