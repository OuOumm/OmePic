"use client";

/* eslint-disable @next/next/no-img-element */

import { useEffect, useId, useRef, type KeyboardEvent, type ReactNode } from "react";

import { Button } from "@/components/ui/Button";
import { cn } from "@/lib/utils";

export type ImageLightboxMetadata = {
  label: string;
  mono?: boolean;
  value: string;
};

export type ImageLightboxItem = {
  alt: string;
  metadata?: ImageLightboxMetadata[];
  src: string;
  subtitle?: string;
  title: string;
};

type ImageLightboxProps = {
  actions?: ReactNode;
  closeLabel: string;
  eyebrowLabel: string;
  image: ImageLightboxItem | null;
  metadataLabel: string;
  onClose: () => void;
};

const focusableSelector = [
  "a[href]",
  "button:not([disabled])",
  "input:not([disabled])",
  "select:not([disabled])",
  "textarea:not([disabled])",
  "[tabindex]:not([tabindex='-1'])"
].join(",");

export function ImageLightbox({
  actions,
  closeLabel,
  eyebrowLabel,
  image,
  metadataLabel,
  onClose
}: ImageLightboxProps) {
  const titleId = useId();
  const descriptionId = useId();
  const closeButtonRef = useRef<HTMLButtonElement>(null);
  const dialogRef = useRef<HTMLDivElement>(null);
  const previousFocusRef = useRef<HTMLElement | null>(null);
  const onCloseRef = useRef(onClose);
  const isOpen = Boolean(image);

  useEffect(() => {
    onCloseRef.current = onClose;
  }, [onClose]);

  useEffect(() => {
    if (!isOpen) {
      return;
    }

    previousFocusRef.current = document.activeElement instanceof HTMLElement
      ? document.activeElement
      : null;
    const previousOverflow = document.body.style.overflow;
    document.body.style.overflow = "hidden";
    window.setTimeout(() => closeButtonRef.current?.focus(), 0);

    function handleKeyDown(event: globalThis.KeyboardEvent) {
      if (event.key === "Escape") {
        event.preventDefault();
        onCloseRef.current();
      }
    }

    document.addEventListener("keydown", handleKeyDown);
    return () => {
      document.body.style.overflow = previousOverflow;
      document.removeEventListener("keydown", handleKeyDown);
      previousFocusRef.current?.focus();
      previousFocusRef.current = null;
    };
  }, [isOpen]);

  if (!image) {
    return null;
  }

  function handleDialogKeyDown(event: KeyboardEvent<HTMLDivElement>) {
    if (event.key !== "Tab" || !dialogRef.current) {
      return;
    }

    const focusable = Array.from(
      dialogRef.current.querySelectorAll<HTMLElement>(focusableSelector)
    ).filter((element) => !element.hasAttribute("disabled") && element.tabIndex !== -1);

    if (focusable.length === 0) {
      event.preventDefault();
      return;
    }

    const first = focusable[0];
    const last = focusable[focusable.length - 1];

    if (event.shiftKey && document.activeElement === first) {
      event.preventDefault();
      last.focus();
      return;
    }

    if (!event.shiftKey && document.activeElement === last) {
      event.preventDefault();
      first.focus();
    }
  }

  return (
    <div
      aria-describedby={descriptionId}
      aria-labelledby={titleId}
      aria-modal="true"
      className="fixed inset-0 z-[80] flex items-center justify-center bg-slate-950/84 p-3 backdrop-blur-2xl animate-fade-in sm:p-6"
      onKeyDown={handleDialogKeyDown}
      onMouseDown={(event) => {
        if (event.target === event.currentTarget) {
          onClose();
        }
      }}
      ref={dialogRef}
      role="dialog"
    >
      <div className="grid max-h-[92vh] w-full max-w-6xl overflow-hidden rounded-[32px] border border-white/15 bg-white/80 shadow-[0_30px_100px_rgba(2,6,23,0.45)] backdrop-blur-2xl dark:bg-slate-950/90 lg:grid-cols-[minmax(0,1fr)_340px]">
        <div className="flex min-h-[280px] items-center justify-center bg-[radial-gradient(circle_at_top,rgba(139,92,246,0.18),transparent_30%),linear-gradient(180deg,rgba(2,6,23,0.98),rgba(15,23,42,0.98))] p-3 sm:p-6">
          <img
            alt={image.alt}
            className="max-h-[72vh] w-full rounded-2xl object-contain shadow-2xl"
            decoding="async"
            src={image.src}
          />
        </div>
        <aside className="flex min-h-0 flex-col border-t border-white/20 bg-white/70 backdrop-blur-xl dark:border-white/10 dark:bg-slate-900/70 lg:border-l lg:border-t-0">
          <div className="flex items-start justify-between gap-3 border-b border-white/40 p-4 sm:p-5 dark:border-white/10">
            <div className="min-w-0 space-y-1">
              <p className="text-xs font-bold uppercase tracking-[0.22em] text-violet-600 dark:text-violet-300">
                {eyebrowLabel}
              </p>
              <h2 className="truncate text-lg font-bold text-slate-900 dark:text-white" id={titleId}>
                {image.title}
              </h2>
              {image.subtitle ? (
                <p className="break-all font-mono text-xs text-muted" id={descriptionId}>
                  {image.subtitle}
                </p>
              ) : (
                <p className="sr-only" id={descriptionId}>
                  {image.title}
                </p>
              )}
            </div>
            <Button
              aria-label={closeLabel}
              onClick={onClose}
              ref={closeButtonRef}
              size="icon"
              variant="secondary"
            >
              <CloseIcon />
            </Button>
          </div>

          {image.metadata?.length ? (
            <div className="overflow-y-auto p-4 sm:p-5">
              <p className="sr-only">{metadataLabel}</p>
              <dl className="grid gap-3 text-sm">
                {image.metadata.map((item) => (
                  <div
                    className="rounded-2xl border border-white/40 bg-white/55 p-3 backdrop-blur-xl dark:border-white/10 dark:bg-slate-950/40"
                    key={`${item.label}-${item.value}`}
                  >
                    <dt className="text-xs font-semibold uppercase tracking-wide text-muted">{item.label}</dt>
                    <dd
                      className={cn(
                        "mt-1 break-all text-slate-800 dark:text-slate-100",
                        item.mono ? "font-mono text-xs" : "text-sm"
                      )}
                    >
                      {item.value}
                    </dd>
                  </div>
                ))}
              </dl>
            </div>
          ) : null}

          {actions ? (
            <div className="mt-auto flex flex-wrap gap-2 border-t border-white/40 p-4 sm:p-5 dark:border-white/10">
              {actions}
            </div>
          ) : null}
        </aside>
      </div>
    </div>
  );
}

function CloseIcon() {
  return (
    <svg aria-hidden="true" className="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2.2}>
      <path d="M6 6l12 12M18 6 6 18" strokeLinecap="round" strokeLinejoin="round" />
    </svg>
  );
}
