package service

import (
	"context"
	"log/slog"
	"path/filepath"
	"strings"
	"testing"

	"omepic/backend/internal/config"
	"omepic/backend/internal/model"
	"omepic/backend/internal/repository"
	"omepic/backend/internal/storage"
)

func TestGetConfigMasksSecrets(t *testing.T) {
	ctx := context.Background()
	adminService, repo := newAdminServiceTestHarness(t)

	if err := repo.CreateStorageConfig(ctx, config.RuntimeStorageConfig{
		StorageKey:       "s3-secondary",
		Name:             "S3 Secondary",
		Backend:          config.StorageBackendS3,
		S3Endpoint:       "127.0.0.1:9000",
		S3Region:         "auto",
		S3Bucket:         "bucket-b",
		S3AccessKey:      "access-b",
		S3SecretKey:      "secret-b",
		S3ForcePathStyle: true,
	}); err != nil {
		t.Fatalf("CreateStorageConfig returned error: %v", err)
	}

	view, err := adminService.GetConfig(ctx)
	if err != nil {
		t.Fatalf("GetConfig returned error: %v", err)
	}
	if view.DefaultStorageKey != "local-default" {
		t.Fatalf("expected default storage key local-default, got %q", view.DefaultStorageKey)
	}

	var found bool
	for _, item := range view.StorageConfigs {
		if item.StorageKey != "s3-secondary" {
			continue
		}
		found = true
		if item.S3AccessKey == "access-b" || item.S3SecretKey == "secret-b" {
			t.Fatalf("expected secrets to be masked, got access=%q secret=%q", item.S3AccessKey, item.S3SecretKey)
		}
	}
	if !found {
		t.Fatalf("expected s3-secondary to appear in config view")
	}
}

func TestDeleteStorageConfigRejectsDefaultAndInUseInstances(t *testing.T) {
	ctx := context.Background()
	adminService, repo := newAdminServiceTestHarness(t)

	if _, err := adminService.DeleteStorageConfig(ctx, "local-default"); err == nil || !containsError(err, ErrConflict) {
		t.Fatalf("expected ErrConflict for deleting default storage, got %v", err)
	}

	if err := repo.CreateStorageConfig(ctx, config.RuntimeStorageConfig{
		StorageKey:       "s3-secondary",
		Name:             "S3 Secondary",
		Backend:          config.StorageBackendS3,
		S3Endpoint:       "127.0.0.1:9000",
		S3Region:         "auto",
		S3Bucket:         "bucket-b",
		S3AccessKey:      "access-b",
		S3SecretKey:      "secret-b",
		S3ForcePathStyle: true,
	}); err != nil {
		t.Fatalf("CreateStorageConfig returned error: %v", err)
	}

	if err := repo.InsertImage(ctx, modelImageRecord("uid-1", "s3-secondary", config.StorageBackendS3)); err != nil {
		t.Fatalf("InsertImage returned error: %v", err)
	}

	if _, err := adminService.DeleteStorageConfig(ctx, "s3-secondary"); err == nil || !containsError(err, ErrConflict) {
		t.Fatalf("expected ErrConflict for deleting in-use storage, got %v", err)
	}
}

func TestUpdateStorageConfigPreservesMaskedSecretsAndReloadsManager(t *testing.T) {
	ctx := context.Background()
	adminService, repo := newAdminServiceTestHarness(t)

	if err := repo.CreateStorageConfig(ctx, config.RuntimeStorageConfig{
		StorageKey:       "s3-secondary",
		Name:             "S3 Secondary",
		Backend:          config.StorageBackendS3,
		S3Endpoint:       "127.0.0.1:9000",
		S3Region:         "auto",
		S3Bucket:         "bucket-b",
		S3AccessKey:      "access-b",
		S3SecretKey:      "secret-b",
		S3ForcePathStyle: true,
	}); err != nil {
		t.Fatalf("CreateStorageConfig returned error: %v", err)
	}
	if err := adminService.reloadStorageManager(ctx); err != nil {
		t.Fatalf("reloadStorageManager returned error: %v", err)
	}

	view, err := adminService.UpdateStorageConfig(ctx, "s3-secondary", AdminStorageConfigUpdateInput{
		Name:        strPtr("S3 Renamed"),
		S3Bucket:    strPtr("bucket-updated"),
		S3SecretKey: strPtr(maskSecret("secret-b")),
	})
	if err != nil {
		t.Fatalf("UpdateStorageConfig returned error: %v", err)
	}

	var updatedView *AdminStorageConfigView
	for _, item := range view.StorageConfigs {
		if item.StorageKey == "s3-secondary" {
			copy := item
			updatedView = &copy
			break
		}
	}
	if updatedView == nil {
		t.Fatalf("expected updated storage config in response")
	}
	if updatedView.Name != "S3 Renamed" || updatedView.S3Bucket != "bucket-updated" {
		t.Fatalf("unexpected updated view: %+v", updatedView)
	}

	stored, err := repo.GetStorageConfigByKey(ctx, "s3-secondary")
	if err != nil {
		t.Fatalf("GetStorageConfigByKey returned error: %v", err)
	}
	if stored.S3SecretKey != "secret-b" {
		t.Fatalf("expected secret key to remain unchanged, got %q", stored.S3SecretKey)
	}
	if stored.Name != "S3 Renamed" || stored.S3Bucket != "bucket-updated" {
		t.Fatalf("unexpected stored config: %+v", stored)
	}

	resolved, err := adminService.storage.ForKey("s3-secondary")
	if err != nil {
		t.Fatalf("storage.ForKey returned error: %v", err)
	}
	if resolved.Config.S3Bucket != "bucket-updated" {
		t.Fatalf("expected manager reload to pick up updated bucket, got %q", resolved.Config.S3Bucket)
	}
}

func TestUpdateConfigPatchesDefaultStorageInstance(t *testing.T) {
	ctx := context.Background()
	adminService, _ := newAdminServiceTestHarness(t)

	nextPath := filepath.Join(t.TempDir(), "next-images")
	view, err := adminService.UpdateConfig(ctx, AdminConfigUpdateInput{
		Name:             strPtr("Renamed Local"),
		LocalStoragePath: strPtr(nextPath),
	})
	if err != nil {
		t.Fatalf("UpdateConfig returned error: %v", err)
	}
	if view.DefaultStorageKey != "local-default" {
		t.Fatalf("expected local-default to remain default, got %q", view.DefaultStorageKey)
	}

	var updated *AdminStorageConfigView
	for _, item := range view.StorageConfigs {
		if item.StorageKey == "local-default" {
			copy := item
			updated = &copy
			break
		}
	}
	if updated == nil {
		t.Fatalf("expected local-default in config view")
	}
	if updated.Name != "Renamed Local" || updated.LocalStoragePath != nextPath {
		t.Fatalf("unexpected updated storage view: %+v", updated)
	}

	resolved, err := adminService.storage.ForKey("local-default")
	if err != nil {
		t.Fatalf("storage.ForKey returned error: %v", err)
	}
	if resolved.Config.LocalStoragePath != nextPath {
		t.Fatalf("expected storage manager reload with path %q, got %q", nextPath, resolved.Config.LocalStoragePath)
	}
}

func TestUpdateConfigSwitchesDefaultStorageInstance(t *testing.T) {
	ctx := context.Background()
	adminService, repo := newAdminServiceTestHarness(t)

	if err := repo.CreateStorageConfig(ctx, config.RuntimeStorageConfig{
		StorageKey:       "local-secondary",
		Name:             "Local Secondary",
		Backend:          config.StorageBackendLocal,
		LocalStoragePath: filepath.Join(t.TempDir(), "secondary"),
	}); err != nil {
		t.Fatalf("CreateStorageConfig returned error: %v", err)
	}
	if err := adminService.reloadStorageManager(ctx); err != nil {
		t.Fatalf("reloadStorageManager returned error: %v", err)
	}

	view, err := adminService.UpdateConfig(ctx, AdminConfigUpdateInput{
		DefaultStorageKey: strPtr("local-secondary"),
	})
	if err != nil {
		t.Fatalf("UpdateConfig returned error: %v", err)
	}
	if view.DefaultStorageKey != "local-secondary" {
		t.Fatalf("expected local-secondary as default, got %q", view.DefaultStorageKey)
	}
	if current := adminService.storage.CurrentKey(); current != "local-secondary" {
		t.Fatalf("expected storage manager current key local-secondary, got %q", current)
	}
}

func TestUpdateConfigRejectsInvalidDefaultBeforePatch(t *testing.T) {
	ctx := context.Background()
	adminService, repo := newAdminServiceTestHarness(t)

	original, err := repo.GetStorageConfigByKey(ctx, "local-default")
	if err != nil {
		t.Fatalf("GetStorageConfigByKey returned error: %v", err)
	}

	_, err = adminService.UpdateConfig(ctx, AdminConfigUpdateInput{
		StorageKey:        strPtr("local-default"),
		DefaultStorageKey: strPtr("missing-storage"),
		Name:              strPtr("Should Not Persist"),
	})
	if err == nil || !containsError(err, ErrNotFound) {
		t.Fatalf("expected ErrNotFound for missing default storage, got %v", err)
	}

	stored, err := repo.GetStorageConfigByKey(ctx, "local-default")
	if err != nil {
		t.Fatalf("GetStorageConfigByKey returned error: %v", err)
	}
	if stored.Name != original.Name {
		t.Fatalf("expected config patch to be rejected before save, got name %q", stored.Name)
	}
}

func TestUpdateConfigRejectsEmptyDefaultBeforePatch(t *testing.T) {
	ctx := context.Background()
	adminService, repo := newAdminServiceTestHarness(t)

	original, err := repo.GetStorageConfigByKey(ctx, "local-default")
	if err != nil {
		t.Fatalf("GetStorageConfigByKey returned error: %v", err)
	}

	_, err = adminService.UpdateConfig(ctx, AdminConfigUpdateInput{
		StorageKey:        strPtr("local-default"),
		DefaultStorageKey: strPtr("   "),
		Name:              strPtr("Should Not Persist"),
	})
	if err == nil || !containsError(err, ErrInvalidInput) {
		t.Fatalf("expected ErrInvalidInput for empty default storage, got %v", err)
	}

	stored, err := repo.GetStorageConfigByKey(ctx, "local-default")
	if err != nil {
		t.Fatalf("GetStorageConfigByKey returned error: %v", err)
	}
	if stored.Name != original.Name {
		t.Fatalf("expected config patch to be rejected before save, got name %q", stored.Name)
	}
}

func TestUpdateStorageConfigRejectsBackendChangeForInUseInstance(t *testing.T) {
	ctx := context.Background()
	adminService, repo := newAdminServiceTestHarness(t)

	if err := repo.CreateStorageConfig(ctx, config.RuntimeStorageConfig{
		StorageKey:       "s3-secondary",
		Name:             "S3 Secondary",
		Backend:          config.StorageBackendS3,
		S3Endpoint:       "127.0.0.1:9000",
		S3Region:         "auto",
		S3Bucket:         "bucket-b",
		S3AccessKey:      "access-b",
		S3SecretKey:      "secret-b",
		S3ForcePathStyle: true,
	}); err != nil {
		t.Fatalf("CreateStorageConfig returned error: %v", err)
	}
	if err := repo.InsertImage(ctx, modelImageRecord("uid-1", "s3-secondary", config.StorageBackendS3)); err != nil {
		t.Fatalf("InsertImage returned error: %v", err)
	}

	_, err := adminService.UpdateStorageConfig(ctx, "s3-secondary", AdminStorageConfigUpdateInput{
		Backend:   strPtr(config.StorageBackendWebDAV),
		WebDAVURL: strPtr("https://dav.example.com/remote.php/dav/files/demo"),
	})
	if err == nil || !containsError(err, ErrConflict) {
		t.Fatalf("expected ErrConflict for backend change on in-use storage, got %v", err)
	}

	stored, getErr := repo.GetStorageConfigByKey(ctx, "s3-secondary")
	if getErr != nil {
		t.Fatalf("GetStorageConfigByKey returned error: %v", getErr)
	}
	if stored.Backend != config.StorageBackendS3 {
		t.Fatalf("expected backend to remain %q, got %q", config.StorageBackendS3, stored.Backend)
	}
}

func newAdminServiceTestHarness(t *testing.T) (*AdminService, *repository.Repository) {
	t.Helper()

	dir := t.TempDir()
	repo, err := repository.New(filepath.Join(dir, "test.sqlite"))
	if err != nil {
		t.Fatalf("repository.New returned error: %v", err)
	}
	t.Cleanup(func() {
		_ = repo.Close()
	})

	ctx := context.Background()
	if err := repo.Migrate(ctx); err != nil {
		t.Fatalf("Migrate returned error: %v", err)
	}

	catalog, err := repo.InitializeStorageCatalog(ctx, config.RuntimeStorageConfig{
		StorageKey:       "local-default",
		Name:             "Default Local Storage",
		IsDefault:        true,
		Backend:          config.StorageBackendLocal,
		LocalStoragePath: filepath.Join(dir, "images"),
	})
	if err != nil {
		t.Fatalf("InitializeStorageCatalog returned error: %v", err)
	}

	manager, err := storage.NewManager(catalog.StorageConfigs)
	if err != nil {
		t.Fatalf("storage.NewManager returned error: %v", err)
	}

	logger := slog.New(slog.NewTextHandler(ioDiscard{}, nil))
	return NewAdminService(repo, manager, NewImageService(repo, newFakeCache(), manager, nil, nil, logger), "admin123", "secret"), repo
}

func modelImageRecord(uid string, storageKey string, backend string) model.ImageRecord {
	return model.ImageRecord{
		UID:            uid,
		Token:          "token",
		StorageKey:     storageKey,
		StorageBackend: backend,
		FilePath:       "2026/04/" + uid + ".avif",
		MIMEType:       "image/avif",
		Size:           1,
		MD5Hash:        "hash-" + uid,
	}
}

func containsError(err error, target error) bool {
	return err != nil && (err == target || strings.Contains(err.Error(), target.Error()))
}

func strPtr(value string) *string {
	return &value
}
