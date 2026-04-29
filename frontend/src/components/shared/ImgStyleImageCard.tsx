/* eslint-disable @next/next/no-img-element */

import type { CSSProperties, ReactNode } from "react";

import { cn } from "@/lib/utils";

type ImgStyleImageCardProps = {
  alt: string;
  animationDelay?: string;
  className?: string;
  onPreview?: () => void;
  previewLabel?: string;
  selected?: boolean;
  sizeLabel: string;
  src: string;
  title: string;
  topLeft?: ReactNode;
};

export function ImgStyleImageCard({
  alt,
  animationDelay,
  className,
  onPreview,
  previewLabel,
  selected = false,
  sizeLabel,
  src,
  title,
  topLeft
}: ImgStyleImageCardProps) {
  const style: CSSProperties | undefined = animationDelay
    ? { animationDelay }
    : undefined;

  return (
    <div
      className={cn(
        "group relative rounded-2xl overflow-hidden bg-white dark:bg-slate-800/60 border border-slate-200/60 dark:border-slate-700/40 shadow-sm hover:shadow-xl dark:hover:shadow-violet-500/10 transition-all duration-300 hover:-translate-y-1 cursor-pointer animate-fade-in",
        selected ? "border-violet-400/70 shadow-glow ring-2 ring-violet-400/25 dark:border-violet-300/70" : null,
        className
      )}
      style={style}
    >
      {onPreview ? (
        <button
          aria-label={previewLabel || title}
          className="block w-full text-left focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-violet-400 focus-visible:ring-offset-2 focus-visible:ring-offset-surface"
          onClick={onPreview}
          type="button"
        >
          <ImageCardVisual alt={alt} sizeLabel={sizeLabel} src={src} title={title} />
        </button>
      ) : (
        <ImageCardVisual alt={alt} sizeLabel={sizeLabel} src={src} title={title} />
      )}
      {topLeft ? <div className="absolute left-3 top-3 z-20">{topLeft}</div> : null}
    </div>
  );
}

function ImageCardVisual({
  alt,
  sizeLabel,
  src,
  title
}: {
  alt: string;
  sizeLabel: string;
  src: string;
  title: string;
}) {
  return (
    <>
      <div className="aspect-square overflow-hidden relative">
        <img
          alt={alt}
          className="w-full h-full object-cover transition-transform duration-[400ms] group-hover:scale-105"
          decoding="async"
          loading="lazy"
          src={src}
        />
        <div className="absolute inset-0 bg-black/0 group-hover:bg-black/40 transition-all duration-300 flex items-center justify-center">
          <div className="opacity-0 group-hover:opacity-100 transition-all duration-300 transform scale-75 group-hover:scale-100">
            <MagnifyIcon />
          </div>
        </div>
      </div>
      <div className="absolute bottom-0 left-0 right-0 px-3 py-2 bg-white/60 dark:bg-slate-900/60 backdrop-blur-md border-t border-white/20 dark:border-slate-700/30 flex items-center justify-between text-xs">
        <span className="text-slate-700 dark:text-slate-300 font-medium truncate max-w-[60%]">
          {title}
        </span>
        <span className="text-slate-500 dark:text-slate-400 flex-shrink-0">{sizeLabel}</span>
      </div>
    </>
  );
}

type ImgGalleryEmptyStateProps = {
  description?: string;
  icon: ReactNode;
  title: string;
};

export function ImgGalleryEmptyState({ description, icon, title }: ImgGalleryEmptyStateProps) {
  return (
    <div className="flex flex-col items-center justify-center gap-4 py-20 text-center">
      <div className="flex h-24 w-24 items-center justify-center rounded-3xl bg-gradient-to-br from-violet-100 to-cyan-100 text-violet-400 shadow-sm dark:from-violet-500/15 dark:to-cyan-500/15 dark:text-violet-300">
        {icon}
      </div>
      <div className="space-y-2">
        <p className="text-lg font-semibold text-slate-800 dark:text-slate-100">{title}</p>
        {description ? <p className="max-w-md text-sm text-muted">{description}</p> : null}
      </div>
    </div>
  );
}

function MagnifyIcon() {
  return (
    <svg
      aria-hidden="true"
      className="w-10 h-10 text-white drop-shadow-lg"
      fill="none"
      viewBox="0 0 24 24"
      stroke="currentColor"
      strokeWidth={1.8}
    >
      <path
        d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0zM10 7v3m0 0v3m0-3h3m-3 0H7"
        strokeLinecap="round"
        strokeLinejoin="round"
      />
    </svg>
  );
}
