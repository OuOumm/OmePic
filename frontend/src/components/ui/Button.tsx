"use client";

import * as React from "react";

import { cn } from "@/lib/utils";

type ButtonProps = React.ButtonHTMLAttributes<HTMLButtonElement> & {
  variant?: "primary" | "secondary" | "ghost" | "danger";
  size?: "sm" | "md" | "icon";
};

const variants: Record<NonNullable<ButtonProps["variant"]>, string> = {
  primary:
    "border border-violet-400/30 bg-gradient-to-r from-violet-600 to-cyan-600 text-white shadow-lg shadow-violet-500/25 hover:from-violet-500 hover:to-cyan-500 hover:shadow-xl hover:shadow-violet-500/30 dark:from-violet-500 dark:to-cyan-500 dark:hover:from-violet-400 dark:hover:to-cyan-400",
  secondary:
    "border border-white/50 bg-white/70 text-slate-700 shadow-sm backdrop-blur-xl hover:border-violet-300/70 hover:bg-white/90 hover:text-violet-700 dark:border-white/10 dark:bg-slate-900/60 dark:text-slate-200 dark:hover:border-violet-400/40 dark:hover:bg-slate-800/70 dark:hover:text-violet-200",
  ghost:
    "border border-transparent bg-transparent text-muted hover:border-white/50 hover:bg-white/60 hover:text-violet-700 dark:hover:border-white/10 dark:hover:bg-white/10 dark:hover:text-violet-200",
  danger:
    "border border-rose-400/40 bg-rose-500/10 text-rose-600 shadow-sm hover:border-rose-400/70 hover:bg-rose-500 hover:text-white hover:shadow-lg hover:shadow-rose-500/20 dark:text-rose-300 dark:hover:text-white"
};

const sizes: Record<NonNullable<ButtonProps["size"]>, string> = {
  sm: "min-h-9 px-3.5 py-2 text-xs",
  md: "min-h-11 px-5 py-2.5 text-sm",
  icon: "h-10 w-10 p-0"
};

export const Button = React.forwardRef<HTMLButtonElement, ButtonProps>(function Button(
  {
    className,
    size = "md",
    variant = "primary",
    type = "button",
    ...props
  },
  ref
) {
  return (
    <button
      className={cn(
        "inline-flex items-center justify-center gap-2 rounded-2xl font-semibold tracking-[0.01em] transition-all duration-200 hover:-translate-y-0.5 active:translate-y-0 active:scale-[0.98] disabled:cursor-not-allowed disabled:opacity-55 disabled:hover:translate-y-0 disabled:active:scale-100 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-violet-400/70 focus-visible:ring-offset-2 focus-visible:ring-offset-surface [&_svg]:shrink-0",
        variants[variant],
        sizes[size],
        className
      )}
      ref={ref}
      type={type}
      {...props}
    />
  );
});
