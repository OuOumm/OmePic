import { clsx, type ClassValue } from 'clsx';
import { twMerge } from 'tailwind-merge';
import type { Language } from '@/types';
import { locale } from '@/i18n';

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

export function getImageUrl(uid: string): string {
  const base = getApiBaseUrl();
  return `${base}/i/${uid}.avif`;
}

export function getAbsoluteImageUrl(uid: string): string {
  return getAbsoluteUrl(getImageUrl(uid));
}
