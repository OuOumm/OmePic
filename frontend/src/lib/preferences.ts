import type { Language, Theme } from "@/types";
import { detectLanguage } from "./i18n";

const PREF_STORAGE_KEY = "omepic-ui-preferences";
const TOKEN_STORAGE_KEY = "omepic-client-token";
const ADMIN_TOKEN_KEY = "omepic-admin-token";

function generateToken(): string {
  const chars = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789";
  let token = "";
  for (let i = 0; i < 32; i++) {
    token += chars.charAt(Math.floor(Math.random() * chars.length));
  }
  return token;
}

export function getClientToken(): string {
  if (typeof window === "undefined") return "";
  let token = localStorage.getItem(TOKEN_STORAGE_KEY);
  if (!token) {
    token = generateToken();
    localStorage.setItem(TOKEN_STORAGE_KEY, token);
  }
  return token;
}

export function getLanguage(): Language {
  if (typeof window === "undefined") return "en";
  try {
    const raw = localStorage.getItem(PREF_STORAGE_KEY);
    if (raw) {
      const prefs = JSON.parse(raw);
      if (prefs.language === "en" || prefs.language === "zh") {
        return prefs.language;
      }
    }
  } catch { /* corrupted */ }
  return detectLanguage();
}

export function setLanguage(lang: Language) {
  if (typeof window === "undefined") return;
  try {
    const raw = localStorage.getItem(PREF_STORAGE_KEY);
    const prefs = raw ? JSON.parse(raw) : {};
    prefs.language = lang;
    localStorage.setItem(PREF_STORAGE_KEY, JSON.stringify(prefs));
  } catch { /* ignore */ }
}

export function getTheme(): Theme {
  if (typeof window === "undefined") return "dark";
  try {
    const raw = localStorage.getItem(PREF_STORAGE_KEY);
    if (raw) {
      const prefs = JSON.parse(raw);
      if (
        prefs.theme === "light" ||
        prefs.theme === "dark" ||
        prefs.theme === "system"
      ) {
        return prefs.theme;
      }
    }
  } catch { /* corrupted */ }
  return "dark";
}

export function setTheme(theme: Theme) {
  if (typeof window === "undefined") return;
  try {
    const raw = localStorage.getItem(PREF_STORAGE_KEY);
    const prefs = raw ? JSON.parse(raw) : {};
    prefs.theme = theme;
    localStorage.setItem(PREF_STORAGE_KEY, JSON.stringify(prefs));
  } catch { /* ignore */ }
}

export function resolveTheme(theme: Theme): "light" | "dark" {
  if (theme === "system") {
    if (typeof window === "undefined") return "dark";
    return window.matchMedia("(prefers-color-scheme: dark)").matches
      ? "dark"
      : "light";
  }
  return theme;
}

export function getAdminToken(): string | null {
  if (typeof window === "undefined") return null;
  try {
    const raw = localStorage.getItem(ADMIN_TOKEN_KEY);
    if (raw) {
      const data = JSON.parse(raw);
      return data.token ?? null;
    }
  } catch { /* corrupted */ }
  return null;
}

export function setAdminToken(token: string) {
  if (typeof window === "undefined") return;
  localStorage.setItem(ADMIN_TOKEN_KEY, JSON.stringify({ token }));
}

export function clearAdminToken() {
  if (typeof window === "undefined") return;
  localStorage.removeItem(ADMIN_TOKEN_KEY);
}
