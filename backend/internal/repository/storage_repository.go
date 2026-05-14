package repository

import (
	"context"
	"errors"
	"strings"
	"time"

	"omepic/backend/internal/config"
)

const storageConfigInsertSQL = `INSERT INTO storage_configs(
	storage_key, name, backend, is_default, local_storage_path, s3_endpoint, s3_region, s3_bucket, s3_access_key, s3_secret_key, s3_use_ssl, s3_force_path_style, webdav_url, webdav_user, webdav_pass, created_at, updated_at
) VALUES(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

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
