import type { RuntimeSettings } from "@/types";

export const DEFAULT_RUNTIME_SETTINGS: RuntimeSettings = {
  public_base_url: "",
  max_upload_size_mb: 0,
  allowed_mime_types: [],
  allow_storage_selection: true,
  maintenance_mode: false,
  maintenance_message: "",
  rate_limit_window_minutes: 1,
  rate_limit_max_requests: 120,
  upload_rate_limit_window_minutes: 10,
  upload_rate_limit_max_requests: 20,
};

export function normalizeRuntimeSettings(settings?: Partial<RuntimeSettings> | null): RuntimeSettings {
  return {
    public_base_url: settings?.public_base_url ?? "",
    max_upload_size_mb: settings?.max_upload_size_mb ?? 0,
    allowed_mime_types: Array.isArray(settings?.allowed_mime_types) ? settings.allowed_mime_types : [],
    allow_storage_selection: settings?.allow_storage_selection ?? true,
    maintenance_mode: settings?.maintenance_mode ?? false,
    maintenance_message: settings?.maintenance_message ?? "",
    rate_limit_window_minutes: settings?.rate_limit_window_minutes ?? 1,
    rate_limit_max_requests: settings?.rate_limit_max_requests ?? 120,
    upload_rate_limit_window_minutes: settings?.upload_rate_limit_window_minutes ?? 10,
    upload_rate_limit_max_requests: settings?.upload_rate_limit_max_requests ?? 20,
  };
}
