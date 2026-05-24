package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
)

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
			status TEXT NOT NULL,
			last_check_at DATETIME NOT NULL,
			latency_ms INTEGER NOT NULL DEFAULT 0,
			error_message TEXT NOT NULL DEFAULT '',
			consecutive_failures INTEGER NOT NULL DEFAULT 0,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		);`,
	}
	imageIndexes := map[string]string{
		"idx_images_uid":                `CREATE INDEX IF NOT EXISTS idx_images_uid ON images(uid);`,
		"idx_images_storage_md5":        `CREATE INDEX IF NOT EXISTS idx_images_storage_md5 ON images(storage_key, md5_hash);`,
		"idx_images_created_at":         `CREATE INDEX IF NOT EXISTS idx_images_created_at ON images(created_at DESC);`,
		"idx_images_token_created_at":   `CREATE INDEX IF NOT EXISTS idx_images_token_created_at ON images(token, created_at DESC);`,
		"idx_images_ip_created_at":      `CREATE INDEX IF NOT EXISTS idx_images_ip_created_at ON images(ip_address, created_at DESC);`,
		"idx_images_storage_created_at": `CREATE INDEX IF NOT EXISTS idx_images_storage_created_at ON images(storage_key, created_at DESC);`,
	}
	indexes := []string{
		`CREATE INDEX IF NOT EXISTS idx_images_md5_hash ON images(md5_hash);`,
		`CREATE INDEX IF NOT EXISTS idx_images_file_path ON images(file_path);`,
		`CREATE INDEX IF NOT EXISTS idx_images_storage_key ON images(storage_key);`,
		`CREATE INDEX IF NOT EXISTS idx_images_ip_address ON images(ip_address);`,
		`CREATE INDEX IF NOT EXISTS idx_images_created_ip ON images(created_at, ip_address);`,
		`CREATE INDEX IF NOT EXISTS idx_images_created_token ON images(created_at, token);`,
		`CREATE INDEX IF NOT EXISTS idx_storage_configs_default ON storage_configs(is_default);`,
		`CREATE INDEX IF NOT EXISTS idx_announcements_public ON announcements(status, starts_at, ends_at, sort_order, created_at);`,
		`CREATE INDEX IF NOT EXISTS idx_ip_bans_ip_hash ON ip_bans(ip_hash);`,
		`CREATE INDEX IF NOT EXISTS idx_ip_bans_expires_at ON ip_bans(expires_at);`,
		`CREATE INDEX IF NOT EXISTS idx_storage_health_checks_storage_key ON storage_health_checks(storage_key, last_check_at DESC);`,
		`CREATE INDEX IF NOT EXISTS idx_storage_health_checks_status ON storage_health_checks(status, last_check_at DESC);`,
	}

	for _, stmt := range schema {
		if _, err := r.db.ExecContext(ctx, stmt); err != nil {
			return err
		}
	}
	for _, stmt := range []string{
		`DROP INDEX IF EXISTS idx_token_usage_updated_at`,
		`DROP INDEX IF EXISTS idx_token_controls_disabled`,
		`DROP TABLE IF EXISTS token_usage`,
		`DROP TABLE IF EXISTS token_controls`,
		`DROP INDEX IF EXISTS idx_config_audit_logs_scope_created`,
		`DROP INDEX IF EXISTS idx_config_audit_logs_created`,
		`DROP TABLE IF EXISTS config_audit_logs`,
	} {
		if _, err := r.db.ExecContext(ctx, stmt); err != nil {
			return err
		}
	}
	if err := r.ensureStorageHealthHistoryTable(ctx); err != nil {
		return err
	}
	if err := r.ensureImageColumn(ctx, "storage_key", "TEXT NOT NULL DEFAULT ''"); err != nil {
		return err
	}
	for _, column := range []struct {
		name string
		ddl  string
	}{
		{name: "deleted_at", ddl: "DATETIME NULL"},
		{name: "deleted_by", ddl: "TEXT NOT NULL DEFAULT ''"},
		{name: "delete_reason", ddl: "TEXT NOT NULL DEFAULT ''"},
		{name: "purge_after", ddl: "DATETIME NULL"},
	} {
		if err := r.ensureImageColumn(ctx, column.name, column.ddl); err != nil {
			return err
		}
	}
	imageIndexes["idx_images_deleted_created_at"] = `CREATE INDEX IF NOT EXISTS idx_images_deleted_created_at ON images(deleted_at, created_at DESC);`
	for name, stmt := range imageIndexes {
		if err := r.ensureIndex(ctx, name, stmt); err != nil {
			return err
		}
	}
	for _, stmt := range indexes {
		if _, err := r.db.ExecContext(ctx, stmt); err != nil {
			return err
		}
	}
	return nil
}

func (r *Repository) ensureStorageHealthHistoryTable(ctx context.Context) error {
	var tableSQL string
	err := r.db.QueryRowContext(ctx, `SELECT sql FROM sqlite_master WHERE type = 'table' AND name = 'storage_health_checks'`).Scan(&tableSQL)
	if err != nil {
		return err
	}
	if !strings.Contains(strings.ToUpper(tableSQL), "STORAGE_KEY TEXT UNIQUE") {
		return nil
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()
	for _, stmt := range []string{
		`DROP INDEX IF EXISTS idx_storage_health_checks_storage_key`,
		`DROP INDEX IF EXISTS idx_storage_health_checks_status`,
		`DROP TABLE IF EXISTS storage_health_checks_legacy`,
		`ALTER TABLE storage_health_checks RENAME TO storage_health_checks_legacy`,
		`CREATE TABLE storage_health_checks (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			storage_key TEXT NOT NULL,
			status TEXT NOT NULL,
			last_check_at DATETIME NOT NULL,
			latency_ms INTEGER NOT NULL DEFAULT 0,
			error_message TEXT NOT NULL DEFAULT '',
			consecutive_failures INTEGER NOT NULL DEFAULT 0,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		`INSERT INTO storage_health_checks(id, storage_key, status, last_check_at, latency_ms, error_message, consecutive_failures, created_at, updated_at)
			SELECT id, storage_key, status, last_check_at, latency_ms, error_message, consecutive_failures, created_at, updated_at FROM storage_health_checks_legacy`,
		`DROP TABLE storage_health_checks_legacy`,
	} {
		if _, err := tx.ExecContext(ctx, stmt); err != nil {
			return err
		}
	}
	return tx.Commit()
}

func (r *Repository) ensureIndex(ctx context.Context, name string, ddl string) error {
	var current string
	err := r.db.QueryRowContext(ctx, `SELECT sql FROM sqlite_master WHERE type = 'index' AND name = ?`, name).Scan(&current)
	if err == nil {
		if normalizeIndexSQL(current) == normalizeIndexSQL(ddl) {
			return nil
		}
		if _, err := r.db.ExecContext(ctx, fmt.Sprintf(`DROP INDEX IF EXISTS %s`, name)); err != nil {
			return err
		}
	} else if err != sql.ErrNoRows {
		return err
	}
	_, err = r.db.ExecContext(ctx, ddl)
	return err
}

func normalizeIndexSQL(sql string) string {
	normalized := strings.TrimSuffix(strings.Join(strings.Fields(sql), " "), ";")
	normalized = strings.ReplaceAll(normalized, " IF NOT EXISTS", "")
	return normalized
}

func (r *Repository) ensureImageColumn(ctx context.Context, column string, ddl string) error {
	exists, err := testTableColumnExists(ctx, r.db, "images", column)
	if err != nil {
		return err
	}
	if exists {
		return nil
	}
	_, err = r.db.ExecContext(ctx, fmt.Sprintf(`ALTER TABLE images ADD COLUMN %s %s`, column, ddl))
	return err
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
