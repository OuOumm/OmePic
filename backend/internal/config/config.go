package config

import (
	"os"
	"strings"

	"github.com/joho/godotenv"
)

const (
	StorageBackendLocal  = "local"
	StorageBackendS3     = "s3"
	StorageBackendWebDAV = "webdav"
)

type AppConfig struct {
	HTTPAddr              string
	DatabasePath          string
	RedisURL              string
	UIDPrefix             string
	UIDEncryptionKey      string // UID obfuscation key (env: UID_ENCRYPTION_KEY). XOR-based ID obfuscation, not a cryptographic boundary.
	JWTSecret             string
	SecretEncryptionKey   string // AES-256-GCM key (env: SECRET_ENCRYPTION_KEY). Encrypts storage credentials in DB.
	AppEnv                string
	PublicBaseURL          string
}

// IsProduction returns true when APP_ENV is not "development".
func (c AppConfig) IsProduction() bool {
	return !strings.EqualFold(strings.TrimSpace(c.AppEnv), "development")
}

type RuntimeStorageConfig struct {
	StorageKey                  string `json:"storage_key"`
	Name                        string `json:"name"`
	IsDefault                   bool   `json:"is_default"`
	Backend                     string `json:"storage_backend"`
	LocalStoragePath             string `json:"local_storage_path"`
	S3Endpoint                  string `json:"s3_endpoint"`
	S3Region                    string `json:"s3_region"`
	S3Bucket                    string `json:"s3_bucket"`
	S3AccessKey                 string `json:"s3_access_key"`
	S3SecretKey                 string `json:"s3_secret_key"`
	S3UseSSL                    bool   `json:"s3_use_ssl"`
	S3ForcePathStyle            bool   `json:"s3_force_path_style"`
	WebDAVURL                   string `json:"webdav_url"`
	WebDAVUser                  string `json:"webdav_user"`
	WebDAVPass                  string `json:"webdav_pass"`
	MaxUploadSizeMB             int    `json:"max_upload_size_mb"`
	AllowedMIMETypes            string `json:"allowed_mime_types"`
	AvifQuality                 int    `json:"avif_quality"`
	AvifSpeed                   int    `json:"avif_speed"`
	MaxImagePixels              int64  `json:"max_image_pixels"`
	AVIFMaxConcurrency          int    `json:"avif_max_concurrency"`
	AVIFConversionTimeoutSeconds int   `json:"avif_conversion_timeout_seconds"`
}

type RuntimeStorageCatalog struct {
	DefaultStorageKey string                 `json:"default_storage_key"`
	StorageConfigs    []RuntimeStorageConfig `json:"storage_configs"`
}

type RuntimeStorageUpdate struct {
	Name                        *string `json:"name"`
	Backend                     *string `json:"storage_backend"`
	LocalStoragePath             *string `json:"local_storage_path"`
	S3Endpoint                  *string `json:"s3_endpoint"`
	S3Region                    *string `json:"s3_region"`
	S3Bucket                    *string `json:"s3_bucket"`
	S3AccessKey                 *string `json:"s3_access_key"`
	S3SecretKey                 *string `json:"s3_secret_key"`
	S3UseSSL                    *bool   `json:"s3_use_ssl"`
	S3ForcePathStyle            *bool   `json:"s3_force_path_style"`
	WebDAVURL                   *string `json:"webdav_url"`
	WebDAVUser                  *string `json:"webdav_user"`
	WebDAVPass                  *string `json:"webdav_pass"`
	MaxUploadSizeMB             *int    `json:"max_upload_size_mb"`
	AllowedMIMETypes            *string `json:"allowed_mime_types"`
	AvifQuality                 *int    `json:"avif_quality"`
	AvifSpeed                   *int    `json:"avif_speed"`
	MaxImagePixels              *int64  `json:"max_image_pixels"`
	AVIFMaxConcurrency          *int    `json:"avif_max_concurrency"`
	AVIFConversionTimeoutSeconds *int   `json:"avif_conversion_timeout_seconds"`
}

// Load reads configuration from environment variables.
// H-01: ADMIN_PASSWORD, JWT_SECRET, and UID_ENCRYPTION_KEY have no defaults.
// They must be set explicitly in .env; startup will fail if any is missing.
func Load() AppConfig {
	// Load .env file if present (does not override existing env vars).
	// Search from current directory and parent directories.
	godotenv.Load()
	godotenv.Load("../.env")
	return AppConfig{
		HTTPAddr:            envOrDefault("HTTP_ADDR", ":8080"),
		DatabasePath:        envOrDefault("DATABASE_PATH", "data/omepic.db"),
		RedisURL:            envOrDefault("REDIS_URL", "redis://localhost:6379/0"),
		UIDPrefix:           envOrDefault("UID_PREFIX", "omeo_"),
		UIDEncryptionKey:    strings.TrimSpace(os.Getenv("UID_ENCRYPTION_KEY")),
		JWTSecret:           strings.TrimSpace(os.Getenv("JWT_SECRET")),
		SecretEncryptionKey: strings.TrimSpace(os.Getenv("SECRET_ENCRYPTION_KEY")),
		AppEnv:              envOrDefault("APP_ENV", ""),
		PublicBaseURL:       strings.TrimRight(strings.TrimSpace(os.Getenv("PUBLIC_BASE_URL")), "/"),
	}
}

func DefaultStorageConfig() RuntimeStorageConfig {
	return RuntimeStorageConfig{
		StorageKey:                  "local-default",
		Name:                        "Default Local Storage",
		IsDefault:                   true,
		Backend:                     "local",
		LocalStoragePath:            "data/images",
		MaxUploadSizeMB:             20,
		AllowedMIMETypes:            "image/avif,image/gif,image/jpeg,image/png,image/webp",
		AvifQuality:                 60,
		AvifSpeed:                   8,
		MaxImagePixels:              40000000,
		AVIFMaxConcurrency:          2,
		AVIFConversionTimeoutSeconds: 30,
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
