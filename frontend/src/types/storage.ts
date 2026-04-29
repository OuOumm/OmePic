export type StorageBackend = "local" | "s3" | "webdav";

export type PublicStorageOption = {
  storage_key: string;
  name: string;
  storage_backend: StorageBackend;
  is_default: boolean;
};

export type PublicStorageOptionsResponse = {
  items: PublicStorageOption[];
};
