"use client";

import { useCallback, useRef } from "react";
import { cn } from "@/lib/utils";
import { t } from "@/lib/i18n";
import { Upload, Image } from "lucide-react";
import type { Language } from "@/types";

interface UploadDropzoneProps {
  disabled?: boolean;
  isDragging: boolean;
  onDragStateChange: (value: boolean) => void;
  onSelectFiles: (files: File[]) => void;
  fileInputRef: React.RefObject<HTMLInputElement | null>;
  language: Language;
}

const ACCEPTED_TYPES = ".avif,.png,.jpg,.jpeg,.gif,.webp,.bmp";

export function UploadDropzone({
  disabled,
  isDragging,
  onDragStateChange,
  onSelectFiles,
  fileInputRef,
  language,
}: UploadDropzoneProps) {
  const dragDepth = useRef(0);

  const handleFiles = useCallback(
    (files: File[]) => {
      if (files.length > 0) onSelectFiles(files);
    },
    [onSelectFiles]
  );

  const onDragEnter = useCallback(
    (e: React.DragEvent) => {
      e.preventDefault();
      e.stopPropagation();
      dragDepth.current++;
      if (dragDepth.current === 1) onDragStateChange(true);
    },
    [onDragStateChange]
  );

  const onDragLeave = useCallback(
    (e: React.DragEvent) => {
      e.preventDefault();
      e.stopPropagation();
      dragDepth.current--;
      if (dragDepth.current <= 0) {
        dragDepth.current = 0;
        onDragStateChange(false);
      }
    },
    [onDragStateChange]
  );

  const onDragOver = useCallback((e: React.DragEvent) => {
    e.preventDefault();
    e.stopPropagation();
  }, []);

  const onDrop = useCallback(
    (e: React.DragEvent) => {
      e.preventDefault();
      e.stopPropagation();
      dragDepth.current = 0;
      onDragStateChange(false);
      if (disabled) return;
      const fileList = e.dataTransfer.files;
      if (fileList && fileList.length > 0) {
        handleFiles(Array.from(fileList));
      }
    },
    [disabled, handleFiles, onDragStateChange]
  );

  const onPaste = useCallback(
    (e: React.ClipboardEvent) => {
      const items = e.clipboardData?.items;
      if (!items) return;
      for (let i = 0; i < items.length; i++) {
        if (items[i].type.startsWith("image/")) {
          e.preventDefault();
          const file = items[i].getAsFile();
          if (file) handleFiles([file]);
          return;
        }
      }
      // No image in clipboard: do not prevent default, let event propagate normally
    },
    [handleFiles]
  );

  return (
    <div
      className={cn(
        "relative flex flex-col items-center justify-center rounded-lg border-2 border-dashed p-8 transition-colors",
        isDragging ? "border-primary bg-primary/5" : "border-muted-foreground/25 hover:border-primary/50",
        disabled && "opacity-50 pointer-events-none"
      )}
      onDragEnter={onDragEnter}
      onDragLeave={onDragLeave}
      onDragOver={onDragOver}
      onDrop={onDrop}
      onPaste={onPaste}
      tabIndex={0}
      role="button"
      aria-label={t(language, "upload.dropzone")}
      data-paste-managed="true"
    >
      <input
        ref={fileInputRef}
        type="file"
        accept={ACCEPTED_TYPES}
        multiple
        className="absolute inset-0 w-full h-full opacity-0 cursor-pointer"
        onChange={(e) => {
          const fileList = e.target.files;
          if (fileList && fileList.length > 0) {
            handleFiles(Array.from(fileList));
          }
          e.target.value = "";
        }}
        disabled={disabled}
        aria-label={t(language, "upload.chooseFile")}
      />
      <div className="flex flex-col items-center gap-2 text-center pointer-events-none">
        {isDragging ? (
          <Image className="h-10 w-10 text-primary" />
        ) : (
          <Upload className="h-10 w-10 text-muted-foreground" />
        )}
        <p className="text-sm font-medium">
          {isDragging ? t(language, "upload.pasting") : t(language, "upload.dropzone")}
        </p>
        <p className="text-xs text-muted-foreground">{t(language, "upload.supportedFormats")}</p>
        <p className="text-xs text-muted-foreground/70">{t(language, "upload.pasteHint")}</p>
      </div>
    </div>
  );
}
