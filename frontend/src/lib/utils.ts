import { clsx, type ClassValue } from 'clsx';
import { twMerge } from 'tailwind-merge';
import type { Language, Theme } from '@/types';
import { locale } from '@/i18n';

const fallbackOrigin = 'http://localhost';
const blockedImageSchemes = new Set(['javascript:', 'data:', 'vbscript:', 'file:']);
const defaultImageAcceptTypes = ['image/avif', 'image/png', 'image/jpeg', 'image/gif', 'image/webp'];

export function cn(...inputs: ClassValue[]) {
  return twMerge(clsx(inputs));
}

export function formatBytes(bytes: number, language: Language = 'en'): string {
  if (bytes === 0) return '0 B';
  const units = ['B', 'KB', 'MB', 'GB', 'TB'];
  const i = Math.min(Math.floor(Math.log(bytes) / Math.log(1024)), units.length - 1);
  const size = bytes / Math.pow(1024, i);
  const formatted = new Intl.NumberFormat(locale(language), { maximumFractionDigits: i === 0 ? 0 : 1 }).format(size);
  return `${formatted} ${units[i]}`;
}

export function formatMegabytes(value: number, language: Language = 'en'): string {
  return `${new Intl.NumberFormat(locale(language), { maximumFractionDigits: 1 }).format(value)} MB`;
}

export function formatDate(dateStr: string, language: Language = 'en'): string {
  return new Intl.DateTimeFormat(locale(language), { dateStyle: 'medium', timeStyle: 'short' }).format(new Date(dateStr));
}

export function getApiBaseUrl(): string {
  if (typeof window === 'undefined') return 'http://localhost:8080';
  const envBase = import.meta.env.VITE_API_BASE_URL;
  if (envBase) return envBase.replace(/\/+$/, '');
  return '';
}

export function getAbsoluteUrl(url: string): string {
  if (/^https?:\/\//i.test(url)) return url;
  if (typeof window !== 'undefined') return new URL(url, window.location.origin).toString();
  const base = getApiBaseUrl();
  return `${base}${url.startsWith('/') ? url : `/${url}`}`;
}

export function getApiExampleBaseUrl(runtimePublicBaseUrl?: string | null): string {
  const runtimeBase = runtimePublicBaseUrl?.trim();
  if (runtimeBase) return runtimeBase.replace(/\/+$/, '');
  if (typeof window !== 'undefined') return window.location.origin;
  return '$ORIGIN';
}

export function getImagePath(uid: string): string {
  return `/i/${uid}.avif`;
}

export function getImageUrl(uid: string): string {
  const base = getApiBaseUrl();
  return `${base}${getImagePath(uid)}`;
}

export function safeImageUrl(value: string, origin = currentOrigin(), allowedOrigins: readonly string[] = []): string | null {
  const trimmed = value.trim();
  if (!trimmed) return null;

  try {
    const parsed = new URL(trimmed, origin);
    if (blockedImageSchemes.has(parsed.protocol)) return null;
    if (parsed.protocol !== 'http:' && parsed.protocol !== 'https:') return null;
    const allowedOriginSet = new Set([origin, ...allowedOrigins].map((item) => originFromUrl(item)).filter((item): item is string => Boolean(item)));
    if (!allowedOriginSet.has(parsed.origin)) return null;
    return trimmed.startsWith('/') ? `${parsed.pathname}${parsed.search}${parsed.hash}` : parsed.toString();
  } catch {
    return null;
  }
}

export function imageUrlAllowedOrigins(publicBaseUrl?: string | null): string[] {
  const origin = originFromUrl(publicBaseUrl ?? '');
  return origin ? [origin] : [];
}

function originFromUrl(value: string): string | null {
  const trimmed = value.trim();
  if (!trimmed) return null;
  try {
    const parsed = new URL(trimmed, currentOrigin());
    return parsed.protocol === 'http:' || parsed.protocol === 'https:' ? parsed.origin : null;
  } catch {
    return null;
  }
}

export function normalizedImageMimeType(mimeType: string): string {
  const normalized = mimeType.split(';', 1)[0].trim().toLowerCase();
  return normalized === 'image/jpg' ? 'image/jpeg' : normalized;
}

export function isBlockedImageMimeType(mimeType: string): boolean {
  return normalizedImageMimeType(mimeType) === 'image/svg+xml';
}

export function normalizedAllowedImageMimeTypes(allowedMimeTypes: readonly string[]): string[] {
  const seen = new Set<string>();
  const result: string[] = [];
  for (const value of allowedMimeTypes) {
    const normalized = normalizedImageMimeType(value);
    if (!normalized || normalized === 'image/svg+xml' || !normalized.startsWith('image/')) continue;
    if (seen.has(normalized)) continue;
    seen.add(normalized);
    result.push(normalized);
  }
  return result;
}

export function imageAcceptFromMimeTypes(allowedMimeTypes?: readonly string[] | null): string {
  const normalized = normalizedAllowedImageMimeTypes(allowedMimeTypes?.length ? allowedMimeTypes : defaultImageAcceptTypes);
  return normalized.join(',');
}

export function isAllowedImageMimeType(mimeType: string, allowedMimeTypes: readonly string[]): boolean {
  const normalized = normalizedImageMimeType(mimeType);
  if (!normalized || normalized === 'image/svg+xml') return false;
  return normalizedAllowedImageMimeTypes(allowedMimeTypes).includes(normalized);
}

export function isAbortError(err: unknown): boolean {
  return err instanceof DOMException && err.name === 'AbortError';
}

export function normalizeDownloadFilename(value: string | null | undefined, fallback: string): string {
  const normalized = (value ?? '').replace(/[\\/:*?"<>|]/g, '').trim();
  return normalized || fallback;
}

export function getInitialThemeScriptTheme(rawPreferences: string | null, systemPrefersDark: boolean): 'light' | 'dark' {
  try {
    const prefs = rawPreferences ? JSON.parse(rawPreferences) as { theme?: unknown } : {};
    if (prefs.theme === 'dark') return 'dark';
    if (prefs.theme === 'system') return systemPrefersDark ? 'dark' : 'light';
    return 'light';
  } catch {
    return 'light';
  }
}

export function initialThemeScript(storageKey = 'omepic-ui-preferences'): string {
  return `(function(){try{var raw=localStorage.getItem(${JSON.stringify(storageKey)});var prefersDark=matchMedia('(prefers-color-scheme: dark)').matches;var prefs=raw?JSON.parse(raw):{};var theme=prefs.theme==='dark'||prefs.theme==='system'&&prefersDark?'dark':'light';document.documentElement.classList.toggle('dark',theme==='dark')}catch{document.documentElement.classList.remove('dark')}})();`;
}

export function normalizeTheme(theme: unknown): Theme {
  return theme === 'light' || theme === 'dark' || theme === 'system' ? theme : 'light';
}

export function markdownSummaryText(content: string, maxLength = 180): string {
  const plain = content
    .replace(/```[\s\S]*?```/g, ' ')
    .replace(/`([^`]*)`/g, '$1')
    .replace(/!\[[^\]]*\]\([^)]*\)/g, ' ')
    .replace(/\[([^\]]+)\]\([^)]*\)/g, '$1')
    .replace(/^\s{0,3}#{1,6}\s+/gm, '')
    .replace(/^\s*[-*+]\s+/gm, '')
    .replace(/^\s*\d+\.\s+/gm, '')
    .replace(/[*_~>#|]/g, '')
    .replace(/\s+/g, ' ')
    .trim();

  const chars = Array.from(plain);
  return chars.length > maxLength ? `${chars.slice(0, maxLength).join('')}…` : plain;
}

function currentOrigin(): string {
  if (typeof window === 'undefined') return fallbackOrigin;
  return window.location.origin;
}

export function getAbsoluteImageUrl(uid: string): string {
  return getAbsoluteUrl(getImageUrl(uid));
}
