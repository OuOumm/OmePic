package service

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCloudflareCachePurgerPurgesSingleURLWithBearerToken(t *testing.T) {
	ctx := context.Background()
	var capturedPath string
	var capturedAuth string
	var capturedFiles []string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedPath = r.URL.Path
		capturedAuth = r.Header.Get("Authorization")
		var payload cloudflarePurgeCacheRequest
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			t.Fatalf("failed to decode request payload: %v", err)
		}
		capturedFiles = payload.Files
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"success":true}`))
	}))
	defer server.Close()

	purger := NewCloudflareCachePurger("zone-123", "token-secret", server.URL, server.Client())
	if err := purger.PurgeURL(ctx, " https://img.example.com/i/uid-1.avif#preview "); err != nil {
		t.Fatalf("PurgeURL returned error: %v", err)
	}

	if capturedPath != "/zones/zone-123/purge_cache" {
		t.Fatalf("unexpected cloudflare endpoint path %q", capturedPath)
	}
	if capturedAuth != "Bearer token-secret" {
		t.Fatalf("unexpected authorization header %q", capturedAuth)
	}
	if len(capturedFiles) != 1 || capturedFiles[0] != "https://img.example.com/i/uid-1.avif" {
		t.Fatalf("expected one normalized purge URL, got %+v", capturedFiles)
	}
}

func TestCloudflareCachePurgerPurgesMultipleURLsInOneFilesRequest(t *testing.T) {
	ctx := context.Background()
	requestCount := 0
	var capturedFiles []string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestCount++
		var payload cloudflarePurgeCacheRequest
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			t.Fatalf("failed to decode request payload: %v", err)
		}
		capturedFiles = payload.Files
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"success":true}`))
	}))
	defer server.Close()

	purger := NewCloudflareCachePurger("zone-123", "token-secret", server.URL, server.Client())
	if err := purger.PurgeURLs(ctx, []string{
		"https://img.example.com/i/uid-1.avif#preview",
		" https://img.example.com/i/uid-2.avif ",
	}); err != nil {
		t.Fatalf("PurgeURLs returned error: %v", err)
	}

	if requestCount != 1 {
		t.Fatalf("expected one Cloudflare request, got %d", requestCount)
	}
	want := []string{"https://img.example.com/i/uid-1.avif", "https://img.example.com/i/uid-2.avif"}
	if len(capturedFiles) != len(want) {
		t.Fatalf("expected files %+v, got %+v", want, capturedFiles)
	}
	for i := range want {
		if capturedFiles[i] != want[i] {
			t.Fatalf("expected files %+v, got %+v", want, capturedFiles)
		}
	}
}

func TestCloudflareCachePurgerRejectsInvalidURLBeforeRequest(t *testing.T) {
	purger := NewCloudflareCachePurger("zone-123", "token-secret", "", nil)
	if err := purger.PurgeURL(context.Background(), "javascript:alert(1)"); err == nil || !containsError(err, ErrInvalidInput) {
		t.Fatalf("expected ErrInvalidInput for invalid URL, got %v", err)
	}
	if err := purger.PurgeURLs(context.Background(), nil); err == nil || !containsError(err, ErrInvalidInput) {
		t.Fatalf("expected ErrInvalidInput for empty URL list, got %v", err)
	}
}

func TestCloudflareCachePurgerPropagatesCloudflareFailure(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadGateway)
		_, _ = w.Write([]byte(`{"success":false}`))
	}))
	defer server.Close()

	purger := NewCloudflareCachePurger("zone-123", "token-secret", server.URL, server.Client())
	if err := purger.PurgeURL(context.Background(), "https://img.example.com/i/uid-1.avif"); err == nil || !containsError(err, ErrDependencyUnavailable) {
		t.Fatalf("expected ErrDependencyUnavailable for cloudflare failure, got %v", err)
	}
}
