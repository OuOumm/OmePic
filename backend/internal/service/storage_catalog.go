package service

import (
	"context"
	"fmt"
	"strings"

	"omepic/backend/internal/config"
	"omepic/backend/internal/repository"
	"omepic/backend/internal/secrets"
	"omepic/backend/internal/storage"
)

type storageCatalog struct {
	repo    *repository.Repository
	manager storageReconfigurer
	encrypt *secrets.SecretEncryptor
}

type storageReconfigurer interface {
	Reconfigure([]config.RuntimeStorageConfig) error
}

type legacyStorageConfigPatch struct {
	TargetStorageKey  string
	DefaultStorageKey *string
	Update            AdminStorageConfigUpdateInput
	HasPatch          bool
}

func (s *AdminService) storageCatalog() storageCatalog {
	return storageCatalog{repo: s.repo, manager: s.storage, encrypt: s.secretEncryptor}
}

func (c storageCatalog) View(ctx context.Context) (AdminConfigView, error) {
	configs, err := c.repo.ListStorageConfigs(ctx)
	if err != nil {
		return AdminConfigView{}, fmt.Errorf("%w: config query failed", ErrDependencyUnavailable)
	}
	return storageCatalogView(configs, c.encrypt), nil
}

func (c storageCatalog) Create(ctx context.Context, input AdminStorageConfigCreateInput) (AdminConfigView, error) {
	next, err := buildStorageConfig(input)
	if err != nil {
		return AdminConfigView{}, err
	}
	if err := storage.ValidateConfig(next); err != nil {
		return AdminConfigView{}, WithUserMessage(ErrInvalidInput, err.Error())
	}

	// Encrypt sensitive fields before storing in DB.
	encrypted, encErr := encryptStorageSecrets(next, c.encrypt)
	if encErr != nil {
		return AdminConfigView{}, encErr
	}

	configs, err := c.repo.ListStorageConfigs(ctx)
	if err != nil {
		return AdminConfigView{}, fmt.Errorf("%w: config query failed", ErrDependencyUnavailable)
	}
	// For validation, use the decrypted (original) values.
	configs = append(configs, next)
	if err := validateStorageCatalogReload(configs); err != nil {
		return AdminConfigView{}, err
	}

	if err := c.repo.CreateStorageConfig(ctx, encrypted); err != nil {
		return AdminConfigView{}, fmt.Errorf("%w: config save failed", ErrDependencyUnavailable)
	}
	if err := c.reload(ctx); err != nil {
		return AdminConfigView{}, err
	}
	return c.View(ctx)
}

func (c storageCatalog) Patch(ctx context.Context, storageKey string, input AdminStorageConfigUpdateInput) (AdminConfigView, error) {
	key := strings.TrimSpace(storageKey)
	if key == "" {
		return AdminConfigView{}, WithUserMessage(ErrInvalidInput, "storage key is required")
	}
	configs, currentIndex, err := c.loadCatalogWithTarget(ctx, key)
	if err != nil {
		return AdminConfigView{}, err
	}

	// Decrypt current config's sensitive fields for comparison with masked values.
	current := decryptStorageSecrets(configs[currentIndex], c.encrypt)

	next := current
	mergeStorageConfig(&next, current, input)
	if err := c.validatePatch(ctx, key, current, next); err != nil {
		return AdminConfigView{}, err
	}
	configs[currentIndex] = next
	if err := validateStorageCatalogReload(configs); err != nil {
		return AdminConfigView{}, err
	}

	// Encrypt sensitive fields before writing to DB.
	encrypted, encErr := encryptStorageSecrets(next, c.encrypt)
	if encErr != nil {
		return AdminConfigView{}, encErr
	}

	if err := c.repo.UpdateStorageConfig(ctx, encrypted); err != nil {
		if repository.IsNotFound(err) {
			return AdminConfigView{}, ErrNotFound
		}
		return AdminConfigView{}, fmt.Errorf("%w: config save failed", ErrDependencyUnavailable)
	}
	if err := c.reload(ctx); err != nil {
		return AdminConfigView{}, err
	}
	return c.View(ctx)
}

func (c storageCatalog) Delete(ctx context.Context, storageKey string) (AdminConfigView, error) {
	key := strings.TrimSpace(storageKey)
	if key == "" {
		return AdminConfigView{}, WithUserMessage(ErrInvalidInput, "storage key is required")
	}
	current, err := c.repo.GetStorageConfigByKey(ctx, key)
	if err != nil {
		if repository.IsNotFound(err) {
			return AdminConfigView{}, ErrNotFound
		}
		return AdminConfigView{}, fmt.Errorf("%w: config query failed", ErrDependencyUnavailable)
	}
	if current.IsDefault {
		return AdminConfigView{}, WithUserMessage(ErrConflict, "default storage instance cannot be deleted")
	}
	count, err := c.repo.CountImagesByStorageKey(ctx, key)
	if err != nil {
		return AdminConfigView{}, fmt.Errorf("%w: image usage lookup failed", ErrDependencyUnavailable)
	}
	if count > 0 {
		return AdminConfigView{}, WithUserMessage(ErrConflict, "storage instance is in use by existing images")
	}

	configs, err := c.repo.ListStorageConfigs(ctx)
	if err != nil {
		return AdminConfigView{}, fmt.Errorf("%w: config query failed", ErrDependencyUnavailable)
	}
	nextConfigs := make([]config.RuntimeStorageConfig, 0, len(configs)-1)
	for _, cfg := range configs {
		if cfg.StorageKey != key {
			nextConfigs = append(nextConfigs, cfg)
		}
	}
	if err := validateStorageCatalogReload(nextConfigs); err != nil {
		return AdminConfigView{}, err
	}

	if err := c.repo.DeleteStorageConfig(ctx, key); err != nil {
		if repository.IsNotFound(err) {
			return AdminConfigView{}, ErrNotFound
		}
		return AdminConfigView{}, fmt.Errorf("%w: config delete failed", ErrDependencyUnavailable)
	}
	if err := c.reload(ctx); err != nil {
		return AdminConfigView{}, err
	}
	return c.View(ctx)
}

func (c storageCatalog) SetDefault(ctx context.Context, storageKey string) (AdminConfigView, error) {
	key := strings.TrimSpace(storageKey)
	if key == "" {
		return AdminConfigView{}, WithUserMessage(ErrInvalidInput, "storage key is required")
	}
	configs, targetIndex, err := c.loadCatalogWithTarget(ctx, key)
	if err != nil {
		return AdminConfigView{}, err
	}
	for index := range configs {
		configs[index].IsDefault = index == targetIndex
	}
	if err := validateStorageCatalogReload(configs); err != nil {
		return AdminConfigView{}, err
	}

	if err := c.repo.SetDefaultStorageConfig(ctx, key); err != nil {
		if repository.IsNotFound(err) {
			return AdminConfigView{}, ErrNotFound
		}
		return AdminConfigView{}, fmt.Errorf("%w: default storage update failed", ErrDependencyUnavailable)
	}
	if err := c.reload(ctx); err != nil {
		return AdminConfigView{}, err
	}
	return c.View(ctx)
}

func (c storageCatalog) ApplyLegacyPatch(ctx context.Context, patch legacyStorageConfigPatch) (AdminConfigView, error) {
	if !patch.HasPatch && patch.DefaultStorageKey == nil {
		return c.View(ctx)
	}

	defaultKey := ""
	if patch.DefaultStorageKey != nil {
		defaultKey = strings.TrimSpace(*patch.DefaultStorageKey)
		if defaultKey == "" {
			return AdminConfigView{}, WithUserMessage(ErrInvalidInput, "default storage key is required")
		}
		if _, _, err := c.loadCatalogWithTarget(ctx, defaultKey); err != nil {
			return AdminConfigView{}, err
		}
	}

	if !patch.HasPatch {
		return c.SetDefault(ctx, defaultKey)
	}

	targetKey := strings.TrimSpace(patch.TargetStorageKey)
	if targetKey == "" {
		targetKey = defaultKey
	}
	if targetKey == "" {
		view, err := c.View(ctx)
		if err != nil {
			return AdminConfigView{}, err
		}
		targetKey = view.DefaultStorageKey
	}
	if targetKey == "" {
		return AdminConfigView{}, WithUserMessage(ErrInvalidInput, "storage key is required")
	}

	configs, targetIndex, err := c.loadCatalogWithTarget(ctx, targetKey)
	if err != nil {
		return AdminConfigView{}, err
	}
	// Decrypt current config's sensitive fields for comparison with masked values.
	current := decryptStorageSecrets(configs[targetIndex], c.encrypt)

	next := current
	mergeStorageConfig(&next, current, patch.Update)
	if err := c.validatePatch(ctx, targetKey, current, next); err != nil {
		return AdminConfigView{}, err
	}
	configs[targetIndex] = next
	if patch.DefaultStorageKey != nil {
		for index := range configs {
			configs[index].IsDefault = configs[index].StorageKey == defaultKey
		}
	}
	if err := validateStorageCatalogReload(configs); err != nil {
		return AdminConfigView{}, err
	}

	// Encrypt sensitive fields before writing to DB.
	encrypted, encErr := encryptStorageSecrets(next, c.encrypt)
	if encErr != nil {
		return AdminConfigView{}, encErr
	}

	if patch.DefaultStorageKey != nil {
		if err := c.repo.UpdateStorageConfigAndSetDefault(ctx, encrypted, defaultKey); err != nil {
			if repository.IsNotFound(err) {
				return AdminConfigView{}, ErrNotFound
			}
			return AdminConfigView{}, fmt.Errorf("%w: config save failed", ErrDependencyUnavailable)
		}
	} else if err := c.repo.UpdateStorageConfig(ctx, encrypted); err != nil {
		if repository.IsNotFound(err) {
			return AdminConfigView{}, ErrNotFound
		}
		return AdminConfigView{}, fmt.Errorf("%w: config save failed", ErrDependencyUnavailable)
	}
	if err := c.reload(ctx); err != nil {
		return AdminConfigView{}, err
	}
	return c.View(ctx)
}

func (c storageCatalog) loadCatalogWithTarget(ctx context.Context, storageKey string) ([]config.RuntimeStorageConfig, int, error) {
	configs, err := c.repo.ListStorageConfigs(ctx)
	if err != nil {
		return nil, -1, fmt.Errorf("%w: config query failed", ErrDependencyUnavailable)
	}
	for index, cfg := range configs {
		if cfg.StorageKey == storageKey {
			return configs, index, nil
		}
	}
	return nil, -1, ErrNotFound
}

func (c storageCatalog) validatePatch(ctx context.Context, storageKey string, current config.RuntimeStorageConfig, next config.RuntimeStorageConfig) error {
	if storageBackendChanged(current.Backend, next.Backend) {
		count, err := c.repo.CountImagesByStorageKey(ctx, storageKey)
		if err != nil {
			return fmt.Errorf("%w: image usage lookup failed", ErrDependencyUnavailable)
		}
		if count > 0 {
			return WithUserMessage(ErrConflict, "storage backend cannot change while images still reference this storage key")
		}
	}
	if strings.TrimSpace(next.Name) == "" {
		return WithUserMessage(ErrInvalidInput, "storage instance name is required")
	}
	if err := storage.ValidateConfig(next); err != nil {
		return WithUserMessage(ErrInvalidInput, err.Error())
	}
	return nil
}

func (c storageCatalog) reload(ctx context.Context) error {
	configs, err := c.repo.ListStorageConfigs(ctx)
	if err != nil {
		return fmt.Errorf("%w: config query failed", ErrDependencyUnavailable)
	}
	// Decrypt credentials before passing to storage manager — it needs real values.
	decrypted := decryptStorageCatalog(configs, c.encrypt)
	if c.manager == nil {
		return nil
	}
	if err := c.manager.Reconfigure(decrypted); err != nil {
		return fmt.Errorf("%w: storage reload failed", ErrDependencyUnavailable)
	}
	return nil
}

func validateStorageCatalogReload(configs []config.RuntimeStorageConfig) error {
	if _, _, err := storage.ValidateCatalog(configs); err != nil {
		return WithUserMessage(ErrInvalidInput, err.Error())
	}
	return nil
}

func storageCatalogView(configs []config.RuntimeStorageConfig, encrypt *secrets.SecretEncryptor) AdminConfigView {
	view := AdminConfigView{
		StorageConfigs: make([]AdminStorageConfigView, 0, len(configs)),
	}
	for _, cfg := range configs {
		if cfg.IsDefault {
			view.DefaultStorageKey = cfg.StorageKey
		}
		// Decrypt before masking for display.
		decrypted := decryptStorageSecrets(cfg, encrypt)
		view.StorageConfigs = append(view.StorageConfigs, maskStorageConfig(decrypted))
	}
	if view.DefaultStorageKey == "" && len(view.StorageConfigs) > 0 {
		view.DefaultStorageKey = view.StorageConfigs[0].StorageKey
	}
	return view
}

// encryptStorageSecrets encrypts the sensitive credential fields (s3_access_key,
// s3_secret_key, webdav_pass) using AES-256-GCM before writing to the database.
// If the encryptor is nil, the config is returned unchanged (no encryption).
// Returns an error if encryption fails — we must not store plaintext.
func encryptStorageSecrets(cfg config.RuntimeStorageConfig, encrypt *secrets.SecretEncryptor) (config.RuntimeStorageConfig, error) {
	if encrypt == nil {
		return cfg, nil
	}
	cfg = cloneStorageConfig(cfg)
	if cfg.S3AccessKey != "" {
		enc, err := encrypt.Encrypt(cfg.S3AccessKey)
		if err != nil {
			return cfg, fmt.Errorf("encrypt s3_access_key: %w", err)
		}
		cfg.S3AccessKey = enc
	}
	if cfg.S3SecretKey != "" {
		enc, err := encrypt.Encrypt(cfg.S3SecretKey)
		if err != nil {
			return cfg, fmt.Errorf("encrypt s3_secret_key: %w", err)
		}
		cfg.S3SecretKey = enc
	}
	if cfg.WebDAVPass != "" {
		enc, err := encrypt.Encrypt(cfg.WebDAVPass)
		if err != nil {
			return cfg, fmt.Errorf("encrypt webdav_pass: %w", err)
		}
		cfg.WebDAVPass = enc
	}
	return cfg, nil
}

// decryptStorageSecrets decrypts the sensitive credential fields using AES-256-GCM
// after reading from the database. If the encryptor is nil or a field looks like
// plaintext (not valid base64 or too short for a nonce), the field is returned unchanged.
func decryptStorageSecrets(cfg config.RuntimeStorageConfig, encrypt *secrets.SecretEncryptor) config.RuntimeStorageConfig {
	if encrypt == nil {
		return cfg
	}
	cfg = cloneStorageConfig(cfg)
	cfg.S3AccessKey = decryptField(cfg.S3AccessKey, encrypt)
	cfg.S3SecretKey = decryptField(cfg.S3SecretKey, encrypt)
	cfg.WebDAVPass = decryptField(cfg.WebDAVPass, encrypt)
	return cfg
}

// decryptField attempts to decrypt a field. If decryption fails (field is
// plaintext or empty), it returns the original value unchanged.
func decryptField(value string, encrypt *secrets.SecretEncryptor) string {
	if value == "" {
		return value
	}
	decrypted, err := encrypt.Decrypt(value)
	if err != nil {
		// Not encrypted ciphertext — return as-is (plaintext fallback).
		return value
	}
	return decrypted
}

// decryptStorageCatalog decrypts all configs in a catalog for use by the
// storage manager, which needs real credential values.
func decryptStorageCatalog(configs []config.RuntimeStorageConfig, encrypt *secrets.SecretEncryptor) []config.RuntimeStorageConfig {
	result := make([]config.RuntimeStorageConfig, len(configs))
	for i, cfg := range configs {
		result[i] = decryptStorageSecrets(cfg, encrypt)
	}
	return result
}

// cloneStorageConfig makes a shallow copy so encryption/decryption
// modifications don't affect the original struct.
func cloneStorageConfig(cfg config.RuntimeStorageConfig) config.RuntimeStorageConfig {
	return cfg
}

// EncryptStorageCatalogSecrets encrypts the sensitive fields of all storage
// configs currently in the database and writes them back. This is used at
// startup to encrypt the initial seed configs that were inserted as plaintext.
func EncryptStorageCatalogSecrets(ctx context.Context, repo *repository.Repository, encrypt *secrets.SecretEncryptor) error {
	configs, err := repo.ListStorageConfigs(ctx)
	if err != nil {
		return err
	}
	for _, cfg := range configs {
		encrypted, encErr := encryptStorageSecrets(cfg, encrypt)
		if encErr != nil {
			return encErr
		}
		if err := repo.UpdateStorageConfig(ctx, encrypted); err != nil {
			return err
		}
	}
	return nil
}

// DecryptStorageCatalogValues decrypts all storage configs in a catalog.
// This is used at startup to provide decrypted values to the storage manager.
func DecryptStorageCatalogValues(configs []config.RuntimeStorageConfig, encrypt *secrets.SecretEncryptor) []config.RuntimeStorageConfig {
	return decryptStorageCatalog(configs, encrypt)
}
