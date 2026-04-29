package storage

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/studio-b12/gowebdav"

	"omepic/backend/internal/config"
)

type Provider interface {
	Name() string
	Save(ctx context.Context, objectKey string, data []byte, contentType string) (string, error)
	Open(ctx context.Context, objectKey string) (OpenResult, error)
	Delete(ctx context.Context, objectKey string) error
}

type OpenResult struct {
	Reader  io.ReadCloser
	Size    int64
	ModTime time.Time
}

type ResolvedProvider struct {
	Config   config.RuntimeStorageConfig
	Provider Provider
}

type Manager struct {
	mu         sync.RWMutex
	configs    map[string]config.RuntimeStorageConfig
	defaultKey string
	providers  map[string]Provider
}

func NewManager(settings []config.RuntimeStorageConfig) (*Manager, error) {
	manager := &Manager{
		configs:   make(map[string]config.RuntimeStorageConfig),
		providers: make(map[string]Provider),
	}
	if err := manager.Reconfigure(settings); err != nil {
		return nil, err
	}
	return manager, nil
}

func (m *Manager) Current() (ResolvedProvider, error) {
	m.mu.RLock()
	defaultKey := m.defaultKey
	m.mu.RUnlock()

	return m.ForKey(defaultKey)
}

func (m *Manager) CurrentKey() string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.defaultKey
}

func (m *Manager) CurrentBackend() string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if cfg, ok := m.configs[m.defaultKey]; ok {
		return cfg.Backend
	}
	return config.StorageBackendLocal
}

func (m *Manager) ForKey(storageKey string) (ResolvedProvider, error) {
	key := strings.TrimSpace(storageKey)
	if key == "" {
		return ResolvedProvider{}, errors.New("storage key is required")
	}

	m.mu.RLock()
	if provider, ok := m.providers[key]; ok {
		cfg := m.configs[key]
		m.mu.RUnlock()
		return ResolvedProvider{Config: cfg, Provider: provider}, nil
	}
	cfg, ok := m.configs[key]
	m.mu.RUnlock()
	if !ok {
		return ResolvedProvider{}, fmt.Errorf("unknown storage key: %s", key)
	}

	provider, err := buildProvider(cfg)
	if err != nil {
		return ResolvedProvider{}, err
	}

	m.mu.Lock()
	m.providers[key] = provider
	m.mu.Unlock()
	return ResolvedProvider{Config: cfg, Provider: provider}, nil
}

func (m *Manager) Reconfigure(settings []config.RuntimeStorageConfig) error {
	normalized, defaultKey, err := normalizeConfigs(settings)
	if err != nil {
		return err
	}
	for _, cfg := range normalized {
		if err := ValidateConfig(cfg); err != nil {
			return err
		}
	}

	configs := make(map[string]config.RuntimeStorageConfig, len(normalized))
	for _, cfg := range normalized {
		configs[cfg.StorageKey] = cfg
	}

	m.mu.Lock()
	defer m.mu.Unlock()
	m.configs = configs
	m.defaultKey = defaultKey
	m.providers = make(map[string]Provider)
	return nil
}

func ValidateConfig(settings config.RuntimeStorageConfig) error {
	_, err := buildProvider(normalizeConfig(settings))
	return err
}

func BuildObjectKey(uid string, extension string) string {
	now := time.Now().UTC()
	cleanExt := strings.TrimPrefix(strings.ToLower(extension), ".")
	if cleanExt == "" {
		cleanExt = "bin"
	}
	return path.Join(
		fmt.Sprintf("%04d", now.Year()),
		fmt.Sprintf("%02d", int(now.Month())),
		uid+"."+cleanExt,
	)
}

type localProvider struct {
	root string
}

func (p *localProvider) Name() string {
	return config.StorageBackendLocal
}

func (p *localProvider) Save(_ context.Context, objectKey string, data []byte, _ string) (string, error) {
	targetPath := filepath.Join(p.root, filepath.FromSlash(objectKey))
	if err := os.MkdirAll(filepath.Dir(targetPath), 0o755); err != nil {
		return "", err
	}
	if err := os.WriteFile(targetPath, data, 0o644); err != nil {
		return "", err
	}
	return objectKey, nil
}

func (p *localProvider) Open(_ context.Context, objectKey string) (OpenResult, error) {
	targetPath := filepath.Join(p.root, filepath.FromSlash(objectKey))
	file, err := os.Open(targetPath)
	if err != nil {
		return OpenResult{}, err
	}
	info, err := file.Stat()
	if err != nil {
		file.Close()
		return OpenResult{}, err
	}
	return OpenResult{
		Reader:  file,
		Size:    info.Size(),
		ModTime: info.ModTime(),
	}, nil
}

func (p *localProvider) Delete(_ context.Context, objectKey string) error {
	targetPath := filepath.Join(p.root, filepath.FromSlash(objectKey))
	if err := os.Remove(targetPath); err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}
	return nil
}

type s3Provider struct {
	client *minio.Client
	bucket string
}

func (p *s3Provider) Name() string {
	return config.StorageBackendS3
}

func (p *s3Provider) Save(ctx context.Context, objectKey string, data []byte, contentType string) (string, error) {
	reader := bytes.NewReader(data)
	_, err := p.client.PutObject(ctx, p.bucket, objectKey, reader, int64(len(data)), minio.PutObjectOptions{ContentType: contentType})
	if err != nil {
		return "", err
	}
	return objectKey, nil
}

func (p *s3Provider) Open(ctx context.Context, objectKey string) (OpenResult, error) {
	object, err := p.client.GetObject(ctx, p.bucket, objectKey, minio.GetObjectOptions{})
	if err != nil {
		return OpenResult{}, err
	}
	info, err := object.Stat()
	if err != nil {
		object.Close()
		return OpenResult{}, err
	}
	return OpenResult{
		Reader:  object,
		Size:    info.Size,
		ModTime: info.LastModified,
	}, nil
}

func (p *s3Provider) Delete(ctx context.Context, objectKey string) error {
	return p.client.RemoveObject(ctx, p.bucket, objectKey, minio.RemoveObjectOptions{})
}

type webdavProvider struct {
	client *gowebdav.Client
}

func (p *webdavProvider) Name() string {
	return config.StorageBackendWebDAV
}

func (p *webdavProvider) Save(_ context.Context, objectKey string, data []byte, _ string) (string, error) {
	if err := p.client.MkdirAll(path.Dir(objectKey), 0o755); err != nil {
		return "", err
	}
	if err := p.client.Write(objectKey, data, 0o644); err != nil {
		return "", err
	}
	return objectKey, nil
}

func (p *webdavProvider) Open(_ context.Context, objectKey string) (OpenResult, error) {
	reader, err := p.client.ReadStream(objectKey)
	if err != nil {
		return OpenResult{}, err
	}
	info, err := p.client.Stat(objectKey)
	if err != nil {
		reader.Close()
		return OpenResult{}, err
	}
	return OpenResult{
		Reader:  reader,
		Size:    info.Size(),
		ModTime: info.ModTime(),
	}, nil
}

func (p *webdavProvider) Delete(_ context.Context, objectKey string) error {
	if err := p.client.Remove(objectKey); err != nil && !strings.Contains(strings.ToLower(err.Error()), "404") {
		return err
	}
	return nil
}

func buildProvider(settings config.RuntimeStorageConfig) (Provider, error) {
	settings = normalizeConfig(settings)
	switch settings.Backend {
	case config.StorageBackendLocal:
		return &localProvider{root: settings.LocalStoragePath}, nil
	case config.StorageBackendS3:
		if settings.S3Endpoint == "" || settings.S3Bucket == "" || settings.S3AccessKey == "" || settings.S3SecretKey == "" {
			return nil, errors.New("s3 storage requires endpoint, bucket, access key, and secret key")
		}
		bucketLookup := minio.BucketLookupAuto
		if settings.S3ForcePathStyle {
			bucketLookup = minio.BucketLookupPath
		}
		client, err := minio.New(settings.S3Endpoint, &minio.Options{
			Creds:        credentials.NewStaticV4(settings.S3AccessKey, settings.S3SecretKey, ""),
			Secure:       settings.S3UseSSL,
			Region:       settings.S3Region,
			BucketLookup: bucketLookup,
		})
		if err != nil {
			return nil, err
		}
		return &s3Provider{client: client, bucket: settings.S3Bucket}, nil
	case config.StorageBackendWebDAV:
		if settings.WebDAVURL == "" {
			return nil, errors.New("webdav storage requires url")
		}
		if _, err := url.Parse(settings.WebDAVURL); err != nil {
			return nil, err
		}
		client := gowebdav.NewClient(settings.WebDAVURL, settings.WebDAVUser, settings.WebDAVPass)
		return &webdavProvider{client: client}, nil
	default:
		return nil, fmt.Errorf("unsupported storage backend: %s", settings.Backend)
	}
}

func normalizeConfigs(settings []config.RuntimeStorageConfig) ([]config.RuntimeStorageConfig, string, error) {
	if len(settings) == 0 {
		return nil, "", errors.New("at least one storage config is required")
	}

	normalized := make([]config.RuntimeStorageConfig, 0, len(settings))
	seen := make(map[string]struct{}, len(settings))
	defaultKey := ""
	for _, raw := range settings {
		cfg := normalizeConfig(raw)
		if cfg.StorageKey == "" {
			return nil, "", errors.New("storage key is required")
		}
		if _, ok := seen[cfg.StorageKey]; ok {
			return nil, "", fmt.Errorf("duplicate storage key: %s", cfg.StorageKey)
		}
		seen[cfg.StorageKey] = struct{}{}
		if cfg.IsDefault {
			if defaultKey != "" {
				return nil, "", errors.New("multiple default storage configs are not allowed")
			}
			defaultKey = cfg.StorageKey
		}
		normalized = append(normalized, cfg)
	}

	if defaultKey == "" {
		normalized[0].IsDefault = true
		defaultKey = normalized[0].StorageKey
	}

	return normalized, defaultKey, nil
}

func normalizeConfig(settings config.RuntimeStorageConfig) config.RuntimeStorageConfig {
	settings.StorageKey = strings.TrimSpace(settings.StorageKey)
	settings.Name = strings.TrimSpace(settings.Name)
	settings.Backend = strings.TrimSpace(strings.ToLower(settings.Backend))
	if settings.Backend == "" {
		settings.Backend = config.StorageBackendLocal
	}
	if settings.Name == "" {
		settings.Name = config.BootstrapStorageName(settings.Backend)
	}
	if settings.LocalStoragePath == "" {
		settings.LocalStoragePath = "data/images"
	}
	if settings.S3Region == "" {
		settings.S3Region = "auto"
	}
	return settings
}
