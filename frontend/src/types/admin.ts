import type { StorageBackend } from "@/types/storage";

export type { StorageBackend } from "@/types/storage";

export type AdminStatus = {
  total_images: number;
  total_storage_size: number;
  today_uploads: number;
  unique_tokens: number;
};

export type AdminImageItem = {
  id: number;
  uid: string;
  token: string;
  storage_key: string;
  storage_backend: StorageBackend;
  mime_type: string;
  size: number;
  md5_hash: string;
  ip_address: string;
  created_at: string;
};

export type AdminImageList = {
  items: AdminImageItem[];
  page: number;
  page_size: number;
  total: number;
};

export type AdminStorageConfig = {
  storage_key: string;
  name: string;
  is_default: boolean;
  storage_backend: StorageBackend;
  local_storage_path: string;
  s3_endpoint: string;
  s3_region: string;
  s3_bucket: string;
  s3_access_key: string;
  s3_secret_key: string;
  s3_use_ssl: boolean;
  s3_force_path_style: boolean;
  webdav_url: string;
  webdav_user: string;
  webdav_pass: string;
};

export type AdminConfig = {
  default_storage_key: string;
  storage_configs: AdminStorageConfig[];
};

export type AdminStorageConfigCreateInput = Omit<AdminStorageConfig, "storage_key" | "is_default">;

export type AdminStorageConfigUpdateInput = Partial<Omit<AdminStorageConfig, "storage_key" | "is_default">>;
