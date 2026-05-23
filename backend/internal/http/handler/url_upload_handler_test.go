package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"image/color"
	"io"
	"log/slog"
	"net"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strconv"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/gin-gonic/gin"

	"omepic/backend/internal/config"
	"omepic/backend/internal/repository"
	"omepic/backend/internal/service"
	"omepic/backend/internal/storage"
)

func TestUploadURLHandlerUploadsPublicMockImage(t *testing.T) {
	h, remoteBase := newURLUploadHandlerHarness(t, service.RuntimeSettings{})
	engine := gin.New()
	engine.POST("/v1/image/url", h.UploadURL)

	recorder := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/v1/image/url", strings.NewReader(`{"url":"`+remoteBase+`/image.png"}`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Token", "token-url")
	engine.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", recorder.Code, recorder.Body.String())
	}
	var payload struct {
		Success bool                 `json:"success"`
		Data    service.UploadOutput `json:"data"`
	}
	if err := json.Unmarshal(recorder.Body.Bytes(), &payload); err != nil {
		t.Fatalf("json.Unmarshal returned error: %v", err)
	}
	if !payload.Success || !strings.Contains(payload.Data.URL, "/i/uid-url-1.avif") {
		t.Fatalf("unexpected upload response: %#v", payload)
	}
}

func TestUploadURLHandlerRejectsRedirectToPrivateAddress(t *testing.T) {
	recorder := postURLUpload(t, service.RuntimeSettings{}, "/redirect-private", "token-url")
	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d: %s", recorder.Code, recorder.Body.String())
	}
}

func TestUploadURLHandlerRejectsContentLengthOverLimit(t *testing.T) {
	recorder := postURLUpload(t, service.RuntimeSettings{MaxUploadSizeMB: 1}, "/too-large", "token-url")
	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d: %s", recorder.Code, recorder.Body.String())
	}
}

func TestUploadURLHandlerRejectsTooManyRedirects(t *testing.T) {
	recorder := postURLUpload(t, service.RuntimeSettings{}, "/redirect?n=6", "token-url")
	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d: %s", recorder.Code, recorder.Body.String())
	}
}

func TestUploadURLHandlerRejectsInvalidScheme(t *testing.T) {
	h, _ := newURLUploadHandlerHarness(t, service.RuntimeSettings{})
	engine := gin.New()
	engine.POST("/v1/image/url", h.UploadURL)

	recorder := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/v1/image/url", strings.NewReader(`{"url":"ftp://example.com/image.png"}`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Token", "token-url")
	engine.ServeHTTP(recorder, req)
	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d: %s", recorder.Code, recorder.Body.String())
	}
}

func postURLUpload(t *testing.T, settings service.RuntimeSettings, remotePath string, token string) *httptest.ResponseRecorder {
	t.Helper()
	h, remoteBase := newURLUploadHandlerHarness(t, settings)
	engine := gin.New()
	engine.POST("/v1/image/url", h.UploadURL)
	recorder := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/v1/image/url", strings.NewReader(`{"url":"`+remoteBase+remotePath+`"}`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Token", token)
	engine.ServeHTTP(recorder, req)
	return recorder
}

func newURLUploadHandlerHarness(t *testing.T, runtime service.RuntimeSettings) (*ImageHandler, string) {
	t.Helper()
	gin.SetMode(gin.TestMode)

	remote := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/image.png":
			payload := mustPNGBytes(t, color.RGBA{R: 90, G: 40, B: 220, A: 255})
			w.Header().Set("Content-Type", "image/png")
			w.Header().Set("Content-Length", strconv.Itoa(len(payload)))
			_, _ = w.Write(payload)
		case "/redirect-private":
			http.Redirect(w, r, "http://127.0.0.1/private.png", http.StatusFound)
		case "/redirect":
			n, _ := strconv.Atoi(r.URL.Query().Get("n"))
			if n <= 0 {
				http.Redirect(w, r, "/image.png", http.StatusFound)
				return
			}
			http.Redirect(w, r, "/redirect?n="+strconv.Itoa(n-1), http.StatusFound)
		case "/too-large":
			w.Header().Set("Content-Type", "image/png")
			w.Header().Set("Content-Length", strconv.Itoa(2*1024*1024))
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("not read"))
		default:
			http.NotFound(w, r)
		}
	}))
	t.Cleanup(remote.Close)

	dir := t.TempDir()
	repo, err := repository.New(filepath.Join(dir, "test.sqlite"))
	if err != nil {
		t.Fatalf("repository.New returned error: %v", err)
	}
	t.Cleanup(func() { _ = repo.Close() })
	if err := repo.Migrate(context.Background()); err != nil {
		t.Fatalf("Migrate returned error: %v", err)
	}
	storageConfig := config.RuntimeStorageConfig{StorageKey: "local-primary", Name: "Local", IsDefault: true, Backend: config.StorageBackendLocal, LocalStoragePath: filepath.Join(dir, "images")}
	if err := repo.CreateStorageConfig(context.Background(), storageConfig); err != nil {
		t.Fatalf("CreateStorageConfig returned error: %v", err)
	}
	manager, err := storage.NewManager([]config.RuntimeStorageConfig{storageConfig})
	if err != nil {
		t.Fatalf("storage.NewManager returned error: %v", err)
	}
	settingsManager := service.NewRuntimeSettingsManager()
	if runtime.MaxUploadSizeMB == 0 {
		runtime = service.RuntimeSettings{MaxUploadSizeMB: 20, AllowedMIMETypes: service.DefaultAllowedMIMETypes(), MaxImagePixels: service.DefaultMaxImagePixels, AVIFMaxConcurrency: service.DefaultAVIFMaxConcurrency, AVIFConversionTimeoutSeconds: service.DefaultAVIFConversionTimeoutSeconds, AvifQuality: service.DefaultAVIFQuality, AvifSpeed: service.DefaultAVIFSpeed}
	} else {
		runtime.AllowedMIMETypes = service.DefaultAllowedMIMETypes()
		runtime.MaxImagePixels = service.DefaultMaxImagePixels
		runtime.AVIFMaxConcurrency = service.DefaultAVIFMaxConcurrency
		runtime.AVIFConversionTimeoutSeconds = service.DefaultAVIFConversionTimeoutSeconds
		runtime.AvifQuality = service.DefaultAVIFQuality
		runtime.AvifSpeed = service.DefaultAVIFSpeed
	}
	settingsManager.Reconfigure(runtime)
	cacheStore := newHandlerFakeCache()
	var counter atomic.Int64
	imageService := service.NewImageServiceWithCaches(repo, cacheStore, nil, cacheStore, nil, manager, settingsManager, func() (string, error) {
		return "uid-url-" + strconv.FormatInt(counter.Add(1), 10), nil
	}, func(uid string) error {
		if !strings.HasPrefix(uid, "uid-url-") {
			return errors.New("invalid uid")
		}
		return nil
	}, slog.New(slog.NewTextHandler(discardWriter{}, nil)))
	imageService.SetRemoteImageFetcher(&service.RemoteImageFetcher{Client: publicHostTestClient(t, remote.Listener.Addr().String()), MaxRedirects: 5})
	return NewImageHandler(imageService, slog.New(slog.NewTextHandler(discardWriter{}, nil)), nil), "http://203.0.113.10"
}

func publicHostTestClient(t *testing.T, target string) *http.Client {
	t.Helper()
	dialer := &net.Dialer{}
	transport := http.DefaultTransport.(*http.Transport).Clone()
	transport.DialContext = func(ctx context.Context, network string, _ string) (net.Conn, error) {
		return dialer.DialContext(ctx, network, target)
	}
	return &http.Client{Transport: transport}
}

var _ io.Reader = bytes.NewReader(nil)
