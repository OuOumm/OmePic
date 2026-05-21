package service

import (
	"context"
	"encoding/json"
	"errors"
	"image/color"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
)

type recordingImageURLCachePurger struct {
	mu         sync.Mutex
	urlBatches [][]string
	err        error
}

func (p *recordingImageURLCachePurger) PurgeURL(ctx context.Context, rawURL string) error {
	return p.PurgeURLs(ctx, []string{rawURL})
}

func (p *recordingImageURLCachePurger) PurgeURLs(_ context.Context, rawURLs []string) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.urlBatches = append(p.urlBatches, append([]string(nil), rawURLs...))
	return p.err
}

func (p *recordingImageURLCachePurger) recordedURLs() []string {
	p.mu.Lock()
	defer p.mu.Unlock()
	var urls []string
	for _, batch := range p.urlBatches {
		urls = append(urls, batch...)
	}
	return urls
}

func (p *recordingImageURLCachePurger) recordedBatches() [][]string {
	p.mu.Lock()
	defer p.mu.Unlock()
	batches := make([][]string, 0, len(p.urlBatches))
	for _, batch := range p.urlBatches {
		batches = append(batches, append([]string(nil), batch...))
	}
	return batches
}

func TestDeletePurgesCloudflareSingleImageURLWhenEnabled(t *testing.T) {
	ctx := context.Background()
	imageService, _, _, _, uidCodec := newImageServiceTestHarness(t)
	uidCodec.Queue("uid-cf")

	settings := defaultRuntimeSettings()
	settings.PublicBaseURL = "https://img.example.com/"
	settings.CloudflarePurgeEnabled = true
	imageService.settings.Reconfigure(settings)
	purger := &recordingImageURLCachePurger{}
	imageService.SetImageURLCachePurger(purger)

	if _, err := imageService.Upload(ctx, UploadInput{
		Token:            "owner-token",
		OriginalFilename: "sample.png",
		MIMEType:         "image/png",
		Bytes:            mustPNGBytes(t, color.RGBA{R: 20, G: 40, B: 60, A: 255}),
		BaseURL:          "http://localhost:8080",
	}); err != nil {
		t.Fatalf("Upload returned error: %v", err)
	}

	if err := imageService.Delete(ctx, "uid-cf"+publicImageExtension, "owner-token", false, ""); err != nil {
		t.Fatalf("Delete returned error: %v", err)
	}

	urls := purger.recordedURLs()
	if len(urls) != 1 || urls[0] != "https://img.example.com/i/uid-cf"+publicImageExtension {
		t.Fatalf("expected one Cloudflare purge URL for the deleted image, got %+v", urls)
	}
}

func TestDeleteStopsBeforeRecordRemovalWhenCloudflarePurgeFails(t *testing.T) {
	ctx := context.Background()
	imageService, repo, _, _, uidCodec := newImageServiceTestHarness(t)
	uidCodec.Queue("uid-cf-fail")

	settings := defaultRuntimeSettings()
	settings.PublicBaseURL = "https://img.example.com"
	settings.CloudflarePurgeEnabled = true
	imageService.settings.Reconfigure(settings)
	imageService.SetImageURLCachePurger(&recordingImageURLCachePurger{err: errors.New("cloudflare down")})

	if _, err := imageService.Upload(ctx, UploadInput{
		Token:            "owner-token",
		OriginalFilename: "sample.png",
		MIMEType:         "image/png",
		Bytes:            mustPNGBytes(t, color.RGBA{R: 80, G: 40, B: 20, A: 255}),
		BaseURL:          "http://localhost:8080",
	}); err != nil {
		t.Fatalf("Upload returned error: %v", err)
	}

	if err := imageService.Delete(ctx, "uid-cf-fail"+publicImageExtension, "owner-token", false, ""); err == nil {
		t.Fatalf("expected Delete to fail when Cloudflare purge fails")
	}
	if _, err := repo.FindByUID(ctx, "uid-cf-fail"); err != nil {
		t.Fatalf("expected record to remain after purge failure, got %v", err)
	}
}

func TestAdminDeleteImagesPurgesCloudflareURLsInOneBatchBeforeDeleting(t *testing.T) {
	ctx := context.Background()
	adminService, repo := newAdminServiceTestHarness(t)
	adminService.imageService.validateUID = func(string) error { return nil }

	settings := defaultRuntimeSettings()
	settings.PublicBaseURL = "https://img.example.com/"
	settings.CloudflarePurgeEnabled = true
	adminService.settings.Reconfigure(settings)
	purger := &recordingImageURLCachePurger{}
	adminService.imageService.SetImageURLCachePurger(purger)

	for _, uid := range []string{"uid-batch-1", "uid-batch-2"} {
		if err := repo.InsertImage(ctx, modelImageRecord(uid, "local-default", "local")); err != nil {
			t.Fatalf("InsertImage %s returned error: %v", uid, err)
		}
	}

	if err := adminService.DeleteImages(ctx, []string{"uid-batch-1", "uid-batch-2"}); err != nil {
		t.Fatalf("DeleteImages returned error: %v", err)
	}

	batches := purger.recordedBatches()
	if len(batches) != 1 {
		t.Fatalf("expected one Cloudflare purge request, got %+v", batches)
	}
	want := []string{
		"https://img.example.com/i/uid-batch-1" + publicImageExtension,
		"https://img.example.com/i/uid-batch-2" + publicImageExtension,
	}
	if len(batches[0]) != len(want) {
		t.Fatalf("expected purge files %+v, got %+v", want, batches[0])
	}
	for i := range want {
		if batches[0][i] != want[i] {
			t.Fatalf("expected purge files %+v, got %+v", want, batches[0])
		}
	}
	for _, uid := range []string{"uid-batch-1", "uid-batch-2"} {
		if _, err := repo.FindByUID(ctx, uid); err == nil {
			t.Fatalf("expected %s to be deleted after successful purge", uid)
		}
	}
}

func TestAdminDeleteImagesStopsBeforeRecordRemovalWhenBatchCloudflarePurgeFails(t *testing.T) {
	ctx := context.Background()
	adminService, repo := newAdminServiceTestHarness(t)
	adminService.imageService.validateUID = func(string) error { return nil }

	settings := defaultRuntimeSettings()
	settings.PublicBaseURL = "https://img.example.com"
	settings.CloudflarePurgeEnabled = true
	adminService.settings.Reconfigure(settings)
	adminService.imageService.SetImageURLCachePurger(&recordingImageURLCachePurger{err: errors.New("cloudflare down")})

	for _, uid := range []string{"uid-batch-fail-1", "uid-batch-fail-2"} {
		if err := repo.InsertImage(ctx, modelImageRecord(uid, "local-default", "local")); err != nil {
			t.Fatalf("InsertImage %s returned error: %v", uid, err)
		}
	}

	if err := adminService.DeleteImages(ctx, []string{"uid-batch-fail-1", "uid-batch-fail-2"}); err == nil {
		t.Fatalf("expected DeleteImages to fail when Cloudflare purge fails")
	}
	for _, uid := range []string{"uid-batch-fail-1", "uid-batch-fail-2"} {
		if _, err := repo.FindByUID(ctx, uid); err != nil {
			t.Fatalf("expected %s to remain after purge failure, got %v", uid, err)
		}
	}
}

func TestAdminPurgeCloudflareImageCachePurgesSingleURL(t *testing.T) {
	ctx := context.Background()
	adminService, _ := newAdminServiceTestHarness(t)
	settings := defaultRuntimeSettings()
	settings.PublicBaseURL = "https://img.example.com"
	settings.CloudflarePurgeEnabled = true
	adminService.settings.Reconfigure(settings)
	purger := &recordingImageURLCachePurger{}
	adminService.imageService.SetImageURLCachePurger(purger)

	result, err := adminService.PurgeCloudflareImageCache(ctx, " https://img.example.com/i/manual.avif#frag ")
	if err != nil {
		t.Fatalf("PurgeCloudflareImageCache returned error: %v", err)
	}
	if result.URL != "https://img.example.com/i/manual.avif" {
		t.Fatalf("expected normalized result URL, got %q", result.URL)
	}
	urls := purger.recordedURLs()
	if len(urls) != 1 || urls[0] != result.URL {
		t.Fatalf("expected manual purge to call purger once with %q, got %+v", result.URL, urls)
	}
}

func TestAdminPurgeCloudflareImageCacheUsesRuntimeConfigHotReload(t *testing.T) {
	ctx := context.Background()
	adminService, _ := newAdminServiceTestHarness(t)
	var requests []struct{ Path, Auth string }
	var mu sync.Mutex
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		requests = append(requests, struct{ Path, Auth string }{Path: r.URL.Path, Auth: r.Header.Get("Authorization")})
		mu.Unlock()
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]bool{"success": true})
	}))
	defer server.Close()

	settings := defaultRuntimeSettings()
	settings.PublicBaseURL = "https://img.example.com"
	settings.CloudflarePurgeEnabled = true
	settings.CloudflareZoneID = "zone-a"
	settings.CloudflareAPIToken = "token-a"
	settings.CloudflareAPIBaseURL = server.URL + "/"
	adminService.settings.Reconfigure(settings)

	if _, err := adminService.PurgeCloudflareImageCache(ctx, "https://img.example.com/i/a.avif"); err != nil {
		t.Fatalf("first purge returned error: %v", err)
	}
	settings.CloudflareZoneID = "zone-b"
	settings.CloudflareAPIToken = "token-b"
	adminService.settings.Reconfigure(settings)
	if _, err := adminService.PurgeCloudflareImageCache(ctx, "https://img.example.com/i/b.avif"); err != nil {
		t.Fatalf("second purge returned error: %v", err)
	}

	mu.Lock()
	defer mu.Unlock()
	if len(requests) != 2 {
		t.Fatalf("expected two Cloudflare requests, got %+v", requests)
	}
	if requests[0].Path != "/zones/zone-a/purge_cache" || requests[0].Auth != "Bearer token-a" {
		t.Fatalf("first purge did not use initial runtime config: %+v", requests[0])
	}
	if requests[1].Path != "/zones/zone-b/purge_cache" || requests[1].Auth != "Bearer token-b" {
		t.Fatalf("second purge did not use hot-reloaded runtime config: %+v", requests[1])
	}
}

func TestDeletePurgesCloudflareWithRuntimeConfigHotReload(t *testing.T) {
	ctx := context.Background()
	imageService, _, _, _, uidCodec := newImageServiceTestHarness(t)
	uidCodec.Queue("uid-hot-a", "uid-hot-b")
	var paths []string
	var mu sync.Mutex
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		paths = append(paths, r.URL.Path+"|"+r.Header.Get("Authorization"))
		mu.Unlock()
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]bool{"success": true})
	}))
	defer server.Close()

	settings := defaultRuntimeSettings()
	settings.PublicBaseURL = "https://img.example.com"
	settings.CloudflarePurgeEnabled = true
	settings.CloudflareZoneID = "zone-a"
	settings.CloudflareAPIToken = "token-a"
	settings.CloudflareAPIBaseURL = server.URL
	imageService.settings.Reconfigure(settings)

	for _, uid := range []string{"uid-hot-a", "uid-hot-b"} {
		if _, err := imageService.Upload(ctx, UploadInput{Token: "owner-token", OriginalFilename: uid + ".png", MIMEType: "image/png", Bytes: mustPNGBytes(t, color.RGBA{R: 20, G: 40, B: 60, A: 255}), BaseURL: "http://localhost:8080"}); err != nil {
			t.Fatalf("Upload %s returned error: %v", uid, err)
		}
	}
	if err := imageService.Delete(ctx, "uid-hot-a"+publicImageExtension, "owner-token", false, ""); err != nil {
		t.Fatalf("first Delete returned error: %v", err)
	}
	settings.CloudflareZoneID = "zone-b"
	settings.CloudflareAPIToken = "token-b"
	imageService.settings.Reconfigure(settings)
	if err := imageService.Delete(ctx, "uid-hot-b"+publicImageExtension, "owner-token", false, ""); err != nil {
		t.Fatalf("second Delete returned error: %v", err)
	}

	mu.Lock()
	defer mu.Unlock()
	if len(paths) != 2 {
		t.Fatalf("expected two purge requests, got %+v", paths)
	}
	if paths[0] != "/zones/zone-a/purge_cache|Bearer token-a" || paths[1] != "/zones/zone-b/purge_cache|Bearer token-b" {
		t.Fatalf("delete purge did not use hot-reloaded runtime config: %+v", paths)
	}
}

func TestValidateRuntimeSettingsInputRequiresPublicBaseURLAndCredentialsForCloudflarePurge(t *testing.T) {
	input := RuntimeSettingsUpdateInput(defaultRuntimeSettings())
	input.CloudflarePurgeEnabled = true
	input.PublicBaseURL = ""
	if _, err := ValidateRuntimeSettingsInput(input); err == nil || !containsError(err, ErrInvalidInput) {
		t.Fatalf("expected ErrInvalidInput when Cloudflare purge is enabled without public_base_url, got %v", err)
	}

	input.PublicBaseURL = "https://img.example.com"
	if _, err := ValidateRuntimeSettingsInput(input); err == nil || !containsError(err, ErrInvalidInput) {
		t.Fatalf("expected ErrInvalidInput when Cloudflare purge is enabled without credentials, got %v", err)
	}

	input.CloudflareZoneID = "zone-123"
	input.CloudflareAPIToken = "token-secret"
	if _, err := ValidateRuntimeSettingsInput(input); err != nil {
		t.Fatalf("expected valid Cloudflare purge settings with public_base_url and credentials, got %v", err)
	}
}
