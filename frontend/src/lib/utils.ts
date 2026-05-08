import { clsx, type ClassValue } from "clsx";
import { twMerge } from "tailwind-merge";

export function cn(...inputs: ClassValue[]) {
  return twMerge(clsx(inputs));
}

export function formatBytes(bytes: number): string {
  if (bytes === 0) return "0 B";
  const units = ["B", "KB", "MB", "GB", "TB"];
  const i = Math.floor(Math.log(bytes) / Math.log(1024));
  const size = bytes / Math.pow(1024, i);
  return `${size.toFixed(i === 0 ? 0 : 1)} ${units[i]}`;
}

export function formatDate(dateStr: string): string {
  const d = new Date(dateStr);
  return d.toLocaleString();
}

export function getApiBaseUrl(): string {
  if (typeof window === "undefined") return "http://localhost:8080";
  if (process.env.NEXT_PUBLIC_API_BASE_URL) {
    return process.env.NEXT_PUBLIC_API_BASE_URL.replace(/\/+$/, "");
  }
  return "";
}

export function getAbsoluteUrl(url: string): string {
  if (/^https?:\/\//i.test(url)) return url;
  if (typeof window !== "undefined") return new URL(url, window.location.origin).toString();
  const base = getApiBaseUrl();
  return `${base}${url.startsWith("/") ? url : `/${url}`}`;
}

export function getImageUrl(uid: string): string {
  const base = getApiBaseUrl();
  return `${base}/i/${uid}.avif`;
}

export function getAbsoluteImageUrl(uid: string): string {
  return getAbsoluteUrl(getImageUrl(uid));
}
