export type ApiResponse<T> =
  | { success: true; data: T }
  | { success: false; error: { message: string } };

export interface UploadResult {
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
  created_at: string;
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
