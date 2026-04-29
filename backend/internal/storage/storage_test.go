package storage

import (
	"path"
	"path/filepath"
	"testing"

	"omepic/backend/internal/config"
)

func TestManagerResolvesHistoricalStorageKeyAfterReconfigure(t *testing.T) {
	rootDir := filepath.Join(t.TempDir(), "images")
	manager, err := NewManager([]config.RuntimeStorageConfig{
		{
			StorageKey:       "local-primary",
			Name:             "Local Primary",
			IsDefault:        true,
			Backend:          config.StorageBackendLocal,
			LocalStoragePath: rootDir,
		},
	})
	if err != nil {
		t.Fatalf("NewManager returned error: %v", err)
	}

	if err := manager.Reconfigure([]config.RuntimeStorageConfig{
		{
			StorageKey:       "local-primary",
			Name:             "Local Primary",
			Backend:          config.StorageBackendLocal,
			LocalStoragePath: rootDir,
		},
		{
			StorageKey:       "s3-primary",
			Name:             "S3 Primary",
			IsDefault:        true,
			Backend:          config.StorageBackendS3,
			LocalStoragePath: rootDir,
			S3Endpoint:       "127.0.0.1:9000",
			S3Bucket:         "omepic",
			S3AccessKey:      "access",
			S3SecretKey:      "secret",
			S3Region:         "auto",
			S3ForcePathStyle: true,
		},
	}); err != nil {
		t.Fatalf("Reconfigure returned error: %v", err)
	}

	localProvider, err := manager.ForKey("local-primary")
	if err != nil {
		t.Fatalf("ForKey(local-primary) returned error: %v", err)
	}
	if localProvider.Config.Backend != config.StorageBackendLocal || localProvider.Provider.Name() != config.StorageBackendLocal {
		t.Fatalf("expected local provider, got backend=%s provider=%s", localProvider.Config.Backend, localProvider.Provider.Name())
	}

	currentProvider, err := manager.Current()
	if err != nil {
		t.Fatalf("Current returned error: %v", err)
	}
	if currentProvider.Config.StorageKey != "s3-primary" {
		t.Fatalf("expected current storage key s3-primary, got %s", currentProvider.Config.StorageKey)
	}
	if manager.CurrentBackend() != config.StorageBackendS3 {
		t.Fatalf("expected current backend s3, got %s", manager.CurrentBackend())
	}
}

func TestManagerAllowsMultipleInstancesOfSameBackend(t *testing.T) {
	manager, err := NewManager([]config.RuntimeStorageConfig{
		{
			StorageKey:       "s3-one",
			Name:             "S3 One",
			IsDefault:        true,
			Backend:          config.StorageBackendS3,
			S3Endpoint:       "127.0.0.1:9000",
			S3Bucket:         "bucket-a",
			S3AccessKey:      "access-a",
			S3SecretKey:      "secret-a",
			S3Region:         "auto",
			S3ForcePathStyle: true,
		},
		{
			StorageKey:       "s3-two",
			Name:             "S3 Two",
			Backend:          config.StorageBackendS3,
			S3Endpoint:       "127.0.0.1:9000",
			S3Bucket:         "bucket-b",
			S3AccessKey:      "access-b",
			S3SecretKey:      "secret-b",
			S3Region:         "auto",
			S3ForcePathStyle: true,
		},
	})
	if err != nil {
		t.Fatalf("NewManager returned error: %v", err)
	}

	first, err := manager.ForKey("s3-one")
	if err != nil {
		t.Fatalf("ForKey(s3-one) returned error: %v", err)
	}
	second, err := manager.ForKey("s3-two")
	if err != nil {
		t.Fatalf("ForKey(s3-two) returned error: %v", err)
	}

	if first.Config.StorageKey == second.Config.StorageKey {
		t.Fatalf("expected distinct storage keys")
	}
	if first.Provider == second.Provider {
		t.Fatalf("expected distinct provider instances per storage key")
	}
}

func TestBuildObjectKeyUsesUIDAsFilenameBase(t *testing.T) {
	objectKey := BuildObjectKey("uid-123", ".avif")

	if got := path.Base(objectKey); got != "uid-123.avif" {
		t.Fatalf("expected uid-based object key basename, got %q", got)
	}
}
