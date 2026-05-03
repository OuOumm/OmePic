"use client";

import { useUiTranslations } from "@/hooks/useUiPreferences";

export function SkipLink() {
  const t = useUiTranslations();

  return (
    <a
      className="sr-only focus:not-sr-only focus:fixed focus:left-4 focus:top-4 focus:z-[70] focus:rounded-md focus:border focus:border-border focus:bg-background focus:px-4 focus:py-2 focus:text-sm focus:font-medium focus:text-foreground focus:shadow-md focus:outline-none focus:ring-2 focus:ring-ring focus:ring-offset-2 focus:ring-offset-background"
      href="#main-content"
    >
      {t.common.skipToContent}
    </a>
  );
}
