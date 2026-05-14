package service

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"image"
	"image/color"
	"image/png"
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

type fakeCache struct {
	mu             sync.Mutex
	images         map[string]model.CachedImage
	md5ToUID       map[string]string
	imageSets      int
	imageBatchSets int
	md5Sets        int
	md5BatchSets   int
}

func newFakeCache() *fakeCache {
	return &fakeCache{
		images:   make(map[string]model.CachedImage),
		md5ToUID: make(map[string]string),
	}
}

func (c *fakeCache) GetImage(_ context.Context, uid string) (*model.CachedImage, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	value, ok := c.images[uid]
	if !ok {
		return nil, nil
	}
	copy := value
	return &copy, nil
}

func (c *fakeCache) SetImage(_ context.Context, record model.ImageRecord) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.images[record.UID] = model.CachedImageFromRecord(record)
	c.imageSets++
	return nil
}

func (c *fakeCache) SetImages(_ context.Context, records []model.ImageRecord) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	for _, record := range records {
		c.images[record.UID] = model.CachedImageFromRecord(record)
		c.imageSets++
	}
	c.imageBatchSets++
	return nil
}

func (c *fakeCache) DeleteImage(_ context.Context, uid string) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.images, uid)
	return nil
}

func (c *fakeCache) GetMD5(_ context.Context, md5Hash string) (string, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.md5ToUID[md5Hash], nil
}

func (c *fakeCache) SetMD5IfAbsent(_ context.Context, md5Hash string, uid string) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if _, ok := c.md5ToUID[md5Hash]; !ok {
		c.md5ToUID[md5Hash] = uid
		c.md5Sets++
	}
	return nil
}

func (c *fakeCache) SetMD5(_ context.Context, md5Hash string, uid string) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.md5ToUID[md5Hash] = uid
	c.md5Sets++
	return nil
}

func (c *fakeCache) SetMD5Mappings(_ context.Context, mappings map[string]string) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	for md5Hash, uid := range mappings {
		c.md5ToUID[md5Hash] = uid
		c.md5Sets++
	}
	c.md5BatchSets++
	return nil
}

func (c *fakeCache) DeleteMD5(_ context.Context, md5Hash string) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.md5ToUID, md5Hash)
	return nil
}

func (c *fakeCache) Ping(_ context.Context) error {
	return nil
}

func (c *fakeCache) cachedMD5(key string) (string, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	value, ok := c.md5ToUID[key]
	return value, ok
}

func (c *fakeCache) setCachedMD5(key string, uid string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.md5ToUID[key] = uid
}

func (c *fakeCache) hasCachedImage(uid string) bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	_, ok := c.images[uid]
	return ok
}

func (c *fakeCache) stats() (imageSets int, md5Sets int, imageBatchSets int, md5BatchSets int) {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.imageSets, c.md5Sets, c.imageBatchSets, c.md5BatchSets
}

type fakeUIDCodec struct {
	mu     sync.Mutex
	queued []string
	valid  map[string]struct{}
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

	if result.UID != "uid-1" {
		t.Fatalf("expected uid uid-1, got %q", result.UID)
	}
	if result.MIMEType != publicImageMIMEType {
		t.Fatalf("expected mime %q, got %q", publicImageMIMEType, result.MIMEType)
	}
	if !strings.HasSuffix(result.URL, "/i/uid-1"+publicImageExtension) {
		t.Fatalf("expected avif url, got %q", result.URL)
	}
	if !strings.Contains(result.MDURL, "sample.png") {
		t.Fatalf("expected markdown output to keep request filename, got %q", result.MDURL)
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
	if result.MIMEType != publicImageMIMEType {
		t.Fatalf("expected mime %q, got %q", publicImageMIMEType, result.MIMEType)
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

	result, err := service.Upload(ctx, UploadInput{
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
	if result.StorageKey != "local-secondary" {
		t.Fatalf("expected selected storage key in response, got %q", result.StorageKey)
	}
	if result.StorageBackend != config.StorageBackendLocal {
		t.Fatalf("expected selected backend in response, got %q", result.StorageBackend)
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

	result, err := service.Upload(ctx, UploadInput{
		Token:            "token-default-switch",
		OriginalFilename: "sample.png",
		MIMEType:         "image/png",
		Bytes:            mustPNGBytes(t, color.RGBA{R: 20, G: 210, B: 80, A: 255}),
		BaseURL:          "http://localhost:8080",
	})
	if err != nil {
		t.Fatalf("Upload returned error: %v", err)
	}
	if result.StorageKey != "local-secondary" {
		t.Fatalf("expected hot-reloaded default storage key local-secondary, got %q", result.StorageKey)
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
	service, repo, _, primaryRoot, uidCodec := newImageServiceTestHarness(t)
	uidCodec.Queue("uid-primary", "uid-secondary", "uid-secondary-dup")

	sourceBytes := mustPNGBytes(t, color.RGBA{R: 10, G: 150, B: 220, A: 255})
	first, err := service.Upload(ctx, UploadInput{
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
	if first.StorageKey != "local-primary" || second.StorageKey != "local-secondary" || third.StorageKey != "local-secondary" {
		t.Fatalf("unexpected response storage keys: %q %q %q", first.StorageKey, second.StorageKey, third.StorageKey)
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

func TestUploadSkipsAVIFConversionForDuplicateUploads(t *testing.T) {
	ctx := context.Background()
	service, repo, _, _, uidCodec := newImageServiceTestHarness(t)
	uidCodec.Queue("uid-1", "uid-2")

	sourceBytes := mustPNGBytes(t, color.RGBA{R: 80, G: 160, B: 40, A: 255})
	conversionCalls := 0
	service.transformer = func(payload []byte, settings AVIFConversionSettings) ([]byte, error) {
		conversionCalls++
		return convertToAVIFWithSettings(payload, settings)
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
	service.transformer = func(payload []byte, settings AVIFConversionSettings) ([]byte, error) {
		started <- md5Hex(payload)
		<-release
		return convertToAVIFWithSettings(payload, settings)
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
	service.transformer = func(payload []byte, settings AVIFConversionSettings) ([]byte, error) {
		copy := settings
		observed = &copy
		return convertToAVIFWithSettings(payload, settings)
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
		t.Fatalf("expected transformer to observe avif settings")
	}
	if observed.Quality != 42 || observed.Speed != 3 {
		t.Fatalf("expected configured avif settings quality=42 speed=3, got %+v", *observed)
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
	if _, ok := cacheStore.cachedMD5(scopedMD5CacheKey(firstRecord.StorageKey, firstRecord.MD5Hash)); !ok {
		t.Fatalf("expected md5 cache to remain while references exist")
	}

	if err := service.Delete(ctx, "uid-b"+publicImageExtension, "token-b", false, ""); err != nil {
		t.Fatalf("Delete second returned error: %v", err)
	}
	if _, err := os.Stat(storedPath); err != nil {
		t.Fatalf("expected physical file to remain after last logical delete: %v", err)
	}
	if _, ok := cacheStore.cachedMD5(scopedMD5CacheKey(firstRecord.StorageKey, firstRecord.MD5Hash)); ok {
		t.Fatalf("expected md5 cache to be removed after last delete")
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
	if got, _ := cacheStore.cachedMD5(scopedMD5CacheKey(record.StorageKey, record.MD5Hash)); got != "uid-a" {
		t.Fatalf("expected initial md5 cache to point at first uid")
	}

	if err := service.Delete(ctx, "uid-a"+publicImageExtension, "token-a", false, ""); err != nil {
		t.Fatalf("Delete returned error: %v", err)
	}

	if got, _ := cacheStore.cachedMD5(scopedMD5CacheKey(record.StorageKey, record.MD5Hash)); got != "uid-b" {
		t.Fatalf("expected md5 cache to repoint to remaining uid, got %q", got)
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
	cacheStore.setCachedMD5(scopedMD5CacheKey(localRecord.StorageKey, localRecord.MD5Hash), "uid-local")

	storedPath := filepath.Join(rootDir, filepath.FromSlash(localRecord.FilePath))
	if err := service.Delete(ctx, "uid-local"+publicImageExtension, "token-local", false, ""); err != nil {
		t.Fatalf("Delete returned error: %v", err)
	}
	if _, err := os.Stat(storedPath); err != nil {
		t.Fatalf("expected local file to remain for deferred cleanup, got err=%v", err)
	}
	if _, ok := cacheStore.cachedMD5(scopedMD5CacheKey(localRecord.StorageKey, localRecord.MD5Hash)); ok {
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

	cacheStore.setCachedMD5(scopedMD5CacheKey("local-primary", "shared-hash"), "stale-uid")

	count, err := service.Preheat(ctx)
	if err != nil {
		t.Fatalf("Preheat returned error: %v", err)
	}
	if count != len(records) {
		t.Fatalf("expected preheat count %d, got %d", len(records), count)
	}
	if got, _ := cacheStore.cachedMD5(scopedMD5CacheKey("local-primary", "shared-hash")); got != "uid-1" {
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
	uidCodec := newFakeUIDCodec()
	logger := slog.New(slog.NewTextHandler(ioDiscard{}, nil))
	settingsManager := NewRuntimeSettingsManager()
	return NewImageService(repo, cacheStore, manager, settingsManager, uidCodec.Generate, uidCodec.Validate, logger), repo, cacheStore, rootDir, uidCodec
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
