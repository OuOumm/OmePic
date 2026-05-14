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

func TestCreateStorageConfigUsesProvidedKeyAndGeneratesOnlyWhenEmpty(t *testing.T) {
	ctx := context.Background()
	adminService, repo := newAdminServiceTestHarness(t)

	customPath := filepath.Join(t.TempDir(), "custom")
	view, err := adminService.CreateStorageConfig(ctx, AdminStorageConfigCreateInput{
		StorageKey:       "custom-local",
		Name:             "Custom Local",
		Backend:          config.StorageBackendLocal,
		LocalStoragePath: customPath,
	})
	if err != nil {
		t.Fatalf("CreateStorageConfig returned error: %v", err)
	}

	var foundCustom bool
	for _, item := range view.StorageConfigs {
		if item.StorageKey == "custom-local" {
			foundCustom = true
			break
		}
	}
	if !foundCustom {
		t.Fatalf("expected custom-local to appear in config view")
	}

	stored, err := repo.GetStorageConfigByKey(ctx, "custom-local")
	if err != nil {
		t.Fatalf("GetStorageConfigByKey returned error: %v", err)
	}
	if stored.Name != "Custom Local" || stored.LocalStoragePath != customPath {
		t.Fatalf("unexpected stored custom config: %+v", stored)
	}

	generatedPath := filepath.Join(t.TempDir(), "generated")
	view, err = adminService.CreateStorageConfig(ctx, AdminStorageConfigCreateInput{
		Name:             "Generated Local",
		Backend:          config.StorageBackendLocal,
		LocalStoragePath: generatedPath,
	})
	if err != nil {
		t.Fatalf("CreateStorageConfig with empty key returned error: %v", err)
	}

	var generatedKey string
	for _, item := range view.StorageConfigs {
		if item.Name == "Generated Local" {
			generatedKey = item.StorageKey
			break
		}
	}
	if generatedKey == "" || generatedKey == "custom-local" {
		t.Fatalf("expected generated non-custom storage key, got %q", generatedKey)
	}
	if _, err := repo.GetStorageConfigByKey(ctx, generatedKey); err != nil {
		t.Fatalf("GetStorageConfigByKey for generated key returned error: %v", err)
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
		RuntimeStorageUpdate: config.RuntimeStorageUpdate{
			Name:             strPtr("Renamed Local"),
			LocalStoragePath: strPtr(nextPath),
		},
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
		RuntimeStorageUpdate: config.RuntimeStorageUpdate{
			Name: strPtr("Should Not Persist"),
		},
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
		RuntimeStorageUpdate: config.RuntimeStorageUpdate{
			Name: strPtr("Should Not Persist"),
		},
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

func TestUpdateSystemSettingsRejectsInvalidAVIFSettingsWithoutPartialSave(t *testing.T) {
	ctx := context.Background()
	adminService, repo := newAdminServiceTestHarness(t)
	if _, err := adminService.UpdateSystemSettings(ctx, RuntimeSettingsUpdateInput(defaultRuntimeSettings())); err != nil {
		t.Fatalf("initial UpdateSystemSettings returned error: %v", err)
	}

	input := RuntimeSettingsUpdateInput(defaultRuntimeSettings())
	input.SiteName = "Should Not Persist"
	input.AvifQuality = 101
	_, err := adminService.UpdateSystemSettings(ctx, input)
	if err == nil || !containsError(err, ErrInvalidInput) {
		t.Fatalf("expected ErrInvalidInput for invalid avif quality, got %v", err)
	}

	values, err := repo.GetAllConfig(ctx)
	if err != nil {
		t.Fatalf("GetAllConfig returned error: %v", err)
	}
	if values["site_name"] == "Should Not Persist" {
		t.Fatalf("expected invalid avif update to avoid partial site_name save")
	}
	if adminService.settings.Current().SiteName == "Should Not Persist" {
		t.Fatalf("expected invalid avif update to avoid in-memory reconfigure")
	}
}

func TestGetSystemSettingsMarksDefaultSecurityValuesAndPasswordBootstrapState(t *testing.T) {
	ctx := context.Background()
	adminService, _ := newAdminServiceTestHarnessWithEnv(t, "change-me-too", "change-me-uid-secret")

	view, err := adminService.GetSystemSettings(ctx)
	if err != nil {
		t.Fatalf("GetSystemSettings returned error: %v", err)
	}
	if !view.Readonly.Security.JWTSecret.UsingDefault {
		t.Fatalf("expected jwt_secret.using_default to be true")
	}
	if !view.Readonly.Security.UIDEncryptionKey.UsingDefault {
		t.Fatalf("expected uid_encryption_key.using_default to be true")
	}
	if view.Readonly.Security.AdminPassword.Configured {
		t.Fatalf("expected admin_password.configured to be false before password bootstrap")
	}

	if _, err := adminService.Login(ctx, DefaultAdminPassword); err != nil {
		t.Fatalf("expected first-boot default login to succeed, got %v", err)
	}
	view, err = adminService.GetSystemSettings(ctx)
	if err != nil {
		t.Fatalf("GetSystemSettings after bootstrap returned error: %v", err)
	}
	if !view.Readonly.Security.AdminPassword.Configured {
		t.Fatalf("expected admin_password.configured to be true after password bootstrap")
	}
	if view.Readonly.Security.AdminPassword.UsingDefault {
		t.Fatalf("expected admin_password.using_default to remain false until explicit support exists")
	}
}

func TestChangePasswordRequiresValidOldPasswordAndStoresBcryptHash(t *testing.T) {
	ctx := context.Background()
	adminService, repo := newAdminServiceTestHarness(t)

	if err := adminService.ChangePassword(ctx, DefaultAdminPassword, "First-secret!"); err != nil {
		t.Fatalf("first-boot ChangePassword returned error: %v", err)
	}
	if _, err := adminService.Login(ctx, "First-secret!"); err != nil {
		t.Fatalf("expected first changed password login to succeed, got %v", err)
	}
	if err := adminService.ChangePassword(ctx, "First-secret!", "Admin123!"); err != nil {
		t.Fatalf("reset ChangePassword returned error: %v", err)
	}

	if err := adminService.ChangePassword(ctx, "wrong-password", "New-secret!"); err == nil || !containsError(err, ErrForbidden) || !strings.Contains(err.Error(), "current password is incorrect") {
		t.Fatalf("expected clear ErrForbidden for wrong old password, got %v", err)
	}
	weakPasswords := []string{"   ", "Short1!", "lowercase!", "UPPERCASE!", "NoSymbol1"}
	for _, password := range weakPasswords {
		if err := adminService.ChangePassword(ctx, "Admin123!", password); err == nil || !containsError(err, ErrInvalidInput) {
			t.Fatalf("expected ErrInvalidInput for weak new password %q, got %v", password, err)
		}
	}
	if err := adminService.ChangePassword(ctx, "Admin123!", "New-secret!"); err != nil {
		t.Fatalf("ChangePassword returned error: %v", err)
	}

	storedHash, err := repo.GetConfigValue(ctx, "admin_password_hash")
	if err != nil {
		t.Fatalf("GetConfigValue returned error: %v", err)
	}
	if storedHash == "New-secret!" || !strings.HasPrefix(storedHash, "$2") {
		t.Fatalf("expected stored bcrypt hash, got %q", storedHash)
	}
	if _, err := adminService.Login(ctx, "Admin123!"); err == nil || !containsError(err, ErrForbidden) {
		t.Fatalf("expected old password to fail after change, got %v", err)
	}
	if _, err := adminService.Login(ctx, "New-secret!"); err != nil {
		t.Fatalf("expected new password login to succeed, got %v", err)
	}
}

func newAdminServiceTestHarness(t *testing.T) (*AdminService, *repository.Repository) {
	return newAdminServiceTestHarnessWithEnv(t, "secret", "uid-secret")
}

func newAdminServiceTestHarnessWithEnv(t *testing.T, jwtSecret string, uidSecret string) (*AdminService, *repository.Repository) {
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
	settingsManager := NewRuntimeSettingsManager()
	if err := settingsManager.Load(ctx, repo); err != nil {
		t.Fatalf("settingsManager.Load returned error: %v", err)
	}
	imageService := NewImageService(repo, newFakeCache(), manager, settingsManager, nil, nil, logger)
	return NewAdminService(repo, manager, settingsManager, imageService, jwtSecret, AdminEnvMetadata{HTTPAddr: ":8080", DatabasePath: ":memory:", RedisURL: "", UIDEncryptionKey: uidSecret}), repo
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
