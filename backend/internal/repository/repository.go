package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	_ "modernc.org/sqlite"

	"omepic/backend/internal/config"
	"omepic/backend/internal/iputil"
	"omepic/backend/internal/model"
)

type Repository struct {
	db *sql.DB
}

type execContexter interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
}

const (
	imageColumns = "id, uid, token, storage_key, storage_backend, file_path, mime_type, size, md5_hash, ip_address, created_at"

	storageConfigInsertSQL = `INSERT INTO storage_configs(
		storage_key, name, backend, is_default, local_storage_path, s3_endpoint, s3_region, s3_bucket, s3_access_key, s3_secret_key, s3_use_ssl, s3_force_path_style, webdav_url, webdav_user, webdav_pass, created_at, updated_at
	) VALUES(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`
)

func New(databasePath string) (*Repository, error) {
	if err := os.MkdirAll(filepath.Dir(databasePath), 0o755); err != nil {
		return nil, err
	}
	db, err := sql.Open("sqlite", databasePath)
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(time.Hour)
	repository := &Repository{db: db}
	if err := repository.configureSQLite(); err != nil {
		_ = db.Close()
		return nil, err
	}
	return repository, nil
}

func (r *Repository) Close() error {
	return r.db.Close()
}

func (r *Repository) Ping(ctx context.Context) error {
	return r.db.PingContext(ctx)
}

func (r *Repository) configureSQLite() error {
	pragmas := []string{
		`PRAGMA journal_mode = WAL;`,
		`PRAGMA synchronous = NORMAL;`,
		`PRAGMA busy_timeout = 5000;`,
		`PRAGMA foreign_keys = ON;`,
		`PRAGMA temp_store = MEMORY;`,
		`PRAGMA mmap_size = 268435456;`,
	}
	for _, pragma := range pragmas {
		if _, err := r.db.Exec(pragma); err != nil {
			return err
		}
	}
	return nil
}

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
			ip_address_masked TEXT NOT NULL,
			reason TEXT NOT NULL,
			expires_at DATETIME NULL,
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
		`CREATE INDEX IF NOT EXISTS idx_images_created_at ON images(created_at);`,
		`CREATE INDEX IF NOT EXISTS idx_images_created_ip ON images(created_at, ip_address);`,
		`CREATE INDEX IF NOT EXISTS idx_images_created_token ON images(created_at, token);`,
		`CREATE INDEX IF NOT EXISTS idx_storage_configs_default ON storage_configs(is_default);`,
		`CREATE INDEX IF NOT EXISTS idx_announcements_public ON announcements(status, starts_at, ends_at, sort_order, created_at);`,
		`CREATE INDEX IF NOT EXISTS idx_ip_bans_ip_hash ON ip_bans(ip_hash);`,
		`CREATE INDEX IF NOT EXISTS idx_ip_bans_expires_at ON ip_bans(expires_at);`,
	}

	for _, stmt := range schema {
		if _, err := r.db.ExecContext(ctx, stmt); err != nil {
			return err
		}
	}
	if err := r.ensureImageColumn(ctx, "storage_key", "TEXT NOT NULL DEFAULT ''"); err != nil {
		return err
	}
	for _, stmt := range indexes {
		if _, err := r.db.ExecContext(ctx, stmt); err != nil {
			return err
		}
	}
	return nil
}

func (r *Repository) InitializeStorageCatalog(ctx context.Context, envDefault config.RuntimeStorageConfig) (config.RuntimeStorageCatalog, error) {
	configs, err := r.ListStorageConfigs(ctx)
	if err != nil {
		return config.RuntimeStorageCatalog{}, err
	}

	if len(configs) == 0 {
		legacy, err := r.GetLegacyStorageConfig(ctx)
		if err != nil {
			return config.RuntimeStorageCatalog{}, err
		}
		configs = initialStorageConfigs(legacy, envDefault)
		if err := r.insertStorageConfigs(ctx, configs); err != nil {
			return config.RuntimeStorageCatalog{}, err
		}
	}

	if err := r.normalizeDefaultStorageConfig(ctx); err != nil {
		return config.RuntimeStorageCatalog{}, err
	}

	configs, err = r.ListStorageConfigs(ctx)
	if err != nil {
		return config.RuntimeStorageCatalog{}, err
	}
	catalog := buildStorageCatalog(configs)

	if err := r.backfillImageStorageKeys(ctx, keyByBackend(configs), catalog.DefaultStorageKey); err != nil {
		return config.RuntimeStorageCatalog{}, err
	}

	return catalog, nil
}

func (r *Repository) GetLegacyStorageConfig(ctx context.Context) (config.RuntimeStorageConfig, error) {
	values, err := r.GetAllConfig(ctx)
	if err != nil {
		return config.RuntimeStorageConfig{}, err
	}
	return config.RuntimeStorageConfig{
		Backend:          values["storage_backend"],
		LocalStoragePath: values["local_storage_path"],
		S3Endpoint:       values["s3_endpoint"],
		S3Region:         values["s3_region"],
		S3Bucket:         values["s3_bucket"],
		S3AccessKey:      values["s3_access_key"],
		S3SecretKey:      values["s3_secret_key"],
		S3UseSSL:         parseBool(values["s3_use_ssl"]),
		S3ForcePathStyle: parseBool(values["s3_force_path_style"]),
		WebDAVURL:        values["webdav_url"],
		WebDAVUser:       values["webdav_user"],
		WebDAVPass:       values["webdav_pass"],
	}, nil
}

func (r *Repository) GetAllConfig(ctx context.Context) (map[string]string, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT key, value FROM config`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	values := make(map[string]string)
	for rows.Next() {
		var key string
		var value string
		if err := rows.Scan(&key, &value); err != nil {
			return nil, err
		}
		values[key] = value
	}
	return values, rows.Err()
}

func (r *Repository) UpsertConfigValues(ctx context.Context, values map[string]string) error {
	for key, value := range values {
		if _, err := r.db.ExecContext(
			ctx,
			`INSERT INTO config(key, value)
			 VALUES(?, ?)
			 ON CONFLICT(key) DO UPDATE SET value = excluded.value`,
			key,
			value,
		); err != nil {
			return err
		}
	}
	return nil
}

func (r *Repository) InsertMissingConfigValues(ctx context.Context, values map[string]string) error {
	for key, value := range values {
		if _, err := r.db.ExecContext(
			ctx,
			`INSERT INTO config(key, value)
			 VALUES(?, ?)
			 ON CONFLICT(key) DO NOTHING`,
			key,
			value,
		); err != nil {
			return err
		}
	}
	return nil
}

func (r *Repository) GetConfigValue(ctx context.Context, key string) (string, error) {
	var value string
	err := r.db.QueryRowContext(ctx, `SELECT value FROM config WHERE key = ?`, key).Scan(&value)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", fmt.Errorf("%w: config key %s", sql.ErrNoRows, key)
		}
		return "", err
	}
	return value, nil
}

func (r *Repository) SetConfigValue(ctx context.Context, key string, value string) error {
	_, err := r.db.ExecContext(ctx, `INSERT INTO config (key, value) VALUES (?, ?) ON CONFLICT(key) DO UPDATE SET value = excluded.value`, key, value)
	return err
}

func (r *Repository) ListStorageConfigs(ctx context.Context) ([]config.RuntimeStorageConfig, error) {
	rows, err := r.db.QueryContext(
		ctx,
		`SELECT storage_key, name, backend, is_default, local_storage_path, s3_endpoint, s3_region, s3_bucket, s3_access_key, s3_secret_key, s3_use_ssl, s3_force_path_style, webdav_url, webdav_user, webdav_pass
		 FROM storage_configs
		 ORDER BY is_default DESC, id ASC`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var configs []config.RuntimeStorageConfig
	for rows.Next() {
		record, err := scanStorageConfig(rows)
		if err != nil {
			return nil, err
		}
		configs = append(configs, record)
	}
	return configs, rows.Err()
}

func (r *Repository) GetStorageConfigByKey(ctx context.Context, storageKey string) (config.RuntimeStorageConfig, error) {
	row := r.db.QueryRowContext(
		ctx,
		`SELECT storage_key, name, backend, is_default, local_storage_path, s3_endpoint, s3_region, s3_bucket, s3_access_key, s3_secret_key, s3_use_ssl, s3_force_path_style, webdav_url, webdav_user, webdav_pass
		 FROM storage_configs
		 WHERE storage_key = ?`,
		storageKey,
	)
	return scanStorageConfig(row)
}

func (r *Repository) CreateStorageConfig(ctx context.Context, cfg config.RuntimeStorageConfig) error {
	return insertStorageConfig(ctx, r.db, cfg, time.Now().UTC().Format(time.RFC3339))
}

func (r *Repository) UpdateStorageConfig(ctx context.Context, cfg config.RuntimeStorageConfig) error {
	result, err := r.db.ExecContext(
		ctx,
		`UPDATE storage_configs
		 SET name = ?, backend = ?, local_storage_path = ?, s3_endpoint = ?, s3_region = ?, s3_bucket = ?, s3_access_key = ?, s3_secret_key = ?, s3_use_ssl = ?, s3_force_path_style = ?, webdav_url = ?, webdav_user = ?, webdav_pass = ?, updated_at = ?
		 WHERE storage_key = ?`,
		cfg.Name,
		cfg.Backend,
		cfg.LocalStoragePath,
		cfg.S3Endpoint,
		cfg.S3Region,
		cfg.S3Bucket,
		cfg.S3AccessKey,
		cfg.S3SecretKey,
		boolString(cfg.S3UseSSL),
		boolString(cfg.S3ForcePathStyle),
		cfg.WebDAVURL,
		cfg.WebDAVUser,
		cfg.WebDAVPass,
		time.Now().UTC().Format(time.RFC3339),
		cfg.StorageKey,
	)
	if err != nil {
		return err
	}
	if err := ensureRowsAffected(result); err != nil {
		return err
	}
	return nil
}

func (r *Repository) DeleteStorageConfig(ctx context.Context, storageKey string) error {
	result, err := r.db.ExecContext(ctx, `DELETE FROM storage_configs WHERE storage_key = ?`, storageKey)
	if err != nil {
		return err
	}
	if err := ensureRowsAffected(result); err != nil {
		return err
	}
	return nil
}

func (r *Repository) SetDefaultStorageConfig(ctx context.Context, storageKey string) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if _, err := tx.ExecContext(ctx, `UPDATE storage_configs SET is_default = 0`); err != nil {
		return err
	}
	result, err := tx.ExecContext(
		ctx,
		`UPDATE storage_configs SET is_default = 1, updated_at = ? WHERE storage_key = ?`,
		time.Now().UTC().Format(time.RFC3339),
		storageKey,
	)
	if err != nil {
		return err
	}
	if err := ensureRowsAffected(result); err != nil {
		return err
	}

	return tx.Commit()
}

func (r *Repository) CountImagesByStorageKey(ctx context.Context, storageKey string) (int64, error) {
	return r.countByQuery(ctx, `SELECT COUNT(1) FROM images WHERE storage_key = ?`, storageKey)
}

func (r *Repository) InsertImage(ctx context.Context, record model.ImageRecord) error {
	if record.CreatedAt.IsZero() {
		record.CreatedAt = time.Now().UTC()
	}
	_, err := r.db.ExecContext(
		ctx,
		`INSERT INTO images(
			uid, token, storage_key, storage_backend, file_path, mime_type, size, md5_hash, ip_address, created_at
		) VALUES(?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		record.UID,
		record.Token,
		record.StorageKey,
		record.StorageBackend,
		record.FilePath,
		record.MIMEType,
		record.Size,
		record.MD5Hash,
		record.IPAddress,
		record.CreatedAt.UTC().Format(time.RFC3339),
	)
	return err
}

func (r *Repository) FindByUID(ctx context.Context, uid string) (*model.ImageRecord, error) {
	row := r.db.QueryRowContext(ctx, `SELECT `+imageColumns+` FROM images WHERE uid = ?`, uid)
	record, err := scanImage(row)
	if err != nil {
		return nil, err
	}
	return &record, nil
}

func (r *Repository) FindByMD5(ctx context.Context, md5Hash string) (*model.ImageRecord, error) {
	row := r.db.QueryRowContext(ctx, `SELECT `+imageColumns+` FROM images WHERE md5_hash = ? ORDER BY id ASC LIMIT 1`, md5Hash)
	record, err := scanImage(row)
	if err != nil {
		return nil, err
	}
	return &record, nil
}

func (r *Repository) FindByMD5AndStorageKey(ctx context.Context, md5Hash string, storageKey string) (*model.ImageRecord, error) {
	row := r.db.QueryRowContext(ctx, `SELECT `+imageColumns+` FROM images WHERE md5_hash = ? AND storage_key = ? ORDER BY id ASC LIMIT 1`, md5Hash, storageKey)
	record, err := scanImage(row)
	if err != nil {
		return nil, err
	}
	return &record, nil
}

func (r *Repository) DeleteByUID(ctx context.Context, uid string) error {
	result, err := r.db.ExecContext(ctx, `DELETE FROM images WHERE uid = ?`, uid)
	if err != nil {
		return err
	}
	if err := ensureRowsAffected(result); err != nil {
		return err
	}
	return nil
}

func (r *Repository) CountByMD5(ctx context.Context, md5Hash string) (int64, error) {
	return r.countByQuery(ctx, `SELECT COUNT(1) FROM images WHERE md5_hash = ?`, md5Hash)
}

func (r *Repository) CountByMD5AndStorageKey(ctx context.Context, md5Hash string, storageKey string) (int64, error) {
	return r.countByQuery(ctx, `SELECT COUNT(1) FROM images WHERE md5_hash = ? AND storage_key = ?`, md5Hash, storageKey)
}

func (r *Repository) CountByStoredFile(ctx context.Context, storageKey string, filePath string) (int64, error) {
	return r.countByQuery(ctx, `SELECT COUNT(1) FROM images WHERE storage_key = ? AND file_path = ?`, storageKey, filePath)
}

func (r *Repository) ImageSummaryByIP(ctx context.Context, ipAddress string) (model.IPImageSummary, error) {
	var summary model.IPImageSummary
	err := r.db.QueryRowContext(ctx, `SELECT COUNT(1), COALESCE(SUM(size), 0) FROM images WHERE ip_address = ?`, ipAddress).Scan(&summary.Count, &summary.TotalSize)
	return summary, err
}

func (r *Repository) ListImagesByIP(ctx context.Context, ipAddress string) ([]model.ImageRecord, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT `+imageColumns+` FROM images WHERE ip_address = ? ORDER BY id ASC`, ipAddress)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanImages(rows)
}

func (r *Repository) ListAllImages(ctx context.Context) ([]model.ImageRecord, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT `+imageColumns+` FROM images ORDER BY id ASC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanImages(rows)
}

func (r *Repository) SearchImages(ctx context.Context, page int, pageSize int, search string) ([]model.ImageRecord, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100
	}

	like := "%" + search + "%"
	where := `WHERE (? = '' OR uid LIKE ? OR token LIKE ? OR ip_address LIKE ? OR md5_hash LIKE ? OR storage_key LIKE ?)`

	total, err := r.countByQuery(ctx, `SELECT COUNT(1) FROM images `+where, search, like, like, like, like, like)
	if err != nil {
		return nil, 0, err
	}

	rows, err := r.db.QueryContext(
		ctx,
		`SELECT `+imageColumns+`
		 FROM images `+where+`
		 ORDER BY id DESC
		 LIMIT ? OFFSET ?`,
		search,
		like,
		like,
		like,
		like,
		like,
		pageSize,
		(page-1)*pageSize,
	)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	records, err := scanImages(rows)
	if err != nil {
		return nil, 0, err
	}
	return records, total, nil
}

func (r *Repository) AggregateStatus(ctx context.Context) (model.AdminStatus, error) {
	var status model.AdminStatus
	if err := r.db.QueryRowContext(ctx, `SELECT COUNT(1), COALESCE(SUM(size), 0), COUNT(DISTINCT token) FROM images`).Scan(&status.TotalImages, &status.TotalStorageSize, &status.UniqueTokens); err != nil {
		return status, err
	}
	if err := r.db.QueryRowContext(ctx, `SELECT COUNT(1) FROM images WHERE DATE(created_at) = DATE('now')`).Scan(&status.TodayUploads); err != nil {
		return status, err
	}
	return status, nil
}

func (r *Repository) CreateIPBan(ctx context.Context, ban model.IPBan) (model.IPBan, error) {
	now := time.Now().UTC()
	if ban.CreatedAt.IsZero() {
		ban.CreatedAt = now
	}
	ban.UpdatedAt = now
	result, err := r.db.ExecContext(
		ctx,
		`INSERT INTO ip_bans(ip_hash, ip_address, ip_address_masked, reason, expires_at, created_at, updated_at)
		 VALUES(?, ?, ?, ?, ?, ?, ?)`,
		ban.IPHash,
		ban.IPAddress,
		ban.IPAddressMasked,
		ban.Reason,
		nullableTimeString(ban.ExpiresAt),
		ban.CreatedAt.Format(time.RFC3339),
		ban.UpdatedAt.Format(time.RFC3339),
	)
	if err != nil {
		return model.IPBan{}, err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return model.IPBan{}, err
	}
	return r.GetIPBan(ctx, id)
}

func (r *Repository) ListIPBans(ctx context.Context) ([]model.IPBan, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT id, ip_hash, ip_address, ip_address_masked, reason, expires_at, created_at, updated_at FROM ip_bans ORDER BY id DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanIPBans(rows)
}

func (r *Repository) GetIPBan(ctx context.Context, id int64) (model.IPBan, error) {
	row := r.db.QueryRowContext(ctx, `SELECT id, ip_hash, ip_address, ip_address_masked, reason, expires_at, created_at, updated_at FROM ip_bans WHERE id = ?`, id)
	return scanIPBan(row)
}

func (r *Repository) FindActiveIPBanByHash(ctx context.Context, ipHash string) (model.IPBan, error) {
	now := time.Now().UTC().Format(time.RFC3339)
	row := r.db.QueryRowContext(ctx, `SELECT id, ip_hash, ip_address, ip_address_masked, reason, expires_at, created_at, updated_at FROM ip_bans WHERE ip_hash = ? AND (expires_at IS NULL OR expires_at = '' OR expires_at > ?) ORDER BY id DESC LIMIT 1`, ipHash, now)
	return scanIPBan(row)
}

func (r *Repository) DeleteIPBan(ctx context.Context, id int64) error {
	result, err := r.db.ExecContext(ctx, `DELETE FROM ip_bans WHERE id = ?`, id)
	if err != nil {
		return err
	}
	if err := ensureRowsAffected(result); err != nil {
		return err
	}
	return nil
}

func (r *Repository) FindActiveIPBanByIP(ctx context.Context, ipAddress string) (model.IPBan, error) {
	return r.FindActiveIPBanByHash(ctx, ipHashValue(ipAddress))
}

func (r *Repository) ActiveIPBansByHash(ctx context.Context) (map[string]model.IPBan, error) {
	now := time.Now().UTC().Format(time.RFC3339)
	rows, err := r.db.QueryContext(ctx, `SELECT id, ip_hash, ip_address, ip_address_masked, reason, expires_at, created_at, updated_at FROM ip_bans WHERE expires_at IS NULL OR expires_at = '' OR expires_at > ? ORDER BY id DESC`, now)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	bans, err := scanIPBans(rows)
	if err != nil {
		return nil, err
	}
	result := make(map[string]model.IPBan, len(bans))
	for _, ban := range bans {
		if _, exists := result[ban.IPHash]; !exists {
			result[ban.IPHash] = ban
		}
	}
	return result, nil
}

func (r *Repository) CountActiveIPBans(ctx context.Context) (int64, error) {
	now := time.Now().UTC().Format(time.RFC3339)
	return r.countByQuery(ctx, `SELECT COUNT(1) FROM ip_bans WHERE expires_at IS NULL OR expires_at = '' OR expires_at > ?`, now)
}

func (r *Repository) AbuseOverviewTotals(ctx context.Context, from time.Time, to time.Time) (int64, int64, error) {
	var count int64
	var totalSize int64
	err := r.db.QueryRowContext(ctx, `SELECT COUNT(1), COALESCE(SUM(size), 0) FROM images WHERE created_at >= ? AND created_at <= ?`, from.UTC().Format(time.RFC3339), to.UTC().Format(time.RFC3339)).Scan(&count, &totalSize)
	return count, totalSize, err
}

func (r *Repository) TopAbuseIPs(ctx context.Context, from time.Time, to time.Time, limit int) ([]model.AbuseIPRankItem, error) {
	if limit < 1 {
		limit = 10
	}
	if limit > 50 {
		limit = 50
	}
	activeBans, err := r.ActiveIPBansByHash(ctx)
	if err != nil {
		return nil, err
	}
	rows, err := r.db.QueryContext(
		ctx,
		`SELECT ip_address, COUNT(1), COALESCE(SUM(size), 0), MAX(created_at)
		 FROM images
		 WHERE created_at >= ? AND created_at <= ? AND ip_address IS NOT NULL AND TRIM(ip_address) != ''
		 GROUP BY ip_address
		 ORDER BY COUNT(1) DESC, COALESCE(SUM(size), 0) DESC, MAX(created_at) DESC
		 LIMIT ?`,
		from.UTC().Format(time.RFC3339),
		to.UTC().Format(time.RFC3339),
		limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := make([]model.AbuseIPRankItem, 0)
	for rows.Next() {
		var item model.AbuseIPRankItem
		var latest string
		if err := rows.Scan(&item.IPAddress, &item.UploadCount, &item.TotalSize, &latest); err != nil {
			return nil, err
		}
		item.IPAddressMasked = maskIPValue(item.IPAddress)
		item.LatestUploadAt = parseTime(latest)
		if ban, exists := activeBans[ipHashValue(item.IPAddress)]; exists {
			item.IsBanned = true
			item.BanID = ban.ID
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *Repository) TopAbuseTokens(ctx context.Context, from time.Time, to time.Time, limit int) ([]model.AbuseTokenRankItem, error) {
	if limit < 1 {
		limit = 10
	}
	if limit > 50 {
		limit = 50
	}
	rows, err := r.db.QueryContext(
		ctx,
		`SELECT token, COUNT(1), COALESCE(SUM(size), 0), MAX(created_at)
		 FROM images
		 WHERE created_at >= ? AND created_at <= ? AND token IS NOT NULL AND TRIM(token) != ''
		 GROUP BY token
		 ORDER BY COUNT(1) DESC, COALESCE(SUM(size), 0) DESC, MAX(created_at) DESC
		 LIMIT ?`,
		from.UTC().Format(time.RFC3339),
		to.UTC().Format(time.RFC3339),
		limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := make([]model.AbuseTokenRankItem, 0)
	for rows.Next() {
		var item model.AbuseTokenRankItem
		var latest string
		if err := rows.Scan(&item.Token, &item.UploadCount, &item.TotalSize, &latest); err != nil {
			return nil, err
		}
		item.TokenPreview = previewValue(item.Token, 8)
		item.LatestUploadAt = parseTime(latest)
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *Repository) IPDetail(ctx context.Context, ipAddress string) (model.AbuseIPDetail, error) {
	summary, err := r.ImageSummaryByIP(ctx, ipAddress)
	if err != nil {
		return model.AbuseIPDetail{}, err
	}
	detail := model.AbuseIPDetail{
		IPAddress:       ipAddress,
		IPAddressMasked: maskIPValue(ipAddress),
		UploadCount:     summary.Count,
		TotalSize:       summary.TotalSize,
	}
	ban, err := r.FindActiveIPBanByIP(ctx, ipAddress)
	if err == nil {
		detail.IsBanned = true
		detail.Ban = &ban
		return detail, nil
	}
	if IsNotFound(err) {
		return detail, nil
	}
	return model.AbuseIPDetail{}, err
}

func (r *Repository) ListPublicAnnouncements(ctx context.Context, limit int) ([]model.Announcement, error) {
	if limit < 1 {
		limit = 10
	}
	if limit > 50 {
		limit = 50
	}
	now := time.Now().UTC().Format(time.RFC3339)
	rows, err := r.db.QueryContext(
		ctx,
		`SELECT id, title, content, status, priority, starts_at, ends_at, sort_order, created_at, updated_at
		 FROM announcements
		 WHERE status = ?
		   AND (starts_at IS NULL OR starts_at = '' OR starts_at <= ?)
		   AND (ends_at IS NULL OR ends_at = '' OR ends_at > ?)
		 ORDER BY CASE priority WHEN 'urgent' THEN 3 WHEN 'important' THEN 2 ELSE 1 END DESC, sort_order DESC, created_at DESC, id DESC
		 LIMIT ?`,
		model.AnnouncementStatusPublished,
		now,
		now,
		limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanAnnouncements(rows)
}

func (r *Repository) ListAnnouncements(ctx context.Context) ([]model.Announcement, error) {
	rows, err := r.db.QueryContext(
		ctx,
		`SELECT id, title, content, status, priority, starts_at, ends_at, sort_order, created_at, updated_at
		 FROM announcements
		 ORDER BY updated_at DESC, id DESC`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanAnnouncements(rows)
}

func (r *Repository) GetAnnouncement(ctx context.Context, id int64) (model.Announcement, error) {
	row := r.db.QueryRowContext(
		ctx,
		`SELECT id, title, content, status, priority, starts_at, ends_at, sort_order, created_at, updated_at
		 FROM announcements
		 WHERE id = ?`,
		id,
	)
	return scanAnnouncement(row)
}

func (r *Repository) CreateAnnouncement(ctx context.Context, announcement model.Announcement) (model.Announcement, error) {
	now := time.Now().UTC()
	if announcement.CreatedAt.IsZero() {
		announcement.CreatedAt = now
	}
	announcement.UpdatedAt = now
	result, err := r.db.ExecContext(
		ctx,
		`INSERT INTO announcements(title, content, status, priority, starts_at, ends_at, sort_order, created_at, updated_at)
		 VALUES(?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		announcement.Title,
		announcement.Content,
		announcement.Status,
		announcement.Priority,
		nullableTimeString(announcement.StartsAt),
		nullableTimeString(announcement.EndsAt),
		announcement.SortOrder,
		announcement.CreatedAt.Format(time.RFC3339),
		announcement.UpdatedAt.Format(time.RFC3339),
	)
	if err != nil {
		return model.Announcement{}, err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return model.Announcement{}, err
	}
	return r.GetAnnouncement(ctx, id)
}

func (r *Repository) UpdateAnnouncement(ctx context.Context, announcement model.Announcement) (model.Announcement, error) {
	announcement.UpdatedAt = time.Now().UTC()
	result, err := r.db.ExecContext(
		ctx,
		`UPDATE announcements
		 SET title = ?, content = ?, status = ?, priority = ?, starts_at = ?, ends_at = ?, sort_order = ?, updated_at = ?
		 WHERE id = ?`,
		announcement.Title,
		announcement.Content,
		announcement.Status,
		announcement.Priority,
		nullableTimeString(announcement.StartsAt),
		nullableTimeString(announcement.EndsAt),
		announcement.SortOrder,
		announcement.UpdatedAt.Format(time.RFC3339),
		announcement.ID,
	)
	if err != nil {
		return model.Announcement{}, err
	}
	if err := ensureRowsAffected(result); err != nil {
		return model.Announcement{}, err
	}
	return r.GetAnnouncement(ctx, announcement.ID)
}

func (r *Repository) DeleteAnnouncement(ctx context.Context, id int64) error {
	result, err := r.db.ExecContext(ctx, `DELETE FROM announcements WHERE id = ?`, id)
	if err != nil {
		return err
	}
	if err := ensureRowsAffected(result); err != nil {
		return err
	}
	return nil
}

func (r *Repository) ArchiveAnnouncement(ctx context.Context, id int64) (model.Announcement, error) {
	result, err := r.db.ExecContext(
		ctx,
		`UPDATE announcements SET status = ?, updated_at = ? WHERE id = ?`,
		model.AnnouncementStatusArchived,
		time.Now().UTC().Format(time.RFC3339),
		id,
	)
	if err != nil {
		return model.Announcement{}, err
	}
	if err := ensureRowsAffected(result); err != nil {
		return model.Announcement{}, err
	}
	return r.GetAnnouncement(ctx, id)
}

func scanImage(scanner interface{ Scan(dest ...any) error }) (model.ImageRecord, error) {
	var record model.ImageRecord
	var createdAt string
	err := scanner.Scan(
		&record.ID,
		&record.UID,
		&record.Token,
		&record.StorageKey,
		&record.StorageBackend,
		&record.FilePath,
		&record.MIMEType,
		&record.Size,
		&record.MD5Hash,
		&record.IPAddress,
		&createdAt,
	)
	if err != nil {
		return model.ImageRecord{}, err
	}
	record.CreatedAt = parseTime(createdAt)
	return record, nil
}

func scanImages(rows *sql.Rows) ([]model.ImageRecord, error) {
	var records []model.ImageRecord
	for rows.Next() {
		record, err := scanImage(rows)
		if err != nil {
			return nil, err
		}
		records = append(records, record)
	}
	return records, rows.Err()
}

func scanIPBan(scanner interface{ Scan(dest ...any) error }) (model.IPBan, error) {
	var ban model.IPBan
	var expiresAt sql.NullString
	var createdAt string
	var updatedAt string
	err := scanner.Scan(
		&ban.ID,
		&ban.IPHash,
		&ban.IPAddress,
		&ban.IPAddressMasked,
		&ban.Reason,
		&expiresAt,
		&createdAt,
		&updatedAt,
	)
	if err != nil {
		return model.IPBan{}, err
	}
	ban.ExpiresAt = parseNullableTime(expiresAt)
	ban.CreatedAt = parseTime(createdAt)
	ban.UpdatedAt = parseTime(updatedAt)
	return ban, nil
}

func scanIPBans(rows *sql.Rows) ([]model.IPBan, error) {
	var bans []model.IPBan
	for rows.Next() {
		ban, err := scanIPBan(rows)
		if err != nil {
			return nil, err
		}
		bans = append(bans, ban)
	}
	return bans, rows.Err()
}

func scanAnnouncement(scanner interface{ Scan(dest ...any) error }) (model.Announcement, error) {
	var record model.Announcement
	var startsAt sql.NullString
	var endsAt sql.NullString
	var createdAt string
	var updatedAt string
	err := scanner.Scan(
		&record.ID,
		&record.Title,
		&record.Content,
		&record.Status,
		&record.Priority,
		&startsAt,
		&endsAt,
		&record.SortOrder,
		&createdAt,
		&updatedAt,
	)
	if err != nil {
		return model.Announcement{}, err
	}
	record.StartsAt = parseNullableTime(startsAt)
	record.EndsAt = parseNullableTime(endsAt)
	record.CreatedAt = parseTime(createdAt)
	record.UpdatedAt = parseTime(updatedAt)
	return record, nil
}

func scanAnnouncements(rows *sql.Rows) ([]model.Announcement, error) {
	var records []model.Announcement
	for rows.Next() {
		record, err := scanAnnouncement(rows)
		if err != nil {
			return nil, err
		}
		records = append(records, record)
	}
	return records, rows.Err()
}

func scanStorageConfig(scanner interface{ Scan(dest ...any) error }) (config.RuntimeStorageConfig, error) {
	var (
		record           config.RuntimeStorageConfig
		isDefault        int
		s3UseSSL         string
		s3ForcePathStyle string
	)
	err := scanner.Scan(
		&record.StorageKey,
		&record.Name,
		&record.Backend,
		&isDefault,
		&record.LocalStoragePath,
		&record.S3Endpoint,
		&record.S3Region,
		&record.S3Bucket,
		&record.S3AccessKey,
		&record.S3SecretKey,
		&s3UseSSL,
		&s3ForcePathStyle,
		&record.WebDAVURL,
		&record.WebDAVUser,
		&record.WebDAVPass,
	)
	if err != nil {
		return config.RuntimeStorageConfig{}, err
	}
	record.IsDefault = isDefault == 1
	record.S3UseSSL = parseBool(s3UseSSL)
	record.S3ForcePathStyle = parseBool(s3ForcePathStyle)
	return record, nil
}

func (r *Repository) countByQuery(ctx context.Context, query string, args ...any) (int64, error) {
	var count int64
	if err := r.db.QueryRowContext(ctx, query, args...).Scan(&count); err != nil {
		return 0, err
	}
	return count, nil
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

func insertStorageConfig(ctx context.Context, execer execContexter, cfg config.RuntimeStorageConfig, timestamp string) error {
	_, err := execer.ExecContext(ctx, storageConfigInsertSQL, storageConfigInsertArgs(cfg, timestamp)...)
	return err
}

func storageConfigInsertArgs(cfg config.RuntimeStorageConfig, timestamp string) []any {
	return []any{
		cfg.StorageKey,
		cfg.Name,
		cfg.Backend,
		boolInt(cfg.IsDefault),
		cfg.LocalStoragePath,
		cfg.S3Endpoint,
		cfg.S3Region,
		cfg.S3Bucket,
		cfg.S3AccessKey,
		cfg.S3SecretKey,
		boolString(cfg.S3UseSSL),
		boolString(cfg.S3ForcePathStyle),
		cfg.WebDAVURL,
		cfg.WebDAVUser,
		cfg.WebDAVPass,
		timestamp,
		timestamp,
	}
}

func ensureRowsAffected(result sql.Result) error {
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func (r *Repository) insertStorageConfigs(ctx context.Context, configs []config.RuntimeStorageConfig) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	now := time.Now().UTC().Format(time.RFC3339)
	for _, cfg := range configs {
		if err := insertStorageConfig(ctx, tx, cfg, now); err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (r *Repository) normalizeDefaultStorageConfig(ctx context.Context) error {
	rows, err := r.db.QueryContext(ctx, `SELECT storage_key, is_default FROM storage_configs ORDER BY id ASC`)
	if err != nil {
		return err
	}
	defer rows.Close()

	keys := make([]string, 0)
	defaultKey := ""
	for rows.Next() {
		var (
			key       string
			isDefault int
		)
		if err := rows.Scan(&key, &isDefault); err != nil {
			return err
		}
		keys = append(keys, key)
		if defaultKey == "" && isDefault == 1 {
			defaultKey = key
		}
	}
	if err := rows.Err(); err != nil {
		return err
	}
	if len(keys) == 0 {
		return errors.New("no storage configs available")
	}
	if defaultKey == "" {
		defaultKey = keys[0]
	}
	return r.SetDefaultStorageConfig(ctx, defaultKey)
}

func (r *Repository) backfillImageStorageKeys(ctx context.Context, backendToKey map[string]string, defaultKey string) error {
	for backend, storageKey := range backendToKey {
		if storageKey == "" {
			continue
		}
		if _, err := r.db.ExecContext(
			ctx,
			`UPDATE images
			 SET storage_key = ?
			 WHERE COALESCE(storage_key, '') = '' AND storage_backend = ?`,
			storageKey,
			backend,
		); err != nil {
			return err
		}
	}
	if defaultKey == "" {
		return nil
	}
	_, err := r.db.ExecContext(
		ctx,
		`UPDATE images
		 SET storage_key = ?
		 WHERE COALESCE(storage_key, '') = ''`,
		defaultKey,
	)
	return err
}

func initialStorageConfigs(legacy config.RuntimeStorageConfig, envDefault config.RuntimeStorageConfig) []config.RuntimeStorageConfig {
	if !hasLegacyStorageSeed(legacy) {
		return []config.RuntimeStorageConfig{normalizeSeedConfig(envDefault)}
	}

	defaultBackend := config.NormalizeStorageBackend(legacy.Backend)
	if defaultBackend == "" {
		defaultBackend = config.NormalizeStorageBackend(envDefault.Backend)
	}
	if defaultBackend == "" {
		defaultBackend = config.StorageBackendLocal
	}

	configs := []config.RuntimeStorageConfig{
		{
			StorageKey:       config.BootstrapStorageKey(config.StorageBackendLocal),
			Name:             config.BootstrapStorageName(config.StorageBackendLocal),
			IsDefault:        defaultBackend == config.StorageBackendLocal,
			Backend:          config.StorageBackendLocal,
			LocalStoragePath: firstNonEmpty(legacy.LocalStoragePath, envDefault.LocalStoragePath, "data/images"),
		},
	}

	if hasS3Seed(legacy) {
		configs = append(configs, config.RuntimeStorageConfig{
			StorageKey:       config.BootstrapStorageKey(config.StorageBackendS3),
			Name:             config.BootstrapStorageName(config.StorageBackendS3),
			IsDefault:        defaultBackend == config.StorageBackendS3,
			Backend:          config.StorageBackendS3,
			LocalStoragePath: firstNonEmpty(legacy.LocalStoragePath, envDefault.LocalStoragePath, "data/images"),
			S3Endpoint:       legacy.S3Endpoint,
			S3Region:         firstNonEmpty(legacy.S3Region, envDefault.S3Region, "auto"),
			S3Bucket:         legacy.S3Bucket,
			S3AccessKey:      legacy.S3AccessKey,
			S3SecretKey:      legacy.S3SecretKey,
			S3UseSSL:         legacy.S3UseSSL,
			S3ForcePathStyle: legacy.S3ForcePathStyle,
		})
	}
	if hasWebDAVSeed(legacy) {
		configs = append(configs, config.RuntimeStorageConfig{
			StorageKey:       config.BootstrapStorageKey(config.StorageBackendWebDAV),
			Name:             config.BootstrapStorageName(config.StorageBackendWebDAV),
			IsDefault:        defaultBackend == config.StorageBackendWebDAV,
			Backend:          config.StorageBackendWebDAV,
			LocalStoragePath: firstNonEmpty(legacy.LocalStoragePath, envDefault.LocalStoragePath, "data/images"),
			WebDAVURL:        legacy.WebDAVURL,
			WebDAVUser:       legacy.WebDAVUser,
			WebDAVPass:       legacy.WebDAVPass,
		})
	}

	if !hasDefaultConfig(configs) {
		configs[0].IsDefault = true
	}
	return configs
}

func normalizeSeedConfig(cfg config.RuntimeStorageConfig) config.RuntimeStorageConfig {
	if strings.TrimSpace(cfg.StorageKey) == "" {
		cfg.StorageKey = config.BootstrapStorageKey(cfg.Backend)
	}
	if strings.TrimSpace(cfg.Name) == "" {
		cfg.Name = config.BootstrapStorageName(cfg.Backend)
	}
	if cfg.Backend == "" {
		cfg.Backend = config.StorageBackendLocal
	}
	cfg.IsDefault = true
	if cfg.LocalStoragePath == "" {
		cfg.LocalStoragePath = "data/images"
	}
	if cfg.S3Region == "" {
		cfg.S3Region = "auto"
	}
	return cfg
}

func buildStorageCatalog(configs []config.RuntimeStorageConfig) config.RuntimeStorageCatalog {
	catalog := config.RuntimeStorageCatalog{
		StorageConfigs: configs,
	}
	for _, cfg := range configs {
		if cfg.IsDefault {
			catalog.DefaultStorageKey = cfg.StorageKey
			break
		}
	}
	if catalog.DefaultStorageKey == "" && len(configs) > 0 {
		catalog.DefaultStorageKey = configs[0].StorageKey
	}
	return catalog
}

func keyByBackend(configs []config.RuntimeStorageConfig) map[string]string {
	values := make(map[string]string, len(configs))
	for _, cfg := range configs {
		if _, exists := values[cfg.Backend]; !exists || cfg.IsDefault {
			values[cfg.Backend] = cfg.StorageKey
		}
	}
	return values
}

func hasLegacyStorageSeed(cfg config.RuntimeStorageConfig) bool {
	return strings.TrimSpace(cfg.Backend) != "" ||
		strings.TrimSpace(cfg.LocalStoragePath) != "" ||
		hasS3Seed(cfg) ||
		hasWebDAVSeed(cfg)
}

func hasS3Seed(cfg config.RuntimeStorageConfig) bool {
	return strings.TrimSpace(cfg.S3Endpoint) != "" &&
		strings.TrimSpace(cfg.S3Bucket) != "" &&
		strings.TrimSpace(cfg.S3AccessKey) != "" &&
		strings.TrimSpace(cfg.S3SecretKey) != ""
}

func hasWebDAVSeed(cfg config.RuntimeStorageConfig) bool {
	return strings.TrimSpace(cfg.WebDAVURL) != ""
}

func hasDefaultConfig(configs []config.RuntimeStorageConfig) bool {
	for _, cfg := range configs {
		if cfg.IsDefault {
			return true
		}
	}
	return false
}

func parseTime(value string) time.Time {
	if value == "" {
		return time.Time{}
	}
	layouts := []string{time.RFC3339Nano, time.RFC3339, "2006-01-02 15:04:05"}
	for _, layout := range layouts {
		if parsed, err := time.Parse(layout, value); err == nil {
			return parsed.UTC()
		}
	}
	return time.Time{}
}

func parseNullableTime(value sql.NullString) *time.Time {
	if !value.Valid || strings.TrimSpace(value.String) == "" {
		return nil
	}
	parsed := parseTime(value.String)
	if parsed.IsZero() {
		return nil
	}
	return &parsed
}

func nullableTimeString(value *time.Time) any {
	if value == nil || value.IsZero() {
		return nil
	}
	return value.UTC().Format(time.RFC3339)
}

func ipHashValue(ipAddress string) string {
	return iputil.Hash(ipAddress)
}

func maskIPValue(ipAddress string) string {
	return iputil.Mask(ipAddress)
}

func previewValue(value string, max int) string {
	trimmed := strings.TrimSpace(value)
	if max < 1 || len(trimmed) <= max {
		return trimmed
	}
	return trimmed[:max] + "..."
}

func boolString(value bool) string {
	if value {
		return "true"
	}
	return "false"
}

func boolInt(value bool) int {
	if value {
		return 1
	}
	return 0
}

func parseBool(value string) bool {
	return value == "true" || value == "1" || value == "yes"
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}
	return ""
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

func IsNotFound(err error) bool {
	return errors.Is(err, sql.ErrNoRows)
}
