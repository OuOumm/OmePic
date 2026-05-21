export type ApiResponse<T> =
  | { success: true; data: T }
  | { success: false; error: { code?: string; message: string } };

export interface UploadResult {
  url: string;
  duplicate: boolean;
}

export interface StorageOption {
  storage_key: string;
  name: string;
  storage_backend: string;
  is_default: boolean;
}

export interface AdminStatus {
  total_images: number;
  total_storage_size: number;
  today_uploads: number;
  unique_tokens: number;
}

export interface AdminImage {
  id: number;
  uid: string;
  token: string;
  storage_key: string;
  storage_backend: string;
  mime_type: string;
  size: number;
  md5_hash: string;
  ip_address: string;
  ip_address_masked: string;
  created_at: string;
}

export interface AdminIPBan {
  id: number;
  ip_hash: string;
  ip_address: string;
  ip_address_masked: string;
  reason: string;
  expires_at: string | null;
  created_at: string;
  updated_at: string;
}

export interface AdminIPBanCreateResult {
  ban: AdminIPBan;
  affected_image_count: number;
  affected_total_size: number;
}

export interface AdminIPBanDeleteImagesResult {
  deleted_count: number;
}

export interface CloudflareImageCachePurgeResult {
  url: string;
}

export interface AdminAbuseOverview {
  from: string;
  to: string;
  upload_count: number;
  upload_size: number;
  active_ip_ban_count: number;
  top_ips: AdminAbuseIPRankItem[];
  top_tokens: AdminAbuseTokenRankItem[];
}

export interface AdminAbuseIPRankItem {
  ip_address: string;
  ip_address_masked: string;
  upload_count: number;
  total_size: number;
  latest_upload_at: string;
  is_banned: boolean;
  ban_id?: number;
}

export interface AdminAbuseTokenRankItem {
  token: string;
  token_preview: string;
  upload_count: number;
  total_size: number;
  latest_upload_at: string;
}

export interface AdminAbuseIPDetail {
  ip_address: string;
  ip_address_masked: string;
  upload_count: number;
  total_size: number;
  is_banned: boolean;
  ban: AdminIPBan | null;
}

export interface AdminImagesResponse {
  items: AdminImage[];
  total: number;
  page: number;
  page_size: number;
}

export interface StorageInstance {
  storage_key: string;
  name: string;
  is_default: boolean;
  storage_backend: "local" | "s3" | "webdav";
  local_storage_path?: string;
  s3_endpoint?: string;
  s3_region?: string;
  s3_bucket?: string;
  s3_access_key?: string;
  s3_secret_key?: string;
  s3_use_ssl?: boolean;
  s3_force_path_style?: boolean;
  webdav_url?: string;
  webdav_user?: string;
  webdav_pass?: string;
}

export interface AdminConfig {
  default_storage_key: string;
  storage_configs: StorageInstance[];
}

export interface RuntimeSettings {
  site_name: string;
  site_tagline: string;
  public_base_url: string;
  cloudflare_purge_enabled: boolean;
  cloudflare_zone_id: string;
  cloudflare_api_token: string;
  cloudflare_api_base_url: string;
  max_upload_size_mb: number;
  allowed_mime_types: string[];
  avif_quality: number;
  avif_speed: number;
  allow_storage_selection: boolean;
  maintenance_mode: boolean;
  maintenance_message: string;
  rate_limit_window_minutes: number;
  rate_limit_max_requests: number;
  upload_rate_limit_window_minutes: number;
  upload_rate_limit_max_requests: number;
}

export interface PublicRuntimeSettings {
  site: {
    name: string;
    tagline: string;
  };
  access: {
    public_base_url: string;
  };
  upload: {
    max_upload_size_mb: number;
    allowed_mime_types: string[];
  };
  features: {
    allow_storage_selection: boolean;
    maintenance_mode: boolean;
    maintenance_message: string;
  };
  storage: {
    options: StorageOption[];
  };
}

export interface AdminSystemSettings {
  runtime: RuntimeSettings;
  readonly: {
    environment: {
      http_addr: string;
      database_path: string;
      redis_configured: boolean;
      public_base_url_source: string;
      runtime_public_base_url_set: boolean;
    };
    security: {
      jwt_secret: SecretStatus;
      admin_password: SecretStatus;
      uid_encryption_key: SecretStatus;
    };
    storage: {
      default_storage_key: string;
      storage_config_count: number;
      allow_storage_selection: boolean;
    };
    service: {
      health: string;
      maintenance_mode: boolean;
      cloudflare_purge_configured: boolean;
    };
  };
}

export interface SecretStatus {
  configured: boolean;
  using_default: boolean;
}

export type AnnouncementStatus = "draft" | "published" | "archived";
export type AnnouncementPriority = "normal" | "important" | "urgent";

export interface Announcement {
  id: number;
  title: string;
  content: string;
  status?: AnnouncementStatus;
  priority: AnnouncementPriority;
  starts_at: string | null;
  ends_at: string | null;
  sort_order?: number;
  created_at: string;
  updated_at: string;
}

export interface AnnouncementListResponse {
  items: Announcement[];
}

export interface AnnouncementInput {
  title: string;
  content: string;
  status: AnnouncementStatus;
  priority: AnnouncementPriority;
  starts_at: string | null;
  ends_at: string | null;
  sort_order: number;
}

export interface UploadHistoryRecord {
  uid: string;
  url: string;
  mime_type: string;
  size: number;
  created_at: string;
  is_duplicate: boolean;
  storage_key: string;
  storage_backend: string;
  markdown: string;
  bbcode: string;
  client_token: string;
  original_filename: string;
  saved_at: string;
}

export type Language = "en" | "zh";
export type Theme = "light" | "dark" | "system";
export type ViewMode = "grid" | "list";
