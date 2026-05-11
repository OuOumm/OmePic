import type { Language, PublicRuntimeSettings, Theme } from '@/types';
import { detectLanguage, htmlLang } from '@/i18n';

const PREF_STORAGE_KEY = 'omepic-ui-preferences';
const UPLOAD_PREF_KEY = 'omepic-upload-preferences';
let inMemoryAdminToken: string | null = null;

export type PreferencesState = {
  language: Language;
  theme: Theme;
  selectedStorageKey: string;
  adminToken: string | null;
  runtimeSettings: PublicRuntimeSettings | null;
};

function clearLegacyAdminToken() {
  if (typeof window === 'undefined') return;
  localStorage.removeItem('omepic-admin-token');
}

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

function syncDocumentLanguage(language: Language) {
  if (typeof document === 'undefined') return;
  document.documentElement.lang = htmlLang(language);
}

const uiPrefs = readJSON(PREF_STORAGE_KEY, { language: detectLanguage(), theme: 'light' as Theme });
const uploadPrefs = readJSON(UPLOAD_PREF_KEY, { selectedStorageKey: '' });

export const preferences = $state<PreferencesState>({
  language: uiPrefs.language === 'en' || uiPrefs.language === 'zh' ? uiPrefs.language : detectLanguage(),
  theme: uiPrefs.theme === 'light' || uiPrefs.theme === 'dark' || uiPrefs.theme === 'system' ? uiPrefs.theme : 'light',
  selectedStorageKey: uploadPrefs.selectedStorageKey || '',
  adminToken: inMemoryAdminToken,
  runtimeSettings: null,
});

syncDocumentLanguage(preferences.language);
clearLegacyAdminToken();

export function setLanguage(language: Language) {
  preferences.language = language;
  syncDocumentLanguage(language);
  writeJSON(PREF_STORAGE_KEY, { language: preferences.language, theme: preferences.theme });
}

export function setTheme(theme: Theme) {
  preferences.theme = theme;
  writeJSON(PREF_STORAGE_KEY, { language: preferences.language, theme });
}

export function setSelectedStorageKey(selectedStorageKey: string) {
  preferences.selectedStorageKey = selectedStorageKey;
  writeJSON(UPLOAD_PREF_KEY, { selectedStorageKey });
}

export function setRuntimeSettings(runtimeSettings: PublicRuntimeSettings | null) {
  preferences.runtimeSettings = runtimeSettings;
}

export function setAdminToken(token: string) {
  inMemoryAdminToken = token;
  preferences.adminToken = token;
}

export function clearAdminToken() {
  inMemoryAdminToken = null;
  preferences.adminToken = null;
}

export function resolvedTheme() {
  if (preferences.theme === 'system') {
    if (typeof window === 'undefined') return 'light';
    return window.matchMedia('(prefers-color-scheme: dark)').matches ? 'dark' : 'light';
  }
  return preferences.theme;
}
