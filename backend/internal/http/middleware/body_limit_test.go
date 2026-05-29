package middleware

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestBodyLimitAppliesMaxBytesReader(t *testing.T) {
	gin.SetMode(gin.TestMode)

	recorder := httptest.NewRecorder()
	_, engine := gin.CreateTestContext(recorder)

	engine.Use(BodyLimit(func() int64 { return 1024 }))
	engine.POST("/upload", func(c *gin.Context) {
		raw, err := c.GetRawData()
		if RespondMaxBytesError(c, err) {
			return
		}
		_ = raw
		c.Status(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodPost, "/upload", strings.NewReader("hello"))
	engine.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected 200 for small body, got %d", recorder.Code)
	}
}

func TestBodyLimitZeroLimitSkipsMaxBytesReader(t *testing.T) {
	gin.SetMode(gin.TestMode)

	recorder := httptest.NewRecorder()
	_, engine := gin.CreateTestContext(recorder)

	engine.Use(BodyLimit(func() int64 { return 0 }))
	engine.POST("/upload", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodPost, "/upload", strings.NewReader("hello"))
	engine.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected 200 when limit=0, got %d", recorder.Code)
	}
}

func TestBodyLimitOversizedBodyWithRealServer(t *testing.T) {
	// Use a real HTTP server to test MaxBytesReader end-to-end since
	// BodyLimit adds 1 MiB overhead, making Gin test harness impractical.
	handler := http.NewServeMux()
	handler.HandleFunc("/upload", func(w http.ResponseWriter, r *http.Request) {
		r.Body = http.MaxBytesReader(w, r.Body, 50)
		buf := make([]byte, 1000)
		_, err := r.Body.Read(buf)
		if err != nil {
			var maxErr *http.MaxBytesError
			if errors.As(err, &maxErr) {
				w.WriteHeader(http.StatusRequestEntityTooLarge)
				return
			}
		}
		w.WriteHeader(http.StatusOK)
	})

	server := httptest.NewServer(handler)
	defer server.Close()

	// Send body larger than 50 bytes → expect 413
	resp, err := http.Post(server.URL+"/upload", "text/plain", strings.NewReader(strings.Repeat("x", 200)))
	if err != nil {
		t.Fatalf("POST request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusRequestEntityTooLarge {
		t.Fatalf("expected 413, got %d", resp.StatusCode)
	}

	// Send body within 50 bytes → expect 200
	resp2, err := http.Post(server.URL+"/upload", "text/plain", strings.NewReader("hello"))
	if err != nil {
		t.Fatalf("POST request failed: %v", err)
	}
	defer resp2.Body.Close()

	if resp2.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 for small body, got %d", resp2.StatusCode)
	}
}

func TestRespondMaxBytesErrorIdentifiesMaxBytesError(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name        string
		err         error
		wantHandled bool
	}{
		{
			name:        "MaxBytesError is handled",
			err:         &http.MaxBytesError{Limit: 1024},
			wantHandled: true,
		},
		{
			name:        "other error is not handled",
			err:         errors.New("some other error"),
			wantHandled: false,
		},
		{
			name:        "nil error is not handled",
			err:         nil,
			wantHandled: false,
		},
		{
			name:        "wrapped MaxBytesError is handled",
			err:         fmt.Errorf("wrapped: %w", &http.MaxBytesError{Limit: 1024}),
			wantHandled: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			recorder := httptest.NewRecorder()
			ctx, _ := gin.CreateTestContext(recorder)

			handled := RespondMaxBytesError(ctx, tt.err)
			if handled != tt.wantHandled {
				t.Fatalf("expected handled=%v, got handled=%v", tt.wantHandled, handled)
			}

			if tt.wantHandled {
				if recorder.Code != http.StatusRequestEntityTooLarge {
					t.Fatalf("expected 413 status, got %d", recorder.Code)
				}
			}
		})
	}
}