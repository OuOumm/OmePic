"use client";

import { useUiTranslations } from "@/hooks/useUiPreferences";

export function SkipLink() {
  const t = useUiTranslations();

  return (
    <a
      className="sr-only focus:not-sr-only focus:fixed focus:left-4 focus:top-4 focus:z-[70] focus:rounded-xl focus:border focus:border-white/50 focus:bg-white/90 focus:px-4 focus:py-2 focus:text-sm focus:font-semibold focus:text-slate-900 focus:shadow-glow focus:outline-none focus:ring-2 focus:ring-violet-400 focus:ring-offset-2 focus:ring-offset-surface dark:focus:border-white/10 dark:focus:bg-slate-900/90 dark:focus:text-white"
      href="#main-content"
    >
      {t.common.skipToContent}
    </a>
  );
}
