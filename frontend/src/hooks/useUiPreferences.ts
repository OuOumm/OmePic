"use client";

import { useEffect } from "react";
import { useUiPreferencesStore } from "@/stores/ui-preferences-store";
import { resolveTheme } from "@/lib/preferences";

export function useUiPreferences() {
  const language = useUiPreferencesStore((state) => state.language);
  const theme = useUiPreferencesStore((state) => state.theme);
  const hasHydrated = useUiPreferencesStore((state) => state.hasHydrated);

  useEffect(() => {
    if (!hasHydrated) return;
    const resolved = resolveTheme(theme);
    document.documentElement.lang = language;
    document.documentElement.dataset.theme = resolved;
    document.documentElement.dataset.themeMode = theme;
    if (resolved === "dark") {
      document.documentElement.classList.add("dark");
    } else {
      document.documentElement.classList.remove("dark");
    }
  }, [language, theme, hasHydrated]);

  return { language, theme, hasHydrated };
}
