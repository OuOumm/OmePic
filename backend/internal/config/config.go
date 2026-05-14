package config

import (
	"os"
	"strings"
)

const (
	StorageBackendLocal  = "local"
	StorageBackendS3     = "s3"
	StorageBackendWebDAV = "webdav"
)

type AppConfig struct {
	HTTPAddr         string
	DatabasePath     string
	RedisURL         string
	UIDPrefix        string
	UIDEncryptionKey string
	JWTSecret        string
}

type RuntimeStorageConfig struct {
	StorageKey       string `json:"storage_key"`
	Name             string `json:"name"`
	IsDefault        bool   `json:"is_default"`
	Backend          string `json:"storage_backend"`
	LocalStoragePath string `json:"local_storage_path"`
	S3Endpoint       string `json:"s3_endpoint"`
	S3Region         string `json:"s3_region"`
	S3Bucket         string `json:"s3_bucket"`
	S3AccessKey      string `json:"s3_access_key"`
	S3SecretKey      string `json:"s3_secret_key"`
	S3UseSSL         bool   `json:"s3_use_ssl"`
	S3ForcePathStyle bool   `json:"s3_force_path_style"`
	WebDAVURL        string `json:"webdav_url"`
	WebDAVUser       string `json:"webdav_user"`
	WebDAVPass       string `json:"webdav_pass"`
}

type RuntimeStorageCatalog struct {
	DefaultStorageKey string                 `json:"default_storage_key"`
	StorageConfigs    []RuntimeStorageConfig `json:"storage_configs"`
}

type RuntimeStorageUpdate struct {
	Name             *string `json:"name"`
	Backend          *string `json:"storage_backend"`
	LocalStoragePath *string `json:"local_storage_path"`
	S3Endpoint       *string `json:"s3_endpoint"`
	S3Region         *string `json:"s3_region"`
	S3Bucket         *string `json:"s3_bucket"`
	S3AccessKey      *string `json:"s3_access_key"`
	S3SecretKey      *string `json:"s3_secret_key"`
	S3UseSSL         *bool   `json:"s3_use_ssl"`
	S3ForcePathStyle *bool   `json:"s3_force_path_style"`
	WebDAVURL        *string `json:"webdav_url"`
	WebDAVUser       *string `json:"webdav_user"`
	WebDAVPass       *string `json:"webdav_pass"`
}

func Load() AppConfig {
	return AppConfig{
		HTTPAddr:         envOrDefault("HTTP_ADDR", ":8080"),
		DatabasePath:     envOrDefault("DATABASE_PATH", "data/omepic.db"),
		RedisURL:         envOrDefault("REDIS_URL", "redis://localhost:6379/0"),
		UIDPrefix:        envOrDefault("UID_PREFIX", "omeo_"),
		UIDEncryptionKey: envOrDefault("UID_ENCRYPTION_KEY", "change-me-uid-secret"),
		JWTSecret:        envOrDefault("JWT_SECRET", "change-me-too"),
	}
}

func DefaultStorageConfig() RuntimeStorageConfig {
	return RuntimeStorageConfig{
		StorageKey:       "local-default",
		Name:             "Default Local Storage",
		IsDefault:        true,
		Backend:          "local",
		LocalStoragePath: "data/images",
	}
}

func BootstrapStorageKey(backend string) string {
	switch normalizedStorageBackendOrDefault(backend) {
	case StorageBackendS3:
		return "s3-default"
	case StorageBackendWebDAV:
		return "webdav-default"
	default:
		return "local-default"
	}
}

func BootstrapStorageName(backend string) string {
	switch normalizedStorageBackendOrDefault(backend) {
	case StorageBackendS3:
		return "Default S3 Storage"
	case StorageBackendWebDAV:
		return "Default WebDAV Storage"
	default:
		return "Default Local Storage"
	}
}

func NormalizeStorageBackend(backend string) string {
	return strings.TrimSpace(strings.ToLower(backend))
}

func normalizedStorageBackendOrDefault(backend string) string {
	value := NormalizeStorageBackend(backend)
	if value == "" {
		return StorageBackendLocal
	}
	return value
}

func envOrDefault(key, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	return value
}
