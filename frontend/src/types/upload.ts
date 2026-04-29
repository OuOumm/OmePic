import type { StorageBackend } from "@/types/storage";

export type UploadResponseData = {
  uid: string;
  url: string;
  md_url: string;
  bbcode: string;
  size: number;
  mime_type: string;
  created_at: string;
  duplicate: boolean;
  storage_key: string;
  storage_backend: StorageBackend;
};

export type UploadHistoryRecord = UploadResponseData & {
  token: string;
  original_filename: string;
};
