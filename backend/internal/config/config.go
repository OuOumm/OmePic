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
	PublicBaseURL    string
	UIDPrefix        string
	UIDEncryptionKey string
	StorageBackend   string
	LocalStoragePath string
	AdminPassword    string
	JWTSecret        string
	S3               S3Config
	WebDAV           WebDAVConfig
}

type S3Config struct {
	Endpoint       string
	Region         string
	Bucket         string
	AccessKey      string
	SecretKey      string
	UseSSL         bool
	ForcePathStyle bool
}

type WebDAVConfig struct {
	URL      string
	Username string
	Password string
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
	uidEncryptionKey := os.Getenv("UID_ENCRYPTION_KEY")
	if uidEncryptionKey == "" {
		uidEncryptionKey = envOrDefault("JWT_SECRET", "change-me")
	}

	return AppConfig{
		HTTPAddr:         envOrDefault("HTTP_ADDR", ":8080"),
		DatabasePath:     envOrDefault("DATABASE_PATH", "data/omepic.db"),
		RedisURL:         envOrDefault("REDIS_URL", "redis://localhost:6379/0"),
		PublicBaseURL:    os.Getenv("PUBLIC_BASE_URL"),
		UIDPrefix:        envOrDefault("UID_PREFIX", "omeo_"),
		UIDEncryptionKey: uidEncryptionKey,
		StorageBackend:   envOrDefault("STORAGE_BACKEND", StorageBackendLocal),
		LocalStoragePath: envOrDefault("LOCAL_STORAGE_PATH", "data/images"),
		AdminPassword:    envOrDefault("ADMIN_PASSWORD", "admin123"),
		JWTSecret:        envOrDefault("JWT_SECRET", "change-me"),
		S3: S3Config{
			Endpoint:       os.Getenv("S3_ENDPOINT"),
			Region:         envOrDefault("S3_REGION", "auto"),
			Bucket:         os.Getenv("S3_BUCKET"),
			AccessKey:      os.Getenv("S3_ACCESS_KEY"),
			SecretKey:      os.Getenv("S3_SECRET_KEY"),
			UseSSL:         envBool("S3_USE_SSL", false),
			ForcePathStyle: envBool("S3_FORCE_PATH_STYLE", true),
		},
		WebDAV: WebDAVConfig{
			URL:      os.Getenv("WEBDAV_URL"),
			Username: os.Getenv("WEBDAV_USER"),
			Password: os.Getenv("WEBDAV_PASS"),
		},
	}
}

func (c AppConfig) DefaultStorageConfig() RuntimeStorageConfig {
	return RuntimeStorageConfig{
		StorageKey:       BootstrapStorageKey(c.StorageBackend),
		Name:             BootstrapStorageName(c.StorageBackend),
		IsDefault:        true,
		Backend:          c.StorageBackend,
		LocalStoragePath: c.LocalStoragePath,
		S3Endpoint:       c.S3.Endpoint,
		S3Region:         c.S3.Region,
		S3Bucket:         c.S3.Bucket,
		S3AccessKey:      c.S3.AccessKey,
		S3SecretKey:      c.S3.SecretKey,
		S3UseSSL:         c.S3.UseSSL,
		S3ForcePathStyle: c.S3.ForcePathStyle,
		WebDAVURL:        c.WebDAV.URL,
		WebDAVUser:       c.WebDAV.Username,
		WebDAVPass:       c.WebDAV.Password,
	}
}

func BootstrapStorageKey(backend string) string {
	switch normalizeBackend(backend) {
	case StorageBackendS3:
		return "s3-default"
	case StorageBackendWebDAV:
		return "webdav-default"
	default:
		return "local-default"
	}
}

func BootstrapStorageName(backend string) string {
	switch normalizeBackend(backend) {
	case StorageBackendS3:
		return "Default S3 Storage"
	case StorageBackendWebDAV:
		return "Default WebDAV Storage"
	default:
		return "Default Local Storage"
	}
}

func normalizeBackend(backend string) string {
	value := strings.TrimSpace(strings.ToLower(backend))
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

func envBool(key string, fallback bool) bool {
	value := os.Getenv(key)
	switch value {
	case "1", "true", "TRUE", "yes", "YES":
		return true
	case "0", "false", "FALSE", "no", "NO":
		return false
	default:
		return fallback
	}
}
