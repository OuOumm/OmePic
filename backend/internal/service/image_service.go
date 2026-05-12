package service

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"sync"
	"time"

	"omepic/backend/internal/cache"
	"omepic/backend/internal/config"
	"omepic/backend/internal/model"
	"omepic/backend/internal/repository"
	"omepic/backend/internal/storage"
)

const maxUploadSizeBytes = 20 * 1024 * 1024

type UploadInput struct {
	Token            string
	OriginalFilename string
	MIMEType         string
	IPAddress        string
	Bytes            []byte
	BaseURL          string
	StorageKey       string
}

type UploadOutput struct {
	UID            string    `json:"uid"`
	URL            string    `json:"url"`
	MDURL          string    `json:"md_url"`
	BBCode         string    `json:"bbcode"`
	Size           int64     `json:"size"`
	MIMEType       string    `json:"mime_type"`
	CreatedAt      time.Time `json:"created_at"`
	Duplicate      bool      `json:"duplicate"`
	StorageKey     string    `json:"storage_key"`
	StorageBackend string    `json:"storage_backend"`
}

type PublicStorageOption struct {
	StorageKey     string `json:"storage_key"`
	Name           string `json:"name"`
	StorageBackend string `json:"storage_backend"`
	IsDefault      bool   `json:"is_default"`
}

type ImageResolverOutput struct {
	Record   model.ImageRecord
	CacheHit bool
}

type UIDGenerator func() (string, error)
type UIDValidator func(string) error

type ImageService struct {
	repo         *repository.Repository
	cache        cache.ImageCache
	storage      *storage.Manager
	settings     *RuntimeSettingsManager
	logger       *slog.Logger
	generateUID  UIDGenerator
	validateUID  UIDValidator
	transformer  func([]byte) ([]byte, error)
	operationMux sync.Mutex
}

func NewImageService(
	repo *repository.Repository,
	imageCache cache.ImageCache,
	storageManager *storage.Manager,
	settingsManager *RuntimeSettingsManager,
	generateUID UIDGenerator,
	validateUID UIDValidator,
	logger *slog.Logger,
) *ImageService {
	if generateUID == nil {
		generateUID = func() (string, error) {
			return "", errors.New("uid generator is not configured")
		}
	}
	if validateUID == nil {
		validateUID = func(string) error {
			return errors.New("uid validator is not configured")
		}
	}

	return &ImageService{
		repo:        repo,
		cache:       imageCache,
		storage:     storageManager,
		settings:    settingsManager,
		logger:      logger,
		generateUID: generateUID,
		validateUID: validateUID,
		transformer: convertToAVIF,
	}
}

func (s *ImageService) SetUIDGenerator(fn UIDGenerator) {
	if fn != nil {
		s.generateUID = fn
	}
}

func (s *ImageService) SetUIDValidator(fn UIDValidator) {
	if fn != nil {
		s.validateUID = fn
	}
}

func (s *ImageService) Upload(ctx context.Context, input UploadInput) (UploadOutput, error) {
	if strings.TrimSpace(input.Token) == "" {
		return UploadOutput{}, ErrMissingToken
	}
	if err := s.ensureIPAllowed(ctx, input.IPAddress); err != nil {
		return UploadOutput{}, err
	}
	runtimeSettings := s.currentRuntimeSettings()
	if runtimeSettings.MaintenanceMode {
		return UploadOutput{}, fmt.Errorf("%w: %s", ErrInvalidInput, runtimeSettings.EffectiveMaintenanceMessage())
	}
	maxBytes := runtimeSettings.MaxUploadSizeBytes()
	if len(input.Bytes) == 0 || (maxBytes > 0 && int64(len(input.Bytes)) > maxBytes) {
		if maxBytes > 0 {
			return UploadOutput{}, fmt.Errorf("%w: file size must be between 1 byte and %d MB", ErrInvalidInput, runtimeSettings.MaxUploadSizeMB)
		}
		return UploadOutput{}, fmt.Errorf("%w: file size must be greater than 0 bytes", ErrInvalidInput)
	}

	if !runtimeSettingsAllowsMIME(runtimeSettings, input.MIMEType) {
		return UploadOutput{}, fmt.Errorf("%w: file MIME type is not allowed", ErrInvalidInput)
	}

	s.operationMux.Lock()
	defer s.operationMux.Unlock()

	resolved, err := s.resolveUploadStorage(input.StorageKey, runtimeSettings.AllowStorageSelect)
	if err != nil {
		return UploadOutput{}, err
	}

	md5Hash := md5Hex(input.Bytes)

	existing, err := s.findExistingByMD5(ctx, resolved.Config.StorageKey, md5Hash)
	if err != nil {
		return UploadOutput{}, err
	}

	now := time.Now().UTC()
	uid, err := s.generateUID()
	if err != nil {
		return UploadOutput{}, fmt.Errorf("%w: uid generation failed", ErrDependencyUnavailable)
	}
	var record model.ImageRecord

	if existing != nil {
		record = model.ImageRecord{
			UID:            uid,
			Token:          input.Token,
			StorageKey:     existing.StorageKey,
			StorageBackend: existing.StorageBackend,
			FilePath:       existing.FilePath,
			MIMEType:       existing.MIMEType,
			Size:           existing.Size,
			MD5Hash:        existing.MD5Hash,
			IPAddress:      input.IPAddress,
			CreatedAt:      now,
		}
		if err := s.repo.InsertImage(ctx, record); err != nil {
			return UploadOutput{}, mapRepoError(err)
		}
		if err := s.cache.SetImage(ctx, record); err != nil {
			return UploadOutput{}, fmt.Errorf("%w: redis uid write failed", ErrDependencyUnavailable)
		}
		return buildUploadOutput(record, input.BaseURL, input.OriginalFilename, true), nil
	}

	convertedBytes, err := s.transformer(input.Bytes)
	if err != nil {
		return UploadOutput{}, err
	}

	objectKey := storage.BuildObjectKey(uid, publicImageExtension)
	storedPath, err := resolved.Provider.Save(ctx, objectKey, convertedBytes, publicImageMIMEType)
	if err != nil {
		return UploadOutput{}, fmt.Errorf("%w: failed to persist file", ErrDependencyUnavailable)
	}

	record = model.ImageRecord{
		UID:            uid,
		Token:          input.Token,
		StorageKey:     resolved.Config.StorageKey,
		StorageBackend: resolved.Config.Backend,
		FilePath:       storedPath,
		MIMEType:       publicImageMIMEType,
		Size:           int64(len(convertedBytes)),
		MD5Hash:        md5Hash,
		IPAddress:      input.IPAddress,
		CreatedAt:      now,
	}

	if err := s.repo.InsertImage(ctx, record); err != nil {
		_ = resolved.Provider.Delete(ctx, storedPath)
		return UploadOutput{}, mapRepoError(err)
	}
	if err := s.cache.SetImage(ctx, record); err != nil {
		return UploadOutput{}, fmt.Errorf("%w: redis uid write failed", ErrDependencyUnavailable)
	}
	if err := s.cache.SetMD5IfAbsent(ctx, scopedMD5CacheKey(record.StorageKey, md5Hash), uid); err != nil {
		return UploadOutput{}, fmt.Errorf("%w: redis md5 write failed", ErrDependencyUnavailable)
	}
	return buildUploadOutput(record, input.BaseURL, input.OriginalFilename, false), nil
}

func (s *ImageService) Delete(ctx context.Context, uid string, token string, isAdmin bool, ipAddress string) error {
	s.operationMux.Lock()
	defer s.operationMux.Unlock()

	if !isAdmin {
		if err := s.ensureIPAllowed(ctx, ipAddress); err != nil {
			return err
		}
	}

	normalizedUID, err := s.normalizeDeleteUID(uid, isAdmin)
	if err != nil {
		return err
	}

	record, err := s.repo.FindByUID(ctx, normalizedUID)
	if err != nil {
		if repository.IsNotFound(err) {
			return ErrNotFound
		}
		return fmt.Errorf("%w: lookup failed", ErrDependencyUnavailable)
	}

	if !isAdmin {
		if strings.TrimSpace(token) == "" {
			return ErrMissingToken
		}
		if token != record.Token {
			return ErrForbidden
		}
	}

	if err := s.repo.DeleteByUID(ctx, normalizedUID); err != nil {
		if repository.IsNotFound(err) {
			return ErrNotFound
		}
		return fmt.Errorf("%w: delete record failed", ErrDependencyUnavailable)
	}

	if err := s.cache.DeleteImage(ctx, normalizedUID); err != nil {
		return fmt.Errorf("%w: redis uid delete failed", ErrDependencyUnavailable)
	}

	md5Count, err := s.repo.CountByMD5AndStorageKey(ctx, record.MD5Hash, record.StorageKey)
	if err != nil {
		return fmt.Errorf("%w: md5 count failed", ErrDependencyUnavailable)
	}
	if md5Count == 0 {
		if err := s.cache.DeleteMD5(ctx, scopedMD5CacheKey(record.StorageKey, record.MD5Hash)); err != nil {
			return fmt.Errorf("%w: redis md5 delete failed", ErrDependencyUnavailable)
		}
		return nil
	}

	if err := s.repairMD5Mapping(ctx, record.StorageKey, record.MD5Hash, record.UID); err != nil {
		return err
	}

	return nil
}

func (s *ImageService) Resolve(ctx context.Context, uid string) (ImageResolverOutput, error) {
	normalizedUID, err := s.normalizeServeUID(uid)
	if err != nil {
		return ImageResolverOutput{}, err
	}

	cached, err := s.cache.GetImage(ctx, normalizedUID)
	if err != nil {
		return ImageResolverOutput{}, fmt.Errorf("%w: redis uid lookup failed", ErrDependencyUnavailable)
	}
	if cached != nil {
		return ImageResolverOutput{
			Record: model.ImageRecord{
				UID:            cached.UID,
				Token:          cached.Token,
				StorageKey:     cached.StorageKey,
				StorageBackend: cached.StorageBackend,
				FilePath:       cached.FilePath,
				MIMEType:       cached.MIMEType,
				Size:           cached.Size,
				MD5Hash:        cached.MD5Hash,
				CreatedAt:      cached.CreatedAt,
			},
			CacheHit: true,
		}, nil
	}

	record, err := s.repo.FindByUID(ctx, normalizedUID)
	if err != nil {
		if repository.IsNotFound(err) {
			return ImageResolverOutput{}, ErrNotFound
		}
		return ImageResolverOutput{}, fmt.Errorf("%w: sqlite uid lookup failed", ErrDependencyUnavailable)
	}

	if err := s.cache.SetImage(ctx, *record); err != nil {
		return ImageResolverOutput{}, fmt.Errorf("%w: redis uid repopulate failed", ErrDependencyUnavailable)
	}
	if err := s.cache.SetMD5IfAbsent(ctx, scopedMD5CacheKey(record.StorageKey, record.MD5Hash), record.UID); err != nil {
		return ImageResolverOutput{}, fmt.Errorf("%w: redis md5 repopulate failed", ErrDependencyUnavailable)
	}

	return ImageResolverOutput{Record: *record, CacheHit: false}, nil
}

func (s *ImageService) Preheat(ctx context.Context) (int, error) {
	records, err := s.repo.ListAllImages(ctx)
	if err != nil {
		return 0, fmt.Errorf("%w: list images failed", ErrDependencyUnavailable)
	}

	count := 0
	seenMD5 := make(map[string]struct{}, len(records))
	mappings := make(map[string]string, len(records))
	for _, record := range records {
		seenKey := scopedMD5SeenKey(record.StorageKey, record.MD5Hash)
		if _, ok := seenMD5[seenKey]; !ok {
			seenMD5[seenKey] = struct{}{}
			mappings[scopedMD5CacheKey(record.StorageKey, record.MD5Hash)] = record.UID
		}
		count++
	}

	if err := s.cache.SetImages(ctx, records); err != nil {
		return 0, fmt.Errorf("%w: redis uid preheat failed", ErrDependencyUnavailable)
	}

	if err := s.cache.SetMD5Mappings(ctx, mappings); err != nil {
		return 0, fmt.Errorf("%w: redis md5 preheat failed", ErrDependencyUnavailable)
	}

	s.logger.Info("redis cache preheated", "records", count)
	return count, nil
}

func (s *ImageService) PublicRuntimeSettings(ctx context.Context) (PublicRuntimeSettingsView, error) {
	settings := s.currentRuntimeSettings()
	configs, err := s.repo.ListStorageConfigs(ctx)
	if err != nil {
		return PublicRuntimeSettingsView{}, fmt.Errorf("%w: config query failed", ErrDependencyUnavailable)
	}

	options := publicStorageOptionsFromConfigs(configs, settings.AllowStorageSelect)
	return PublicRuntimeSettingsView{
		Site: PublicSiteSettingsView{
			Name:    settings.SiteName,
			Tagline: settings.SiteTagline,
		},
		Access: PublicAccessSettingsView{
			PublicBaseURL: s.EffectivePublicBaseURL(""),
		},
		Upload: PublicUploadSettingsView{
			MaxUploadSizeMB:           settings.MaxUploadSizeMB,
			AllowedMIMETypes:          append([]string(nil), settings.AllowedMIMETypes...),
			EffectiveAllowedMIMETypes: settings.EffectiveAllowedMIMETypes(),
		},
		Features: PublicFeatureSettingsView{
			AllowStorageSelection: settings.AllowStorageSelect,
			MaintenanceMode:       settings.MaintenanceMode,
			MaintenanceMessage:    settings.EffectiveMaintenanceMessage(),
		},
		Storage: PublicStorageSettingsView{
			Options: options,
		},
	}, nil
}

func (s *ImageService) findExistingByMD5(ctx context.Context, storageKey string, md5Hash string) (*model.ImageRecord, error) {
	cacheKey := scopedMD5CacheKey(storageKey, md5Hash)
	cachedUID, err := s.cache.GetMD5(ctx, cacheKey)
	if err != nil {
		return nil, fmt.Errorf("%w: redis md5 lookup failed", ErrDependencyUnavailable)
	}
	if cachedUID != "" {
		record, err := s.repo.FindByUID(ctx, cachedUID)
		switch {
		case err == nil && record.StorageKey == storageKey:
			return record, nil
		case err == nil:
		case repository.IsNotFound(err):
		default:
			return nil, fmt.Errorf("%w: sqlite uid lookup failed", ErrDependencyUnavailable)
		}
	}

	record, err := s.repo.FindByMD5AndStorageKey(ctx, md5Hash, storageKey)
	if err != nil {
		if repository.IsNotFound(err) {
			if cachedUID != "" {
				if err := s.cache.DeleteMD5(ctx, cacheKey); err != nil {
					return nil, fmt.Errorf("%w: redis md5 stale delete failed", ErrDependencyUnavailable)
				}
			}
			return nil, nil
		}
		return nil, fmt.Errorf("%w: sqlite md5 lookup failed", ErrDependencyUnavailable)
	}
	if cachedUID != record.UID {
		if err := s.cache.SetMD5(ctx, cacheKey, record.UID); err != nil {
			return nil, fmt.Errorf("%w: redis md5 repair failed", ErrDependencyUnavailable)
		}
	}
	return record, nil
}

func (s *ImageService) repairMD5Mapping(ctx context.Context, storageKey string, md5Hash string, deletedUID string) error {
	cacheKey := scopedMD5CacheKey(storageKey, md5Hash)
	cachedUID, err := s.cache.GetMD5(ctx, cacheKey)
	if err != nil {
		return fmt.Errorf("%w: redis md5 lookup failed", ErrDependencyUnavailable)
	}
	if cachedUID != "" && cachedUID != deletedUID {
		record, err := s.repo.FindByUID(ctx, cachedUID)
		switch {
		case err == nil && record.StorageKey == storageKey:
			return nil
		case err == nil:
		case repository.IsNotFound(err):
		default:
			return fmt.Errorf("%w: sqlite uid lookup failed", ErrDependencyUnavailable)
		}
	}

	replacement, err := s.repo.FindByMD5AndStorageKey(ctx, md5Hash, storageKey)
	if err != nil {
		if repository.IsNotFound(err) {
			return nil
		}
		return fmt.Errorf("%w: sqlite md5 lookup failed", ErrDependencyUnavailable)
	}
	if err := s.cache.SetMD5(ctx, cacheKey, replacement.UID); err != nil {
		return fmt.Errorf("%w: redis md5 repair failed", ErrDependencyUnavailable)
	}
	return nil
}

func (s *ImageService) resolveUploadStorage(storageKey string, allowStorageSelection bool) (storage.ResolvedProvider, error) {
	key := strings.TrimSpace(storageKey)
	if key == "" || !allowStorageSelection {
		resolved, err := s.storage.Current()
		if err != nil {
			return storage.ResolvedProvider{}, fmt.Errorf("%w: active storage backend is invalid", ErrDependencyUnavailable)
		}
		return resolved, nil
	}

	resolved, err := s.storage.ForKey(key)
	if err != nil {
		if strings.Contains(err.Error(), "unknown storage key") || strings.Contains(err.Error(), "storage key is required") {
			return storage.ResolvedProvider{}, fmt.Errorf("%w: storage instance not found", ErrNotFound)
		}
		return storage.ResolvedProvider{}, fmt.Errorf("%w: selected storage backend is invalid", ErrDependencyUnavailable)
	}
	return resolved, nil
}

func buildUploadOutput(record model.ImageRecord, baseURL string, originalFilename string, duplicate bool) UploadOutput {
	url := strings.TrimRight(baseURL, "/") + "/i/" + record.UID + publicImageExtension
	altText := strings.TrimSpace(originalFilename)
	if altText == "" {
		altText = "image"
	}
	return UploadOutput{
		UID:            record.UID,
		URL:            url,
		MDURL:          fmt.Sprintf("![%s](%s)", altText, url),
		BBCode:         fmt.Sprintf("[img]%s[/img]", url),
		Size:           record.Size,
		MIMEType:       record.MIMEType,
		CreatedAt:      record.CreatedAt,
		Duplicate:      duplicate,
		StorageKey:     record.StorageKey,
		StorageBackend: record.StorageBackend,
	}
}

func mapRepoError(err error) error {
	if errors.Is(err, ErrConflict) {
		return ErrConflict
	}
	return fmt.Errorf("%w: sqlite write failed", ErrDependencyUnavailable)
}

func MaxUploadSizeBytes() int64 {
	return maxUploadSizeBytes
}

func (s *ImageService) MaxUploadSizeBytes() int64 {
	settings := s.currentRuntimeSettings()
	if value := settings.MaxUploadSizeBytes(); value > 0 {
		return value
	}
	return maxUploadSizeBytes
}

func (s *ImageService) EffectivePublicBaseURL(requestBase string) string {
	if s.settings == nil {
		return strings.TrimRight(requestBase, "/")
	}
	return s.settings.EffectivePublicBaseURL(requestBase)
}

func (s *ImageService) currentRuntimeSettings() RuntimeSettings {
	if s.settings == nil {
		return defaultRuntimeSettings()
	}
	return s.settings.Current()
}

func (s *ImageService) ensureIPAllowed(ctx context.Context, ipAddress string) error {
	trimmed := strings.TrimSpace(ipAddress)
	if trimmed == "" {
		return nil
	}
	_, err := s.repo.FindActiveIPBanByHash(ctx, ipHash(trimmed))
	if err == nil {
		return ErrIPBanned
	}
	if repository.IsNotFound(err) {
		return nil
	}
	return fmt.Errorf("%w: ip ban lookup failed", ErrDependencyUnavailable)
}

func runtimeSettingsAllowsMIME(settings RuntimeSettings, mimeType string) bool {
	candidate := strings.ToLower(strings.TrimSpace(strings.Split(mimeType, ";")[0]))
	if candidate == "" {
		return false
	}
	for _, allowed := range settings.EffectiveAllowedMIMETypes() {
		if candidate == allowed {
			return true
		}
	}
	return false
}

func publicStorageOptionsFromConfigs(configs []config.RuntimeStorageConfig, allowStorageSelection bool) []PublicStorageOption {
	options := make([]PublicStorageOption, 0, len(configs))
	for _, cfg := range configs {
		if !allowStorageSelection && !cfg.IsDefault {
			continue
		}
		options = append(options, PublicStorageOption{
			StorageKey:     cfg.StorageKey,
			Name:           cfg.Name,
			StorageBackend: cfg.Backend,
			IsDefault:      cfg.IsDefault,
		})
	}
	return options
}

func md5Hex(payload []byte) string {
	hash := md5.Sum(payload)
	return hex.EncodeToString(hash[:])
}

func scopedMD5CacheKey(storageKey string, md5Hash string) string {
	return strings.TrimSpace(storageKey) + ":" + md5Hash
}

func scopedMD5SeenKey(storageKey string, md5Hash string) string {
	return strings.TrimSpace(storageKey) + "\x00" + md5Hash
}

func (s *ImageService) normalizeDeleteUID(rawUID string, isAdmin bool) (string, error) {
	if isAdmin {
		return s.normalizeStoredUID(rawUID)
	}
	return s.normalizeServeUID(rawUID)
}

func (s *ImageService) normalizeServeUID(rawUID string) (string, error) {
	value := strings.TrimSpace(rawUID)
	if value == "" || !strings.HasSuffix(value, publicImageExtension) {
		return "", ErrNotFound
	}

	return s.normalizeStoredUID(value[:len(value)-len(publicImageExtension)])
}

func (s *ImageService) normalizeStoredUID(rawUID string) (string, error) {
	value := strings.TrimSpace(rawUID)
	if value == "" {
		return "", ErrNotFound
	}
	if err := s.validateUID(value); err != nil {
		return "", ErrNotFound
	}
	return value, nil
}
