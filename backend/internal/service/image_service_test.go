package service

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"image"
	"image/color"
	"image/png"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/gen2brain/avif"

	"omepic/backend/internal/config"
	"omepic/backend/internal/model"
	"omepic/backend/internal/repository"
	"omepic/backend/internal/storage"
)

type fakeImageCache struct {
	mu             sync.Mutex
	images         map[string]model.CachedImage
	imageSets      int
	imageBatchSets int
}

func newFakeImageCache() *fakeImageCache {
	return &fakeImageCache{images: make(map[string]model.CachedImage)}
}

func (c *fakeImageCache) GetImage(_ context.Context, uid string) (*model.CachedImage, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	value, ok := c.images[uid]
	if !ok {
		return nil, nil
	}
	copy := value
	return &copy, nil
}

func (c *fakeImageCache) SetImage(_ context.Context, record model.ImageRecord) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.images[record.UID] = model.CachedImageFromRecord(record)
	c.imageSets++
	return nil
}

func (c *fakeImageCache) SetImages(_ context.Context, records []model.ImageRecord) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	for _, record := range records {
		c.images[record.UID] = model.CachedImageFromRecord(record)
		c.imageSets++
	}
	c.imageBatchSets++
	return nil
}

func (c *fakeImageCache) DeleteImage(_ context.Context, uid string) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.images, uid)
	return nil
}

func (c *fakeImageCache) hasCachedImage(uid string) bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	_, ok := c.images[uid]
	return ok
}

func (c *fakeImageCache) stats() (imageSets int, imageBatchSets int) {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.imageSets, c.imageBatchSets
}

type fakeMD5MappingCache struct {
	mu                sync.Mutex
	md5ToUID          map[model.MD5MappingKey]string
	md5Sets           int
	md5BatchSets      int
	getMD5Err         error
	setMD5Err         error
	setMD5IfAbsentErr error
	setMD5MappingsErr error
	deleteMD5Err      error
}

func newFakeMD5MappingCache() *fakeMD5MappingCache {
	return &fakeMD5MappingCache{md5ToUID: make(map[model.MD5MappingKey]string)}
}

func (c *fakeMD5MappingCache) GetMD5(_ context.Context, key model.MD5MappingKey) (string, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.getMD5Err != nil {
		return "", c.getMD5Err
	}
	return c.md5ToUID[key], nil
}

func (c *fakeMD5MappingCache) SetMD5IfAbsent(_ context.Context, key model.MD5MappingKey, uid string) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.setMD5IfAbsentErr != nil {
		return c.setMD5IfAbsentErr
	}
	if _, ok := c.md5ToUID[key]; !ok {
		c.md5ToUID[key] = uid
		c.md5Sets++
	}
	return nil
}

func (c *fakeMD5MappingCache) SetMD5(_ context.Context, key model.MD5MappingKey, uid string) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.setMD5Err != nil {
		return c.setMD5Err
	}
	c.md5ToUID[key] = uid
	c.md5Sets++
	return nil
}

func (c *fakeMD5MappingCache) SetMD5Mappings(_ context.Context, mappings []model.MD5Mapping) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.setMD5MappingsErr != nil {
		return c.setMD5MappingsErr
	}
	for _, mapping := range mappings {
		c.md5ToUID[mapping.Key] = mapping.UID
		c.md5Sets++
	}
	c.md5BatchSets++
	return nil
}

func (c *fakeMD5MappingCache) DeleteMD5(_ context.Context, key model.MD5MappingKey) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.deleteMD5Err != nil {
		return c.deleteMD5Err
	}
	delete(c.md5ToUID, key)
	return nil
}

func (c *fakeMD5MappingCache) cachedMD5(key model.MD5MappingKey) (string, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	value, ok := c.md5ToUID[key]
	return value, ok
}

func (c *fakeMD5MappingCache) setCachedMD5(key model.MD5MappingKey, uid string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.md5ToUID[key] = uid
}

func (c *fakeMD5MappingCache) clearMD5Mappings() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.md5ToUID = make(map[model.MD5MappingKey]string)
}

func (c *fakeMD5MappingCache) stats() (md5Sets int, md5BatchSets int) {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.md5Sets, c.md5BatchSets
}

type fakeCache struct {
	images *fakeImageCache
	md5    *fakeMD5MappingCache
}

func newFakeCache() *fakeCache {
	return &fakeCache{images: newFakeImageCache(), md5: newFakeMD5MappingCache()}
}

func (c *fakeCache) cachedMD5(key model.MD5MappingKey) (string, bool) {
	return c.md5.cachedMD5(key)
}

func (c *fakeCache) setCachedMD5(key model.MD5MappingKey, uid string) {
	c.md5.setCachedMD5(key, uid)
}

func (c *fakeCache) hasCachedImage(uid string) bool {
	return c.images.hasCachedImage(uid)
}

func (c *fakeCache) stats() (imageSets int, md5Sets int, imageBatchSets int, md5BatchSets int) {
	imageSets, imageBatchSets = c.images.stats()
	md5Sets, md5BatchSets = c.md5.stats()
	return imageSets, md5Sets, imageBatchSets, md5BatchSets
}

func (c *fakeCache) GetImage(ctx context.Context, uid string) (*model.CachedImage, error) {
	return c.images.GetImage(ctx, uid)
}

func (c *fakeCache) SetImage(ctx context.Context, record model.ImageRecord) error {
	return c.images.SetImage(ctx, record)
}

func (c *fakeCache) SetImages(ctx context.Context, records []model.ImageRecord) error {
	return c.images.SetImages(ctx, records)
}

func (c *fakeCache) DeleteImage(ctx context.Context, uid string) error {
	return c.images.DeleteImage(ctx, uid)
}

func (c *fakeCache) GetMD5(ctx context.Context, key model.MD5MappingKey) (string, error) {
	return c.md5.GetMD5(ctx, key)
}

func (c *fakeCache) SetMD5(ctx context.Context, key model.MD5MappingKey, uid string) error {
	return c.md5.SetMD5(ctx, key, uid)
}

func (c *fakeCache) SetMD5IfAbsent(ctx context.Context, key model.MD5MappingKey, uid string) error {
	return c.md5.SetMD5IfAbsent(ctx, key, uid)
}

func (c *fakeCache) SetMD5Mappings(ctx context.Context, mappings []model.MD5Mapping) error {
	return c.md5.SetMD5Mappings(ctx, mappings)
}

func (c *fakeCache) DeleteMD5(ctx context.Context, key model.MD5MappingKey) error {
	return c.md5.DeleteMD5(ctx, key)
}

type fakeUIDCodec struct {
	mu     sync.Mutex
	queued []string
	valid  map[string]struct{}
}

type fakeFailingProvider struct {
	name       string
	readBytes  int
	saveErr    error
	readCalled bool
}

type fakeStreamStorageProvider struct {
	mu              sync.Mutex
	savedPaths      []string
	deletedPaths    []string
	saveStreamCalls int
}

type fakeUploadStorageResolver struct {
	resolved storage.ResolvedProvider
}

func newFakeUIDCodec() *fakeUIDCodec {
	return &fakeUIDCodec{
		valid: make(map[string]struct{}),
	}
}

func (c *fakeUIDCodec) Queue(ids ...string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.queued = append(c.queued, ids...)
	for _, id := range ids {
		c.valid[id] = struct{}{}
	}
}

func (c *fakeUIDCodec) Add(ids ...string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	for _, id := range ids {
		c.valid[id] = struct{}{}
	}
}

func (c *fakeUIDCodec) Generate() (string, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if len(c.queued) == 0 {
		return "", errors.New("no queued uid values")
	}
	value := c.queued[0]
	c.queued = c.queued[1:]
	c.valid[value] = struct{}{}
	return value, nil
}

func (c *fakeUIDCodec) Validate(uid string) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if _, ok := c.valid[uid]; !ok {
		return errors.New("invalid uid")
	}
	return nil
}

func (p *fakeFailingProvider) Name() string {
	if p.name != "" {
		return p.name
	}
	return config.StorageBackendLocal
}

func (p *fakeFailingProvider) Save(_ context.Context, _ string, _ []byte, _ string) (string, error) {
	return "", p.saveErr
}

func (p *fakeFailingProvider) SaveStream(_ context.Context, _ string, reader io.Reader, _ int64, _ string) (string, error) {
	if p.readBytes > 0 {
		buf := make([]byte, p.readBytes)
		_, _ = io.ReadFull(reader, buf)
		p.readCalled = true
	}
	return "", p.saveErr
}

func (p *fakeFailingProvider) Open(_ context.Context, _ string) (storage.OpenResult, error) {
	return storage.OpenResult{}, errors.New("not implemented")
}

func (p *fakeFailingProvider) Delete(_ context.Context, _ string) error {
	return nil
}

func (p *fakeStreamStorageProvider) Name() string {
	return config.StorageBackendLocal
}

func (p *fakeStreamStorageProvider) Save(ctx context.Context, objectKey string, data []byte, contentType string) (string, error) {
	return p.SaveStream(ctx, objectKey, bytes.NewReader(data), int64(len(data)), contentType)
}

func (p *fakeStreamStorageProvider) SaveStream(_ context.Context, objectKey string, reader io.Reader, _ int64, _ string) (string, error) {
	if _, err := io.Copy(io.Discard, reader); err != nil {
		return "", err
	}
	p.mu.Lock()
	defer p.mu.Unlock()
	p.saveStreamCalls++
	p.savedPaths = append(p.savedPaths, objectKey)
	return objectKey, nil
}

func (p *fakeStreamStorageProvider) Open(_ context.Context, _ string) (storage.OpenResult, error) {
	return storage.OpenResult{}, errors.New("not implemented")
}

func (p *fakeStreamStorageProvider) Delete(_ context.Context, objectKey string) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.deletedPaths = append(p.deletedPaths, objectKey)
	return nil
}

func (p *fakeStreamStorageProvider) stats() (saveStreamCalls int, savedPaths []string, deletedPaths []string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.saveStreamCalls, append([]string(nil), p.savedPaths...), append([]string(nil), p.deletedPaths...)
}

func (r fakeUploadStorageResolver) Current() (storage.ResolvedProvider, error) {
	return r.resolved, nil
}

func (r fakeUploadStorageResolver) ForKey(string) (storage.ResolvedProvider, error) {
	return r.resolved, nil
}

func (r fakeUploadStorageResolver) Reconfigure([]config.RuntimeStorageConfig) error {
	return nil
}

func TestUploadConvertsToAVIFAndBuildsAVIFURL(t *testing.T) {
	ctx := context.Background()
	service, repo, _, rootDir, uidCodec := newImageServiceTestHarness(t)
	uidCodec.Queue("uid-1")

	result, err := service.Upload(ctx, UploadInput{
		Token:            "token-a",
		OriginalFilename: "sample.png",
		MIMEType:         "image/png",
		Bytes:            mustPNGBytes(t, color.RGBA{R: 255, G: 80, B: 32, A: 255}),
		BaseURL:          "http://localhost:8080",
	})
	if err != nil {
		t.Fatalf("Upload returned error: %v", err)
	}

	if !strings.HasSuffix(result.URL, "/i/uid-1"+publicImageExtension) {
		t.Fatalf("expected avif url, got %q", result.URL)
	}

	record, err := repo.FindByUID(ctx, "uid-1")
	if err != nil {
		t.Fatalf("FindByUID returned error: %v", err)
	}
	if record.MIMEType != publicImageMIMEType {
		t.Fatalf("expected stored mime %q, got %q", publicImageMIMEType, record.MIMEType)
	}
	if record.StorageKey != "local-primary" {
		t.Fatalf("expected stored storage key local-primary, got %q", record.StorageKey)
	}
	if got := filepath.Base(record.FilePath); got != "uid-1"+publicImageExtension {
		t.Fatalf("expected stored file basename %q, got %q", "uid-1"+publicImageExtension, got)
	}
	if !strings.HasSuffix(record.FilePath, publicImageExtension) {
		t.Fatalf("expected stored file path to end with .avif, got %q", record.FilePath)
	}

	storedBytes, err := os.ReadFile(filepath.Join(rootDir, filepath.FromSlash(record.FilePath)))
	if err != nil {
		t.Fatalf("ReadFile returned error: %v", err)
	}
	if _, err := avif.Decode(bytes.NewReader(storedBytes)); err != nil {
		t.Fatalf("expected stored bytes to decode as avif: %v", err)
	}
}

func TestUploadAcceptsAVIFSourceAndStoresAVIFOutput(t *testing.T) {
	ctx := context.Background()
	service, repo, _, rootDir, uidCodec := newImageServiceTestHarness(t)
	uidCodec.Queue("uid-avif-source")

	sourceBytes := mustAVIFBytes(t, color.RGBA{R: 10, G: 120, B: 240, A: 255})
	result, err := service.Upload(ctx, UploadInput{
		Token:            "token-a",
		OriginalFilename: "remote-image.avif",
		MIMEType:         "image/avif",
		Bytes:            sourceBytes,
		BaseURL:          "http://localhost:8080",
	})
	if err != nil {
		t.Fatalf("Upload returned error: %v", err)
	}
	if !strings.HasSuffix(result.URL, "/i/uid-avif-source"+publicImageExtension) {
		t.Fatalf("expected avif url, got %q", result.URL)
	}

	record, err := repo.FindByUID(ctx, "uid-avif-source")
	if err != nil {
		t.Fatalf("FindByUID returned error: %v", err)
	}
	storedBytes, err := os.ReadFile(filepath.Join(rootDir, filepath.FromSlash(record.FilePath)))
	if err != nil {
		t.Fatalf("ReadFile returned error: %v", err)
	}
	if _, err := avif.Decode(bytes.NewReader(storedBytes)); err != nil {
		t.Fatalf("expected stored bytes to decode as avif: %v", err)
	}
}

func TestUploadDeduplicatesByOriginalBytesAndSharesStoredFile(t *testing.T) {
	ctx := context.Background()
	service, repo, _, rootDir, uidCodec := newImageServiceTestHarness(t)
	uidCodec.Queue("uid-1", "uid-2")

	sourceBytes := mustPNGBytes(t, color.RGBA{R: 40, G: 100, B: 220, A: 255})
	first, err := service.Upload(ctx, UploadInput{
		Token:            "token-a",
		OriginalFilename: "sample.png",
		MIMEType:         "image/png",
		Bytes:            sourceBytes,
		BaseURL:          "http://localhost:8080",
	})
	if err != nil {
		t.Fatalf("first upload returned error: %v", err)
	}
	if first.Duplicate {
		t.Fatalf("expected first upload to not be duplicate")
	}

	second, err := service.Upload(ctx, UploadInput{
		Token:            "token-b",
		OriginalFilename: "sample.png",
		MIMEType:         "image/png",
		Bytes:            sourceBytes,
		BaseURL:          "http://localhost:8080",
	})
	if err != nil {
		t.Fatalf("second upload returned error: %v", err)
	}
	if !second.Duplicate {
		t.Fatalf("expected second upload to be duplicate")
	}

	firstRecord, err := repo.FindByUID(ctx, "uid-1")
	if err != nil {
		t.Fatalf("FindByUID first failed: %v", err)
	}
	secondRecord, err := repo.FindByUID(ctx, "uid-2")
	if err != nil {
		t.Fatalf("FindByUID second failed: %v", err)
	}
	if firstRecord.FilePath != secondRecord.FilePath {
		t.Fatalf("expected duplicate uploads to share file path")
	}
	if firstRecord.StorageKey != secondRecord.StorageKey {
		t.Fatalf("expected duplicate uploads to share storage key")
	}
	if got := filepath.Base(firstRecord.FilePath); got != "uid-1"+publicImageExtension {
		t.Fatalf("expected first stored basename %q, got %q", "uid-1"+publicImageExtension, got)
	}
	if firstRecord.MD5Hash != secondRecord.MD5Hash {
		t.Fatalf("expected duplicate uploads to share original-bytes md5 hash")
	}
	if firstRecord.MD5Hash != md5Hex(sourceBytes) {
		t.Fatalf("expected stored md5 hash to match original upload bytes")
	}

	files, err := os.ReadDir(filepath.Join(rootDir, filepath.Dir(firstRecord.FilePath)))
	if err != nil {
		t.Fatalf("ReadDir returned error: %v", err)
	}
	if len(files) != 1 {
		t.Fatalf("expected exactly one stored file, got %d", len(files))
	}
}

func TestUploadUsesSelectedStorageKeyAndReturnsMetadata(t *testing.T) {
	ctx := context.Background()
	service, repo, _, primaryRoot, uidCodec := newImageServiceTestHarness(t)
	uidCodec.Queue("uid-selected")

	_, err := service.Upload(ctx, UploadInput{
		Token:            "token-selected",
		OriginalFilename: "sample.png",
		MIMEType:         "image/png",
		Bytes:            mustPNGBytes(t, color.RGBA{R: 60, G: 10, B: 220, A: 255}),
		BaseURL:          "http://localhost:8080",
		StorageKey:       "local-secondary",
	})
	if err != nil {
		t.Fatalf("Upload returned error: %v", err)
	}
	record, err := repo.FindByUID(ctx, "uid-selected")
	if err != nil {
		t.Fatalf("FindByUID returned error: %v", err)
	}
	if record.StorageKey != "local-secondary" {
		t.Fatalf("expected selected storage key in record, got %q", record.StorageKey)
	}

	secondaryRoot := filepath.Join(filepath.Dir(primaryRoot), "images-secondary")
	if _, err := os.Stat(filepath.Join(secondaryRoot, filepath.FromSlash(record.FilePath))); err != nil {
		t.Fatalf("expected selected storage file to exist: %v", err)
	}
	if _, err := os.Stat(filepath.Join(primaryRoot, filepath.FromSlash(record.FilePath))); err == nil {
		t.Fatalf("expected selected storage write to avoid the default storage root")
	}
}

func TestUploadDefaultFallbackUsesHotReloadedDefaultStorage(t *testing.T) {
	ctx := context.Background()
	service, repo, _, primaryRoot, uidCodec := newImageServiceTestHarness(t)
	uidCodec.Queue("uid-default-switch")

	if err := repo.SetDefaultStorageConfig(ctx, "local-secondary"); err != nil {
		t.Fatalf("SetDefaultStorageConfig returned error: %v", err)
	}
	configs, err := repo.ListStorageConfigs(ctx)
	if err != nil {
		t.Fatalf("ListStorageConfigs returned error: %v", err)
	}
	if err := service.storage.Reconfigure(configs); err != nil {
		t.Fatalf("Reconfigure returned error: %v", err)
	}

	_, err = service.Upload(ctx, UploadInput{
		Token:            "token-default-switch",
		OriginalFilename: "sample.png",
		MIMEType:         "image/png",
		Bytes:            mustPNGBytes(t, color.RGBA{R: 20, G: 210, B: 80, A: 255}),
		BaseURL:          "http://localhost:8080",
	})
	if err != nil {
		t.Fatalf("Upload returned error: %v", err)
	}
	record, err := repo.FindByUID(ctx, "uid-default-switch")
	if err != nil {
		t.Fatalf("FindByUID returned error: %v", err)
	}
	secondaryRoot := filepath.Join(filepath.Dir(primaryRoot), "images-secondary")
	if _, err := os.Stat(filepath.Join(secondaryRoot, filepath.FromSlash(record.FilePath))); err != nil {
		t.Fatalf("expected fallback upload file to exist in the hot-reloaded default storage: %v", err)
	}
	if _, err := os.Stat(filepath.Join(primaryRoot, filepath.FromSlash(record.FilePath))); err == nil {
		t.Fatalf("expected fallback upload to avoid the previous default storage root")
	}
}

func TestUploadDeduplicatesWithinSelectedStorageOnly(t *testing.T) {
	ctx := context.Background()
	service, repo, cacheStore, primaryRoot, uidCodec := newImageServiceTestHarness(t)
	uidCodec.Queue("uid-primary", "uid-secondary", "uid-secondary-dup")

	sourceBytes := mustPNGBytes(t, color.RGBA{R: 10, G: 150, B: 220, A: 255})
	_, err := service.Upload(ctx, UploadInput{
		Token:            "token-primary",
		OriginalFilename: "sample.png",
		MIMEType:         "image/png",
		Bytes:            sourceBytes,
		BaseURL:          "http://localhost:8080",
	})
	if err != nil {
		t.Fatalf("first upload returned error: %v", err)
	}

	second, err := service.Upload(ctx, UploadInput{
		Token:            "token-secondary",
		OriginalFilename: "sample.png",
		MIMEType:         "image/png",
		Bytes:            sourceBytes,
		BaseURL:          "http://localhost:8080",
		StorageKey:       "local-secondary",
	})
	if err != nil {
		t.Fatalf("second upload returned error: %v", err)
	}
	if second.Duplicate {
		t.Fatalf("expected same bytes on a different selected storage to create a new physical object")
	}

	third, err := service.Upload(ctx, UploadInput{
		Token:            "token-secondary-dup",
		OriginalFilename: "sample.png",
		MIMEType:         "image/png",
		Bytes:            sourceBytes,
		BaseURL:          "http://localhost:8080",
		StorageKey:       "local-secondary",
	})
	if err != nil {
		t.Fatalf("third upload returned error: %v", err)
	}
	if !third.Duplicate {
		t.Fatalf("expected same bytes on the same selected storage to deduplicate")
	}

	firstRecord, err := repo.FindByUID(ctx, "uid-primary")
	if err != nil {
		t.Fatalf("FindByUID first returned error: %v", err)
	}
	secondRecord, err := repo.FindByUID(ctx, "uid-secondary")
	if err != nil {
		t.Fatalf("FindByUID second returned error: %v", err)
	}
	thirdRecord, err := repo.FindByUID(ctx, "uid-secondary-dup")
	if err != nil {
		t.Fatalf("FindByUID third returned error: %v", err)
	}
	if firstRecord.FilePath == secondRecord.FilePath {
		t.Fatalf("expected different selected storage uploads to use different physical object keys")
	}
	if secondRecord.FilePath != thirdRecord.FilePath {
		t.Fatalf("expected same selected storage duplicate to reuse physical object")
	}
	if got, _ := cacheStore.cachedMD5(model.NewMD5MappingKey(firstRecord.StorageKey, firstRecord.MD5Hash)); got != "uid-primary" {
		t.Fatalf("expected primary storage md5 mapping to point at primary uid, got %q", got)
	}
	if got, _ := cacheStore.cachedMD5(model.NewMD5MappingKey(secondRecord.StorageKey, secondRecord.MD5Hash)); got != "uid-secondary" {
		t.Fatalf("expected secondary storage md5 mapping to point at secondary uid, got %q", got)
	}
	secondaryRoot := filepath.Join(filepath.Dir(primaryRoot), "images-secondary")
	if _, err := os.Stat(filepath.Join(primaryRoot, filepath.FromSlash(firstRecord.FilePath))); err != nil {
		t.Fatalf("expected default storage file to exist: %v", err)
	}
	if _, err := os.Stat(filepath.Join(secondaryRoot, filepath.FromSlash(secondRecord.FilePath))); err != nil {
		t.Fatalf("expected selected storage file to exist: %v", err)
	}
}

func TestUploadRejectsUnknownStorageKeyBeforeWrite(t *testing.T) {
	ctx := context.Background()
	service, repo, _, rootDir, _ := newImageServiceTestHarness(t)

	_, err := service.Upload(ctx, UploadInput{
		Token:            "token-a",
		OriginalFilename: "sample.png",
		MIMEType:         "image/png",
		Bytes:            mustPNGBytes(t, color.RGBA{R: 10, G: 30, B: 90, A: 255}),
		BaseURL:          "http://localhost:8080",
		StorageKey:       "missing-storage",
	})
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected ErrNotFound for unknown storage key, got %v", err)
	}

	records, err := repo.ListAllImages(ctx)
	if err != nil {
		t.Fatalf("ListAllImages returned error: %v", err)
	}
	if len(records) != 0 {
		t.Fatalf("expected unknown storage upload to avoid DB writes, got %d records", len(records))
	}
	if _, err := os.Stat(rootDir); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("expected unknown storage upload to avoid physical writes, stat err=%v", err)
	}
}

func TestUploadDuplicateReusesStoredObjectWithoutStreamWrite(t *testing.T) {
	ctx := context.Background()
	service, _, _, _, uidCodec := newImageServiceTestHarness(t)
	uidCodec.Queue("uid-stream-first", "uid-stream-duplicate")
	provider := &fakeStreamStorageProvider{}
	service.storage = fakeUploadStorageResolver{resolved: storage.ResolvedProvider{
		Config:   config.RuntimeStorageConfig{StorageKey: "local-primary", Backend: config.StorageBackendLocal, IsDefault: true},
		Provider: provider,
	}}

	sourceBytes := mustPNGBytes(t, color.RGBA{R: 210, G: 20, B: 90, A: 255})
	first, err := service.Upload(ctx, UploadInput{
		Token:            "token-stream-first",
		OriginalFilename: "sample.png",
		MIMEType:         "image/png",
		Bytes:            sourceBytes,
		BaseURL:          "http://localhost:8080",
	})
	if err != nil {
		t.Fatalf("first upload returned error: %v", err)
	}
	if first.Duplicate {
		t.Fatalf("expected first upload to create a physical object")
	}

	second, err := service.Upload(ctx, UploadInput{
		Token:            "token-stream-duplicate",
		OriginalFilename: "sample.png",
		MIMEType:         "image/png",
		Bytes:            sourceBytes,
		BaseURL:          "http://localhost:8080",
	})
	if err != nil {
		t.Fatalf("second upload returned error: %v", err)
	}
	if !second.Duplicate {
		t.Fatalf("expected second upload to reuse the existing physical object")
	}

	saveCalls, savedPaths, deletedPaths := provider.stats()
	if saveCalls != 1 {
		t.Fatalf("expected duplicate upload not to call SaveStream again, got %d calls", saveCalls)
	}
	if len(savedPaths) != 1 || !strings.HasSuffix(savedPaths[0], "uid-stream-first"+publicImageExtension) {
		t.Fatalf("expected first uid object to be the only saved path, got %+v", savedPaths)
	}
	if len(deletedPaths) != 0 {
		t.Fatalf("expected duplicate reuse not to delete physical objects, got %+v", deletedPaths)
	}
}

func TestUploadCleansNewPhysicalObjectWhenRecordCommitFails(t *testing.T) {
	ctx := context.Background()
	service, repo, cacheStore, _, uidCodec := newImageServiceTestHarness(t)
	uidCodec.Queue("uid-conflict")
	if err := repo.InsertImage(ctx, model.ImageRecord{
		UID:            "uid-conflict",
		Token:          "token-existing",
		StorageKey:     "local-primary",
		StorageBackend: config.StorageBackendLocal,
		FilePath:       "2026/05/existing.avif",
		MIMEType:       publicImageMIMEType,
		Size:           1,
		MD5Hash:        "other-md5",
		IPAddress:      "127.0.0.1",
		CreatedAt:      time.Now().UTC(),
	}); err != nil {
		t.Fatalf("InsertImage existing returned error: %v", err)
	}

	provider := &fakeStreamStorageProvider{}
	service.storage = fakeUploadStorageResolver{resolved: storage.ResolvedProvider{
		Config:   config.RuntimeStorageConfig{StorageKey: "local-primary", Backend: config.StorageBackendLocal, IsDefault: true},
		Provider: provider,
	}}
	sourceBytes := mustPNGBytes(t, color.RGBA{R: 20, G: 210, B: 120, A: 255})
	_, err := service.Upload(ctx, UploadInput{
		Token:            "token-conflict",
		OriginalFilename: "sample.png",
		MIMEType:         "image/png",
		Bytes:            sourceBytes,
		BaseURL:          "http://localhost:8080",
	})
	if err == nil {
		t.Fatalf("expected upload error when SQLite record commit fails")
	}

	saveCalls, savedPaths, deletedPaths := provider.stats()
	if saveCalls != 1 || len(savedPaths) != 1 {
		t.Fatalf("expected one new physical SaveStream before commit failure, calls=%d paths=%+v", saveCalls, savedPaths)
	}
	if len(deletedPaths) != 1 || deletedPaths[0] != savedPaths[0] {
		t.Fatalf("expected failed commit to delete the newly saved object %q, got %+v", savedPaths[0], deletedPaths)
	}
	if got, ok := cacheStore.cachedMD5(model.NewMD5MappingKey("local-primary", md5Hex(sourceBytes))); ok {
		t.Fatalf("expected failed new physical commit not to publish md5 mapping, got %q", got)
	}
}

func TestUploadSkipsAVIFConversionForDuplicateUploads(t *testing.T) {
	ctx := context.Background()
	service, repo, _, _, uidCodec := newImageServiceTestHarness(t)
	uidCodec.Queue("uid-1", "uid-2")
	service.settings.Reconfigure(RuntimeSettings{
		SiteName:                     DefaultSiteName,
		SiteTagline:                  DefaultSiteTagline,
		MaxUploadSizeMB:              20,
		AllowedMIMETypes:             DefaultAllowedMIMETypes(),
		AvifQuality:                  35,
		AvifSpeed:                    2,
		AllowStorageSelect:           true,
		RateLimitWindowMinutes:       DefaultRateLimitWindowMinutes,
		RateLimitMaxRequests:         DefaultRateLimitMaxRequests,
		UploadRateLimitWindowMinutes: DefaultUploadRateLimitWindowMinutes,
		UploadRateLimitMaxRequests:   DefaultUploadRateLimitMaxRequests,
	})

	sourceBytes := mustPNGBytes(t, color.RGBA{R: 80, G: 160, B: 40, A: 255})
	conversionCalls := 0
	var observed []AVIFConversionSettings
	service.encoder = func(source io.Reader, target io.Writer, settings AVIFConversionSettings) error {
		conversionCalls++
		observed = append(observed, settings)
		return encodeAVIFToWriter(source, target, settings)
	}

	if _, err := service.Upload(ctx, UploadInput{
		Token:            "token-a",
		OriginalFilename: "sample.png",
		MIMEType:         "image/png",
		Bytes:            sourceBytes,
		BaseURL:          "http://localhost:8080",
	}); err != nil {
		t.Fatalf("first upload returned error: %v", err)
	}

	service.settings.Reconfigure(RuntimeSettings{
		SiteName:                     DefaultSiteName,
		SiteTagline:                  DefaultSiteTagline,
		MaxUploadSizeMB:              20,
		AllowedMIMETypes:             DefaultAllowedMIMETypes(),
		AvifQuality:                  90,
		AvifSpeed:                    10,
		AllowStorageSelect:           true,
		RateLimitWindowMinutes:       DefaultRateLimitWindowMinutes,
		RateLimitMaxRequests:         DefaultRateLimitMaxRequests,
		UploadRateLimitWindowMinutes: DefaultUploadRateLimitWindowMinutes,
		UploadRateLimitMaxRequests:   DefaultUploadRateLimitMaxRequests,
	})
	result, err := service.Upload(ctx, UploadInput{
		Token:            "token-b",
		OriginalFilename: "sample.png",
		MIMEType:         "image/png",
		Bytes:            sourceBytes,
		BaseURL:          "http://localhost:8080",
	})
	if err != nil {
		t.Fatalf("second upload returned error: %v", err)
	}
	if !result.Duplicate {
		t.Fatalf("expected second upload to be duplicate")
	}
	if conversionCalls != 1 {
		t.Fatalf("expected avif conversion to run once, got %d", conversionCalls)
	}
	if len(observed) != 1 || observed[0].Quality != 35 || observed[0].Speed != 2 {
		t.Fatalf("expected only first upload to convert with initial avif settings, got %+v", observed)
	}

	secondRecord, err := repo.FindByUID(ctx, "uid-2")
	if err != nil {
		t.Fatalf("FindByUID second failed: %v", err)
	}
	if secondRecord.MD5Hash != md5Hex(sourceBytes) {
		t.Fatalf("expected duplicate record md5 hash to match original upload bytes")
	}
}

func TestUploadSerializesSameStorageMD5WithoutGlobalUploadLock(t *testing.T) {
	ctx := context.Background()
	service, _, _, _, uidCodec := newImageServiceTestHarness(t)
	uidCodec.Queue("uid-a", "uid-b", "uid-c", "uid-d")
	firstBytes := mustPNGBytes(t, color.RGBA{R: 100, G: 20, B: 40, A: 255})
	secondBytes := mustPNGBytes(t, color.RGBA{R: 20, G: 100, B: 40, A: 255})

	started := make(chan string, 3)
	release := make(chan struct{})
	service.encoder = func(source io.Reader, target io.Writer, settings AVIFConversionSettings) error {
		payload, err := io.ReadAll(source)
		if err != nil {
			return err
		}
		started <- md5Hex(payload)
		<-release
		return encodeAVIFToWriter(bytes.NewReader(payload), target, settings)
	}

	runUpload := func(token string, payload []byte) <-chan error {
		done := make(chan error, 1)
		go func() {
			_, err := service.Upload(ctx, UploadInput{
				Token:            token,
				OriginalFilename: "sample.png",
				MIMEType:         "image/png",
				Bytes:            payload,
				BaseURL:          "http://localhost:8080",
			})
			done <- err
		}()
		return done
	}

	firstDone := runUpload("token-a", firstBytes)
	if got := <-started; got != md5Hex(firstBytes) {
		t.Fatalf("expected first upload to start first transform, got %q", got)
	}

	secondDone := runUpload("token-b", secondBytes)
	select {
	case got := <-started:
		if got != md5Hex(secondBytes) {
			t.Fatalf("expected different md5 upload to start while first is blocked, got %q", got)
		}
	case <-time.After(time.Second):
		t.Fatalf("expected different md5 upload to proceed without global upload lock")
	}

	duplicateDone := runUpload("token-c", firstBytes)
	select {
	case got := <-started:
		t.Fatalf("same storage/md5 upload should not start another transform while first is blocked, got %q", got)
	case <-time.After(100 * time.Millisecond):
	}

	close(release)
	for _, done := range []<-chan error{firstDone, secondDone, duplicateDone} {
		if err := <-done; err != nil {
			t.Fatalf("Upload returned error: %v", err)
		}
	}
}

func TestSaveConvertedAVIFPrefersEncoderInvalidInputOverSaveFailure(t *testing.T) {
	ctx := context.Background()
	service, _, _, _, _ := newImageServiceTestHarness(t)
	provider := &fakeFailingProvider{readBytes: 1, saveErr: errors.New("save failed after encoder failure")}
	service.encoder = func(io.Reader, io.Writer, AVIFConversionSettings) error {
		return WithUserMessage(ErrInvalidInput, "bad source image")
	}

	_, _, err := service.saveConvertedAVIF(ctx, provider, "bad.avif", strings.NewReader("not an image"), AVIFConversionSettings{Quality: 60, Speed: 8})
	if err == nil || !containsError(err, ErrInvalidInput) {
		t.Fatalf("expected encoder invalid input to take priority over save failure, got %v", err)
	}
	if !provider.readCalled {
		t.Fatalf("expected provider to observe stream closure after encoder failure")
	}
}

func TestAVIFStreamErrorTreatsSaveFailureClosedPipeAsDependencyUnavailable(t *testing.T) {
	saveErr := errors.New("save failed")
	err := avifStreamError(io.ErrClosedPipe, saveErr)
	if err == nil || !containsError(err, ErrDependencyUnavailable) {
		t.Fatalf("expected save failure to map to dependency unavailable when encoder only saw closed pipe, got %v", err)
	}
}

func TestUploadReturnsQuicklyWhenSaveStreamFailsImmediately(t *testing.T) {
	ctx := context.Background()
	service, repo, _, _, uidCodec := newImageServiceTestHarness(t)
	uidCodec.Queue("uid-save-fail")

	provider := &fakeFailingProvider{saveErr: errors.New("save failed immediately")}
	service.storage = fakeUploadStorageResolver{resolved: storage.ResolvedProvider{
		Config:   config.RuntimeStorageConfig{StorageKey: "local-primary", Backend: config.StorageBackendLocal, IsDefault: true},
		Provider: provider,
	}}

	done := make(chan error, 1)
	go func() {
		_, err := service.Upload(ctx, UploadInput{
			Token:            "token-a",
			OriginalFilename: "sample.png",
			MIMEType:         "image/png",
			Bytes:            mustPNGBytes(t, color.RGBA{R: 200, G: 100, B: 20, A: 255}),
			BaseURL:          "http://localhost:8080",
		})
		done <- err
	}()

	select {
	case err := <-done:
		if err == nil {
			t.Fatalf("expected upload error when SaveStream fails immediately")
		}
		if records, listErr := repo.ListAllImages(ctx); listErr != nil {
			t.Fatalf("ListAllImages returned error: %v", listErr)
		} else if len(records) != 0 {
			t.Fatalf("expected no image rows after failed save, got %d", len(records))
		}
	case <-time.After(2 * time.Second):
		t.Fatal("Upload hung when SaveStream failed immediately")
	}
}

func TestUploadReturnsQuicklyWhenSaveStreamFailsAfterPartialRead(t *testing.T) {
	ctx := context.Background()
	service, repo, _, _, uidCodec := newImageServiceTestHarness(t)
	uidCodec.Queue("uid-save-partial-fail")

	provider := &fakeFailingProvider{readBytes: 1, saveErr: errors.New("save failed after partial read")}
	service.storage = fakeUploadStorageResolver{resolved: storage.ResolvedProvider{
		Config:   config.RuntimeStorageConfig{StorageKey: "local-primary", Backend: config.StorageBackendLocal, IsDefault: true},
		Provider: provider,
	}}

	done := make(chan error, 1)
	go func() {
		_, err := service.Upload(ctx, UploadInput{
			Token:            "token-a",
			OriginalFilename: "sample.png",
			MIMEType:         "image/png",
			Bytes:            mustPNGBytes(t, color.RGBA{R: 20, G: 120, B: 220, A: 255}),
			BaseURL:          "http://localhost:8080",
		})
		done <- err
	}()

	select {
	case err := <-done:
		if err == nil {
			t.Fatalf("expected upload error when SaveStream fails after partial read")
		}
		if !provider.readCalled {
			t.Fatalf("expected fake provider to read from stream before failing")
		}
		if records, listErr := repo.ListAllImages(ctx); listErr != nil {
			t.Fatalf("ListAllImages returned error: %v", listErr)
		} else if len(records) != 0 {
			t.Fatalf("expected no image rows after failed partial save, got %d", len(records))
		}
	case <-time.After(2 * time.Second):
		t.Fatal("Upload hung when SaveStream failed after partial read")
	}
}

func TestPrepareUploadSourceSpoolsReaderToTempFileAndComputesMD5(t *testing.T) {
	service, _, _, _, _ := newImageServiceTestHarness(t)
	source := mustPNGBytes(t, color.RGBA{R: 200, G: 20, B: 20, A: 255})
	prepared, err := service.prepareUploadSource(UploadInput{
		Source:       bytes.NewReader(source),
		DeclaredSize: int64(len(source)),
	}, defaultRuntimeSettings().MaxUploadSizeBytes())
	if err != nil {
		t.Fatalf("prepareUploadSource returned error: %v", err)
	}
	defer prepared.Cleanup()
	if prepared.tempPath == "" {
		t.Fatalf("expected temp file path for reader-backed upload source")
	}
	if prepared.size != int64(len(source)) {
		t.Fatalf("expected size %d, got %d", len(source), prepared.size)
	}
	if prepared.originalMD5 != md5Hex(source) {
		t.Fatalf("expected md5 %q, got %q", md5Hex(source), prepared.originalMD5)
	}
	reader, err := prepared.Open()
	if err != nil {
		t.Fatalf("prepared.Open returned error: %v", err)
	}
	defer reader.Close()
	payload, err := io.ReadAll(reader)
	if err != nil {
		t.Fatalf("io.ReadAll returned error: %v", err)
	}
	if !bytes.Equal(payload, source) {
		t.Fatalf("expected payload bytes to round-trip")
	}
}

func TestPrepareUploadSourceRejectsOversizeReader(t *testing.T) {
	service, _, _, _, _ := newImageServiceTestHarness(t)
	_, err := service.prepareUploadSource(UploadInput{
		Source:       strings.NewReader("abcdef"),
		DeclaredSize: 6,
	}, 4)
	if err == nil || !containsError(err, ErrInvalidInput) {
		t.Fatalf("expected ErrInvalidInput for oversize upload source, got %v", err)
	}
}

func TestUploadUsesConfiguredAVIFConversionSettings(t *testing.T) {
	ctx := context.Background()
	service, _, _, _, uidCodec := newImageServiceTestHarness(t)
	uidCodec.Queue("uid-1")
	service.settings.Reconfigure(RuntimeSettings{
		SiteName:                     DefaultSiteName,
		SiteTagline:                  DefaultSiteTagline,
		MaxUploadSizeMB:              20,
		AllowedMIMETypes:             DefaultAllowedMIMETypes(),
		AvifQuality:                  42,
		AvifSpeed:                    3,
		AllowStorageSelect:           true,
		RateLimitWindowMinutes:       DefaultRateLimitWindowMinutes,
		RateLimitMaxRequests:         DefaultRateLimitMaxRequests,
		UploadRateLimitWindowMinutes: DefaultUploadRateLimitWindowMinutes,
		UploadRateLimitMaxRequests:   DefaultUploadRateLimitMaxRequests,
	})

	var observed *AVIFConversionSettings
	service.encoder = func(source io.Reader, target io.Writer, settings AVIFConversionSettings) error {
		copy := settings
		observed = &copy
		return encodeAVIFToWriter(source, target, settings)
	}

	if _, err := service.Upload(ctx, UploadInput{
		Token:            "token-a",
		OriginalFilename: "sample.png",
		MIMEType:         "image/png",
		Bytes:            mustPNGBytes(t, color.RGBA{R: 120, G: 80, B: 40, A: 255}),
		BaseURL:          "http://localhost:8080",
	}); err != nil {
		t.Fatalf("Upload returned error: %v", err)
	}
	if observed == nil {
		t.Fatalf("expected encoder to observe avif settings")
	}
	if observed.Quality != 42 || observed.Speed != 3 {
		t.Fatalf("expected configured avif settings quality=42 speed=3, got %+v", *observed)
	}
}

func TestUploadPrefersProvidedOriginalMD5(t *testing.T) {
	ctx := context.Background()
	service, repo, _, _, uidCodec := newImageServiceTestHarness(t)
	uidCodec.Queue("uid-1", "uid-2")

	sourceBytes := mustPNGBytes(t, color.RGBA{R: 90, G: 40, B: 10, A: 255})
	providedHash := md5Hex(sourceBytes)

	if _, err := service.Upload(ctx, UploadInput{
		Token:            "token-a",
		OriginalFilename: "sample.png",
		MIMEType:         "image/png",
		Bytes:            sourceBytes,
		OriginalMD5:      strings.ToUpper(providedHash),
		BaseURL:          "http://localhost:8080",
	}); err != nil {
		t.Fatalf("first upload returned error: %v", err)
	}

	result, err := service.Upload(ctx, UploadInput{
		Token:            "token-b",
		OriginalFilename: "sample.png",
		MIMEType:         "image/png",
		Bytes:            sourceBytes,
		OriginalMD5:      providedHash,
		BaseURL:          "http://localhost:8080",
	})
	if err != nil {
		t.Fatalf("second upload returned error: %v", err)
	}
	if !result.Duplicate {
		t.Fatalf("expected second upload to deduplicate using provided original md5")
	}

	secondRecord, err := repo.FindByUID(ctx, "uid-2")
	if err != nil {
		t.Fatalf("FindByUID second failed: %v", err)
	}
	if secondRecord.MD5Hash != providedHash {
		t.Fatalf("expected stored md5 hash %q, got %q", providedHash, secondRecord.MD5Hash)
	}
}

func TestDeleteRetainsPhysicalFileAfterLastReferenceDeletion(t *testing.T) {
	ctx := context.Background()
	service, repo, cacheStore, rootDir, uidCodec := newImageServiceTestHarness(t)
	uidCodec.Queue("uid-a", "uid-b")

	sourceBytes := mustPNGBytes(t, color.RGBA{R: 10, G: 40, B: 200, A: 255})
	if _, err := service.Upload(ctx, UploadInput{
		Token:            "token-a",
		OriginalFilename: "sample.png",
		MIMEType:         "image/png",
		Bytes:            sourceBytes,
		BaseURL:          "http://localhost:8080",
	}); err != nil {
		t.Fatalf("first upload returned error: %v", err)
	}
	if _, err := service.Upload(ctx, UploadInput{
		Token:            "token-b",
		OriginalFilename: "sample.png",
		MIMEType:         "image/png",
		Bytes:            sourceBytes,
		BaseURL:          "http://localhost:8080",
	}); err != nil {
		t.Fatalf("second upload returned error: %v", err)
	}

	firstRecord, err := repo.FindByUID(ctx, "uid-a")
	if err != nil {
		t.Fatalf("FindByUID first failed: %v", err)
	}
	storedPath := filepath.Join(rootDir, filepath.FromSlash(firstRecord.FilePath))

	if err := service.Delete(ctx, "uid-a"+publicImageExtension, "token-a", false, ""); err != nil {
		t.Fatalf("Delete first returned error: %v", err)
	}
	if _, err := os.Stat(storedPath); err != nil {
		t.Fatalf("expected physical file to remain after first delete: %v", err)
	}
	if _, ok := cacheStore.cachedMD5(model.NewMD5MappingKey(firstRecord.StorageKey, firstRecord.MD5Hash)); !ok {
		t.Fatalf("expected md5 cache to remain while references exist")
	}

	if err := service.Delete(ctx, "uid-b"+publicImageExtension, "token-b", false, ""); err != nil {
		t.Fatalf("Delete second returned error: %v", err)
	}
	if _, err := os.Stat(storedPath); err != nil {
		t.Fatalf("expected physical file to remain after last logical delete: %v", err)
	}
	if _, ok := cacheStore.cachedMD5(model.NewMD5MappingKey(firstRecord.StorageKey, firstRecord.MD5Hash)); ok {
		t.Fatalf("expected md5 cache to be removed after last delete")
	}
}

func TestUploadRepairsStaleMD5MappingWhenCachedUIDBelongsToDifferentMD5(t *testing.T) {
	ctx := context.Background()
	service, repo, cacheStore, _, uidCodec := newImageServiceTestHarness(t)
	uidCodec.Add("uid-wrong")
	uidCodec.Queue("uid-correct", "uid-dup")

	if err := repo.InsertImage(ctx, model.ImageRecord{
		UID:            "uid-wrong",
		Token:          "token-wrong",
		StorageKey:     "local-primary",
		StorageBackend: config.StorageBackendLocal,
		FilePath:       "2026/04/wrong.avif",
		MIMEType:       publicImageMIMEType,
		Size:           1,
		MD5Hash:        "different-md5",
	}); err != nil {
		t.Fatalf("InsertImage wrong record returned error: %v", err)
	}

	sourceBytes := mustPNGBytes(t, color.RGBA{R: 80, G: 20, B: 180, A: 255})
	md5Key := model.NewMD5MappingKey("local-primary", md5Hex(sourceBytes))
	cacheStore.setCachedMD5(md5Key, "uid-wrong")

	first, err := service.Upload(ctx, UploadInput{
		Token:            "token-correct",
		OriginalFilename: "sample.png",
		MIMEType:         "image/png",
		Bytes:            sourceBytes,
		BaseURL:          "http://localhost:8080",
	})
	if err != nil {
		t.Fatalf("first upload returned error: %v", err)
	}
	if first.Duplicate {
		t.Fatalf("expected stale wrong-md5 mapping to avoid false duplicate reuse")
	}
	if got, _ := cacheStore.cachedMD5(md5Key); got != "uid-correct" {
		t.Fatalf("expected md5 cache to repair to correct uid, got %q", got)
	}

	second, err := service.Upload(ctx, UploadInput{
		Token:            "token-dup",
		OriginalFilename: "sample.png",
		MIMEType:         "image/png",
		Bytes:            sourceBytes,
		BaseURL:          "http://localhost:8080",
	})
	if err != nil {
		t.Fatalf("second upload returned error: %v", err)
	}
	if !second.Duplicate {
		t.Fatalf("expected repaired mapping to deduplicate later upload")
	}
}

func TestDeleteRepointsMD5CacheToRemainingReference(t *testing.T) {
	ctx := context.Background()
	service, repo, cacheStore, _, uidCodec := newImageServiceTestHarness(t)
	uidCodec.Queue("uid-a", "uid-b")

	sourceBytes := mustPNGBytes(t, color.RGBA{R: 120, G: 200, B: 30, A: 255})
	if _, err := service.Upload(ctx, UploadInput{
		Token:            "token-a",
		OriginalFilename: "sample.png",
		MIMEType:         "image/png",
		Bytes:            sourceBytes,
		BaseURL:          "http://localhost:8080",
	}); err != nil {
		t.Fatalf("first upload returned error: %v", err)
	}
	if _, err := service.Upload(ctx, UploadInput{
		Token:            "token-b",
		OriginalFilename: "sample.png",
		MIMEType:         "image/png",
		Bytes:            sourceBytes,
		BaseURL:          "http://localhost:8080",
	}); err != nil {
		t.Fatalf("second upload returned error: %v", err)
	}

	record, err := repo.FindByUID(ctx, "uid-a")
	if err != nil {
		t.Fatalf("FindByUID returned error: %v", err)
	}
	if got, _ := cacheStore.cachedMD5(model.NewMD5MappingKey(record.StorageKey, record.MD5Hash)); got != "uid-a" {
		t.Fatalf("expected initial md5 cache to point at first uid")
	}

	if err := service.Delete(ctx, "uid-a"+publicImageExtension, "token-a", false, ""); err != nil {
		t.Fatalf("Delete returned error: %v", err)
	}

	if got, _ := cacheStore.cachedMD5(model.NewMD5MappingKey(record.StorageKey, record.MD5Hash)); got != "uid-b" {
		t.Fatalf("expected md5 cache to repoint to remaining uid, got %q", got)
	}
}

func TestDeleteRepointsMD5CacheWhenCachedUIDHasSameStorageButDifferentMD5(t *testing.T) {
	ctx := context.Background()
	service, repo, cacheStore, _, uidCodec := newImageServiceTestHarness(t)
	uidCodec.Add("uid-wrong")
	uidCodec.Queue("uid-a", "uid-b")

	if err := repo.InsertImage(ctx, model.ImageRecord{
		UID:            "uid-wrong",
		Token:          "token-wrong",
		StorageKey:     "local-primary",
		StorageBackend: config.StorageBackendLocal,
		FilePath:       "2026/04/wrong.avif",
		MIMEType:       publicImageMIMEType,
		Size:           1,
		MD5Hash:        "different-md5",
	}); err != nil {
		t.Fatalf("InsertImage wrong record returned error: %v", err)
	}

	sourceBytes := mustPNGBytes(t, color.RGBA{R: 30, G: 160, B: 90, A: 255})
	if _, err := service.Upload(ctx, UploadInput{
		Token:            "token-a",
		OriginalFilename: "sample.png",
		MIMEType:         "image/png",
		Bytes:            sourceBytes,
		BaseURL:          "http://localhost:8080",
	}); err != nil {
		t.Fatalf("first upload returned error: %v", err)
	}
	if _, err := service.Upload(ctx, UploadInput{
		Token:            "token-b",
		OriginalFilename: "sample.png",
		MIMEType:         "image/png",
		Bytes:            sourceBytes,
		BaseURL:          "http://localhost:8080",
	}); err != nil {
		t.Fatalf("second upload returned error: %v", err)
	}

	record, err := repo.FindByUID(ctx, "uid-a")
	if err != nil {
		t.Fatalf("FindByUID returned error: %v", err)
	}
	md5Key := model.NewMD5MappingKey(record.StorageKey, record.MD5Hash)
	cacheStore.setCachedMD5(md5Key, "uid-wrong")

	if err := service.Delete(ctx, "uid-a"+publicImageExtension, "token-a", false, ""); err != nil {
		t.Fatalf("Delete returned error: %v", err)
	}
	if got, _ := cacheStore.cachedMD5(md5Key); got != "uid-b" {
		t.Fatalf("expected md5 cache to repoint away from wrong-md5 cached uid, got %q", got)
	}
}

func TestDeleteRetainsPhysicalFileWhenOnlyCrossBackendReferenceTextMatches(t *testing.T) {
	ctx := context.Background()
	service, repo, cacheStore, rootDir, uidCodec := newImageServiceTestHarness(t)
	uidCodec.Queue("uid-local")

	sourceBytes := mustPNGBytes(t, color.RGBA{R: 200, G: 20, B: 60, A: 255})
	if _, err := service.Upload(ctx, UploadInput{
		Token:            "token-local",
		OriginalFilename: "sample.png",
		MIMEType:         "image/png",
		Bytes:            sourceBytes,
		BaseURL:          "http://localhost:8080",
	}); err != nil {
		t.Fatalf("upload returned error: %v", err)
	}

	localRecord, err := repo.FindByUID(ctx, "uid-local")
	if err != nil {
		t.Fatalf("FindByUID returned error: %v", err)
	}
	uidCodec.Add("uid-s3")
	if err := repo.InsertImage(ctx, model.ImageRecord{
		UID:            "uid-s3",
		Token:          "token-s3",
		StorageKey:     "s3-legacy",
		StorageBackend: config.StorageBackendS3,
		FilePath:       localRecord.FilePath,
		MIMEType:       localRecord.MIMEType,
		Size:           localRecord.Size,
		MD5Hash:        localRecord.MD5Hash,
	}); err != nil {
		t.Fatalf("InsertImage returned error: %v", err)
	}
	cacheStore.setCachedMD5(model.NewMD5MappingKey(localRecord.StorageKey, localRecord.MD5Hash), "uid-local")

	storedPath := filepath.Join(rootDir, filepath.FromSlash(localRecord.FilePath))
	if err := service.Delete(ctx, "uid-local"+publicImageExtension, "token-local", false, ""); err != nil {
		t.Fatalf("Delete returned error: %v", err)
	}
	if _, err := os.Stat(storedPath); err != nil {
		t.Fatalf("expected local file to remain for deferred cleanup, got err=%v", err)
	}
	if _, ok := cacheStore.cachedMD5(model.NewMD5MappingKey(localRecord.StorageKey, localRecord.MD5Hash)); ok {
		t.Fatalf("expected md5 cache to clear instead of repointing to another storage key")
	}
}

func TestDeleteRejectsTokenMismatch(t *testing.T) {
	ctx := context.Background()
	service, _, _, _, uidCodec := newImageServiceTestHarness(t)
	uidCodec.Queue("uid-1")

	if _, err := service.Upload(ctx, UploadInput{
		Token:            "owner-token",
		OriginalFilename: "sample.png",
		MIMEType:         "image/png",
		Bytes:            mustPNGBytes(t, color.RGBA{R: 0, G: 120, B: 255, A: 255}),
		BaseURL:          "http://localhost:8080",
	}); err != nil {
		t.Fatalf("upload returned error: %v", err)
	}

	if err := service.Delete(ctx, "uid-1"+publicImageExtension, "other-token", false, ""); err != ErrForbidden {
		t.Fatalf("expected ErrForbidden, got %v", err)
	}
}

func TestResolveRequiresAVIFURLAndRehydratesCache(t *testing.T) {
	ctx := context.Background()
	service, repo, cacheStore, _, uidCodec := newImageServiceTestHarness(t)
	uidCodec.Add("uid-r")

	record := model.ImageRecord{
		UID:            "uid-r",
		Token:          "token-r",
		StorageKey:     "local-primary",
		StorageBackend: config.StorageBackendLocal,
		FilePath:       "2026/04/hash.avif",
		MIMEType:       publicImageMIMEType,
		Size:           12,
		MD5Hash:        "abc",
	}
	if err := repo.InsertImage(ctx, record); err != nil {
		t.Fatalf("InsertImage returned error: %v", err)
	}

	if _, err := service.Resolve(ctx, "uid-r"); err != ErrNotFound {
		t.Fatalf("expected bare uid route to return ErrNotFound, got %v", err)
	}

	result, err := service.Resolve(ctx, "uid-r"+publicImageExtension)
	if err != nil {
		t.Fatalf("Resolve returned error: %v", err)
	}
	if result.CacheHit {
		t.Fatalf("expected repository fallback on first resolve")
	}
	if !cacheStore.hasCachedImage("uid-r") {
		t.Fatalf("expected resolve fallback to hydrate cache")
	}

	second, err := service.Resolve(ctx, "uid-r"+publicImageExtension)
	if err != nil {
		t.Fatalf("second Resolve returned error: %v", err)
	}
	if !second.CacheHit {
		t.Fatalf("expected second resolve to hit cache")
	}
}

func TestResolveRejectsInvalidPublicUID(t *testing.T) {
	ctx := context.Background()
	service, _, _, _, _ := newImageServiceTestHarness(t)

	if _, err := service.Resolve(ctx, "not-a-token"+publicImageExtension); err != ErrNotFound {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestResolveRejectsNonCanonicalAVIFSuffix(t *testing.T) {
	ctx := context.Background()
	service, _, _, _, uidCodec := newImageServiceTestHarness(t)
	uidCodec.Add("uid-r")

	if _, err := service.Resolve(ctx, "uid-r.AVIF"); err != ErrNotFound {
		t.Fatalf("expected ErrNotFound for mixed-case suffix, got %v", err)
	}
}

func TestDeleteRejectsNonCanonicalAVIFSuffix(t *testing.T) {
	ctx := context.Background()
	service, _, _, _, uidCodec := newImageServiceTestHarness(t)
	uidCodec.Queue("uid-1")

	if _, err := service.Upload(ctx, UploadInput{
		Token:            "owner-token",
		OriginalFilename: "sample.png",
		MIMEType:         "image/png",
		Bytes:            mustPNGBytes(t, color.RGBA{R: 0, G: 120, B: 255, A: 255}),
		BaseURL:          "http://localhost:8080",
	}); err != nil {
		t.Fatalf("upload returned error: %v", err)
	}

	if err := service.Delete(ctx, "uid-1.AVIF", "owner-token", false, ""); err != ErrNotFound {
		t.Fatalf("expected ErrNotFound for mixed-case suffix, got %v", err)
	}
}

func TestPreheatWarmsUIDAndMD5Keys(t *testing.T) {
	ctx := context.Background()
	service, repo, cacheStore, _, uidCodec := newImageServiceTestHarness(t)
	uidCodec.Add("uid-1", "uid-2")

	records := []model.ImageRecord{
		{
			UID:            "uid-1",
			Token:          "token-1",
			StorageKey:     "local-primary",
			StorageBackend: config.StorageBackendLocal,
			FilePath:       "2026/04/one.avif",
			MIMEType:       publicImageMIMEType,
			Size:           1,
			MD5Hash:        "hash-1",
		},
		{
			UID:            "uid-2",
			Token:          "token-2",
			StorageKey:     "local-primary",
			StorageBackend: config.StorageBackendLocal,
			FilePath:       "2026/04/two.avif",
			MIMEType:       publicImageMIMEType,
			Size:           2,
			MD5Hash:        "hash-2",
		},
	}
	for _, record := range records {
		if err := repo.InsertImage(ctx, record); err != nil {
			t.Fatalf("InsertImage returned error: %v", err)
		}
	}

	count, err := service.Preheat(ctx)
	if err != nil {
		t.Fatalf("Preheat returned error: %v", err)
	}
	if count != len(records) {
		t.Fatalf("expected preheat count %d, got %d", len(records), count)
	}
	imageSets, md5Sets, imageBatchSets, md5BatchSets := cacheStore.stats()
	if imageSets != len(records) || md5Sets != len(records) {
		t.Fatalf("expected preheat to populate cache for all records")
	}
	if imageBatchSets != 1 || md5BatchSets != 1 {
		t.Fatalf("expected preheat to use batched redis writes, got image batches %d and md5 batches %d", imageBatchSets, md5BatchSets)
	}
}

func TestPreheatRepairsStaleMD5MappingFromSQLite(t *testing.T) {
	ctx := context.Background()
	service, repo, cacheStore, _, uidCodec := newImageServiceTestHarness(t)
	uidCodec.Add("uid-1", "uid-2")

	records := []model.ImageRecord{
		{
			UID:            "uid-1",
			Token:          "token-1",
			StorageKey:     "local-primary",
			StorageBackend: config.StorageBackendLocal,
			FilePath:       "2026/04/shared.avif",
			MIMEType:       publicImageMIMEType,
			Size:           1,
			MD5Hash:        "shared-hash",
		},
		{
			UID:            "uid-2",
			Token:          "token-2",
			StorageKey:     "local-primary",
			StorageBackend: config.StorageBackendLocal,
			FilePath:       "2026/04/shared.avif",
			MIMEType:       publicImageMIMEType,
			Size:           1,
			MD5Hash:        "shared-hash",
		},
	}
	for _, record := range records {
		if err := repo.InsertImage(ctx, record); err != nil {
			t.Fatalf("InsertImage returned error: %v", err)
		}
	}

	cacheStore.setCachedMD5(model.NewMD5MappingKey("local-primary", "shared-hash"), "stale-uid")

	count, err := service.Preheat(ctx)
	if err != nil {
		t.Fatalf("Preheat returned error: %v", err)
	}
	if count != len(records) {
		t.Fatalf("expected preheat count %d, got %d", len(records), count)
	}
	if got, _ := cacheStore.cachedMD5(model.NewMD5MappingKey("local-primary", "shared-hash")); got != "uid-1" {
		t.Fatalf("expected preheat to repair md5 cache to first sqlite uid, got %q", got)
	}
	_, md5Sets, _, _ := cacheStore.stats()
	if md5Sets != 1 {
		t.Fatalf("expected preheat to write one md5 mapping for duplicate hash, got %d", md5Sets)
	}
}

func TestPublicRuntimeSettingsExposeOnlySafeStorageFields(t *testing.T) {
	ctx := context.Background()
	service, repo, _, _, _ := newImageServiceTestHarness(t)

	if err := repo.CreateStorageConfig(ctx, config.RuntimeStorageConfig{
		StorageKey:       "s3-private",
		Name:             "S3 Private",
		Backend:          config.StorageBackendS3,
		S3Endpoint:       "s3.example.test",
		S3Region:         "auto",
		S3Bucket:         "private-bucket",
		S3AccessKey:      "private-access",
		S3SecretKey:      "private-secret",
		S3ForcePathStyle: true,
	}); err != nil {
		t.Fatalf("CreateStorageConfig s3 returned error: %v", err)
	}
	if err := repo.CreateStorageConfig(ctx, config.RuntimeStorageConfig{
		StorageKey: "webdav-private",
		Name:       "WebDAV Private",
		Backend:    config.StorageBackendWebDAV,
		WebDAVURL:  "https://dav.example.test/remote.php/dav/files/demo",
		WebDAVUser: "private-user",
		WebDAVPass: "private-pass",
	}); err != nil {
		t.Fatalf("CreateStorageConfig webdav returned error: %v", err)
	}

	settings, err := service.PublicRuntimeSettings(ctx)
	if err != nil {
		t.Fatalf("PublicRuntimeSettings returned error: %v", err)
	}
	if len(settings.Upload.AllowedMIMETypes) == 0 {
		t.Fatalf("expected public upload settings to expose allowed_mime_types")
	}
	payload, err := json.Marshal(settings.Upload)
	if err != nil {
		t.Fatalf("json.Marshal upload returned error: %v", err)
	}
	var uploadFields map[string]any
	if err := json.Unmarshal(payload, &uploadFields); err != nil {
		t.Fatalf("json.Unmarshal upload returned error: %v", err)
	}
	if _, exists := uploadFields["effective_allowed_mime_types"]; exists {
		t.Fatalf("public upload settings leaked effective_allowed_mime_types in %+v", uploadFields)
	}
	options := settings.Storage.Options
	if len(options) != 4 {
		t.Fatalf("expected four public storage options, got %d", len(options))
	}
	if options[0].StorageKey != "local-primary" || !options[0].IsDefault {
		t.Fatalf("expected default option first, got %+v", options[0])
	}
	if options[0].Name == "" || options[0].StorageBackend != config.StorageBackendLocal {
		t.Fatalf("expected safe display fields, got %+v", options[0])
	}

	allowedFields := map[string]struct{}{
		"storage_key":     {},
		"name":            {},
		"storage_backend": {},
		"is_default":      {},
	}
	for _, option := range options {
		payload, err := json.Marshal(option)
		if err != nil {
			t.Fatalf("json.Marshal returned error: %v", err)
		}
		var fields map[string]any
		if err := json.Unmarshal(payload, &fields); err != nil {
			t.Fatalf("json.Unmarshal returned error: %v", err)
		}
		for field := range fields {
			if _, ok := allowedFields[field]; !ok {
				t.Fatalf("public storage option leaked field %q in %+v", field, fields)
			}
		}
		for _, forbidden := range []string{
			"local_storage_path",
			"s3_endpoint",
			"s3_region",
			"s3_bucket",
			"s3_access_key",
			"s3_secret_key",
			"webdav_url",
			"webdav_user",
			"webdav_pass",
		} {
			if _, exists := fields[forbidden]; exists {
				t.Fatalf("public storage option leaked %s in %+v", forbidden, fields)
			}
		}
	}
}

func newImageServiceTestHarness(t *testing.T) (*ImageService, *repository.Repository, *fakeCache, string, *fakeUIDCodec) {
	t.Helper()

	dir := t.TempDir()
	repo, err := repository.New(filepath.Join(dir, "test.sqlite"))
	if err != nil {
		t.Fatalf("repository.New returned error: %v", err)
	}
	t.Cleanup(func() {
		_ = repo.Close()
	})

	ctx := context.Background()
	if err := repo.Migrate(ctx); err != nil {
		t.Fatalf("Migrate returned error: %v", err)
	}

	rootDir := filepath.Join(dir, "images")
	storageConfigs := []config.RuntimeStorageConfig{
		{
			StorageKey:       "local-primary",
			Name:             "Local Primary",
			IsDefault:        true,
			Backend:          config.StorageBackendLocal,
			LocalStoragePath: rootDir,
		},
		{
			StorageKey:       "local-secondary",
			Name:             "Local Secondary",
			IsDefault:        false,
			Backend:          config.StorageBackendLocal,
			LocalStoragePath: filepath.Join(dir, "images-secondary"),
		},
	}
	for _, cfg := range storageConfigs {
		if err := repo.CreateStorageConfig(ctx, cfg); err != nil {
			t.Fatalf("CreateStorageConfig returned error: %v", err)
		}
	}

	manager, err := storage.NewManager(storageConfigs)
	if err != nil {
		t.Fatalf("NewManager returned error: %v", err)
	}

	cacheStore := newFakeCache()
	imageCache := cacheStore.images
	md5Cache := cacheStore.md5
	uidCodec := newFakeUIDCodec()
	logger := slog.New(slog.NewTextHandler(ioDiscard{}, nil))
	settingsManager := NewRuntimeSettingsManager()
	return NewImageServiceWithCaches(repo, imageCache, imageCache, md5Cache, md5Cache, manager, settingsManager, uidCodec.Generate, uidCodec.Validate, logger), repo, cacheStore, rootDir, uidCodec
}

type ioDiscard struct{}

func (ioDiscard) Write(p []byte) (int, error) {
	return len(p), nil
}

func mustPNGBytes(t *testing.T, fill color.Color) []byte {
	t.Helper()

	img := image.NewNRGBA(image.Rect(0, 0, 2, 2))
	for y := 0; y < 2; y++ {
		for x := 0; x < 2; x++ {
			img.Set(x, y, fill)
		}
	}

	var output bytes.Buffer
	if err := png.Encode(&output, img); err != nil {
		t.Fatalf("png.Encode returned error: %v", err)
	}
	return output.Bytes()
}

func mustAVIFBytes(t *testing.T, fill color.Color) []byte {
	t.Helper()

	img := image.NewNRGBA(image.Rect(0, 0, 2, 2))
	for y := 0; y < 2; y++ {
		for x := 0; x < 2; x++ {
			img.Set(x, y, fill)
		}
	}

	var output bytes.Buffer
	if err := avif.Encode(&output, img, avif.Options{
		Quality: 60,
		Speed:   8,
	}); err != nil {
		t.Fatalf("avif.Encode returned error: %v", err)
	}
	return output.Bytes()
}
