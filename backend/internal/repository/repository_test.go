package repository

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"omepic/backend/internal/config"
	"omepic/backend/internal/model"
)

func TestMigrateCreatesImagesSchemaWithoutOriginalFilenameColumnAndWithStorageKey(t *testing.T) {
	ctx := context.Background()
	repo, err := New(filepath.Join(t.TempDir(), "test.sqlite"))
	if err != nil {
		t.Fatalf("New returned error: %v", err)
	}
	t.Cleanup(func() {
		_ = repo.Close()
	})

	if err := repo.Migrate(ctx); err != nil {
		t.Fatalf("Migrate returned error: %v", err)
	}

	exists, err := testTableColumnExists(ctx, repo.db, "images", "original_filename")
	if err != nil {
		t.Fatalf("testTableColumnExists returned error: %v", err)
	}
	if exists {
		t.Fatalf("expected images schema to omit original_filename")
	}

	exists, err = testTableColumnExists(ctx, repo.db, "images", "storage_key")
	if err != nil {
		t.Fatalf("testTableColumnExists returned error: %v", err)
	}
	if !exists {
		t.Fatalf("expected images schema to include storage_key")
	}

	exists, err = testTableColumnExists(ctx, repo.db, "ip_bans", "ip_address_masked")
	if err != nil {
		t.Fatalf("testTableColumnExists returned error: %v", err)
	}
	if exists {
		t.Fatalf("expected ip_bans schema to omit ip_address_masked")
	}
}

func TestMigrateCreatesCoreImageIndexesIdempotently(t *testing.T) {
	ctx := context.Background()
	repo, err := New(filepath.Join(t.TempDir(), "test.sqlite"))
	if err != nil {
		t.Fatalf("New returned error: %v", err)
	}
	t.Cleanup(func() {
		_ = repo.Close()
	})

	if err := repo.Migrate(ctx); err != nil {
		t.Fatalf("first Migrate returned error: %v", err)
	}
	if err := repo.Migrate(ctx); err != nil {
		t.Fatalf("second Migrate returned error: %v", err)
	}

	expected := map[string]string{
		"idx_images_uid":                "CREATE INDEX idx_images_uid ON images(uid)",
		"idx_images_storage_md5":        "CREATE INDEX idx_images_storage_md5 ON images(storage_key, md5_hash)",
		"idx_images_created_at":         "CREATE INDEX idx_images_created_at ON images(created_at DESC)",
		"idx_images_token_created_at":   "CREATE INDEX idx_images_token_created_at ON images(token, created_at DESC)",
		"idx_images_ip_created_at":      "CREATE INDEX idx_images_ip_created_at ON images(ip_address, created_at DESC)",
		"idx_images_storage_created_at": "CREATE INDEX idx_images_storage_created_at ON images(storage_key, created_at DESC)",
	}
	for name, want := range expected {
		got, err := imageIndexSQL(ctx, repo, name)
		if err != nil {
			t.Fatalf("imageIndexSQL(%s) returned error: %v", name, err)
		}
		if normalizeIndexSQL(got) != normalizeIndexSQL(want) {
			t.Fatalf("index %s mismatch:\nwant %s\n got %s", name, want, got)
		}
	}

	got, err := imageIndexSQL(ctx, repo, "idx_images_deleted_created_at")
	if err != nil {
		t.Fatalf("imageIndexSQL(idx_images_deleted_created_at) returned error: %v", err)
	}
	want := "CREATE INDEX idx_images_deleted_created_at ON images(deleted_at, created_at DESC)"
	if normalizeIndexSQL(got) != normalizeIndexSQL(want) {
		t.Fatalf("deleted_at index mismatch:\nwant %s\n got %s", want, got)
	}
}

func TestMigrateCreatesSoftDeleteColumnsIdempotently(t *testing.T) {
	ctx := context.Background()
	repo, err := New(filepath.Join(t.TempDir(), "test.sqlite"))
	if err != nil {
		t.Fatalf("New returned error: %v", err)
	}
	t.Cleanup(func() {
		_ = repo.Close()
	})

	if err := repo.Migrate(ctx); err != nil {
		t.Fatalf("first Migrate returned error: %v", err)
	}
	if err := repo.Migrate(ctx); err != nil {
		t.Fatalf("second Migrate returned error: %v", err)
	}

	for _, column := range []string{"deleted_at", "deleted_by", "delete_reason", "purge_after"} {
		exists, err := testTableColumnExists(ctx, repo.db, "images", column)
		if err != nil {
			t.Fatalf("testTableColumnExists(%s) returned error: %v", column, err)
		}
		if !exists {
			t.Fatalf("expected images.%s to exist", column)
		}
	}
}

func TestMigrateDoesNotRebuildLegacyImagesTableForDroppedColumns(t *testing.T) {
	ctx := context.Background()
	repo, err := New(filepath.Join(t.TempDir(), "test.sqlite"))
	if err != nil {
		t.Fatalf("New returned error: %v", err)
	}
	t.Cleanup(func() {
		_ = repo.Close()
	})

	if err := createLegacyImagesTable(ctx, repo); err != nil {
		t.Fatalf("createLegacyImagesTable returned error: %v", err)
	}

	if err := repo.Migrate(ctx); err != nil {
		t.Fatalf("Migrate returned error: %v", err)
	}

	exists, err := testTableColumnExists(ctx, repo.db, "images", "original_filename")
	if err != nil {
		t.Fatalf("testTableColumnExists returned error: %v", err)
	}
	if !exists {
		t.Fatalf("expected legacy schema to remain untouched; stale dev databases must be reset manually")
	}

	exists, err = testTableColumnExists(ctx, repo.db, "images", "storage_key")
	if err != nil {
		t.Fatalf("testTableColumnExists returned error: %v", err)
	}
	if !exists {
		t.Fatalf("expected storage_key column to be added to legacy schema")
	}
}

func TestInsertImagePersistsStorageKeyWithoutOriginalFilenameColumn(t *testing.T) {
	ctx := context.Background()
	repo, err := New(filepath.Join(t.TempDir(), "test.sqlite"))
	if err != nil {
		t.Fatalf("New returned error: %v", err)
	}
	t.Cleanup(func() {
		_ = repo.Close()
	})

	if err := repo.Migrate(ctx); err != nil {
		t.Fatalf("Migrate returned error: %v", err)
	}

	record := model.ImageRecord{
		UID:            "uid-2",
		Token:          "token-2",
		StorageKey:     "local-default",
		StorageBackend: "local",
		FilePath:       "2026/04/two.avif",
		MIMEType:       "image/avif",
		Size:           64,
		MD5Hash:        "hash-2",
		IPAddress:      "127.0.0.1",
	}
	if err := repo.InsertImage(ctx, record); err != nil {
		t.Fatalf("InsertImage returned error: %v", err)
	}

	stored, err := repo.FindByUID(ctx, "uid-2")
	if err != nil {
		t.Fatalf("FindByUID returned error: %v", err)
	}
	if stored.MIMEType != "image/avif" || stored.FilePath != "2026/04/two.avif" || stored.StorageKey != "local-default" {
		t.Fatalf("stored row mismatch: %+v", stored)
	}
}

func TestSearchImagesIgnoresLegacyOriginalFilenameColumnAndIncludesStorageKey(t *testing.T) {
	ctx := context.Background()
	repo, err := New(filepath.Join(t.TempDir(), "test.sqlite"))
	if err != nil {
		t.Fatalf("New returned error: %v", err)
	}
	t.Cleanup(func() {
		_ = repo.Close()
	})

	if err := createLegacyImagesTable(ctx, repo); err != nil {
		t.Fatalf("createLegacyImagesTable returned error: %v", err)
	}

	if err := repo.Migrate(ctx); err != nil {
		t.Fatalf("Migrate returned error: %v", err)
	}

	if _, err := repo.db.ExecContext(
		ctx,
		`INSERT INTO images(
			uid, token, storage_key, storage_backend, file_path, mime_type, size, md5_hash, ip_address, original_filename, created_at
		) VALUES(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		"uid-legacy",
		"token-legacy",
		"local-default",
		"local",
		"2026/04/legacy.avif",
		"image/avif",
		64,
		"hash-legacy",
		"127.0.0.1",
		"client-only-name.png",
		"2026-04-27T00:00:00Z",
	); err != nil {
		t.Fatalf("legacy insert returned error: %v", err)
	}

	matches, total, err := repo.SearchImages(ctx, 1, 20, "client-only-name")
	if err != nil {
		t.Fatalf("SearchImages returned error: %v", err)
	}
	if total != 0 || len(matches) != 0 {
		t.Fatalf("expected legacy original_filename to be ignored, got total=%d len=%d", total, len(matches))
	}

	matches, total, err = repo.SearchImages(ctx, 1, 20, "local-default")
	if err != nil {
		t.Fatalf("SearchImages returned error: %v", err)
	}
	if total != 1 || len(matches) != 1 {
		t.Fatalf("expected storage_key to remain searchable, got total=%d len=%d", total, len(matches))
	}
}

func TestSoftDeleteExcludesDefaultQueriesAndSupportsTrashRestore(t *testing.T) {
	ctx := context.Background()
	repo, err := New(filepath.Join(t.TempDir(), "test.sqlite"))
	if err != nil {
		t.Fatalf("New returned error: %v", err)
	}
	t.Cleanup(func() { _ = repo.Close() })
	if err := repo.Migrate(ctx); err != nil {
		t.Fatalf("Migrate returned error: %v", err)
	}

	record := model.ImageRecord{
		UID:            "uid-trash",
		Token:          "token-trash",
		StorageKey:     "local-default",
		StorageBackend: config.StorageBackendLocal,
		FilePath:       "2026/05/trash.avif",
		MIMEType:       "image/avif",
		Size:           10,
		MD5Hash:        "hash-trash",
		IPAddress:      "127.0.0.1",
	}
	if err := repo.InsertImage(ctx, record); err != nil {
		t.Fatalf("InsertImage returned error: %v", err)
	}
	if err := repo.SoftDeleteByUID(ctx, record.UID, "admin", "cleanup", time.Now().UTC().Add(24*time.Hour)); err != nil {
		t.Fatalf("SoftDeleteByUID returned error: %v", err)
	}
	if _, err := repo.FindByUID(ctx, record.UID); !IsNotFound(err) {
		t.Fatalf("expected active FindByUID to miss soft-deleted row, got %v", err)
	}
	if count, err := repo.CountByMD5AndStorageKey(ctx, record.MD5Hash, record.StorageKey); err != nil || count != 0 {
		t.Fatalf("expected active md5 count 0, count=%d err=%v", count, err)
	}
	items, total, err := repo.SearchImages(ctx, 1, 20, "uid-trash")
	if err != nil || total != 0 || len(items) != 0 {
		t.Fatalf("expected default search to exclude deleted rows, total=%d len=%d err=%v", total, len(items), err)
	}
	trash, total, err := repo.SearchDeletedImages(ctx, 1, 20, "uid-trash")
	if err != nil || total != 1 || len(trash) != 1 {
		t.Fatalf("expected trash search to include deleted row, total=%d len=%d err=%v", total, len(trash), err)
	}
	if trash[0].DeletedAt == nil || trash[0].DeletedBy != "admin" || trash[0].DeleteReason != "cleanup" || trash[0].PurgeAfter == nil {
		t.Fatalf("trash metadata mismatch: %+v", trash[0])
	}
	if err := repo.RestoreByUID(ctx, record.UID); err != nil {
		t.Fatalf("RestoreByUID returned error: %v", err)
	}
	restored, err := repo.FindByUID(ctx, record.UID)
	if err != nil {
		t.Fatalf("FindByUID after restore returned error: %v", err)
	}
	if restored.DeletedAt != nil || restored.PurgeAfter != nil || restored.DeletedBy != "" || restored.DeleteReason != "" {
		t.Fatalf("expected restore to clear delete metadata, got %+v", restored)
	}
}

func TestInitializeStorageCatalogSeedsLegacyBackendsAndBackfillsImageStorageKeys(t *testing.T) {
	ctx := context.Background()
	repo, err := New(filepath.Join(t.TempDir(), "test.sqlite"))
	if err != nil {
		t.Fatalf("New returned error: %v", err)
	}
	t.Cleanup(func() {
		_ = repo.Close()
	})

	if err := repo.Migrate(ctx); err != nil {
		t.Fatalf("Migrate returned error: %v", err)
	}

	if err := repo.UpsertConfigValues(ctx, map[string]string{
		"storage_backend":     config.StorageBackendS3,
		"local_storage_path":  "data/images",
		"s3_endpoint":         "127.0.0.1:9000",
		"s3_region":           "auto",
		"s3_bucket":           "omepic",
		"s3_access_key":       "access",
		"s3_secret_key":       "secret",
		"s3_use_ssl":          "false",
		"s3_force_path_style": "true",
	}); err != nil {
		t.Fatalf("UpsertConfigValues returned error: %v", err)
	}

	if err := repo.InsertImage(ctx, model.ImageRecord{
		UID:            "uid-local",
		Token:          "token-local",
		StorageBackend: config.StorageBackendLocal,
		FilePath:       "2026/04/local.avif",
		MIMEType:       "image/avif",
		Size:           1,
		MD5Hash:        "hash-local",
	}); err != nil {
		t.Fatalf("InsertImage local returned error: %v", err)
	}
	if err := repo.InsertImage(ctx, model.ImageRecord{
		UID:            "uid-s3",
		Token:          "token-s3",
		StorageBackend: config.StorageBackendS3,
		FilePath:       "2026/04/s3.avif",
		MIMEType:       "image/avif",
		Size:           1,
		MD5Hash:        "hash-s3",
	}); err != nil {
		t.Fatalf("InsertImage s3 returned error: %v", err)
	}

	catalog, err := repo.InitializeStorageCatalog(ctx, config.RuntimeStorageConfig{
		StorageKey:       "local-default",
		Name:             "Default Local Storage",
		IsDefault:        true,
		Backend:          config.StorageBackendLocal,
		LocalStoragePath: "data/images",
	})
	if err != nil {
		t.Fatalf("InitializeStorageCatalog returned error: %v", err)
	}

	if len(catalog.StorageConfigs) != 2 {
		t.Fatalf("expected 2 seeded storage configs, got %d", len(catalog.StorageConfigs))
	}
	if catalog.DefaultStorageKey != "s3-default" {
		t.Fatalf("expected s3-default to be the default storage key, got %q", catalog.DefaultStorageKey)
	}

	localRecord, err := repo.FindByUID(ctx, "uid-local")
	if err != nil {
		t.Fatalf("FindByUID local returned error: %v", err)
	}
	if localRecord.StorageKey != "local-default" {
		t.Fatalf("expected local record to backfill local-default, got %q", localRecord.StorageKey)
	}

	s3Record, err := repo.FindByUID(ctx, "uid-s3")
	if err != nil {
		t.Fatalf("FindByUID s3 returned error: %v", err)
	}
	if s3Record.StorageKey != "s3-default" {
		t.Fatalf("expected s3 record to backfill s3-default, got %q", s3Record.StorageKey)
	}
}

func imageIndexSQL(ctx context.Context, repo *Repository, name string) (string, error) {
	var sql string
	err := repo.db.QueryRowContext(ctx, `SELECT sql FROM sqlite_master WHERE type = 'index' AND name = ?`, name).Scan(&sql)
	return sql, err
}

func createLegacyImagesTable(ctx context.Context, repo *Repository) error {
	_, err := repo.db.ExecContext(ctx, `CREATE TABLE images (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		uid TEXT UNIQUE NOT NULL,
		token TEXT NOT NULL,
		storage_backend TEXT DEFAULT 'local',
		file_path TEXT,
		mime_type TEXT,
		size INTEGER,
		md5_hash TEXT NOT NULL,
		ip_address TEXT,
		original_filename TEXT,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);`)
	return err
}
