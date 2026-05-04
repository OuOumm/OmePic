import type {
  AdminConfig,
  AdminImageList,
  AdminStatus,
  AdminStorageConfigCreateInput,
  AdminStorageConfigUpdateInput
} from "@/types/admin";
import type { ApiResponse } from "@/types/api";
import type { PublicStorageOptionsResponse } from "@/types/storage";
import type { UploadResponseData } from "@/types/upload";

const configuredApiBase = process.env.NEXT_PUBLIC_API_BASE_URL?.trim();

function resolveApiBase() {
  if (configuredApiBase) {
    return configuredApiBase.replace(/\/+$/, "");
  }
  if (process.env.NODE_ENV === "development") {
    return "http://localhost:8080";
  }
  return "";
}

const API_BASE = resolveApiBase();

type RequestOptions = {
  method?: "GET" | "POST" | "PUT" | "DELETE";
  body?: BodyInit | null;
  headers?: HeadersInit;
  signal?: AbortSignal;
};

export function apiUrl(path: string) {
  return `${API_BASE}${path}`;
}

export async function requestJson<T>(path: string, options: RequestOptions = {}): Promise<T> {
  const response = await fetch(apiUrl(path), {
    method: options.method || "GET",
    body: options.body,
    headers: options.headers,
    cache: "no-store",
    signal: options.signal
  });

  const payload = (await response.json()) as ApiResponse<T>;
  if (!response.ok || !payload.success) {
    const message = payload.success ? "Request failed" : payload.error.message;
    throw new Error(message);
  }
  return payload.data;
}

export function uploadImageWithProgress(
  file: File,
  token: string,
  onProgress: (progress: number) => void,
  storageKey?: string
) {
  return new Promise<UploadResponseData>((resolve, reject) => {
    const xhr = new XMLHttpRequest();
    xhr.open("POST", apiUrl("/v1/image"));
    xhr.setRequestHeader("X-Token", token);
    xhr.upload.onprogress = (event) => {
      if (!event.lengthComputable) {
        return;
      }
      onProgress(Math.round((event.loaded / event.total) * 100));
    };
    xhr.onload = () => {
      try {
        const payload = JSON.parse(xhr.responseText) as ApiResponse<UploadResponseData>;
        if (xhr.status >= 200 && xhr.status < 300 && payload.success) {
          resolve(payload.data);
          return;
        }
        reject(new Error(payload.success ? "Upload failed" : payload.error.message));
      } catch {
        reject(new Error("Upload failed"));
      }
    };
    xhr.onerror = () => reject(new Error("Upload failed"));
    const data = new FormData();
    data.append("file", file);
    if (storageKey) {
      data.append("storage_key", storageKey);
    }
    xhr.send(data);
  });
}

export async function publicStorageOptions(signal?: AbortSignal) {
  return requestJson<PublicStorageOptionsResponse>("/v1/storage-options", { signal });
}

export async function deleteImage(uid: string, token: string) {
  await requestJson<Record<string, never>>(`/i/${uid}.avif`, {
    method: "DELETE",
    headers: {
      "X-Token": token
    }
  });
}

export async function adminLogin(password: string) {
  const result = await requestJson<{ token: string }>("/admin/login", {
    method: "POST",
    body: JSON.stringify({ password }),
    headers: {
      "Content-Type": "application/json"
    }
  });
  return result.token;
}

export async function adminStatus(token: string) {
  return requestJson<AdminStatus>("/admin/status", {
    headers: adminHeaders(token)
  });
}

export async function adminImages(token: string, search: string, page: number, signal?: AbortSignal) {
  const params = new URLSearchParams({
    page: String(page),
    pageSize: "20",
    search
  });
  return requestJson<AdminImageList>(`/admin/images?${params.toString()}`, {
    headers: adminHeaders(token),
    signal
  });
}

export async function adminDeleteImages(token: string, uids: string[]) {
  return requestJson<Record<string, never>>("/admin/images", {
    method: "DELETE",
    body: JSON.stringify({ uids }),
    headers: {
      ...adminHeaders(token),
      "Content-Type": "application/json"
    }
  });
}

export async function adminGetConfig(token: string) {
  return requestJson<AdminConfig>("/admin/config", {
    headers: adminHeaders(token)
  });
}

export async function adminCreateStorageConfig(token: string, config: AdminStorageConfigCreateInput) {
  return requestJson<AdminConfig>("/admin/config/storage-instances", {
    method: "POST",
    body: JSON.stringify(config),
    headers: {
      ...adminHeaders(token),
      "Content-Type": "application/json"
    }
  });
}

export async function adminUpdateStorageConfig(
  token: string,
  storageKey: string,
  config: AdminStorageConfigUpdateInput
) {
  return requestJson<AdminConfig>(`/admin/config/storage-instances/${storageKey}`, {
    method: "PUT",
    body: JSON.stringify(config),
    headers: {
      ...adminHeaders(token),
      "Content-Type": "application/json"
    }
  });
}

export async function adminDeleteStorageConfig(token: string, storageKey: string) {
  return requestJson<AdminConfig>(`/admin/config/storage-instances/${storageKey}`, {
    method: "DELETE",
    headers: adminHeaders(token)
  });
}

export async function adminSetDefaultStorageConfig(token: string, storageKey: string) {
  return requestJson<AdminConfig>("/admin/config/default", {
    method: "POST",
    body: JSON.stringify({ storage_key: storageKey }),
    headers: {
      ...adminHeaders(token),
      "Content-Type": "application/json"
    }
  });
}

export function adminHeaders(token: string): HeadersInit {
  return {
    Authorization: `Bearer ${token}`
  };
}
