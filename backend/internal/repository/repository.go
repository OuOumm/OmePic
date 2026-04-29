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
	"omepic/backend/internal/model"
)

type Repository struct {
	db *sql.DB
}

func New(databasePath string) (*Repository, error) {
	if err := os.MkdirAll(filepath.Dir(databasePath), 0o755); err != nil {
		return nil, err
	}
	db, err := sql.Open("sqlite", databasePath)
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(1)
	repository := &Repository{db: db}
	if _, err := db.Exec(`PRAGMA foreign_keys = ON;`); err != nil {
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
	}
	indexes := []string{
		`CREATE INDEX IF NOT EXISTS idx_images_uid ON images(uid);`,
		`CREATE INDEX IF NOT EXISTS idx_images_md5_hash ON images(md5_hash);`,
		`CREATE INDEX IF NOT EXISTS idx_images_file_path ON images(file_path);`,
		`CREATE INDEX IF NOT EXISTS idx_images_storage_key ON images(storage_key);`,
		`CREATE INDEX IF NOT EXISTS idx_storage_configs_default ON storage_configs(is_default);`,
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
	_, err := r.db.ExecContext(
		ctx,
		`INSERT INTO storage_configs(
			storage_key, name, backend, is_default, local_storage_path, s3_endpoint, s3_region, s3_bucket, s3_access_key, s3_secret_key, s3_use_ssl, s3_force_path_style, webdav_url, webdav_user, webdav_pass, created_at, updated_at
		) VALUES(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
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
		time.Now().UTC().Format(time.RFC3339),
		time.Now().UTC().Format(time.RFC3339),
	)
	return err
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
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func (r *Repository) DeleteStorageConfig(ctx context.Context, storageKey string) error {
	result, err := r.db.ExecContext(ctx, `DELETE FROM storage_configs WHERE storage_key = ?`, storageKey)
	if err != nil {
		return err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return sql.ErrNoRows
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
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return sql.ErrNoRows
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
	row := r.db.QueryRowContext(ctx, `SELECT id, uid, token, storage_key, storage_backend, file_path, mime_type, size, md5_hash, ip_address, created_at FROM images WHERE uid = ?`, uid)
	record, err := scanImage(row)
	if err != nil {
		return nil, err
	}
	return &record, nil
}

func (r *Repository) FindByMD5(ctx context.Context, md5Hash string) (*model.ImageRecord, error) {
	row := r.db.QueryRowContext(ctx, `SELECT id, uid, token, storage_key, storage_backend, file_path, mime_type, size, md5_hash, ip_address, created_at FROM images WHERE md5_hash = ? ORDER BY id ASC LIMIT 1`, md5Hash)
	record, err := scanImage(row)
	if err != nil {
		return nil, err
	}
	return &record, nil
}

func (r *Repository) FindByMD5AndStorageKey(ctx context.Context, md5Hash string, storageKey string) (*model.ImageRecord, error) {
	row := r.db.QueryRowContext(ctx, `SELECT id, uid, token, storage_key, storage_backend, file_path, mime_type, size, md5_hash, ip_address, created_at FROM images WHERE md5_hash = ? AND storage_key = ? ORDER BY id ASC LIMIT 1`, md5Hash, storageKey)
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
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return sql.ErrNoRows
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

func (r *Repository) ListAllImages(ctx context.Context) ([]model.ImageRecord, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT id, uid, token, storage_key, storage_backend, file_path, mime_type, size, md5_hash, ip_address, created_at FROM images ORDER BY id ASC`)
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
		`SELECT id, uid, token, storage_key, storage_backend, file_path, mime_type, size, md5_hash, ip_address, created_at
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

func (r *Repository) insertStorageConfigs(ctx context.Context, configs []config.RuntimeStorageConfig) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	now := time.Now().UTC().Format(time.RFC3339)
	for _, cfg := range configs {
		if _, err := tx.ExecContext(
			ctx,
			`INSERT INTO storage_configs(
				storage_key, name, backend, is_default, local_storage_path, s3_endpoint, s3_region, s3_bucket, s3_access_key, s3_secret_key, s3_use_ssl, s3_force_path_style, webdav_url, webdav_user, webdav_pass, created_at, updated_at
			) VALUES(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
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
			now,
			now,
		); err != nil {
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

	defaultBackend := strings.TrimSpace(strings.ToLower(legacy.Backend))
	if defaultBackend == "" {
		defaultBackend = envDefault.Backend
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
