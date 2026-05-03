/* eslint-disable @next/next/no-img-element */

import type { CSSProperties, ReactNode } from "react";
import { ZoomIn } from "lucide-react";

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
        "group relative animate-fade-in cursor-pointer overflow-hidden rounded-lg border border-border bg-card shadow-sm transition-colors hover:bg-muted/30",
        selected ? "border-primary ring-2 ring-ring/25" : null,
        className
      )}
      style={style}
    >
      {onPreview ? (
        <button
          aria-label={previewLabel || title}
          className="block w-full text-left focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 focus-visible:ring-offset-background"
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
      <div className="aspect-square overflow-hidden relative bg-muted">
        <img
          alt={alt}
          className="w-full h-full object-cover transition-transform duration-300 group-hover:scale-[1.03]"
          decoding="async"
          loading="lazy"
          src={src}
        />
        <div className="absolute inset-0 flex items-center justify-center bg-black/0 transition-colors duration-200 group-hover:bg-black/35">
          <div className="scale-95 rounded-md border border-border bg-background p-2 text-foreground opacity-0 shadow-sm transition-all duration-200 group-hover:scale-100 group-hover:opacity-100">
            <ZoomIn aria-hidden="true" className="h-5 w-5" />
          </div>
        </div>
      </div>
      <div className="absolute bottom-0 left-0 right-0 flex items-center justify-between border-t border-border bg-background/95 px-3 py-2 text-xs">
        <span className="max-w-[60%] truncate font-medium text-foreground">
          {title}
        </span>
        <span className="flex-shrink-0 text-muted-foreground">{sizeLabel}</span>
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
      <div className="flex h-20 w-20 items-center justify-center rounded-lg border border-border bg-muted text-muted-foreground">
        {icon}
      </div>
      <div className="space-y-2">
        <p className="text-lg font-semibold text-foreground">{title}</p>
        {description ? <p className="max-w-md text-sm text-muted-foreground">{description}</p> : null}
      </div>
    </div>
  );
}
