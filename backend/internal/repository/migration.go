package repository

import (
	"context"
	"database/sql"
	"fmt"
)

// CurrentSchemaVersion is the expected PRAGMA user_version for this codebase.
// Bump this value whenever the schema changes (new table, column, index, etc.).
const CurrentSchemaVersion = 2

func (r *Repository) Migrate(ctx context.Context) error {
	schema := []string{
		`CREATE TABLE IF NOT EXISTS images (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			uid TEXT UNIQUE NOT NULL,
			token TEXT NOT NULL,
			storage_key TEXT NOT NULL DEFAULT '',
			storage_backend TEXT DEFAULT 'local',
			file_path TEXT,
			mime_type TEXT,
			size INTEGER,
			md5_hash TEXT NOT NULL,
			ip_address TEXT,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		);`,
		`CREATE TABLE IF NOT EXISTS config (
			key TEXT PRIMARY KEY,
			value TEXT
		);`,
		`CREATE TABLE IF NOT EXISTS storage_configs (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			storage_key TEXT UNIQUE NOT NULL,
			name TEXT NOT NULL,
			backend TEXT NOT NULL,
			is_default INTEGER NOT NULL DEFAULT 0,
			local_storage_path TEXT,
			s3_endpoint TEXT,
			s3_region TEXT,
			s3_bucket TEXT,
			s3_access_key TEXT,
			s3_secret_key TEXT,
			s3_use_ssl TEXT,
			s3_force_path_style TEXT,
			webdav_url TEXT,
			webdav_user TEXT,
			webdav_pass TEXT,
			max_upload_size_mb INTEGER NOT NULL DEFAULT 20,
			allowed_mime_types TEXT NOT NULL DEFAULT 'image/avif,image/gif,image/jpeg,image/png,image/webp',
			avif_quality INTEGER NOT NULL DEFAULT 60,
			avif_speed INTEGER NOT NULL DEFAULT 8,
			max_image_pixels INTEGER NOT NULL DEFAULT 40000000,
			avif_max_concurrency INTEGER NOT NULL DEFAULT 2,
			avif_conversion_timeout_seconds INTEGER NOT NULL DEFAULT 30,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		);`,
		`CREATE TABLE IF NOT EXISTS announcements (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			title TEXT NOT NULL,
			content TEXT NOT NULL,
			status TEXT NOT NULL DEFAULT 'draft',
			priority TEXT NOT NULL DEFAULT 'normal',
			starts_at DATETIME NULL,
			ends_at DATETIME NULL,
			sort_order INTEGER NOT NULL DEFAULT 0,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		);`,
		`CREATE TABLE IF NOT EXISTS ip_bans (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			ip_hash TEXT NOT NULL,
			ip_address TEXT NOT NULL,
			reason TEXT NOT NULL,
			expires_at DATETIME NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		);`,
		`CREATE TABLE IF NOT EXISTS storage_health_checks (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			storage_key TEXT NOT NULL,
			status INTEGER NOT NULL DEFAULT 0,
			latency_ms INTEGER NOT NULL DEFAULT 0,
			error_message TEXT NOT NULL DEFAULT '',
			consecutive_failures INTEGER NOT NULL DEFAULT 0,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		);`,
	}
	indexes := []string{
		`CREATE INDEX IF NOT EXISTS idx_images_uid ON images(uid);`,
		`CREATE INDEX IF NOT EXISTS idx_images_md5_hash ON images(md5_hash);`,
		`CREATE INDEX IF NOT EXISTS idx_images_file_path ON images(file_path);`,
		`CREATE INDEX IF NOT EXISTS idx_images_storage_key ON images(storage_key);`,
		`CREATE INDEX IF NOT EXISTS idx_images_ip_address ON images(ip_address);`,
		`CREATE INDEX IF NOT EXISTS idx_images_created_ip ON images(created_at, ip_address);`,
		`CREATE INDEX IF NOT EXISTS idx_images_created_token ON images(created_at, token);`,
		`CREATE INDEX IF NOT EXISTS idx_images_storage_md5 ON images(storage_key, md5_hash);`,
		`CREATE INDEX IF NOT EXISTS idx_images_created_at ON images(created_at DESC);`,
		`CREATE INDEX IF NOT EXISTS idx_images_token_created_at ON images(token, created_at DESC);`,
		`CREATE INDEX IF NOT EXISTS idx_images_ip_created_at ON images(ip_address, created_at DESC);`,
		`CREATE INDEX IF NOT EXISTS idx_images_storage_created_at ON images(storage_key, created_at DESC);`,
		`CREATE INDEX IF NOT EXISTS idx_storage_configs_default ON storage_configs(is_default);`,
		`CREATE INDEX IF NOT EXISTS idx_announcements_public ON announcements(status, starts_at, ends_at, sort_order, created_at);`,
		`CREATE INDEX IF NOT EXISTS idx_ip_bans_ip_hash ON ip_bans(ip_hash);`,
		`CREATE INDEX IF NOT EXISTS idx_ip_bans_expires_at ON ip_bans(expires_at);`,
		`CREATE INDEX IF NOT EXISTS idx_storage_health_checks_storage_key ON storage_health_checks(storage_key, updated_at DESC);`,
		`CREATE INDEX IF NOT EXISTS idx_storage_health_checks_status ON storage_health_checks(status, updated_at DESC);`,
	}

	for _, stmt := range schema {
		if _, err := r.db.ExecContext(ctx, stmt); err != nil {
			return err
		}
	}
	for _, stmt := range indexes {
		if _, err := r.db.ExecContext(ctx, stmt); err != nil {
			return err
		}
	}

	// Q-02: set schema version via PRAGMA user_version.
	if _, err := r.db.ExecContext(ctx, fmt.Sprintf("PRAGMA user_version = %d;", CurrentSchemaVersion)); err != nil {
		return fmt.Errorf("set user_version: %w", err)
	}
	return nil
}

// SchemaVersion returns the current PRAGMA user_version from the database.
func (r *Repository) SchemaVersion(ctx context.Context) (int, error) {
	var version int
	if err := r.db.QueryRowContext(ctx, "PRAGMA user_version;").Scan(&version); err != nil {
		return 0, fmt.Errorf("query user_version: %w", err)
	}
	return version, nil
}

func testTableColumnExists(ctx context.Context, db *sql.DB, table string, column string) (bool, error) {
	rows, err := db.QueryContext(ctx, fmt.Sprintf(`PRAGMA table_info(%s);`, table))
	if err != nil {
		return false, err
	}
	defer rows.Close()

	for rows.Next() {
		var (
			cid        int
			name       string
			columnType string
			notNull    int
			defaultVal any
			primaryKey int
		)
		if err := rows.Scan(&cid, &name, &columnType, &notNull, &defaultVal, &primaryKey); err != nil {
			return false, err
		}
		if name == column {
			return true, nil
		}
	}
	return false, rows.Err()
}
