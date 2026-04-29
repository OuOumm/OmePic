import type { HTMLAttributes } from "react";

import { cn } from "@/lib/utils";

type CardProps = HTMLAttributes<HTMLDivElement> & {
  variant?: "default" | "strong" | "subtle";
};

const variants: Record<NonNullable<CardProps["variant"]>, string> = {
  default: "glass-panel rounded-[28px]",
  strong:
    "glass-panel-strong rounded-[30px] border-white/60 shadow-[0_24px_80px_rgba(15,23,42,0.16)] dark:border-white/10 dark:shadow-[0_24px_80px_rgba(2,6,23,0.46)]",
  subtle:
    "rounded-[26px] border border-white/45 bg-white/60 shadow-sm backdrop-blur-xl dark:border-white/10 dark:bg-slate-950/40"
};

export function Card({ className, variant = "default", ...props }: CardProps) {
  return (
    <div
      className={cn(
        variants[variant],
        className
      )}
      {...props}
    />
  );
}
