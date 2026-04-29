import {
  isLanguage,
  isThemeMode,
  languageToDocumentLang,
  type Language,
  type ResolvedTheme,
  type ThemeMode
} from "@/types/preferences";

export const UI_PREFERENCES_STORAGE_KEY = "omepic-ui-preferences";

type StoredUiPreferences = {
  language?: Language;
  theme?: ThemeMode;
};

export function detectBrowserLanguage(value: string | undefined) {
  return value?.toLowerCase().startsWith("zh") ? "zh" : "en";
}

export function getSystemTheme(matchesDark: boolean): ResolvedTheme {
  return matchesDark ? "dark" : "light";
}

export function resolveThemeMode(theme: ThemeMode, matchesDark: boolean): ResolvedTheme {
  return theme === "system" ? getSystemTheme(matchesDark) : theme;
}

export function readStoredUiPreferences(): StoredUiPreferences {
  if (typeof window === "undefined") {
    return {};
  }

  try {
    const raw = window.localStorage.getItem(UI_PREFERENCES_STORAGE_KEY);
    if (!raw) {
      return {};
    }

    const parsed = JSON.parse(raw) as { state?: { language?: unknown; theme?: unknown } };
    const language = isLanguage(parsed.state?.language) ? parsed.state.language : undefined;
    const theme = isThemeMode(parsed.state?.theme) ? parsed.state.theme : undefined;

    return { language, theme };
  } catch {
    return {};
  }
}

export function createPreferenceInitScript() {
  return `(() => {
    const storageKey = ${JSON.stringify(UI_PREFERENCES_STORAGE_KEY)};
    const detectLanguage = (value) => typeof value === "string" && value.toLowerCase().startsWith("zh") ? "zh" : "en";
    const normalizeTheme = (value) => value === "light" || value === "dark" || value === "system" ? value : "dark";
    const normalizeLanguage = (value) => value === "en" || value === "zh" ? value : detectLanguage(window.navigator.language);
    const apply = (language, themeMode) => {
      const root = document.documentElement;
      const resolvedTheme = themeMode === "system"
        ? (window.matchMedia("(prefers-color-scheme: dark)").matches ? "dark" : "light")
        : themeMode;
      root.lang = language === "zh" ? ${JSON.stringify(languageToDocumentLang("zh"))} : ${JSON.stringify(languageToDocumentLang("en"))};
      root.dataset.theme = resolvedTheme;
      root.dataset.themeMode = themeMode;
      root.style.colorScheme = resolvedTheme;
      root.classList.toggle("dark", resolvedTheme === "dark");
    };

    try {
      const raw = window.localStorage.getItem(storageKey);
      const parsed = raw ? JSON.parse(raw) : null;
      const state = parsed && typeof parsed === "object" ? parsed.state : null;
      apply(normalizeLanguage(state?.language), normalizeTheme(state?.theme || "dark"));
    } catch {
      apply(detectLanguage(window.navigator.language), "dark");
    }
  })();`;
}
