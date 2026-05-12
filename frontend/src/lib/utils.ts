import { clsx, type ClassValue } from 'clsx';
import { twMerge } from 'tailwind-merge';
import type { Language, Theme } from '@/types';
import { locale } from '@/i18n';

const fallbackOrigin = 'http://localhost';
const blockedImageSchemes = new Set(['javascript:', 'data:', 'vbscript:', 'file:']);

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

export function safeImageUrl(value: string, origin = currentOrigin()): string | null {
  const trimmed = value.trim();
  if (!trimmed) return null;

  try {
    const parsed = new URL(trimmed, origin);
    if (blockedImageSchemes.has(parsed.protocol)) return null;
    if (parsed.protocol !== 'http:' && parsed.protocol !== 'https:') return null;
    if (parsed.origin !== origin) return null;
    return trimmed.startsWith('/') ? `${parsed.pathname}${parsed.search}${parsed.hash}` : parsed.toString();
  } catch {
    return null;
  }
}

export function isAllowedImageMimeType(mimeType: string, allowedMimeTypes: string[]): boolean {
  const normalized = mimeType.split(';', 1)[0].trim().toLowerCase();
  if (normalized === 'image/svg+xml') return false;
  return allowedMimeTypes.map((value) => value.trim().toLowerCase()).includes(normalized);
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
