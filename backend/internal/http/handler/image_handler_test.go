package handler

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
	"net/http/httptest"
	"path/filepath"
	"testing"

	"github.com/gen2brain/avif"
	"github.com/gin-gonic/gin"

	"omepic/backend/internal/config"
	"omepic/backend/internal/model"
	"omepic/backend/internal/repository"
	"omepic/backend/internal/service"
	"omepic/backend/internal/storage"
)

type handlerFakeCache struct {
	images   map[string]model.CachedImage
	md5ToUID map[string]string
}

func newHandlerFakeCache() *handlerFakeCache {
	return &handlerFakeCache{
		images:   make(map[string]model.CachedImage),
		md5ToUID: make(map[string]string),
	}
}

func (c *handlerFakeCache) GetImage(_ context.Context, uid string) (*model.CachedImage, error) {
	record, ok := c.images[uid]
	if !ok {
		return nil, nil
	}
	copy := record
	return &copy, nil
}

func (c *handlerFakeCache) SetImage(_ context.Context, record model.ImageRecord) error {
	c.images[record.UID] = model.CachedImageFromRecord(record)
	return nil
}

func (c *handlerFakeCache) SetImages(_ context.Context, records []model.ImageRecord) error {
	for _, record := range records {
		c.images[record.UID] = model.CachedImageFromRecord(record)
	}
	return nil
}

func (c *handlerFakeCache) DeleteImage(_ context.Context, uid string) error {
	delete(c.images, uid)
	return nil
}

func (c *handlerFakeCache) GetMD5(_ context.Context, md5Hash string) (string, error) {
	return c.md5ToUID[md5Hash], nil
}

func (c *handlerFakeCache) SetMD5IfAbsent(_ context.Context, md5Hash string, uid string) error {
	if _, ok := c.md5ToUID[md5Hash]; !ok {
		c.md5ToUID[md5Hash] = uid
	}
	return nil
}

func (c *handlerFakeCache) SetMD5(_ context.Context, md5Hash string, uid string) error {
	c.md5ToUID[md5Hash] = uid
	return nil
}

func (c *handlerFakeCache) SetMD5Mappings(_ context.Context, mappings map[string]string) error {
	for md5Hash, uid := range mappings {
		c.md5ToUID[md5Hash] = uid
	}
	return nil
}

func (c *handlerFakeCache) DeleteMD5(_ context.Context, md5Hash string) error {
	delete(c.md5ToUID, md5Hash)
	return nil
}

func (c *handlerFakeCache) Ping(_ context.Context) error {
	return nil
}

func TestServeStreamsStoredAVIFByUIDRoute(t *testing.T) {
	gin.SetMode(gin.TestMode)

	imageHandler, uploadResult := newImageHandlerTestHarness(t)
	engine := gin.New()
	engine.GET("/i/:uid", imageHandler.Serve)

	recorder := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/i/"+uploadResult.UID+".avif", nil)
	engine.ServeHTTP(recorder, req)

	if recorder.Code != 200 {
		t.Fatalf("expected status 200, got %d", recorder.Code)
	}
	if contentType := recorder.Header().Get("Content-Type"); contentType != "image/avif" {
		t.Fatalf("expected content type image/avif, got %q", contentType)
	}
	if _, err := avif.Decode(bytes.NewReader(recorder.Body.Bytes())); err != nil {
		t.Fatalf("expected handler response body to be AVIF-decodable: %v", err)
	}
}

func TestServeRejectsBareUIDRoute(t *testing.T) {
	gin.SetMode(gin.TestMode)

	imageHandler, uploadResult := newImageHandlerTestHarness(t)
	engine := gin.New()
	engine.GET("/i/:uid", imageHandler.Serve)

	recorder := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/i/"+uploadResult.UID, nil)
	engine.ServeHTTP(recorder, req)

	if recorder.Code != 404 {
		t.Fatalf("expected status 404, got %d", recorder.Code)
	}
}

func TestDetectContentTypeRecognizesAVIF(t *testing.T) {
	if contentType := detectContentType("remote-image.avif"); contentType != "image/avif" {
		t.Fatalf("expected image/avif, got %q", contentType)
	}
}

func TestRuntimeSettingsReturnsSafePublicCatalog(t *testing.T) {
	gin.SetMode(gin.TestMode)

	imageHandler, _ := newImageHandlerTestHarness(t)
	engine := gin.New()
	engine.GET("/v1/runtime-settings", imageHandler.RuntimeSettings)

	recorder := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/v1/runtime-settings", nil)
	engine.ServeHTTP(recorder, req)

	if recorder.Code != 200 {
		t.Fatalf("expected status 200, got %d", recorder.Code)
	}

	var payload map[string]any
	if err := json.Unmarshal(recorder.Body.Bytes(), &payload); err != nil {
		t.Fatalf("json.Unmarshal returned error: %v", err)
	}
	data, ok := payload["data"].(map[string]any)
	if !ok {
		t.Fatalf("expected data object, got %#v", payload["data"])
	}
	storageView, ok := data["storage"].(map[string]any)
	if !ok {
		t.Fatalf("expected storage object, got %#v", data["storage"])
	}
	items, ok := storageView["options"].([]any)
	if !ok || len(items) != 2 {
		t.Fatalf("expected two storage options, got %#v", data["items"])
	}
	first, ok := items[0].(map[string]any)
	if !ok {
		t.Fatalf("expected first item object, got %#v", items[0])
	}
	for _, secretField := range []string{"local_storage_path", "s3_access_key", "s3_secret_key", "webdav_pass"} {
		if _, exists := first[secretField]; exists {
			t.Fatalf("public storage option leaked %s", secretField)
		}
	}
	if first["storage_key"] != "local-primary" || first["name"] != "Local Primary" || first["storage_backend"] != config.StorageBackendLocal || first["is_default"] != true {
		t.Fatalf("unexpected first public option: %#v", first)
	}
}

func newImageHandlerTestHarness(t *testing.T) (*ImageHandler, service.UploadOutput) {
	t.Helper()

	dir := t.TempDir()
	repo, err := repository.New(filepath.Join(dir, "test.sqlite"))
	if err != nil {
		t.Fatalf("repository.New returned error: %v", err)
	}
	t.Cleanup(func() {
		_ = repo.Close()
	})

	if err := repo.Migrate(context.Background()); err != nil {
		t.Fatalf("Migrate returned error: %v", err)
	}

	storageConfigs := []config.RuntimeStorageConfig{
		{
			StorageKey:       "local-primary",
			Name:             "Local Primary",
			IsDefault:        true,
			Backend:          config.StorageBackendLocal,
			LocalStoragePath: filepath.Join(dir, "images"),
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
		if err := repo.CreateStorageConfig(context.Background(), cfg); err != nil {
			t.Fatalf("CreateStorageConfig returned error: %v", err)
		}
	}

	manager, err := storage.NewManager(storageConfigs)
	if err != nil {
		t.Fatalf("storage.NewManager returned error: %v", err)
	}

	cacheStore := newHandlerFakeCache()
	validUID := "uid-handler"
	settingsManager := service.NewRuntimeSettingsManager("")
	imageService := service.NewImageService(
		repo,
		cacheStore,
		manager,
		settingsManager,
		func() (string, error) { return validUID, nil },
		func(uid string) error {
			if uid != validUID {
				return errors.New("invalid uid")
			}
			return nil
		},
		slog.New(slog.NewTextHandler(discardWriter{}, nil)),
	)

	uploadResult, err := imageService.Upload(context.Background(), service.UploadInput{
		Token:            "token-handler",
		OriginalFilename: "sample.png",
		MIMEType:         "image/png",
		Bytes:            mustPNGBytes(t, color.RGBA{R: 30, G: 180, B: 220, A: 255}),
		BaseURL:          "http://localhost:8080",
	})
	if err != nil {
		t.Fatalf("Upload returned error: %v", err)
	}

	return NewImageHandler(
		imageService,
		manager,
		slog.New(slog.NewTextHandler(discardWriter{}, nil)),
		nil,
	), uploadResult
}

type discardWriter struct{}

func (discardWriter) Write(p []byte) (int, error) {
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

var _ io.Writer = discardWriter{}
