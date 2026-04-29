"use client";

import { useEffect } from "react";

import {
  detectBrowserLanguage,
  readStoredUiPreferences,
  resolveThemeMode
} from "@/lib/preferences";
import { useUiPreferencesStore } from "@/stores/ui-preferences-store";
import { languageToDocumentLang } from "@/types/preferences";

export function UiPreferenceSync() {
  const language = useUiPreferencesStore((state) => state.language);
  const theme = useUiPreferencesStore((state) => state.theme);
  const hasHydrated = useUiPreferencesStore((state) => state.hasHydrated);
  const setLanguage = useUiPreferencesStore((state) => state.setLanguage);

  useEffect(() => {
    if (!hasHydrated) {
      return;
    }

    const stored = readStoredUiPreferences();
    if (!stored.language) {
      setLanguage(detectBrowserLanguage(window.navigator.language));
    }
  }, [hasHydrated, setLanguage]);

  useEffect(() => {
    document.documentElement.lang = languageToDocumentLang(language);
  }, [language]);

  useEffect(() => {
    const mediaQuery = window.matchMedia("(prefers-color-scheme: dark)");

    const applyTheme = () => {
      const resolvedTheme = resolveThemeMode(theme, mediaQuery.matches);
      document.documentElement.dataset.theme = resolvedTheme;
      document.documentElement.dataset.themeMode = theme;
      document.documentElement.style.colorScheme = resolvedTheme;
      document.documentElement.classList.toggle("dark", resolvedTheme === "dark");
    };

    applyTheme();

    if (theme !== "system") {
      return;
    }

    mediaQuery.addEventListener("change", applyTheme);
    return () => {
      mediaQuery.removeEventListener("change", applyTheme);
    };
  }, [theme]);

  return null;
}
