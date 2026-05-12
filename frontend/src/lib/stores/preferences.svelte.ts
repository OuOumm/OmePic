import type { Language, PublicRuntimeSettings, Theme } from '@/types';
import { detectLanguage, htmlLang } from '@/i18n';
import { normalizeTheme } from '@/utils';

const PREF_STORAGE_KEY = 'omepic-ui-preferences';
const UPLOAD_PREF_KEY = 'omepic-upload-preferences';
const ADMIN_TOKEN_KEY = 'omepic-admin-token';

export type PreferencesState = {
  language: Language;
  theme: Theme;
  selectedStorageKey: string;
  adminToken: string | null;
  runtimeSettings: PublicRuntimeSettings | null;
};

function readJSON<T>(key: string, fallback: T): T {
  if (typeof window === 'undefined') return fallback;
  try {
    const raw = localStorage.getItem(key);
    return raw ? { ...fallback, ...JSON.parse(raw) } : fallback;
  } catch {
    return fallback;
  }
}

function writeJSON<T>(key: string, value: T) {
  if (typeof window === 'undefined') return;
  localStorage.setItem(key, JSON.stringify(value));
}

function readString(key: string): string | null {
  if (typeof window === 'undefined') return null;
  try {
    return localStorage.getItem(key);
  } catch {
    return null;
  }
}

function writeString(key: string, value: string) {
  if (typeof window === 'undefined') return;
  localStorage.setItem(key, value);
}

function removeStorageItem(key: string) {
  if (typeof window === 'undefined') return;
  localStorage.removeItem(key);
}

function syncDocumentLanguage(language: Language) {
  if (typeof document === 'undefined') return;
  document.documentElement.lang = htmlLang(language);
}

const uiPrefs = readJSON(PREF_STORAGE_KEY, { language: detectLanguage(), theme: 'light' as Theme });
const uploadPrefs = readJSON(UPLOAD_PREF_KEY, { selectedStorageKey: '' });
const initialLanguage = uiPrefs.language === 'en' || uiPrefs.language === 'zh' ? uiPrefs.language : detectLanguage();
const initialTheme = normalizeTheme(uiPrefs.theme);
const initialAdminToken = readString(ADMIN_TOKEN_KEY);

export const preferences = $state<PreferencesState>({
  language: initialLanguage,
  theme: initialTheme,
  selectedStorageKey: uploadPrefs.selectedStorageKey || '',
  adminToken: initialAdminToken,
  runtimeSettings: null,
});

syncDocumentLanguage(preferences.language);

export function setLanguage(language: Language) {
  preferences.language = language;
  syncDocumentLanguage(language);
  writeJSON(PREF_STORAGE_KEY, { language: preferences.language, theme: preferences.theme });
}

export function setTheme(theme: Theme) {
  preferences.theme = normalizeTheme(theme);
  writeJSON(PREF_STORAGE_KEY, { language: preferences.language, theme: preferences.theme });
}

export function setSelectedStorageKey(selectedStorageKey: string) {
  preferences.selectedStorageKey = selectedStorageKey;
  writeJSON(UPLOAD_PREF_KEY, { selectedStorageKey });
}

export function setRuntimeSettings(runtimeSettings: PublicRuntimeSettings | null) {
  preferences.runtimeSettings = runtimeSettings;
}

export function setAdminToken(token: string) {
  preferences.adminToken = token;
  writeString(ADMIN_TOKEN_KEY, token);
}

export function clearAdminToken() {
  preferences.adminToken = null;
  removeStorageItem(ADMIN_TOKEN_KEY);
}

export function resolvedTheme() {
  if (preferences.theme === 'system') {
    if (typeof window === 'undefined') return 'light';
    return window.matchMedia('(prefers-color-scheme: dark)').matches ? 'dark' : 'light';
  }
  return preferences.theme;
}
