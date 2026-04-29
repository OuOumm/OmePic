import type { HTMLAttributes } from "react";

import { cn } from "@/lib/utils";

export function Badge({ className, ...props }: HTMLAttributes<HTMLSpanElement>) {
  return (
    <span
      className={cn(
        "inline-flex items-center rounded-full border border-violet-300/35 bg-violet-500/15 px-2.5 py-1 text-[11px] font-semibold uppercase tracking-[0.14em] text-violet-700 shadow-sm dark:border-violet-300/20 dark:bg-violet-400/15 dark:text-violet-200",
        className
      )}
      {...props}
    />
  );
}
