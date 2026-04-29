import type { TextareaHTMLAttributes } from "react";

import { cn } from "@/lib/utils";

export function Textarea({ className, ...props }: TextareaHTMLAttributes<HTMLTextAreaElement>) {
  return (
    <textarea
      className={cn(
        "min-h-24 w-full rounded-2xl border border-slate-300/70 bg-white/70 px-4 py-3 text-sm text-ink shadow-sm outline-none backdrop-blur-xl transition-all duration-200 placeholder:text-muted/70 hover:border-violet-300/70 hover:bg-white/85 focus:border-violet-500 focus:bg-white focus:ring-2 focus:ring-violet-500/40 disabled:cursor-not-allowed disabled:bg-slate-200/50 disabled:opacity-70 dark:border-slate-700/70 dark:bg-slate-900/50 dark:hover:border-violet-400/50 dark:hover:bg-slate-900/65 dark:focus:border-violet-400 dark:focus:bg-slate-950/70 dark:focus:ring-violet-500/40",
        className
      )}
      {...props}
    />
  );
}
