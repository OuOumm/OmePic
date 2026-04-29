"use client";

import { getDictionary, getLocaleForLanguage } from "@/lib/i18n";
import { useUiPreferencesStore } from "@/stores/ui-preferences-store";

export function useLanguage() {
  return useUiPreferencesStore((state) => state.language);
}

export function useThemeMode() {
  return useUiPreferencesStore((state) => state.theme);
}

export function useUiTranslations() {
  const language = useLanguage();
  return getDictionary(language);
}

export function useUiLocale() {
  const language = useLanguage();
  return getLocaleForLanguage(language);
}
