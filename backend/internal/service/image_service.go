package service

import (
	"bytes"
	"context"
	"crypto/md5"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"omepic/backend/internal/cache"
	"omepic/backend/internal/config"
	"omepic/backend/internal/model"
	"omepic/backend/internal/repository"
	"omepic/backend/internal/storage"
)

var filenameReplacer = strings.NewReplacer("\\", "\\\\", "\"", "\\\"", "\r", "", "\n", "")

type UploadInput struct {
	Token            string
	OriginalFilename string
	MIMEType         string
	IPAddress        string
	// Bytes is a compatibility/testing path. Production request flows should prefer Source + DeclaredSize
	// so service-layer upload preparation can spool large uploads to a temp file instead of keeping the
	// original image fully resident in memory.
	Bytes []byte
	// Source is the preferred production upload path. Service will read it once, compute original-byte MD5,
	// and spool to a temp file when needed so later dedup/convert steps can reopen the same original payload.
	Source       io.Reader
	DeclaredSize int64
	OriginalMD5  string
	BaseURL      string
	StorageKey   string
}

func NewUploadInputFromBytes(token string, originalFilename string, mimeType string, payload []byte, baseURL string) UploadInput {
	return UploadInput{
		Token:            token,
		OriginalFilename: originalFilename,
		MIMEType:         mimeType,
		Bytes:            payload,
		DeclaredSize:     int64(len(payload)),
		BaseURL:          baseURL,
	}
}

type UploadOutput struct {
	URL       string `json:"url"`
	Duplicate bool   `json:"duplicate"`
}

func (in UploadInput) payloadSizeHint() int64 {
	if len(in.Bytes) > 0 {
		return int64(len(in.Bytes))
	}
	if in.DeclaredSize > 0 {
		return in.DeclaredSize
	}
	return 0
}

type preparedUploadSource struct {
	bytes       []byte
	tempPath    string
	size        int64
	originalMD5 string
}

func (src preparedUploadSource) Open() (io.ReadCloser, error) {
	if len(src.bytes) > 0 {
		return io.NopCloser(bytes.NewReader(src.bytes)), nil
	}
	if strings.TrimSpace(src.tempPath) == "" {
		return nil, errors.New("upload source is empty")
	}
	return os.Open(src.tempPath)
}

func (src preparedUploadSource) Cleanup() {
	if strings.TrimSpace(src.tempPath) != "" {
		_ = os.Remove(src.tempPath)
	}
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

type ImageOpenOutput struct {
	Reader             io.ReadCloser
	Size               int64
	MIMEType           string
	ContentDisposition string
	Record             model.ImageRecord
	CacheHit           bool
}

type UIDGenerator func() (string, error)
type UIDValidator func(string) error

type uploadStorageResolver interface {
	Current() (storage.ResolvedProvider, error)
	ForKey(string) (storage.ResolvedProvider, error)
	Reconfigure([]config.RuntimeStorageConfig) error
}

type ImageService struct {
	repo        *repository.Repository
	imageCache  cache.ImageLookupCache
	preheat     cache.ImagePreheatCache
	md5Cache    cache.MD5MappingCache
	md5Preheat  cache.MD5MappingPreheatCache
	storage     uploadStorageResolver
	settings    *RuntimeSettingsManager
	logger      *slog.Logger
	generateUID UIDGenerator
	validateUID UIDValidator
	encoder     func(io.Reader, io.Writer, AVIFConversionSettings) error
	hashLocks   *keyedMutex
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
	return NewImageServiceWithCaches(repo, imageCache, imageCache, imageCache, imageCache, storageManager, settingsManager, generateUID, validateUID, logger)
}

func NewImageServiceWithCaches(
	repo *repository.Repository,
	imageCache cache.ImageLookupCache,
	preheatCache cache.ImagePreheatCache,
	md5Cache cache.MD5MappingCache,
	md5PreheatCache cache.MD5MappingPreheatCache,
	storageManager uploadStorageResolver,
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
		imageCache:  imageCache,
		preheat:     preheatCache,
		md5Cache:    md5Cache,
		md5Preheat:  md5PreheatCache,
		storage:     storageManager,
		settings:    settingsManager,
		logger:      logger,
		generateUID: generateUID,
		validateUID: validateUID,
		encoder:     encodeAVIFToWriter,
		hashLocks:   newKeyedMutex(),
	}
}

func (s *ImageService) Upload(ctx context.Context, input UploadInput) (UploadOutput, error) {
	return s.newUploadFlow(ctx, input).Run()
}

func (s *ImageService) Delete(ctx context.Context, uid string, token string, isAdmin bool, ipAddress string) error {
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

	if err := s.imageCache.DeleteImage(ctx, normalizedUID); err != nil {
		return fmt.Errorf("%w: redis uid delete failed", ErrDependencyUnavailable)
	}

	if err := s.md5Mappings().RepairAfterDelete(ctx, md5MappingKeyForRecord(*record), record.UID); err != nil {
		return err
	}

	return nil
}

func (s *ImageService) Open(ctx context.Context, uid string) (ImageOpenOutput, error) {
	result, err := s.Resolve(ctx, uid)
	if err != nil {
		return ImageOpenOutput{}, err
	}

	resolved, err := s.storage.ForKey(result.Record.StorageKey)
	if err != nil {
		return ImageOpenOutput{}, fmt.Errorf("%w: storage backend resolution failed", ErrDependencyUnavailable)
	}

	file, err := resolved.Provider.Open(ctx, result.Record.FilePath)
	if err != nil {
		return ImageOpenOutput{}, fmt.Errorf("%w: image open failed", ErrDependencyUnavailable)
	}

	return ImageOpenOutput{
		Reader:             file.Reader,
		Size:               file.Size,
		MIMEType:           result.Record.MIMEType,
		ContentDisposition: contentDispositionForPath(result.Record.FilePath),
		Record:             result.Record,
		CacheHit:           result.CacheHit,
	}, nil
}

func (s *ImageService) Resolve(ctx context.Context, uid string) (ImageResolverOutput, error) {
	normalizedUID, err := s.normalizeServeUID(uid)
	if err != nil {
		return ImageResolverOutput{}, err
	}

	cached, err := s.imageCache.GetImage(ctx, normalizedUID)
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

	if err := s.imageCache.SetImage(ctx, *record); err != nil {
		return ImageResolverOutput{}, fmt.Errorf("%w: redis uid repopulate failed", ErrDependencyUnavailable)
	}
	if err := s.md5Mappings().BackfillFromRecord(ctx, *record); err != nil {
		return ImageResolverOutput{}, err
	}

	return ImageResolverOutput{Record: *record, CacheHit: false}, nil
}

func (s *ImageService) Preheat(ctx context.Context) (int, error) {
	records, err := s.repo.ListAllImages(ctx)
	if err != nil {
		return 0, fmt.Errorf("%w: list images failed", ErrDependencyUnavailable)
	}

	if err := s.preheat.SetImages(ctx, records); err != nil {
		return 0, fmt.Errorf("%w: redis uid preheat failed", ErrDependencyUnavailable)
	}

	if err := s.md5Mappings().Preheat(ctx, records); err != nil {
		return 0, err
	}

	count := len(records)
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
			return storage.ResolvedProvider{}, WithUserMessage(ErrNotFound, "storage instance not found")
		}
		return storage.ResolvedProvider{}, fmt.Errorf("%w: selected storage backend is invalid", ErrDependencyUnavailable)
	}
	return resolved, nil
}

func contentDispositionForPath(filePath string) string {
	filename := filenameReplacer.Replace(filepath.Base(filePath))
	return "inline; filename=\"" + filename + "\""
}

type countingWriter struct {
	writer io.Writer
	size   int64
}

func (w *countingWriter) Write(p []byte) (int, error) {
	n, err := w.writer.Write(p)
	w.size += int64(n)
	return n, err
}

func (s *ImageService) prepareUploadSource(input UploadInput, maxBytes int64) (preparedUploadSource, error) {
	if len(input.Bytes) > 0 {
		size := int64(len(input.Bytes))
		if maxBytes > 0 && size > maxBytes {
			return preparedUploadSource{}, WithUserMessage(ErrInvalidInput, fmt.Sprintf("file size must be between 1 byte and %d MB", maxBytes/(1024*1024)))
		}
		md5Hash := strings.TrimSpace(strings.ToLower(input.OriginalMD5))
		if md5Hash == "" {
			md5Hash = md5Hex(input.Bytes)
		}
		return preparedUploadSource{bytes: input.Bytes, size: size, originalMD5: md5Hash}, nil
	}
	if input.Source == nil {
		return preparedUploadSource{}, WithUserMessage(ErrInvalidInput, "file size must be greater than 0 bytes")
	}

	readLimit := MaxUploadSizeBytes() + 1
	if maxBytes > 0 {
		readLimit = maxBytes + 1
	}
	tempFile, err := os.CreateTemp("", "omepic-upload-*.img")
	if err != nil {
		return preparedUploadSource{}, fmt.Errorf("%w: failed to create temporary upload file", ErrDependencyUnavailable)
	}
	defer func() {
		_ = tempFile.Close()
	}()

	hasher := md5.New()
	writer := io.MultiWriter(tempFile, hasher)
	size, err := io.Copy(writer, io.LimitReader(input.Source, readLimit))
	if err != nil {
		_ = os.Remove(tempFile.Name())
		return preparedUploadSource{}, fmt.Errorf("%w: failed to read upload source", ErrDependencyUnavailable)
	}
	if maxBytes > 0 && size > maxBytes {
		_ = os.Remove(tempFile.Name())
		return preparedUploadSource{}, WithUserMessage(ErrInvalidInput, fmt.Sprintf("file size must be between 1 byte and %d MB", maxBytes/(1024*1024)))
	}
	return preparedUploadSource{tempPath: tempFile.Name(), size: size, originalMD5: hex.EncodeToString(hasher.Sum(nil))}, nil
}

func (s *ImageService) saveConvertedAVIF(ctx context.Context, provider storage.Provider, objectKey string, source io.Reader, settings AVIFConversionSettings) (int64, string, error) {
	return avifStreamConversion{
		encoder:  s.encoder,
		provider: provider,
		settings: settings,
	}.save(ctx, objectKey, source)
}

func buildUploadOutput(record model.ImageRecord, baseURL string, _ string, duplicate bool) UploadOutput {
	url := strings.TrimRight(baseURL, "/") + "/i/" + record.UID + publicImageExtension
	return UploadOutput{
		URL:       url,
		Duplicate: duplicate,
	}
}

func mapRepoError(err error) error {
	if errors.Is(err, ErrConflict) {
		return ErrConflict
	}
	return fmt.Errorf("%w: sqlite write failed", ErrDependencyUnavailable)
}

func MaxUploadSizeBytes() int64 {
	return defaultRuntimeSettings().MaxUploadSizeBytes()
}

func (s *ImageService) MaxUploadSizeBytes() int64 {
	settings := s.currentRuntimeSettings()
	if value := settings.MaxUploadSizeBytes(); value > 0 {
		return value
	}
	return defaultRuntimeSettings().MaxUploadSizeBytes()
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

type keyedMutex struct {
	mu    sync.Mutex
	locks map[string]*refCountedMutex
}

type refCountedMutex struct {
	mu   sync.Mutex
	refs int
}

func newKeyedMutex() *keyedMutex {
	return &keyedMutex{locks: make(map[string]*refCountedMutex)}
}

func (m *keyedMutex) Lock(key string) func() {
	m.mu.Lock()
	lock := m.locks[key]
	if lock == nil {
		lock = &refCountedMutex{}
		m.locks[key] = lock
	}
	lock.refs++
	m.mu.Unlock()

	lock.mu.Lock()
	return func() {
		lock.mu.Unlock()
		m.mu.Lock()
		lock.refs--
		if lock.refs == 0 {
			delete(m.locks, key)
		}
		m.mu.Unlock()
	}
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
