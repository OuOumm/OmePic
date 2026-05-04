package router

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestFrontendFallbackServesStaticExportPages(t *testing.T) {
	gin.SetMode(gin.TestMode)
	webDir := t.TempDir()
	writeTestFile(t, filepath.Join(webDir, "index.html"), "<!doctype html><title>home</title>")
	writeTestFile(t, filepath.Join(webDir, "history.html"), "<!doctype html><title>history</title>")
	writeTestFile(t, filepath.Join(webDir, "_next", "static", "app.js"), "console.log('ok')")

	engine := gin.New()
	registerFrontendRoutes(engine, webDir, nil)

	tests := []struct {
		path string
		want string
	}{
		{path: "/", want: "<!doctype html><title>home</title>"},
		{path: "/history", want: "<!doctype html><title>history</title>"},
		{path: "/admin/dashboard", want: "<!doctype html><title>home</title>"},
		{path: "/_next/static/app.js", want: "console.log('ok')"},
	}

	for _, test := range tests {
		t.Run(test.path, func(t *testing.T) {
			recorder := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, test.path, nil)

			engine.ServeHTTP(recorder, req)

			if recorder.Code != http.StatusOK {
				t.Fatalf("expected status 200, got %d", recorder.Code)
			}
			if got := recorder.Body.String(); got != test.want {
				t.Fatalf("unexpected response body: got %q want %q", got, test.want)
			}
		})
	}
}

func TestFrontendFallbackReturnsNotFoundForMissingAsset(t *testing.T) {
	gin.SetMode(gin.TestMode)
	webDir := t.TempDir()
	writeTestFile(t, filepath.Join(webDir, "index.html"), "<!doctype html><title>home</title>")

	engine := gin.New()
	registerFrontendRoutes(engine, webDir, nil)

	recorder := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/missing.js", nil)

	engine.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusNotFound {
		t.Fatalf("expected status 404, got %d", recorder.Code)
	}
}

func TestFrontendFallbackPreservesAPINotFoundBehavior(t *testing.T) {
	gin.SetMode(gin.TestMode)
	webDir := t.TempDir()
	writeTestFile(t, filepath.Join(webDir, "index.html"), "<!doctype html><title>home</title>")

	engine := gin.New()
	registerFrontendRoutes(engine, webDir, nil)

	for _, path := range []string{
		"/health",
		"/v1/missing",
		"/i/missing.avif",
		"/admin/status",
		"/admin/images",
		"/admin/config",
	} {
		t.Run(path, func(t *testing.T) {
			recorder := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, path, nil)

			engine.ServeHTTP(recorder, req)

			if recorder.Code != http.StatusNotFound {
				t.Fatalf("expected status 404, got %d", recorder.Code)
			}
			if got := recorder.Body.String(); got == "<!doctype html><title>home</title>" {
				t.Fatalf("api route was shadowed by frontend fallback")
			}
		})
	}
}

func TestFrontendFallbackPreservesAPINotFoundBehaviorByMethod(t *testing.T) {
	gin.SetMode(gin.TestMode)
	webDir := t.TempDir()
	writeTestFile(t, filepath.Join(webDir, "index.html"), "<!doctype html><title>home</title>")
	writeTestFile(t, filepath.Join(webDir, "admin", "login.html"), "<!doctype html><title>login</title>")

	engine := gin.New()
	registerFrontendRoutes(engine, webDir, nil)

	tests := []struct {
		name   string
		method string
		path   string
		status int
		body   string
	}{
		{name: "admin login api post", method: http.MethodPost, path: "/admin/login", status: http.StatusNotFound},
		{name: "admin login page get", method: http.MethodGet, path: "/admin/login", status: http.StatusOK, body: "<!doctype html><title>login</title>"},
		{name: "health api head", method: http.MethodHead, path: "/health", status: http.StatusNotFound},
		{name: "admin status api head", method: http.MethodHead, path: "/admin/status", status: http.StatusNotFound},
		{name: "admin config instances api post", method: http.MethodPost, path: "/admin/config/storage-instances", status: http.StatusNotFound},
		{name: "admin config instances page get", method: http.MethodGet, path: "/admin/config/storage-instances", status: http.StatusOK, body: "<!doctype html><title>home</title>"},
		{name: "admin config instance api put", method: http.MethodPut, path: "/admin/config/storage-instances/local", status: http.StatusNotFound},
		{name: "admin config default api post", method: http.MethodPost, path: "/admin/config/default", status: http.StatusNotFound},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			recorder := httptest.NewRecorder()
			req := httptest.NewRequest(test.method, test.path, nil)

			engine.ServeHTTP(recorder, req)

			if recorder.Code != test.status {
				t.Fatalf("expected status %d, got %d", test.status, recorder.Code)
			}
			if test.body != "" && recorder.Body.String() != test.body {
				t.Fatalf("unexpected response body: got %q want %q", recorder.Body.String(), test.body)
			}
			if test.status == http.StatusNotFound && recorder.Body.String() == "<!doctype html><title>home</title>" {
				t.Fatalf("api route was shadowed by frontend fallback")
			}
		})
	}
}

func TestFrontendFallbackDisabledWhenBuildMissing(t *testing.T) {
	gin.SetMode(gin.TestMode)

	engine := gin.New()
	registerFrontendRoutes(engine, t.TempDir(), nil)

	recorder := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/history", nil)

	engine.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusNotFound {
		t.Fatalf("expected status 404, got %d", recorder.Code)
	}
}

func writeTestFile(t *testing.T, filePath string, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(filePath), 0o755); err != nil {
		t.Fatalf("failed to create parent directory: %v", err)
	}
	if err := os.WriteFile(filePath, []byte(content), 0o644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}
}
