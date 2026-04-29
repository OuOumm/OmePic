package service

import (
	"context"
	"fmt"
	"strings"
	"time"

	"omepic/backend/internal/auth"
	"omepic/backend/internal/config"
	"omepic/backend/internal/model"
	"omepic/backend/internal/repository"
	"omepic/backend/internal/storage"
)

type AdminStorageConfigView struct {
	StorageKey       string `json:"storage_key"`
	Name             string `json:"name"`
	IsDefault        bool   `json:"is_default"`
	StorageBackend   string `json:"storage_backend"`
	LocalStoragePath string `json:"local_storage_path"`
	S3Endpoint       string `json:"s3_endpoint"`
	S3Region         string `json:"s3_region"`
	S3Bucket         string `json:"s3_bucket"`
	S3AccessKey      string `json:"s3_access_key"`
	S3SecretKey      string `json:"s3_secret_key"`
	S3UseSSL         bool   `json:"s3_use_ssl"`
	S3ForcePathStyle bool   `json:"s3_force_path_style"`
	WebDAVURL        string `json:"webdav_url"`
	WebDAVUser       string `json:"webdav_user"`
	WebDAVPass       string `json:"webdav_pass"`
}

type AdminConfigView struct {
	DefaultStorageKey string                   `json:"default_storage_key"`
	StorageConfigs    []AdminStorageConfigView `json:"storage_configs"`
}

type AdminConfigUpdateInput struct {
	DefaultStorageKey *string `json:"default_storage_key"`
	StorageKey        *string `json:"storage_key"`
	Name              *string `json:"name"`
	Backend           *string `json:"storage_backend"`
	LocalStoragePath  *string `json:"local_storage_path"`
	S3Endpoint        *string `json:"s3_endpoint"`
	S3Region          *string `json:"s3_region"`
	S3Bucket          *string `json:"s3_bucket"`
	S3AccessKey       *string `json:"s3_access_key"`
	S3SecretKey       *string `json:"s3_secret_key"`
	S3UseSSL          *bool   `json:"s3_use_ssl"`
	S3ForcePathStyle  *bool   `json:"s3_force_path_style"`
	WebDAVURL         *string `json:"webdav_url"`
	WebDAVUser        *string `json:"webdav_user"`
	WebDAVPass        *string `json:"webdav_pass"`
}

type AdminStorageConfigCreateInput struct {
	Name             string `json:"name"`
	Backend          string `json:"storage_backend"`
	LocalStoragePath string `json:"local_storage_path"`
	S3Endpoint       string `json:"s3_endpoint"`
	S3Region         string `json:"s3_region"`
	S3Bucket         string `json:"s3_bucket"`
	S3AccessKey      string `json:"s3_access_key"`
	S3SecretKey      string `json:"s3_secret_key"`
	S3UseSSL         bool   `json:"s3_use_ssl"`
	S3ForcePathStyle bool   `json:"s3_force_path_style"`
	WebDAVURL        string `json:"webdav_url"`
	WebDAVUser       string `json:"webdav_user"`
	WebDAVPass       string `json:"webdav_pass"`
}

type AdminStorageConfigUpdateInput struct {
	Name             *string `json:"name"`
	Backend          *string `json:"storage_backend"`
	LocalStoragePath *string `json:"local_storage_path"`
	S3Endpoint       *string `json:"s3_endpoint"`
	S3Region         *string `json:"s3_region"`
	S3Bucket         *string `json:"s3_bucket"`
	S3AccessKey      *string `json:"s3_access_key"`
	S3SecretKey      *string `json:"s3_secret_key"`
	S3UseSSL         *bool   `json:"s3_use_ssl"`
	S3ForcePathStyle *bool   `json:"s3_force_path_style"`
	WebDAVURL        *string `json:"webdav_url"`
	WebDAVUser       *string `json:"webdav_user"`
	WebDAVPass       *string `json:"webdav_pass"`
}

type AdminSetDefaultStorageInput struct {
	StorageKey string `json:"storage_key"`
}

type AdminImageList struct {
	Items    []AdminImageItem `json:"items"`
	Page     int              `json:"page"`
	PageSize int              `json:"page_size"`
	Total    int64            `json:"total"`
}

type AdminImageItem struct {
	ID             int64     `json:"id"`
	UID            string    `json:"uid"`
	Token          string    `json:"token"`
	StorageKey     string    `json:"storage_key"`
	StorageBackend string    `json:"storage_backend"`
	MIMEType       string    `json:"mime_type"`
	Size           int64     `json:"size"`
	MD5Hash        string    `json:"md5_hash"`
	IPAddress      string    `json:"ip_address"`
	CreatedAt      time.Time `json:"created_at"`
}

type AdminService struct {
	repo         *repository.Repository
	storage      *storage.Manager
	imageService *ImageService
	adminPass    string
	jwtSecret    string
}

func NewAdminService(repo *repository.Repository, storageManager *storage.Manager, imageService *ImageService, adminPassword string, jwtSecret string) *AdminService {
	return &AdminService{
		repo:         repo,
		storage:      storageManager,
		imageService: imageService,
		adminPass:    adminPassword,
		jwtSecret:    jwtSecret,
	}
}

func (s *AdminService) Login(password string) (string, error) {
	if strings.TrimSpace(password) == "" {
		return "", ErrInvalidInput
	}
	if password != s.adminPass {
		return "", ErrForbidden
	}
	token, err := auth.GenerateJWT(s.jwtSecret, 24*time.Hour)
	if err != nil {
		return "", fmt.Errorf("%w: jwt sign failed", ErrDependencyUnavailable)
	}
	return token, nil
}

func (s *AdminService) Status(ctx context.Context) (model.AdminStatus, error) {
	status, err := s.repo.AggregateStatus(ctx)
	if err != nil {
		return model.AdminStatus{}, fmt.Errorf("%w: status query failed", ErrDependencyUnavailable)
	}
	return status, nil
}

func (s *AdminService) Images(ctx context.Context, page int, pageSize int, search string) (AdminImageList, error) {
	items, total, err := s.repo.SearchImages(ctx, page, pageSize, search)
	if err != nil {
		return AdminImageList{}, fmt.Errorf("%w: image list query failed", ErrDependencyUnavailable)
	}

	viewItems := make([]AdminImageItem, 0, len(items))
	for _, item := range items {
		viewItems = append(viewItems, AdminImageItem{
			ID:             item.ID,
			UID:            item.UID,
			Token:          item.Token,
			StorageKey:     item.StorageKey,
			StorageBackend: item.StorageBackend,
			MIMEType:       item.MIMEType,
			Size:           item.Size,
			MD5Hash:        item.MD5Hash,
			IPAddress:      item.IPAddress,
			CreatedAt:      item.CreatedAt,
		})
	}

	return AdminImageList{
		Items:    viewItems,
		Page:     page,
		PageSize: pageSize,
		Total:    total,
	}, nil
}

func (s *AdminService) DeleteImages(ctx context.Context, uids []string) error {
	if len(uids) == 0 {
		return ErrInvalidInput
	}
	for _, uid := range uids {
		if err := s.imageService.Delete(ctx, uid, "", true); err != nil {
			return err
		}
	}
	return nil
}

func (s *AdminService) GetConfig(ctx context.Context) (AdminConfigView, error) {
	return s.loadConfigView(ctx)
}

func (s *AdminService) UpdateConfig(ctx context.Context, input AdminConfigUpdateInput) (AdminConfigView, error) {
	hasPatch := hasStorageConfigPatch(input)
	if !hasPatch && input.DefaultStorageKey == nil {
		return s.loadConfigView(ctx)
	}

	defaultStorageKey := ""
	if input.DefaultStorageKey != nil {
		defaultStorageKey = trimStringPointer(input.DefaultStorageKey)
		if defaultStorageKey == "" {
			return AdminConfigView{}, fmt.Errorf("%w: default storage key is required", ErrInvalidInput)
		}
		if err := s.ensureStorageConfigExists(ctx, defaultStorageKey); err != nil {
			return AdminConfigView{}, err
		}
	}

	if hasPatch {
		targetKey := trimStringPointer(input.StorageKey)
		if targetKey == "" {
			targetKey = defaultStorageKey
		}
		if targetKey == "" {
			view, err := s.loadConfigView(ctx)
			if err != nil {
				return AdminConfigView{}, err
			}
			targetKey = view.DefaultStorageKey
		}
		if targetKey == "" {
			return AdminConfigView{}, fmt.Errorf("%w: storage key is required", ErrInvalidInput)
		}

		view, err := s.UpdateStorageConfig(ctx, targetKey, storageUpdateFromConfigPatch(input))
		if err != nil {
			return AdminConfigView{}, err
		}
		if input.DefaultStorageKey == nil {
			return view, nil
		}
	}

	return s.SetDefaultStorageConfig(ctx, defaultStorageKey)
}

func (s *AdminService) CreateStorageConfig(ctx context.Context, input AdminStorageConfigCreateInput) (AdminConfigView, error) {
	next, err := buildStorageConfig(input)
	if err != nil {
		return AdminConfigView{}, err
	}
	if err := storage.ValidateConfig(next); err != nil {
		return AdminConfigView{}, fmt.Errorf("%w: %s", ErrInvalidInput, err.Error())
	}
	if err := s.repo.CreateStorageConfig(ctx, next); err != nil {
		return AdminConfigView{}, fmt.Errorf("%w: config save failed", ErrDependencyUnavailable)
	}
	if err := s.reloadStorageManager(ctx); err != nil {
		return AdminConfigView{}, err
	}
	return s.loadConfigView(ctx)
}

func (s *AdminService) UpdateStorageConfig(ctx context.Context, storageKey string, input AdminStorageConfigUpdateInput) (AdminConfigView, error) {
	current, err := s.repo.GetStorageConfigByKey(ctx, storageKey)
	if err != nil {
		if repository.IsNotFound(err) {
			return AdminConfigView{}, ErrNotFound
		}
		return AdminConfigView{}, fmt.Errorf("%w: config query failed", ErrDependencyUnavailable)
	}

	next := current
	mergeStorageConfig(&next, current, input)
	if storageBackendChanged(current.Backend, next.Backend) {
		count, err := s.repo.CountImagesByStorageKey(ctx, storageKey)
		if err != nil {
			return AdminConfigView{}, fmt.Errorf("%w: image usage lookup failed", ErrDependencyUnavailable)
		}
		if count > 0 {
			return AdminConfigView{}, fmt.Errorf("%w: storage backend cannot change while images still reference this storage key", ErrConflict)
		}
	}
	if strings.TrimSpace(next.Name) == "" {
		return AdminConfigView{}, fmt.Errorf("%w: storage instance name is required", ErrInvalidInput)
	}
	if err := storage.ValidateConfig(next); err != nil {
		return AdminConfigView{}, fmt.Errorf("%w: %s", ErrInvalidInput, err.Error())
	}
	if err := s.repo.UpdateStorageConfig(ctx, next); err != nil {
		if repository.IsNotFound(err) {
			return AdminConfigView{}, ErrNotFound
		}
		return AdminConfigView{}, fmt.Errorf("%w: config save failed", ErrDependencyUnavailable)
	}
	if err := s.reloadStorageManager(ctx); err != nil {
		return AdminConfigView{}, err
	}
	return s.loadConfigView(ctx)
}

func (s *AdminService) DeleteStorageConfig(ctx context.Context, storageKey string) (AdminConfigView, error) {
	current, err := s.repo.GetStorageConfigByKey(ctx, storageKey)
	if err != nil {
		if repository.IsNotFound(err) {
			return AdminConfigView{}, ErrNotFound
		}
		return AdminConfigView{}, fmt.Errorf("%w: config query failed", ErrDependencyUnavailable)
	}
	if current.IsDefault {
		return AdminConfigView{}, fmt.Errorf("%w: default storage instance cannot be deleted", ErrConflict)
	}
	count, err := s.repo.CountImagesByStorageKey(ctx, storageKey)
	if err != nil {
		return AdminConfigView{}, fmt.Errorf("%w: image usage lookup failed", ErrDependencyUnavailable)
	}
	if count > 0 {
		return AdminConfigView{}, fmt.Errorf("%w: storage instance is in use by existing images", ErrConflict)
	}
	if err := s.repo.DeleteStorageConfig(ctx, storageKey); err != nil {
		if repository.IsNotFound(err) {
			return AdminConfigView{}, ErrNotFound
		}
		return AdminConfigView{}, fmt.Errorf("%w: config delete failed", ErrDependencyUnavailable)
	}
	if err := s.reloadStorageManager(ctx); err != nil {
		return AdminConfigView{}, err
	}
	return s.loadConfigView(ctx)
}

func (s *AdminService) SetDefaultStorageConfig(ctx context.Context, storageKey string) (AdminConfigView, error) {
	if strings.TrimSpace(storageKey) == "" {
		return AdminConfigView{}, fmt.Errorf("%w: storage key is required", ErrInvalidInput)
	}
	if err := s.repo.SetDefaultStorageConfig(ctx, storageKey); err != nil {
		if repository.IsNotFound(err) {
			return AdminConfigView{}, ErrNotFound
		}
		return AdminConfigView{}, fmt.Errorf("%w: default storage update failed", ErrDependencyUnavailable)
	}
	if err := s.reloadStorageManager(ctx); err != nil {
		return AdminConfigView{}, err
	}
	return s.loadConfigView(ctx)
}

func (s *AdminService) loadConfigView(ctx context.Context) (AdminConfigView, error) {
	configs, err := s.repo.ListStorageConfigs(ctx)
	if err != nil {
		return AdminConfigView{}, fmt.Errorf("%w: config query failed", ErrDependencyUnavailable)
	}

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
	return view, nil
}

func (s *AdminService) reloadStorageManager(ctx context.Context) error {
	configs, err := s.repo.ListStorageConfigs(ctx)
	if err != nil {
		return fmt.Errorf("%w: config query failed", ErrDependencyUnavailable)
	}
	if err := s.storage.Reconfigure(configs); err != nil {
		return fmt.Errorf("%w: storage reload failed", ErrDependencyUnavailable)
	}
	return nil
}

func (s *AdminService) ensureStorageConfigExists(ctx context.Context, storageKey string) error {
	if _, err := s.repo.GetStorageConfigByKey(ctx, storageKey); err != nil {
		if repository.IsNotFound(err) {
			return ErrNotFound
		}
		return fmt.Errorf("%w: config query failed", ErrDependencyUnavailable)
	}
	return nil
}

func buildStorageConfig(input AdminStorageConfigCreateInput) (config.RuntimeStorageConfig, error) {
	name := strings.TrimSpace(input.Name)
	backend := strings.TrimSpace(strings.ToLower(input.Backend))
	if name == "" {
		return config.RuntimeStorageConfig{}, fmt.Errorf("%w: storage instance name is required", ErrInvalidInput)
	}
	if backend == "" {
		return config.RuntimeStorageConfig{}, fmt.Errorf("%w: storage backend is required", ErrInvalidInput)
	}

	return config.RuntimeStorageConfig{
		StorageKey:       newStorageKey(name, backend),
		Name:             name,
		Backend:          backend,
		LocalStoragePath: input.LocalStoragePath,
		S3Endpoint:       input.S3Endpoint,
		S3Region:         input.S3Region,
		S3Bucket:         input.S3Bucket,
		S3AccessKey:      input.S3AccessKey,
		S3SecretKey:      input.S3SecretKey,
		S3UseSSL:         input.S3UseSSL,
		S3ForcePathStyle: input.S3ForcePathStyle,
		WebDAVURL:        input.WebDAVURL,
		WebDAVUser:       input.WebDAVUser,
		WebDAVPass:       input.WebDAVPass,
	}, nil
}

func maskStorageConfig(cfg config.RuntimeStorageConfig) AdminStorageConfigView {
	return AdminStorageConfigView{
		StorageKey:       cfg.StorageKey,
		Name:             cfg.Name,
		IsDefault:        cfg.IsDefault,
		StorageBackend:   cfg.Backend,
		LocalStoragePath: cfg.LocalStoragePath,
		S3Endpoint:       cfg.S3Endpoint,
		S3Region:         cfg.S3Region,
		S3Bucket:         cfg.S3Bucket,
		S3AccessKey:      maskSecret(cfg.S3AccessKey),
		S3SecretKey:      maskSecret(cfg.S3SecretKey),
		S3UseSSL:         cfg.S3UseSSL,
		S3ForcePathStyle: cfg.S3ForcePathStyle,
		WebDAVURL:        cfg.WebDAVURL,
		WebDAVUser:       cfg.WebDAVUser,
		WebDAVPass:       maskSecret(cfg.WebDAVPass),
	}
}

func mergeStorageConfig(target *config.RuntimeStorageConfig, current config.RuntimeStorageConfig, update AdminStorageConfigUpdateInput) {
	if update.Name != nil {
		target.Name = strings.TrimSpace(*update.Name)
	}
	if update.Backend != nil {
		target.Backend = strings.TrimSpace(strings.ToLower(*update.Backend))
	}
	if update.LocalStoragePath != nil {
		target.LocalStoragePath = *update.LocalStoragePath
	}
	if update.S3Endpoint != nil {
		target.S3Endpoint = *update.S3Endpoint
	}
	if update.S3Region != nil {
		target.S3Region = *update.S3Region
	}
	if update.S3Bucket != nil {
		target.S3Bucket = *update.S3Bucket
	}
	if update.S3AccessKey != nil && *update.S3AccessKey != maskSecret(current.S3AccessKey) {
		target.S3AccessKey = *update.S3AccessKey
	}
	if update.S3SecretKey != nil && *update.S3SecretKey != maskSecret(current.S3SecretKey) {
		target.S3SecretKey = *update.S3SecretKey
	}
	if update.S3UseSSL != nil {
		target.S3UseSSL = *update.S3UseSSL
	}
	if update.S3ForcePathStyle != nil {
		target.S3ForcePathStyle = *update.S3ForcePathStyle
	}
	if update.WebDAVURL != nil {
		target.WebDAVURL = *update.WebDAVURL
	}
	if update.WebDAVUser != nil {
		target.WebDAVUser = *update.WebDAVUser
	}
	if update.WebDAVPass != nil && *update.WebDAVPass != maskSecret(current.WebDAVPass) {
		target.WebDAVPass = *update.WebDAVPass
	}
}

func maskSecret(value string) string {
	if value == "" {
		return ""
	}
	if len(value) <= 4 {
		return "****"
	}
	return strings.Repeat("*", len(value)-4) + value[len(value)-4:]
}

func newStorageKey(name string, backend string) string {
	base := slugify(fmt.Sprintf("%s-%s", backend, name))
	if base == "" {
		base = backend
	}
	return fmt.Sprintf("%s-%x", base, time.Now().UnixNano())
}

func slugify(value string) string {
	var builder strings.Builder
	lastDash := false
	for _, r := range strings.ToLower(value) {
		switch {
		case r >= 'a' && r <= 'z':
			builder.WriteRune(r)
			lastDash = false
		case r >= '0' && r <= '9':
			builder.WriteRune(r)
			lastDash = false
		default:
			if !lastDash {
				builder.WriteByte('-')
				lastDash = true
			}
		}
	}
	return strings.Trim(builder.String(), "-")
}

func storageBackendChanged(current string, next string) bool {
	return normalizeStorageBackend(current) != normalizeStorageBackend(next)
}

func normalizeStorageBackend(value string) string {
	return strings.TrimSpace(strings.ToLower(value))
}

func hasStorageConfigPatch(input AdminConfigUpdateInput) bool {
	return input.StorageKey != nil ||
		input.Name != nil ||
		input.Backend != nil ||
		input.LocalStoragePath != nil ||
		input.S3Endpoint != nil ||
		input.S3Region != nil ||
		input.S3Bucket != nil ||
		input.S3AccessKey != nil ||
		input.S3SecretKey != nil ||
		input.S3UseSSL != nil ||
		input.S3ForcePathStyle != nil ||
		input.WebDAVURL != nil ||
		input.WebDAVUser != nil ||
		input.WebDAVPass != nil
}

func storageUpdateFromConfigPatch(input AdminConfigUpdateInput) AdminStorageConfigUpdateInput {
	return AdminStorageConfigUpdateInput{
		Name:             input.Name,
		Backend:          input.Backend,
		LocalStoragePath: input.LocalStoragePath,
		S3Endpoint:       input.S3Endpoint,
		S3Region:         input.S3Region,
		S3Bucket:         input.S3Bucket,
		S3AccessKey:      input.S3AccessKey,
		S3SecretKey:      input.S3SecretKey,
		S3UseSSL:         input.S3UseSSL,
		S3ForcePathStyle: input.S3ForcePathStyle,
		WebDAVURL:        input.WebDAVURL,
		WebDAVUser:       input.WebDAVUser,
		WebDAVPass:       input.WebDAVPass,
	}
}

func trimStringPointer(value *string) string {
	if value == nil {
		return ""
	}
	return strings.TrimSpace(*value)
}
