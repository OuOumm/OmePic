import { getApiBaseUrl, getImagePath } from "./utils";

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

type ApiFetchOptions = RequestInit & { params?: Record<string, string> };

type PendingGetRequest = {
  controller: AbortController;
  promise: Promise<unknown>;
  subscribers: number;
  done: boolean;
};

const pendingGetRequests = new Map<string, PendingGetRequest>();

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

function buildApiUrl(path: string, params?: Record<string, string>): string {
  const base = getApiBaseUrl();
  let url = `${base}${path}`;
  if (params) {
    const searchParams = new URLSearchParams(params);
    url += `?${searchParams.toString()}`;
  }

  return url;
}

function requestMethod(options: RequestInit): string {
  return (options.method ?? "GET").toUpperCase();
}

function normalizedHeaders(headers: HeadersInit | undefined): [string, string][] {
  return Array.from(new Headers(headers).entries()).sort(([left], [right]) => left.localeCompare(right));
}

function pendingRequestKey(method: string, url: string, options: RequestInit): string {
  return JSON.stringify([method, url, normalizedHeaders(options.headers), options.credentials ?? ""]);
}

function abortError(): DOMException {
  return new DOMException("The operation was aborted.", "AbortError");
}

function attachPendingRequest<T>(entry: PendingGetRequest, signal?: AbortSignal | null): Promise<T> {
  if (signal?.aborted) return Promise.reject(abortError());

  entry.subscribers += 1;
  return new Promise<T>((resolve, reject) => {
    let settled = false;
    const cleanup = () => {
      if (settled) return;
      settled = true;
      signal?.removeEventListener("abort", handleAbort);
      entry.subscribers -= 1;
      if (entry.subscribers === 0 && !entry.done) entry.controller.abort();
    };
    const handleAbort = () => {
      cleanup();
      reject(abortError());
    };

    signal?.addEventListener("abort", handleAbort, { once: true });
    entry.promise.then(
      (value) => {
        cleanup();
        resolve(value as T);
      },
      (err) => {
        cleanup();
        reject(err);
      }
    );
  });
}

async function fetchApiResponse<T>(url: string, requestOptions: RequestInit, signal?: AbortSignal | null): Promise<T> {
  const fetchOptions: RequestInit = signal
    ? { cache: "no-store", ...requestOptions, signal }
    : { cache: "no-store", ...requestOptions };
  const res = await fetch(url, fetchOptions);
  let json: ApiResponse<T>;
  try {
    json = await res.json() as ApiResponse<T>;
  } catch {
    throw new ApiError(res.ok ? "Invalid response from server" : `HTTP ${res.status}`, { status: res.status });
  }

  if (!res.ok || !json.success) {
    throw apiErrorFromResponse(json, res.status, `HTTP ${res.status}`);
  }
  return json.data as T;
}

async function apiFetch<T>(path: string, options: ApiFetchOptions = {}): Promise<T> {
  const { params, ...requestOptions } = options;
  const url = buildApiUrl(path, params);
  const method = requestMethod(requestOptions);

  if (method === "GET") {
    if (requestOptions.signal?.aborted) return Promise.reject(abortError());
    const key = pendingRequestKey(method, url, requestOptions);
    let entry = pendingGetRequests.get(key);
    if (!entry) {
      const controller = new AbortController();
      const pending: PendingGetRequest = {
        controller,
        promise: Promise.resolve(),
        subscribers: 0,
        done: false,
      };
      pending.promise = fetchApiResponse<T>(url, requestOptions, controller.signal)
        .finally(() => {
          pending.done = true;
          pendingGetRequests.delete(key);
        });
      entry = pending;
      pendingGetRequests.set(key, entry);
    }
    return attachPendingRequest<T>(entry, requestOptions.signal);
  }

  return fetchApiResponse<T>(url, requestOptions, requestOptions.signal);
}

function apiErrorFromResponse<T>(
  json: ApiResponse<T>,
  status: number,
  fallbackMessage: string,
  retryAfter?: number
): ApiError {
  if ("error" in json) {
    return new ApiError(json.error.message, { code: json.error.code, status, retryAfter });
  }
  return new ApiError(fallbackMessage, { status, retryAfter });
}

function uploadResponseError(xhr: XMLHttpRequest): ApiError {
  const retryAfter = parseRetryAfter(xhr.getResponseHeader("Retry-After"));
  try {
    const json: ApiResponse<UploadResult> = JSON.parse(xhr.responseText);
    return apiErrorFromResponse(json, xhr.status, `Upload failed: HTTP ${xhr.status}`, retryAfter);
  } catch {
    return new ApiError(`Upload failed: HTTP ${xhr.status}`, { status: xhr.status, retryAfter });
  }
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
      if (xhr.status < 200 || xhr.status >= 300) {
        reject(uploadResponseError(xhr));
        return;
      }

      let json: ApiResponse<UploadResult>;
      try {
        json = JSON.parse(xhr.responseText);
      } catch {
        reject(new Error("Invalid response from server"));
        return;
      }

      if (json.success) {
        resolve(json.data as UploadResult);
        return;
      }

      reject(apiErrorFromResponse(json, xhr.status, "Upload failed", parseRetryAfter(xhr.getResponseHeader("Retry-After"))));
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

export async function getRuntimeSettings(signal?: AbortSignal): Promise<PublicRuntimeSettings> {
  return apiFetch<PublicRuntimeSettings>("/v1/runtime-settings", { signal });
}

export async function getAnnouncements(signal?: AbortSignal): Promise<Announcement[]> {
  const data = await apiFetch<AnnouncementListResponse>("/v1/announcements", { signal });
  return data.items;
}

export async function deleteImageByUid(uid: string, token: string): Promise<void> {
  await apiFetch<Record<string, never> | null>(getImagePath(uid), {
    method: "DELETE",
    headers: { "X-Token": token },
  });
}

function adminAuthHeaders(token: string): HeadersInit {
  return {
    Authorization: `Bearer ${token}`,
  };
}

function adminHeaders(token: string): HeadersInit {
  return {
    ...adminAuthHeaders(token),
    "Content-Type": "application/json",
  };
}

export async function adminLogin(password: string, signal?: AbortSignal): Promise<string> {
  const data = await apiFetch<{ token: string }>("/admin/login", {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ password }),
    signal,
  });
  return data.token;
}

export async function adminGetStatus(token: string, signal?: AbortSignal): Promise<AdminStatus> {
  return apiFetch<AdminStatus>("/admin/status", {
    headers: adminAuthHeaders(token),
    signal,
  });
}

export async function adminGetImages(
  token: string,
  page: number,
  pageSize: number,
  search?: string,
  signal?: AbortSignal
): Promise<AdminImagesResponse> {
  return apiFetch<AdminImagesResponse>("/admin/images", {
    headers: adminHeaders(token),
    signal,
    params: {
      page: String(page),
      pageSize: String(pageSize),
      ...(search ? { search } : {}),
    },
  });
}

export async function adminDeleteImages(token: string, uids: string[]): Promise<void> {
  await apiFetch<Record<string, never> | null>("/admin/images", {
    method: "DELETE",
    headers: adminHeaders(token),
    body: JSON.stringify({ uids }),
  });
}

export async function adminCreateIPBan(
  token: string,
  input: { uid?: string; ip_address?: string; duration_hours: number | null; reason?: string }
): Promise<AdminIPBanCreateResult> {
  return apiFetch<AdminIPBanCreateResult>("/admin/ip-bans", {
    method: "POST",
    headers: adminHeaders(token),
    body: JSON.stringify(input),
  });
}

export async function adminGetIPBans(token: string, signal?: AbortSignal): Promise<AdminIPBan[]> {
  return apiFetch<AdminIPBan[]>("/admin/ip-bans", {
    headers: adminAuthHeaders(token),
    signal,
  });
}

export async function adminDeleteIPBan(token: string, id: number): Promise<void> {
  await apiFetch<Record<string, never>>(`/admin/ip-bans/${id}`, {
    method: "DELETE",
    headers: adminAuthHeaders(token),
  });
}

export async function adminDeleteIPBanImages(
  token: string,
  id: number
): Promise<AdminIPBanDeleteImagesResult> {
  return apiFetch<AdminIPBanDeleteImagesResult>(`/admin/ip-bans/${id}/images`, {
    method: "DELETE",
    headers: adminAuthHeaders(token),
  });
}

export async function adminGetAbuseOverview(
  token: string,
  from?: string,
  to?: string,
  signal?: AbortSignal
): Promise<AdminAbuseOverview> {
  return apiFetch<AdminAbuseOverview>("/admin/abuse/overview", {
    headers: adminAuthHeaders(token),
    signal,
    params: {
      ...(from ? { from } : {}),
      ...(to ? { to } : {}),
    },
  });
}

export async function adminGetAbuseIPDetail(token: string, ip: string, signal?: AbortSignal): Promise<AdminAbuseIPDetail> {
  return apiFetch<AdminAbuseIPDetail>("/admin/abuse/ip", {
    headers: adminAuthHeaders(token),
    signal,
    params: { ip },
  });
}

export async function adminGetConfig(token: string, signal?: AbortSignal): Promise<AdminConfig> {
  return apiFetch<AdminConfig>("/admin/config", {
    headers: adminAuthHeaders(token),
    signal,
  });
}

export async function adminCreateStorageInstance(
  token: string,
  instance: Partial<StorageInstance>
): Promise<AdminConfig> {
  return apiFetch<AdminConfig>("/admin/config/storage-instances", {
    method: "POST",
    headers: adminHeaders(token),
    body: JSON.stringify(instance),
  });
}

export async function adminUpdateStorageInstance(
  token: string,
  storageKey: string,
  instance: Partial<StorageInstance>
): Promise<AdminConfig> {
  return apiFetch<AdminConfig>(`/admin/config/storage-instances/${storageKey}`, {
    method: "PUT",
    headers: adminHeaders(token),
    body: JSON.stringify(instance),
  });
}

export async function adminDeleteStorageInstance(
  token: string,
  storageKey: string
): Promise<AdminConfig> {
  return apiFetch<AdminConfig>(`/admin/config/storage-instances/${storageKey}`, {
    method: "DELETE",
    headers: adminAuthHeaders(token),
  });
}

export async function adminSetDefaultStorage(
  token: string,
  storageKey: string
): Promise<AdminConfig> {
  return apiFetch<AdminConfig>("/admin/config/default", {
    method: "POST",
    headers: adminHeaders(token),
    body: JSON.stringify({ storage_key: storageKey }),
  });
}

export async function adminGetSystemSettings(token: string, signal?: AbortSignal): Promise<AdminSystemSettings> {
  return apiFetch<AdminSystemSettings>("/admin/system-settings", {
    headers: adminAuthHeaders(token),
    signal,
  });
}

export async function adminUpdateSystemSettings(
  token: string,
  settings: RuntimeSettings
): Promise<AdminSystemSettings> {
  return apiFetch<AdminSystemSettings>("/admin/system-settings", {
    method: "PUT",
    headers: adminHeaders(token),
    body: JSON.stringify(settings),
  });
}

export async function adminGetAnnouncements(token: string, signal?: AbortSignal): Promise<Announcement[]> {
  const data = await apiFetch<AnnouncementListResponse>("/admin/announcements", {
    headers: adminAuthHeaders(token),
    signal,
  });
  return data.items;
}

export async function adminCreateAnnouncement(
  token: string,
  announcement: AnnouncementInput
): Promise<Announcement> {
  return apiFetch<Announcement>("/admin/announcements", {
    method: "POST",
    headers: adminHeaders(token),
    body: JSON.stringify(announcement),
  });
}

export async function adminUpdateAnnouncement(
  token: string,
  id: number,
  announcement: AnnouncementInput
): Promise<Announcement> {
  return apiFetch<Announcement>(`/admin/announcements/${id}`, {
    method: "PUT",
    headers: adminHeaders(token),
    body: JSON.stringify(announcement),
  });
}

export async function adminDeleteAnnouncement(token: string, id: number): Promise<void> {
  await apiFetch<Record<string, never>>(`/admin/announcements/${id}`, {
    method: "DELETE",
    headers: adminAuthHeaders(token),
  });
}

export async function adminArchiveAnnouncement(token: string, id: number): Promise<Announcement> {
  return apiFetch<Announcement>(`/admin/announcements/${id}/archive`, {
    method: "POST",
    headers: adminAuthHeaders(token),
  });
}
