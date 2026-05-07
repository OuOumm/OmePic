import { getApiBaseUrl } from "./utils";

import type {
  ApiResponse,
  UploadResult,
  StorageOption,
  AdminStatus,
  AdminImagesResponse,
  AdminConfig,
  StorageInstance,
} from "@/types";

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
    const msg =
      "error" in json ? json.error.message : `HTTP ${res.status}`;
    throw new Error(msg);
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
            reject(new Error("error" in json ? json.error.message : "Upload failed"));
          }
        } catch {
          reject(new Error("Invalid response from server"));
        }
      } else {
        reject(new Error(`Upload failed: HTTP ${xhr.status}`));
      }
    });

    xhr.addEventListener("error", () => reject(new Error("Network error during upload")));
    xhr.addEventListener("abort", () => reject(new Error("Upload aborted")));

    xhr.open("POST", `${base}/v1/image`);
    xhr.setRequestHeader("X-Token", token);
    xhr.send(formData);
  });
}

// Public endpoints
export async function getStorageOptions(signal?: AbortSignal): Promise<StorageOption[]> {
  const data = await apiFetch<{ items: StorageOption[] }>("/v1/storage-options", { signal });
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
