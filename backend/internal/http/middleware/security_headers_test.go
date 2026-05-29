package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestSecurityHeaders(t *testing.T) {
	tests := []struct {
		name           string
		path           string
		wantNosniff    bool // X-Content-Type-Options: nosniff
		wantReferrer   bool // Referrer-Policy
		wantCSP        bool // Content-Security-Policy (frontend only)
		wantFrameOpts  bool // X-Frame-Options: DENY (frontend only)
		wantNoStore    bool // Cache-Control: no-store (API only)
		wantNoOverride bool // No Cache-Control override (images only)
	}{
		{
			name:          "frontend path gets CSP and X-Frame-Options",
			path:          "/dashboard",
			wantNosniff:   true,
			wantReferrer:  true,
			wantCSP:       true,
			wantFrameOpts: true,
		},
		{
			name:          "root path gets CSP and X-Frame-Options",
			path:          "/",
			wantNosniff:   true,
			wantReferrer:  true,
			wantCSP:       true,
			wantFrameOpts: true,
		},
		{
			name:          "API path gets Cache-Control no-store but not CSP",
			path:          "/v1/admin/login",
			wantNosniff:   true,
			wantReferrer:  true,
			wantCSP:       false,
			wantFrameOpts: false,
			wantNoStore:   true,
		},
		{
			name:          "image path gets no Cache-Control override",
			path:          "/i/omeo_abc123",
			wantNosniff:   true,
			wantReferrer:  true,
			wantCSP:       false,
			wantFrameOpts: false,
			wantNoOverride: true,
		},
		{
			name:          "API path /v1/ prefix only",
			path:          "/v1/images",
			wantNosniff:   true,
			wantReferrer:  true,
			wantNoStore:   true,
			wantCSP:       false,
			wantFrameOpts: false,
		},
		{
			name:          "static asset path is frontend",
			path:          "/assets/app.js",
			wantNosniff:   true,
			wantReferrer:  true,
			wantCSP:       true,
			wantFrameOpts: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gin.SetMode(gin.TestMode)
			recorder := httptest.NewRecorder()
			_, engine := gin.CreateTestContext(recorder)

			engine.Use(SecurityHeaders())
			engine.Any(tt.path, func(c *gin.Context) {
				c.Status(http.StatusOK)
			})

			req := httptest.NewRequest(http.MethodGet, tt.path, nil)
			engine.ServeHTTP(recorder, req)

			if tt.wantNosniff {
				if got := recorder.Header().Get("X-Content-Type-Options"); got != "nosniff" {
					t.Errorf("expected X-Content-Type-Options nosniff, got %q", got)
				}
			}

			if tt.wantReferrer {
				if got := recorder.Header().Get("Referrer-Policy"); got != "strict-origin-when-cross-origin" {
					t.Errorf("expected Referrer-Policy strict-origin-when-cross-origin, got %q", got)
				}
			}

			if tt.wantCSP {
				if got := recorder.Header().Get("Content-Security-Policy"); got == "" {
					t.Error("expected Content-Security-Policy header for frontend path")
				}
			} else {
				if got := recorder.Header().Get("Content-Security-Policy"); got != "" {
					t.Errorf("expected no CSP for non-frontend path, got %q", got)
				}
			}

			if tt.wantFrameOpts {
				if got := recorder.Header().Get("X-Frame-Options"); got != "DENY" {
					t.Errorf("expected X-Frame-Options DENY, got %q", got)
				}
			} else {
				if got := recorder.Header().Get("X-Frame-Options"); got != "" {
					t.Errorf("expected no X-Frame-Options for non-frontend path, got %q", got)
				}
			}

			if tt.wantNoStore {
				if got := recorder.Header().Get("Cache-Control"); got != "no-store" {
					t.Errorf("expected Cache-Control no-store, got %q", got)
				}
			}

			if tt.wantNoOverride {
				if got := recorder.Header().Get("Cache-Control"); got != "" {
					t.Errorf("expected no Cache-Control override for image path, got %q", got)
				}
			}
		})
	}
}

func TestSecurityHeadersAppliedToAllResponses(t *testing.T) {
	gin.SetMode(gin.TestMode)
	recorder := httptest.NewRecorder()
	_, engine := gin.CreateTestContext(recorder)

	engine.Use(SecurityHeaders())
	engine.GET("/anything", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/anything", nil)
	engine.ServeHTTP(recorder, req)

	if recorder.Header().Get("X-Content-Type-Options") != "nosniff" {
		t.Error("X-Content-Type-Options should be set on all responses")
	}
	if recorder.Header().Get("Referrer-Policy") != "strict-origin-when-cross-origin" {
		t.Error("Referrer-Policy should be set on all responses")
	}
}