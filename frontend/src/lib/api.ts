import { getApiBaseUrl } from "./utils";

import type {
  ApiResponse,
  UploadResult,
  AdminStatus,
  AdminImagesResponse,
  AdminConfig,
  StorageInstance,
  PublicRuntimeSettings,
  AdminSystemSettings,
  RuntimeSettings,
  AnnouncementListResponse,
  Announcement,
  AnnouncementInput,
  AdminIPBan,
  AdminIPBanCreateResult,
  AdminIPBanDeleteImagesResult,
  AdminAbuseOverview,
  AdminAbuseIPDetail,
} from "@/types";

export class ApiError extends Error {
  code?: string;
  status?: number;
  retryAfter?: number;

  constructor(message: string, options: { code?: string; status?: number; retryAfter?: number } = {}) {
    super(message);
    this.name = "ApiError";
    this.code = options.code;
    this.status = options.status;
    this.retryAfter = options.retryAfter;
  }
}

async function apiFetch<T>(
  path: string,
  options: RequestInit & { params?: Record<string, string> } = {}
): Promise<T> {
  const base = getApiBaseUrl();
  let url = `${base}${path}`;
  if (options.params) {
    const searchParams = new URLSearchParams(options.params);
    url += `?${searchParams.toString()}`;
  }
  delete options.params;

  const res = await fetch(url, { cache: "no-store", ...options });
  const json: ApiResponse<T> = await res.json();
  if (!res.ok || !json.success) {
    if ("error" in json) {
      throw new ApiError(json.error.message, { code: json.error.code, status: res.status });
    }
    throw new ApiError(`HTTP ${res.status}`, { status: res.status });
  }
  return json.data as T;
}

export function uploadImageWithProgress(
  file: File,
  token: string,
  onProgress: (pct: number) => void,
  storageKey?: string
): Promise<UploadResult> {
  return new Promise((resolve, reject) => {
    const xhr = new XMLHttpRequest();
    const base = getApiBaseUrl();
    const formData = new FormData();
    formData.append("file", file);
    if (storageKey) formData.append("storage_key", storageKey);

    xhr.upload.addEventListener("progress", (e) => {
      if (e.lengthComputable) {
        onProgress(Math.round((e.loaded / e.total) * 100));
      }
    });

    xhr.addEventListener("load", () => {
      if (xhr.status >= 200 && xhr.status < 300) {
        try {
          const json: ApiResponse<UploadResult> = JSON.parse(xhr.responseText);
          if (json.success) {
            resolve(json.data as UploadResult);
          } else {
            reject(new ApiError("error" in json ? json.error.message : "Upload failed", {
              code: "error" in json ? json.error.code : undefined,
              status: xhr.status,
              retryAfter: parseRetryAfter(xhr.getResponseHeader("Retry-After")),
            }));
          }
        } catch {
          reject(new Error("Invalid response from server"));
        }
      } else {
        try {
          const json: ApiResponse<UploadResult> = JSON.parse(xhr.responseText);
          if ("error" in json) {
            reject(new ApiError(json.error.message, {
              code: json.error.code,
              status: xhr.status,
              retryAfter: parseRetryAfter(xhr.getResponseHeader("Retry-After")),
            }));
            return;
          }
        } catch {
          reject(new ApiError(`Upload failed: HTTP ${xhr.status}`, {
            status: xhr.status,
            retryAfter: parseRetryAfter(xhr.getResponseHeader("Retry-After")),
          }));
          return;
        }
        reject(new ApiError(`Upload failed: HTTP ${xhr.status}`, {
          status: xhr.status,
          retryAfter: parseRetryAfter(xhr.getResponseHeader("Retry-After")),
        }));
      }
    });

    xhr.addEventListener("error", () => {
      reject(new ApiError("Network error during upload", {
        code: xhr.status === 429 || xhr.getResponseHeader("Retry-After") ? "rate_limited" : "network_error",
        status: xhr.status || undefined,
        retryAfter: parseRetryAfter(xhr.getResponseHeader("Retry-After")),
      }));
    });
    xhr.addEventListener("abort", () => reject(new ApiError("Upload aborted", { code: "upload_aborted" })));

    xhr.open("POST", `${base}/v1/image`);
    xhr.setRequestHeader("X-Token", token);
    xhr.send(formData);
  });
}

function parseRetryAfter(value: string | null): number | undefined {
  if (!value) return undefined;
  const parsed = Number(value);
  return Number.isFinite(parsed) && parsed > 0 ? parsed : undefined;
}

// Public endpoints
export async function getRuntimeSettings(signal?: AbortSignal): Promise<PublicRuntimeSettings> {
  return apiFetch<PublicRuntimeSettings>("/v1/runtime-settings", { signal });
}

export async function getAnnouncements(signal?: AbortSignal): Promise<Announcement[]> {
  const data = await apiFetch<AnnouncementListResponse>("/v1/announcements", { signal });
  return data.items;
}

export async function deleteImageByUid(uid: string, token: string): Promise<void> {
  const base = getApiBaseUrl();
  const res = await fetch(`${base}/i/${uid}.avif`, {
    method: "DELETE",
    headers: { "X-Token": token },
    cache: "no-store",
  });
  const json: ApiResponse<null> = await res.json();
  if (!res.ok || !json.success) {
    const msg = "error" in json ? json.error.message : `HTTP ${res.status}`;
    throw new Error(msg);
  }
}

// Admin endpoints
function adminHeaders(token: string): HeadersInit {
  return {
    Authorization: `Bearer ${token}`,
    "Content-Type": "application/json",
  };
}

export async function adminLogin(password: string): Promise<string> {
  const base = getApiBaseUrl();
  const res = await fetch(`${base}/admin/login`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ password }),
    cache: "no-store",
  });
  const json: ApiResponse<{ token: string }> = await res.json();
  if (!res.ok || !json.success) {
    const msg = "error" in json ? json.error.message : `HTTP ${res.status}`;
    throw new Error(msg);
  }
  return json.data.token;
}

export async function adminGetStatus(token: string): Promise<AdminStatus> {
  const base = getApiBaseUrl();
  const res = await fetch(`${base}/admin/status`, {
    headers: { Authorization: `Bearer ${token}` },
    cache: "no-store",
  });
  const json: ApiResponse<AdminStatus> = await res.json();
  if (!res.ok || !json.success) {
    const msg = "error" in json ? json.error.message : `HTTP ${res.status}`;
    throw new Error(msg);
  }
  return json.data;
}

export async function adminGetImages(
  token: string,
  page: number,
  pageSize: number,
  search?: string
): Promise<AdminImagesResponse> {
  return apiFetch<AdminImagesResponse>("/admin/images", {
    headers: adminHeaders(token),
    params: {
      page: String(page),
      pageSize: String(pageSize),
      ...(search ? { search } : {}),
    },
  });
}

export async function adminDeleteImages(token: string, uids: string[]): Promise<void> {
  const base = getApiBaseUrl();
  const res = await fetch(`${base}/admin/images`, {
    method: "DELETE",
    headers: adminHeaders(token),
    body: JSON.stringify({ uids }),
    cache: "no-store",
  });
  const json: ApiResponse<null> = await res.json();
  if (!res.ok || !json.success) {
    const msg = "error" in json ? json.error.message : `HTTP ${res.status}`;
    throw new Error(msg);
  }
}

export async function adminCreateIPBan(
  token: string,
  input: { uid?: string; ip_address?: string; duration_hours: number; reason?: string }
): Promise<AdminIPBanCreateResult> {
  return apiFetch<AdminIPBanCreateResult>("/admin/ip-bans", {
    method: "POST",
    headers: adminHeaders(token),
    body: JSON.stringify(input),
  });
}

export async function adminGetIPBans(token: string): Promise<AdminIPBan[]> {
  return apiFetch<AdminIPBan[]>("/admin/ip-bans", {
    headers: adminHeaders(token),
  });
}

export async function adminDeleteIPBan(token: string, id: number): Promise<void> {
  await apiFetch<Record<string, never>>(`/admin/ip-bans/${id}`, {
    method: "DELETE",
    headers: adminHeaders(token),
  });
}

export async function adminDeleteIPBanImages(
  token: string,
  id: number
): Promise<AdminIPBanDeleteImagesResult> {
  return apiFetch<AdminIPBanDeleteImagesResult>(`/admin/ip-bans/${id}/images`, {
    method: "DELETE",
    headers: adminHeaders(token),
  });
}

export async function adminGetAbuseOverview(
  token: string,
  from?: string,
  to?: string
): Promise<AdminAbuseOverview> {
  return apiFetch<AdminAbuseOverview>("/admin/abuse/overview", {
    headers: adminHeaders(token),
    params: {
      ...(from ? { from } : {}),
      ...(to ? { to } : {}),
    },
  });
}

export async function adminGetAbuseIPDetail(token: string, ip: string): Promise<AdminAbuseIPDetail> {
  return apiFetch<AdminAbuseIPDetail>("/admin/abuse/ip", {
    headers: adminHeaders(token),
    params: { ip },
  });
}

export async function adminGetConfig(token: string): Promise<AdminConfig> {
  return apiFetch<AdminConfig>("/admin/config", {
    headers: { Authorization: `Bearer ${token}` },
  });
}

export async function adminCreateStorageInstance(
  token: string,
  instance: Partial<StorageInstance>
): Promise<AdminConfig> {
  const base = getApiBaseUrl();
  const res = await fetch(`${base}/admin/config/storage-instances`, {
    method: "POST",
    headers: adminHeaders(token),
    body: JSON.stringify(instance),
    cache: "no-store",
  });
  const json: ApiResponse<AdminConfig> = await res.json();
  if (!res.ok || !json.success) {
    const msg = "error" in json ? json.error.message : `HTTP ${res.status}`;
    throw new Error(msg);
  }
  return json.data;
}

export async function adminUpdateStorageInstance(
  token: string,
  storageKey: string,
  instance: Partial<StorageInstance>
): Promise<AdminConfig> {
  const base = getApiBaseUrl();
  const res = await fetch(`${base}/admin/config/storage-instances/${storageKey}`, {
    method: "PUT",
    headers: adminHeaders(token),
    body: JSON.stringify(instance),
    cache: "no-store",
  });
  const json: ApiResponse<AdminConfig> = await res.json();
  if (!res.ok || !json.success) {
    const msg = "error" in json ? json.error.message : `HTTP ${res.status}`;
    throw new Error(msg);
  }
  return json.data;
}

export async function adminDeleteStorageInstance(
  token: string,
  storageKey: string
): Promise<AdminConfig> {
  const base = getApiBaseUrl();
  const res = await fetch(`${base}/admin/config/storage-instances/${storageKey}`, {
    method: "DELETE",
    headers: { Authorization: `Bearer ${token}` },
    cache: "no-store",
  });
  const json: ApiResponse<AdminConfig> = await res.json();
  if (!res.ok || !json.success) {
    const msg = "error" in json ? json.error.message : `HTTP ${res.status}`;
    throw new Error(msg);
  }
  return json.data;
}

export async function adminSetDefaultStorage(
  token: string,
  storageKey: string
): Promise<AdminConfig> {
  const base = getApiBaseUrl();
  const res = await fetch(`${base}/admin/config/default`, {
    method: "POST",
    headers: adminHeaders(token),
    body: JSON.stringify({ storage_key: storageKey }),
    cache: "no-store",
  });
  const json: ApiResponse<AdminConfig> = await res.json();
  if (!res.ok || !json.success) {
    const msg = "error" in json ? json.error.message : `HTTP ${res.status}`;
    throw new Error(msg);
  }
  return json.data;
}

export async function adminGetSystemSettings(token: string): Promise<AdminSystemSettings> {
  return apiFetch<AdminSystemSettings>("/admin/system-settings", {
    headers: { Authorization: `Bearer ${token}` },
  });
}

export async function adminUpdateSystemSettings(
  token: string,
  settings: RuntimeSettings
): Promise<AdminSystemSettings> {
  const base = getApiBaseUrl();
  const res = await fetch(`${base}/admin/system-settings`, {
    method: "PUT",
    headers: adminHeaders(token),
    body: JSON.stringify(settings),
    cache: "no-store",
  });
  const json: ApiResponse<AdminSystemSettings> = await res.json();
  if (!res.ok || !json.success) {
    const msg = "error" in json ? json.error.message : `HTTP ${res.status}`;
    throw new Error(msg);
  }
  return json.data;
}

export async function adminGetAnnouncements(token: string): Promise<Announcement[]> {
  const data = await apiFetch<AnnouncementListResponse>("/admin/announcements", {
    headers: { Authorization: `Bearer ${token}` },
  });
  return data.items;
}

export async function adminCreateAnnouncement(
  token: string,
  announcement: AnnouncementInput
): Promise<Announcement> {
  const base = getApiBaseUrl();
  const res = await fetch(`${base}/admin/announcements`, {
    method: "POST",
    headers: adminHeaders(token),
    body: JSON.stringify(announcement),
    cache: "no-store",
  });
  const json: ApiResponse<Announcement> = await res.json();
  if (!res.ok || !json.success) {
    const msg = "error" in json ? json.error.message : `HTTP ${res.status}`;
    throw new Error(msg);
  }
  return json.data;
}

export async function adminUpdateAnnouncement(
  token: string,
  id: number,
  announcement: AnnouncementInput
): Promise<Announcement> {
  const base = getApiBaseUrl();
  const res = await fetch(`${base}/admin/announcements/${id}`, {
    method: "PUT",
    headers: adminHeaders(token),
    body: JSON.stringify(announcement),
    cache: "no-store",
  });
  const json: ApiResponse<Announcement> = await res.json();
  if (!res.ok || !json.success) {
    const msg = "error" in json ? json.error.message : `HTTP ${res.status}`;
    throw new Error(msg);
  }
  return json.data;
}

export async function adminDeleteAnnouncement(token: string, id: number): Promise<void> {
  const base = getApiBaseUrl();
  const res = await fetch(`${base}/admin/announcements/${id}`, {
    method: "DELETE",
    headers: { Authorization: `Bearer ${token}` },
    cache: "no-store",
  });
  const json: ApiResponse<Record<string, never>> = await res.json();
  if (!res.ok || !json.success) {
    const msg = "error" in json ? json.error.message : `HTTP ${res.status}`;
    throw new Error(msg);
  }
}

export async function adminArchiveAnnouncement(token: string, id: number): Promise<Announcement> {
  const base = getApiBaseUrl();
  const res = await fetch(`${base}/admin/announcements/${id}/archive`, {
    method: "POST",
    headers: { Authorization: `Bearer ${token}` },
    cache: "no-store",
  });
  const json: ApiResponse<Announcement> = await res.json();
  if (!res.ok || !json.success) {
    const msg = "error" in json ? json.error.message : `HTTP ${res.status}`;
    throw new Error(msg);
  }
  return json.data;
}

