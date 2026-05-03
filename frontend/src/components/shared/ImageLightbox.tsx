"use client";

/* eslint-disable @next/next/no-img-element */

import type { ReactNode } from "react";
import { X } from "lucide-react";

import { Button } from "@/components/ui/Button";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogTitle
} from "@/components/ui/Dialog";
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

export function ImageLightbox({
  actions,
  closeLabel,
  eyebrowLabel,
  image,
  metadataLabel,
  onClose
}: ImageLightboxProps) {
  return (
    <Dialog open={Boolean(image)} onOpenChange={(open) => {
      if (!open) {
        onClose();
      }
    }}>
      {image ? (
        <DialogContent
          className="grid max-h-[92vh] w-[calc(100%-1.5rem)] max-w-6xl gap-0 overflow-hidden rounded-lg border-border bg-popover p-0 shadow-lg lg:grid-cols-[minmax(0,1fr)_340px]"
          closeLabel={closeLabel}
          showCloseButton={false}
        >
          <div className="flex min-h-[280px] items-center justify-center bg-muted p-3 sm:p-6">
            <img
              alt={image.alt}
              className="max-h-[72vh] w-full rounded-md object-contain shadow-sm"
              decoding="async"
              src={image.src}
            />
          </div>
          <aside className="flex min-h-0 flex-col border-t border-border bg-popover lg:border-l lg:border-t-0">
            <div className="flex items-start justify-between gap-3 border-b border-border p-4 sm:p-5">
              <div className="min-w-0 space-y-1">
                <p className="text-xs font-medium uppercase tracking-wider text-muted-foreground">
                  {eyebrowLabel}
                </p>
                <DialogTitle className="truncate text-lg font-semibold text-foreground">
                  {image.title}
                </DialogTitle>
                <DialogDescription
                  className={cn(
                    image.subtitle ? "break-all font-mono text-xs text-muted-foreground" : "sr-only"
                  )}
                >
                  {image.subtitle ?? image.title}
                </DialogDescription>
              </div>
              <Button aria-label={closeLabel} onClick={onClose} size="icon" variant="secondary">
                <X aria-hidden="true" className="h-5 w-5" />
              </Button>
            </div>

            {image.metadata?.length ? (
              <div className="overflow-y-auto p-4 sm:p-5">
                <p className="sr-only">{metadataLabel}</p>
                <dl className="grid gap-3 text-sm">
                  {image.metadata.map((item) => (
                    <div
                      className="rounded-md border border-border bg-muted/30 p-3"
                      key={`${item.label}-${item.value}`}
                    >
                      <dt className="text-xs font-medium uppercase tracking-wide text-muted-foreground">
                        {item.label}
                      </dt>
                      <dd
                        className={cn(
                          "mt-1 break-all text-foreground",
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
              <div className="mt-auto flex flex-wrap gap-2 border-t border-border p-4 sm:p-5">
                {actions}
              </div>
            ) : null}
          </aside>
        </DialogContent>
      ) : null}
    </Dialog>
  );
}
