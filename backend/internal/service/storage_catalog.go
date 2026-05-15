package service

import (
	"context"
	"fmt"
	"strings"

	"omepic/backend/internal/config"
	"omepic/backend/internal/repository"
	"omepic/backend/internal/storage"
)

type storageCatalog struct {
	repo    *repository.Repository
	manager storageReconfigurer
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
	return storageCatalog{repo: s.repo, manager: s.storage}
}

func (c storageCatalog) View(ctx context.Context) (AdminConfigView, error) {
	configs, err := c.repo.ListStorageConfigs(ctx)
	if err != nil {
		return AdminConfigView{}, fmt.Errorf("%w: config query failed", ErrDependencyUnavailable)
	}
	return storageCatalogView(configs), nil
}

func (c storageCatalog) Create(ctx context.Context, input AdminStorageConfigCreateInput) (AdminConfigView, error) {
	next, err := buildStorageConfig(input)
	if err != nil {
		return AdminConfigView{}, err
	}
	if err := storage.ValidateConfig(next); err != nil {
		return AdminConfigView{}, WithUserMessage(ErrInvalidInput, err.Error())
	}

	configs, err := c.repo.ListStorageConfigs(ctx)
	if err != nil {
		return AdminConfigView{}, fmt.Errorf("%w: config query failed", ErrDependencyUnavailable)
	}
	configs = append(configs, next)
	if err := validateStorageCatalogReload(configs); err != nil {
		return AdminConfigView{}, err
	}

	if err := c.repo.CreateStorageConfig(ctx, next); err != nil {
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

	current := configs[currentIndex]
	next := current
	mergeStorageConfig(&next, current, input)
	if err := c.validatePatch(ctx, key, current, next); err != nil {
		return AdminConfigView{}, err
	}
	configs[currentIndex] = next
	if err := validateStorageCatalogReload(configs); err != nil {
		return AdminConfigView{}, err
	}

	if err := c.repo.UpdateStorageConfig(ctx, next); err != nil {
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
	current := configs[targetIndex]
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

	if patch.DefaultStorageKey != nil {
		if err := c.repo.UpdateStorageConfigAndSetDefault(ctx, next, defaultKey); err != nil {
			if repository.IsNotFound(err) {
				return AdminConfigView{}, ErrNotFound
			}
			return AdminConfigView{}, fmt.Errorf("%w: config save failed", ErrDependencyUnavailable)
		}
	} else if err := c.repo.UpdateStorageConfig(ctx, next); err != nil {
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
	if c.manager == nil {
		return nil
	}
	if err := c.manager.Reconfigure(configs); err != nil {
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

func storageCatalogView(configs []config.RuntimeStorageConfig) AdminConfigView {
	view := AdminConfigView{
		StorageConfigs: make([]AdminStorageConfigView, 0, len(configs)),
	}
	for _, cfg := range configs {
		if cfg.IsDefault {
			view.DefaultStorageKey = cfg.StorageKey
		}
		view.StorageConfigs = append(view.StorageConfigs, maskStorageConfig(cfg))
	}
	if view.DefaultStorageKey == "" && len(view.StorageConfigs) > 0 {
		view.DefaultStorageKey = view.StorageConfigs[0].StorageKey
	}
	return view
}
