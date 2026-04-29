import type { ReactNode } from "react";

import { Card } from "@/components/ui/Card";
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
    <Card className={cn("relative overflow-hidden p-6 sm:p-7 lg:p-8", className)} variant="strong">
      <div className="pointer-events-none absolute inset-0 bg-[radial-gradient(circle_at_top_left,rgba(139,92,246,0.18),transparent_34%),radial-gradient(circle_at_88%_22%,rgba(34,211,238,0.14),transparent_28%)]" />
      <div className="relative grid gap-6 xl:grid-cols-[minmax(0,1fr)_minmax(260px,0.42fr)] xl:items-end">
        <div className="space-y-4">
          <div className="space-y-3">
            <p className="eyebrow-label">{eyebrow}</p>
            <div className="space-y-3">
              <h1 className="max-w-4xl text-3xl font-bold tracking-tight text-slate-950 dark:text-white sm:text-4xl lg:text-[2.85rem] lg:leading-[1.05]">
                {title}
              </h1>
              {description ? (
                <p className="max-w-3xl text-sm leading-7 text-muted sm:text-[15px]">
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
      </div>
    </Card>
  );
}

export function PageDetailPill({ className, label, value }: PageDetailPillProps) {
  return (
    <div
      className={cn(
        "rounded-2xl border border-white/45 bg-white/60 px-4 py-3 shadow-sm backdrop-blur-xl dark:border-white/10 dark:bg-slate-950/45",
        className
      )}
    >
      <p className="text-[11px] font-bold uppercase tracking-[0.18em] text-muted">{label}</p>
      <div className="mt-2 text-sm font-semibold text-slate-900 dark:text-slate-100">{value}</div>
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
        "flex flex-col gap-4 border-b border-white/45 pb-4 dark:border-white/10 lg:flex-row lg:items-start lg:justify-between",
        className
      )}
    >
      <div className="min-w-0 space-y-2">
        <div className="flex flex-wrap items-center gap-3">
          {icon ? (
            <span className="flex h-10 w-10 items-center justify-center rounded-2xl bg-gradient-to-br from-violet-500/18 to-cyan-500/18 text-violet-600 dark:text-violet-200">
              {icon}
            </span>
          ) : null}
          <div className="min-w-0">
            <h2 className="text-lg font-semibold tracking-tight text-slate-900 dark:text-slate-100 sm:text-xl">
              {title}
            </h2>
            {description ? (
              <p className="mt-1 max-w-3xl text-sm leading-6 text-muted">{description}</p>
            ) : null}
          </div>
          {badge ? <div className="shrink-0">{badge}</div> : null}
        </div>
      </div>
      {actions ? <div className="flex flex-wrap items-center gap-2">{actions}</div> : null}
    </div>
  );
}
