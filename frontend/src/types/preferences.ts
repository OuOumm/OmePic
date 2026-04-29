export const languages = ["en", "zh"] as const;
export const themeModes = ["light", "dark", "system"] as const;

export type Language = (typeof languages)[number];
export type ThemeMode = (typeof themeModes)[number];
export type ResolvedTheme = Exclude<ThemeMode, "system">;

export function isLanguage(value: unknown): value is Language {
  return typeof value === "string" && languages.includes(value as Language);
}

export function isThemeMode(value: unknown): value is ThemeMode {
  return typeof value === "string" && themeModes.includes(value as ThemeMode);
}

export function languageToDocumentLang(language: Language) {
  return language === "zh" ? "zh-CN" : "en";
}

export function languageToLocale(language: Language) {
  return language === "zh" ? "zh-CN" : "en-US";
}
