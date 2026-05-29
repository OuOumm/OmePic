package service

import (
	"context"
	"fmt"
	"strings"
	"time"
	"unicode"

	"golang.org/x/crypto/bcrypt"

	"omepic/backend/internal/auth"
	"omepic/backend/internal/config"
	"omepic/backend/internal/model"
	"omepic/backend/internal/repository"
	"omepic/backend/internal/storage"
)

type AdminStorageConfigView struct {
	StorageKey                  string   `json:"storage_key"`
	Name                        string   `json:"name"`
	IsDefault                   bool     `json:"is_default"`
	StorageBackend              string   `json:"storage_backend"`
	LocalStoragePath            string   `json:"local_storage_path"`
	S3Endpoint                  string   `json:"s3_endpoint"`
	S3Region                    string   `json:"s3_region"`
	S3Bucket                    string   `json:"s3_bucket"`
	S3AccessKey                 string   `json:"s3_access_key"`
	S3SecretKey                 string   `json:"s3_secret_key"`
	S3UseSSL                    bool     `json:"s3_use_ssl"`
	S3ForcePathStyle            bool     `json:"s3_force_path_style"`
	WebDAVURL                   string   `json:"webdav_url"`
	WebDAVUser                  string   `json:"webdav_user"`
	WebDAVPass                  string   `json:"webdav_pass"`
	MaxUploadSizeMB             int      `json:"max_upload_size_mb"`
	AllowedMIMETypes            []string `json:"allowed_mime_types"`
	AvifQuality                 int      `json:"avif_quality"`
	AvifSpeed                   int      `json:"avif_speed"`
	MaxImagePixels              int64    `json:"max_image_pixels"`
	AVIFMaxConcurrency          int      `json:"avif_max_concurrency"`
	AVIFConversionTimeoutSeconds int     `json:"avif_conversion_timeout_seconds"`
}

type AdminConfigView struct {
	DefaultStorageKey string                   `json:"default_storage_key"`
	StorageConfigs    []AdminStorageConfigView `json:"storage_configs"`
}

type AdminConfigUpdateInput struct {
	DefaultStorageKey *string `json:"default_storage_key"`
	StorageKey        *string `json:"storage_key"`
	config.RuntimeStorageUpdate
}

type AdminStorageConfigCreateInput struct {
	StorageKey                  string `json:"storage_key"`
	Name                        string `json:"name"`
	Backend                     string `json:"storage_backend"`
	LocalStoragePath            string `json:"local_storage_path"`
	S3Endpoint                  string `json:"s3_endpoint"`
	S3Region                    string `json:"s3_region"`
	S3Bucket                    string `json:"s3_bucket"`
	S3AccessKey                 string `json:"s3_access_key"`
	S3SecretKey                 string `json:"s3_secret_key"`
	S3UseSSL                    bool   `json:"s3_use_ssl"`
	S3ForcePathStyle            bool   `json:"s3_force_path_style"`
	WebDAVURL                   string `json:"webdav_url"`
	WebDAVUser                  string `json:"webdav_user"`
	WebDAVPass                  string `json:"webdav_pass"`
	MaxUploadSizeMB             int    `json:"max_upload_size_mb"`
	AllowedMIMETypes            string `json:"allowed_mime_types"`
	AvifQuality                 int    `json:"avif_quality"`
	AvifSpeed                   int    `json:"avif_speed"`
	MaxImagePixels              int64  `json:"max_image_pixels"`
	AVIFMaxConcurrency          int    `json:"avif_max_concurrency"`
	AVIFConversionTimeoutSeconds int   `json:"avif_conversion_timeout_seconds"`
}

type AdminStorageConfigUpdateInput = config.RuntimeStorageUpdate

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

type AdminIPBanCreateInput struct {
	UID           string `json:"uid"`
	IPAddress     string `json:"ip_address"`
	DurationHours int    `json:"duration_hours"`
	Reason        string `json:"reason"`
}

type AdminAbuseOverviewInput struct {
	From time.Time
	To   time.Time
}

type AdminIPBanCreateResult struct {
	Ban                model.IPBan `json:"ban"`
	AffectedImageCount int64       `json:"affected_image_count"`
	AffectedTotalSize  int64       `json:"affected_total_size"`
}

type AdminIPBanDeleteImagesResult struct {
	DeletedCount int `json:"deleted_count"`
}

type CloudflareImageCachePurgeResult struct {
	URL string `json:"url"`
}

const (
	adminPasswordHashConfigKey = "admin_password_hash"
)

type adminStorageManager interface {
	storageReconfigurer
	ForKey(string) (storage.ResolvedProvider, error)
	CurrentKey() string
}

type AdminService struct {
	repo         *repository.Repository
	storage      adminStorageManager
	settings     *RuntimeSettingsManager
	imageService *ImageService
	jwtSecret    string
	adminEnv     AdminEnvMetadata
}

type AdminEnvMetadata struct {
	HTTPAddr         string
	DatabasePath     string
	RedisURL         string
	UIDEncryptionKey string
	AdminPassword    string
}

func NewAdminService(repo *repository.Repository, storageManager *storage.Manager, settingsManager *RuntimeSettingsManager, imageService *ImageService, jwtSecret string, adminEnv AdminEnvMetadata) *AdminService {
	return &AdminService{
		repo:         repo,
		storage:      storageManager,
		settings:     settingsManager,
		imageService: imageService,
		jwtSecret:    jwtSecret,
		adminEnv:     adminEnv,
	}
}

func (s *AdminService) isPasswordSet(ctx context.Context) bool {
	_, err := s.repo.GetConfigValue(ctx, adminPasswordHashConfigKey)
	return err == nil
}

func (s *AdminService) Login(ctx context.Context, password string) (string, error) {
	if err := s.verifyAdminPassword(ctx, password); err != nil {
		return "", err
	}

	token, err := auth.GenerateJWT(s.jwtSecret, 24*time.Hour)
	if err != nil {
		return "", fmt.Errorf("%w: jwt sign failed", ErrDependencyUnavailable)
	}
	return token, nil
}

func (s *AdminService) ChangePassword(ctx context.Context, oldPassword string, newPassword string) error {
	if err := validateAdminPasswordStrength(newPassword); err != nil {
		return err
	}
	if err := s.verifyAdminPassword(ctx, oldPassword); err != nil {
		if err == ErrForbidden {
			return WithUserMessage(ErrForbidden, "current password is incorrect")
		}
		return err
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("%w: password hash failed", ErrDependencyUnavailable)
	}
	if err := s.repo.SetConfigValue(ctx, adminPasswordHashConfigKey, string(hash)); err != nil {
		return fmt.Errorf("%w: password save failed", ErrDependencyUnavailable)
	}
	return nil
}

func validateAdminPasswordStrength(password string) error {
	if strings.TrimSpace(password) == "" {
		return WithUserMessage(ErrInvalidInput, "new password is required")
	}
	if len([]rune(password)) < 8 {
		return WithUserMessage(ErrInvalidInput, "new password must be at least 8 characters and include uppercase, lowercase, and symbol characters")
	}

	hasUpper := false
	hasLower := false
	hasSymbol := false
	for _, r := range password {
		switch {
		case unicode.IsUpper(r):
			hasUpper = true
		case unicode.IsLower(r):
			hasLower = true
		case unicode.IsPunct(r) || unicode.IsSymbol(r):
			hasSymbol = true
		}
	}
	if !hasUpper || !hasLower || !hasSymbol {
		return WithUserMessage(ErrInvalidInput, "new password must be at least 8 characters and include uppercase, lowercase, and symbol characters")
	}
	return nil
}

func (s *AdminService) verifyAdminPassword(ctx context.Context, password string) error {
	if strings.TrimSpace(password) == "" {
		return ErrInvalidInput
	}
	storedHash, err := s.repo.GetConfigValue(ctx, adminPasswordHashConfigKey)
	if err != nil {
		if !repository.IsNotFound(err) {
			return fmt.Errorf("%w: password lookup failed", ErrDependencyUnavailable)
		}
		// First boot: no hash stored yet, bootstrap from ADMIN_PASSWORD env.
		if s.adminEnv.AdminPassword == "" {
			return fmt.Errorf("%w: admin password not configured", ErrDependencyUnavailable)
		}
		bootstrapHash, hashErr := bcrypt.GenerateFromPassword([]byte(s.adminEnv.AdminPassword), bcrypt.DefaultCost)
		if hashErr != nil {
			return fmt.Errorf("%w: password hash failed", ErrDependencyUnavailable)
		}
		if err := s.repo.SetConfigValue(ctx, adminPasswordHashConfigKey, string(bootstrapHash)); err != nil {
			return fmt.Errorf("%w: password save failed", ErrDependencyUnavailable)
		}
		storedHash = string(bootstrapHash)
	}

	if err := bcrypt.CompareHashAndPassword([]byte(storedHash), []byte(password)); err != nil {
		return ErrForbidden
	}
	return nil
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
		viewItems = append(viewItems, adminImageItemFromRecord(item))
	}

	return AdminImageList{
		Items:    viewItems,
		Page:     page,
		PageSize: pageSize,
		Total:    total,
	}, nil
}

func adminImageItemFromRecord(item model.ImageRecord) AdminImageItem {
	return AdminImageItem{
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
	}
}

func (s *AdminService) DeleteImages(ctx context.Context, uids []string) error {
	if len(uids) == 0 {
		return ErrInvalidInput
	}
	if s.imageService == nil {
		return fmt.Errorf("%w: image service is not configured", ErrDependencyUnavailable)
	}

	records := make([]model.ImageRecord, 0, len(uids))
	seen := make(map[string]struct{}, len(uids))
	for _, uid := range uids {
		normalizedUID, err := s.imageService.normalizeStoredUID(uid)
		if err != nil {
			return err
		}
		if _, ok := seen[normalizedUID]; ok {
			continue
		}
		seen[normalizedUID] = struct{}{}
		record, err := s.repo.FindByUID(ctx, normalizedUID)
		if err != nil {
			if repository.IsNotFound(err) {
				return ErrNotFound
			}
			return fmt.Errorf("%w: image lookup failed", ErrDependencyUnavailable)
		}
		records = append(records, *record)
	}

	if err := s.imageService.purgeImageURLCachesForRecords(ctx, records); err != nil {
		return err
	}

	for _, record := range records {
		if err := s.imageService.deleteRecord(ctx, record); err != nil {
			return err
		}
	}
	return nil
}

func (s *AdminService) PurgeCloudflareImageCache(ctx context.Context, rawURL string) (CloudflareImageCachePurgeResult, error) {
	settings := defaultRuntimeSettings()
	if s.settings != nil {
		settings = s.settings.Current()
	}
	if !settings.CloudflarePurgeEnabled {
		return CloudflareImageCachePurgeResult{}, WithUserMessage(ErrInvalidInput, "cloudflare purge is not enabled")
	}
	if strings.TrimSpace(settings.PublicBaseURL) == "" {
		return CloudflareImageCachePurgeResult{}, WithUserMessage(ErrInvalidInput, "public base url is required when cloudflare purge is enabled")
	}
	if s.imageService == nil {
		return CloudflareImageCachePurgeResult{}, fmt.Errorf("%w: image service is not configured", ErrDependencyUnavailable)
	}
	normalizedURL, err := s.imageService.PurgeImageURLCache(ctx, rawURL)
	if err != nil {
		return CloudflareImageCachePurgeResult{}, err
	}
	return CloudflareImageCachePurgeResult{URL: normalizedURL}, nil
}

func (s *AdminService) CreateIPBan(ctx context.Context, input AdminIPBanCreateInput) (AdminIPBanCreateResult, error) {
	return newSecurityAnalysis(s.repo).CreateIPBan(ctx, input)
}

func (s *AdminService) IPBans(ctx context.Context) ([]model.IPBan, error) {
	bans, err := s.repo.ListIPBans(ctx)
	if err != nil {
		return nil, fmt.Errorf("%w: ip bans query failed", ErrDependencyUnavailable)
	}
	return bans, nil
}

func (s *AdminService) AbuseOverview(ctx context.Context, input AdminAbuseOverviewInput) (model.AbuseOverview, error) {
	return newSecurityAnalysis(s.repo).Overview(ctx, input)
}

func (s *AdminService) AbuseIPDetail(ctx context.Context, ipAddress string) (model.AbuseIPDetail, error) {
	return newSecurityAnalysis(s.repo).IPDetail(ctx, ipAddress)
}

func (s *AdminService) DeleteIPBan(ctx context.Context, id int64) error {
	if id < 1 {
		return ErrInvalidInput
	}
	if err := s.repo.DeleteIPBan(ctx, id); err != nil {
		if repository.IsNotFound(err) {
			return ErrNotFound
		}
		return fmt.Errorf("%w: ip ban delete failed", ErrDependencyUnavailable)
	}
	return nil
}

func (s *AdminService) DeleteImagesByIPBan(ctx context.Context, id int64) (AdminIPBanDeleteImagesResult, error) {
	if id < 1 {
		return AdminIPBanDeleteImagesResult{}, ErrInvalidInput
	}
	ban, err := s.repo.GetIPBan(ctx, id)
	if err != nil {
		if repository.IsNotFound(err) {
			return AdminIPBanDeleteImagesResult{}, ErrNotFound
		}
		return AdminIPBanDeleteImagesResult{}, fmt.Errorf("%w: ip ban lookup failed", ErrDependencyUnavailable)
	}
	images, err := s.repo.ListImagesByIP(ctx, ban.IPAddress)
	if err != nil {
		return AdminIPBanDeleteImagesResult{}, fmt.Errorf("%w: ip image list failed", ErrDependencyUnavailable)
	}
	uids := make([]string, 0, len(images))
	for _, image := range images {
		uids = append(uids, image.UID)
	}
	if len(uids) == 0 {
		return AdminIPBanDeleteImagesResult{DeletedCount: 0}, nil
	}
	if err := s.DeleteImages(ctx, uids); err != nil {
		return AdminIPBanDeleteImagesResult{}, err
	}
	return AdminIPBanDeleteImagesResult{DeletedCount: len(uids)}, nil
}

func (s *AdminService) GetConfig(ctx context.Context) (AdminConfigView, error) {
	return s.storageCatalog().View(ctx)
}

func (s *AdminService) UpdateConfig(ctx context.Context, input AdminConfigUpdateInput) (AdminConfigView, error) {
	view, err := s.storageCatalog().ApplyLegacyPatch(ctx, legacyStorageConfigPatch{
		TargetStorageKey:  trimStringPointer(input.StorageKey),
		DefaultStorageKey: input.DefaultStorageKey,
		Update:            storageUpdateFromConfigPatch(input),
		HasPatch:          hasStorageConfigPatch(input),
	})
	if err != nil {
		return AdminConfigView{}, err
	}
	return view, nil
}

func (s *AdminService) CreateStorageConfig(ctx context.Context, input AdminStorageConfigCreateInput) (AdminConfigView, error) {
	view, err := s.storageCatalog().Create(ctx, input)
	if err != nil {
		return AdminConfigView{}, err
	}
	return view, nil
}

func (s *AdminService) UpdateStorageConfig(ctx context.Context, storageKey string, input AdminStorageConfigUpdateInput) (AdminConfigView, error) {
	view, err := s.storageCatalog().Patch(ctx, storageKey, input)
	if err != nil {
		return AdminConfigView{}, err
	}
	return view, nil
}

func (s *AdminService) DeleteStorageConfig(ctx context.Context, storageKey string) (AdminConfigView, error) {
	view, err := s.storageCatalog().Delete(ctx, storageKey)
	if err != nil {
		return AdminConfigView{}, err
	}
	return view, nil
}

func (s *AdminService) SetDefaultStorageConfig(ctx context.Context, storageKey string) (AdminConfigView, error) {
	view, err := s.storageCatalog().SetDefault(ctx, storageKey)
	if err != nil {
		return AdminConfigView{}, err
	}
	return view, nil
}

func (s *AdminService) StorageHealth(ctx context.Context) ([]model.StorageHealthCheck, error) {
	return NewStorageHealthService(s.repo, s.storage, nil).List(ctx)
}

func (s *AdminService) CheckStorageHealth(ctx context.Context, storageKey string) (model.StorageHealthCheck, error) {
	return NewStorageHealthService(s.repo, s.storage, nil).Check(ctx, storageKey)
}

func (s *AdminService) CheckAllStorageHealth(ctx context.Context) ([]model.StorageHealthCheck, error) {
	return NewStorageHealthService(s.repo, s.storage, nil).CheckAll(ctx)
}

func (s *AdminService) StorageHealthHistory(ctx context.Context, storageKey string, since time.Time) ([]model.StorageHealthCheck, error) {
	return NewStorageHealthService(s.repo, s.storage, nil).History(ctx, storageKey, since)
}

func (s *AdminService) GetSystemSettings(ctx context.Context) (AdminSystemSettingsView, error) {
	return s.loadSystemSettingsView(ctx)
}

func (s *AdminService) reloadStorageManager(ctx context.Context) error {
	return s.storageCatalog().reload(ctx)
}

func (s *AdminService) UpdateSystemSettings(ctx context.Context, input RuntimeSettingsUpdateInput) (AdminSystemSettingsView, error) {
	current := defaultRuntimeSettings()
	if s.settings != nil {
		current = s.settings.Current()
	}
	if input.CloudflareAPIToken == maskSecret(current.CloudflareAPIToken) {
		input.CloudflareAPIToken = current.CloudflareAPIToken
	}

	settings, err := ValidateRuntimeSettingsInput(input)
	if err != nil {
		return AdminSystemSettingsView{}, err
	}
	if err := s.repo.UpsertConfigValues(ctx, RuntimeSettingsToConfigValues(settings)); err != nil {
		return AdminSystemSettingsView{}, fmt.Errorf("%w: settings save failed", ErrDependencyUnavailable)
	}
	if s.settings != nil {
		s.settings.Reconfigure(settings)
	}
	return s.loadSystemSettingsView(ctx)
}

func (s *AdminService) loadSystemSettingsView(ctx context.Context) (AdminSystemSettingsView, error) {
	settings := defaultRuntimeSettings()
	if s.settings != nil {
		settings = s.settings.Current()
	}
	configs, err := s.repo.ListStorageConfigs(ctx)
	if err != nil {
		return AdminSystemSettingsView{}, fmt.Errorf("%w: config query failed", ErrDependencyUnavailable)
	}
	defaultKey := ""
	for _, cfg := range configs {
		if cfg.IsDefault {
			defaultKey = cfg.StorageKey
			break
		}
	}
	return AdminSystemSettingsView{
		Runtime: maskRuntimeSettings(settings),
		Readonly: AdminReadonlySettings{
			Environment: AdminEnvironmentStatus{
				HTTPAddr:                s.adminEnv.HTTPAddr,
				DatabasePath:            s.adminEnv.DatabasePath,
				RedisConfigured:         strings.TrimSpace(s.adminEnv.RedisURL) != "",
				PublicBaseURLSource:     s.publicBaseURLSource(),
				RuntimePublicBaseURLSet: settings.PublicBaseURL != "",
			},
			Security: AdminSecurityStatus{
				JWTSecret: SecretStatus{
					Configured: strings.TrimSpace(s.jwtSecret) != "",
				},
				AdminPassword: SecretStatus{
					Configured: s.isPasswordSet(ctx),
				},
				UIDEncryptionKey: SecretStatus{
					Configured: strings.TrimSpace(s.adminEnv.UIDEncryptionKey) != "",
				},
			},
			Storage: AdminStorageStatus{
				DefaultStorageKey:     defaultKey,
				StorageConfigCount:    len(configs),
				AllowStorageSelection: settings.AllowStorageSelect,
			},
			Service: AdminServiceStatus{
				Health:                    "ok",
				MaintenanceMode:           settings.MaintenanceMode,
				CloudflarePurgeConfigured: s.cloudflarePurgeConfigured(),
			},
		},
	}, nil
}

func maskRuntimeSettings(settings RuntimeSettings) RuntimeSettings {
	settings = cloneRuntimeSettings(settings)
	settings.CloudflareAPIToken = maskSecret(settings.CloudflareAPIToken)
	return settings
}

func (s *AdminService) publicBaseURLSource() string {
	if s.settings != nil {
		return s.settings.PublicBaseURLSource()
	}
	return "request_host"
}

func (s *AdminService) cloudflarePurgeConfigured() bool {
	return s.imageService != nil && s.imageService.CloudflarePurgeConfigured()
}

func buildStorageConfig(input AdminStorageConfigCreateInput) (config.RuntimeStorageConfig, error) {
	storageKey := strings.TrimSpace(input.StorageKey)
	name := strings.TrimSpace(input.Name)
	backend := config.NormalizeStorageBackend(input.Backend)
	if name == "" {
		return config.RuntimeStorageConfig{}, WithUserMessage(ErrInvalidInput, "storage instance name is required")
	}
	if backend == "" {
		return config.RuntimeStorageConfig{}, WithUserMessage(ErrInvalidInput, "storage backend is required")
	}
	if storageKey == "" {
		storageKey = newStorageKey(name, backend)
	}

	maxUploadSizeMB := input.MaxUploadSizeMB
	if maxUploadSizeMB <= 0 {
		maxUploadSizeMB = 20
	}
	allowedMIMETypes := strings.TrimSpace(input.AllowedMIMETypes)
	if allowedMIMETypes == "" {
		allowedMIMETypes = "image/avif,image/gif,image/jpeg,image/png,image/webp"
	}
	avifQuality := input.AvifQuality
	if avifQuality <= 0 {
		avifQuality = 60
	}
	avifSpeed := input.AvifSpeed
	if avifSpeed <= 0 {
		avifSpeed = 8
	}
	maxImagePixels := input.MaxImagePixels
	if maxImagePixels <= 0 {
		maxImagePixels = 40000000
	}
	avifMaxConcurrency := input.AVIFMaxConcurrency
	if avifMaxConcurrency <= 0 {
		avifMaxConcurrency = 2
	}
	avifTimeout := input.AVIFConversionTimeoutSeconds
	if avifTimeout <= 0 {
		avifTimeout = 30
	}

	return config.RuntimeStorageConfig{
		StorageKey:                  storageKey,
		Name:                        name,
		Backend:                     backend,
		LocalStoragePath:            input.LocalStoragePath,
		S3Endpoint:                  input.S3Endpoint,
		S3Region:                    input.S3Region,
		S3Bucket:                    input.S3Bucket,
		S3AccessKey:                 input.S3AccessKey,
		S3SecretKey:                 input.S3SecretKey,
		S3UseSSL:                    input.S3UseSSL,
		S3ForcePathStyle:            input.S3ForcePathStyle,
		WebDAVURL:                   input.WebDAVURL,
		WebDAVUser:                  input.WebDAVUser,
		WebDAVPass:                  input.WebDAVPass,
		MaxUploadSizeMB:             maxUploadSizeMB,
		AllowedMIMETypes:            allowedMIMETypes,
		AvifQuality:                 avifQuality,
		AvifSpeed:                   avifSpeed,
		MaxImagePixels:              maxImagePixels,
		AVIFMaxConcurrency:          avifMaxConcurrency,
		AVIFConversionTimeoutSeconds: avifTimeout,
	}, nil
}

func maskStorageConfig(cfg config.RuntimeStorageConfig) AdminStorageConfigView {
	return AdminStorageConfigView{
		StorageKey:                  cfg.StorageKey,
		Name:                        cfg.Name,
		IsDefault:                   cfg.IsDefault,
		StorageBackend:              cfg.Backend,
		LocalStoragePath:            cfg.LocalStoragePath,
		S3Endpoint:                  cfg.S3Endpoint,
		S3Region:                    cfg.S3Region,
		S3Bucket:                    cfg.S3Bucket,
		S3AccessKey:                 maskSecret(cfg.S3AccessKey),
		S3SecretKey:                 maskSecret(cfg.S3SecretKey),
		S3UseSSL:                    cfg.S3UseSSL,
		S3ForcePathStyle:            cfg.S3ForcePathStyle,
		WebDAVURL:                   cfg.WebDAVURL,
		WebDAVUser:                  cfg.WebDAVUser,
		WebDAVPass:                  maskSecret(cfg.WebDAVPass),
		MaxUploadSizeMB:             cfg.MaxUploadSizeMB,
		AllowedMIMETypes:            splitCSV(cfg.AllowedMIMETypes),
		AvifQuality:                 cfg.AvifQuality,
		AvifSpeed:                   cfg.AvifSpeed,
		MaxImagePixels:              cfg.MaxImagePixels,
		AVIFMaxConcurrency:          cfg.AVIFMaxConcurrency,
		AVIFConversionTimeoutSeconds: cfg.AVIFConversionTimeoutSeconds,
	}
}

func mergeStorageConfig(target *config.RuntimeStorageConfig, current config.RuntimeStorageConfig, update AdminStorageConfigUpdateInput) {
	if update.Name != nil {
		target.Name = strings.TrimSpace(*update.Name)
	}
	if update.Backend != nil {
		target.Backend = config.NormalizeStorageBackend(*update.Backend)
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
	if update.MaxUploadSizeMB != nil {
		target.MaxUploadSizeMB = *update.MaxUploadSizeMB
	}
	if update.AllowedMIMETypes != nil {
		target.AllowedMIMETypes = *update.AllowedMIMETypes
	}
	if update.AvifQuality != nil {
		target.AvifQuality = *update.AvifQuality
	}
	if update.AvifSpeed != nil {
		target.AvifSpeed = *update.AvifSpeed
	}
	if update.MaxImagePixels != nil {
		target.MaxImagePixels = *update.MaxImagePixels
	}
	if update.AVIFMaxConcurrency != nil {
		target.AVIFMaxConcurrency = *update.AVIFMaxConcurrency
	}
	if update.AVIFConversionTimeoutSeconds != nil {
		target.AVIFConversionTimeoutSeconds = *update.AVIFConversionTimeoutSeconds
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
	return config.NormalizeStorageBackend(current) != config.NormalizeStorageBackend(next)
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
		input.WebDAVPass != nil ||
		input.MaxUploadSizeMB != nil ||
		input.AllowedMIMETypes != nil ||
		input.AvifQuality != nil ||
		input.AvifSpeed != nil ||
		input.MaxImagePixels != nil ||
		input.AVIFMaxConcurrency != nil ||
		input.AVIFConversionTimeoutSeconds != nil
}

func storageUpdateFromConfigPatch(input AdminConfigUpdateInput) AdminStorageConfigUpdateInput {
	return AdminStorageConfigUpdateInput{
		Name:                        input.Name,
		Backend:                     input.Backend,
		LocalStoragePath:            input.LocalStoragePath,
		S3Endpoint:                  input.S3Endpoint,
		S3Region:                    input.S3Region,
		S3Bucket:                    input.S3Bucket,
		S3AccessKey:                 input.S3AccessKey,
		S3SecretKey:                 input.S3SecretKey,
		S3UseSSL:                    input.S3UseSSL,
		S3ForcePathStyle:            input.S3ForcePathStyle,
		WebDAVURL:                   input.WebDAVURL,
		WebDAVUser:                  input.WebDAVUser,
		WebDAVPass:                  input.WebDAVPass,
		MaxUploadSizeMB:             input.MaxUploadSizeMB,
		AllowedMIMETypes:            input.AllowedMIMETypes,
		AvifQuality:                 input.AvifQuality,
		AvifSpeed:                   input.AvifSpeed,
		MaxImagePixels:              input.MaxImagePixels,
		AVIFMaxConcurrency:          input.AVIFMaxConcurrency,
		AVIFConversionTimeoutSeconds: input.AVIFConversionTimeoutSeconds,
	}
}

func trimStringPointer(value *string) string {
	if value == nil {
		return ""
	}
	return strings.TrimSpace(*value)
}
