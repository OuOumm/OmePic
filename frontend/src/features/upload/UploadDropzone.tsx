"use client";

import { useEffect, useRef } from "react";

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
        "group relative overflow-hidden rounded-3xl border-2 border-dashed p-8 text-center shadow-panel backdrop-blur-xl transition-all duration-300 sm:p-10 lg:p-12",
        disabled
          ? "cursor-not-allowed border-slate-300/60 bg-white/50 opacity-70 dark:border-slate-700/60 dark:bg-slate-900/40"
          : "cursor-pointer border-slate-300/80 bg-white/50 hover:-translate-y-0.5 hover:border-violet-400 hover:bg-violet-50/60 hover:shadow-glow dark:border-slate-600/80 dark:bg-slate-900/40 dark:hover:border-violet-500 dark:hover:bg-violet-500/5",
        isDragging &&
          "animate-pulse-glow border-violet-400 bg-violet-50/70 shadow-glow dark:border-violet-400 dark:bg-violet-500/10"
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
      <div className="absolute inset-0 rounded-3xl bg-[radial-gradient(circle_at_top,rgba(139,92,246,0.12),transparent_34%),linear-gradient(135deg,rgba(255,255,255,0.16),rgba(255,255,255,0.02))] transition-all duration-500 group-hover:from-violet-400/10 group-hover:to-cyan-400/10" />
      <div className="relative mx-auto flex max-w-2xl flex-col items-center gap-5">
        <div className="flex h-20 w-20 items-center justify-center rounded-[26px] bg-gradient-to-br from-violet-100 via-white to-cyan-100 text-violet-500 shadow-[0_18px_48px_rgba(139,92,246,0.16)] transition-transform duration-300 group-hover:scale-110 dark:from-violet-500/20 dark:via-slate-900 dark:to-cyan-500/20 dark:text-violet-300 dark:shadow-[0_18px_48px_rgba(76,29,149,0.34)]">
          <UploadCloudIcon />
        </div>
        <div className="space-y-2">
          <h2 className="text-2xl font-bold tracking-tight text-slate-800 dark:text-slate-100" id={titleId}>
            {t.upload.dropTitle}
          </h2>
          <p className="mx-auto max-w-2xl text-sm leading-6 text-slate-500 dark:text-slate-400" id={descriptionId}>
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
          <UploadArrowIcon />
          {t.upload.chooseFile}
        </Button>
        <div className="flex flex-wrap items-center justify-center gap-2 text-[11px] font-semibold uppercase tracking-[0.16em] text-muted">
          {["PNG", "JPG", "GIF", "WEBP", "BMP", "AVIF"].map((format) => (
            <span
              className="rounded-full border border-white/50 bg-white/55 px-3 py-1 backdrop-blur-md dark:border-white/10 dark:bg-slate-950/45"
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

function UploadCloudIcon() {
  return (
    <svg aria-hidden="true" className="h-10 w-10" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={1.8}>
      <path d="M7 16.5a4.5 4.5 0 0 1 .6-8.96A5.5 5.5 0 0 1 18 9.5a3.5 3.5 0 0 1-.5 6.96M12 12v8m0-8-3 3m3-3 3 3" strokeLinecap="round" strokeLinejoin="round" />
    </svg>
  );
}

function UploadArrowIcon() {
  return (
    <svg aria-hidden="true" className="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2.3}>
      <path d="M12 16V4m0 0L8 8m4-4 4 4M5 16v2a2 2 0 0 0 2 2h10a2 2 0 0 0 2-2v-2" strokeLinecap="round" strokeLinejoin="round" />
    </svg>
  );
}
