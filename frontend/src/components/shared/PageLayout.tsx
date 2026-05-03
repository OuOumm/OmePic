import type { ReactNode } from "react";

import { cn } from "@/lib/utils";

type PageIntroProps = {
  actions?: ReactNode;
  aside?: ReactNode;
  className?: string;
  description?: ReactNode;
  eyebrow: string;
  title: ReactNode;
};

type PageDetailPillProps = {
  className?: string;
  label: string;
  value: ReactNode;
};

type PageSectionHeaderProps = {
  actions?: ReactNode;
  badge?: ReactNode;
  className?: string;
  description?: ReactNode;
  icon?: ReactNode;
  title: ReactNode;
};

export function PageIntro({
  actions,
  aside,
  className,
  description,
  eyebrow,
  title
}: PageIntroProps) {
  return (
    <section className={cn("grid gap-6 border-b border-border pb-6 xl:grid-cols-[minmax(0,1fr)_minmax(260px,0.42fr)] xl:items-end", className)}>
        <div className="space-y-4">
          <div className="space-y-3">
            <p className="eyebrow-label">{eyebrow}</p>
            <div className="space-y-3">
              <h1 className="max-w-4xl text-3xl font-semibold tracking-tight text-foreground sm:text-4xl lg:text-[2.4rem] lg:leading-tight">
                {title}
              </h1>
              {description ? (
                <p className="max-w-3xl text-sm leading-6 text-muted-foreground sm:text-[15px]">
                  {description}
                </p>
              ) : null}
            </div>
          </div>
          {actions ? (
            <div className="flex flex-wrap items-center gap-3">
              {actions}
            </div>
          ) : null}
        </div>
        {aside ? <div className="grid gap-3 sm:grid-cols-2 xl:grid-cols-1">{aside}</div> : null}
    </section>
  );
}

export function PageDetailPill({ className, label, value }: PageDetailPillProps) {
  return (
    <div
      className={cn(
        "rounded-lg border border-border bg-card px-4 py-3 shadow-sm",
        className
      )}
    >
      <p className="text-[11px] font-medium uppercase tracking-wider text-muted-foreground">{label}</p>
      <div className="mt-2 text-sm font-medium text-foreground">{value}</div>
    </div>
  );
}

export function PageSectionHeader({
  actions,
  badge,
  className,
  description,
  icon,
  title
}: PageSectionHeaderProps) {
  return (
    <div
      className={cn(
        "flex flex-col gap-4 border-b border-border pb-4 lg:flex-row lg:items-start lg:justify-between",
        className
      )}
    >
      <div className="min-w-0 space-y-2">
        <div className="flex flex-wrap items-center gap-3">
          {icon ? (
            <span className="flex h-9 w-9 items-center justify-center rounded-md border border-border bg-muted text-muted-foreground">
              {icon}
            </span>
          ) : null}
          <div className="min-w-0">
            <h2 className="text-lg font-semibold tracking-tight text-foreground sm:text-xl">
              {title}
            </h2>
            {description ? (
              <p className="mt-1 max-w-3xl text-sm leading-6 text-muted-foreground">{description}</p>
            ) : null}
          </div>
          {badge ? <div className="shrink-0">{badge}</div> : null}
        </div>
      </div>
      {actions ? <div className="flex flex-wrap items-center gap-2">{actions}</div> : null}
    </div>
  );
}
