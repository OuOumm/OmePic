package middleware

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"

	"omepic/backend/internal/ratelimit"
)

// errorLimiter always returns an error to simulate Redis unavailability.
type errorLimiter struct{}

func (errorLimiter) Allow(_ context.Context, _ string, _ int, _ time.Duration) (ratelimit.Result, error) {
	return ratelimit.Result{}, context.DeadlineExceeded
}

func setupRateLimitTest(failClosed bool) (*httptest.ResponseRecorder, *gin.Context) {
	gin.SetMode(gin.TestMode)
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/test", nil)
	ctx, engine := gin.CreateTestContext(recorder)
	ctx.Request = request

	policy := RateLimitPolicy{
		Scope:      "test",
		FailClosed: failClosed,
		LimitFunc:  func() (int, time.Duration) { return 10, time.Minute },
	}
	engine.Use(RateLimit(errorLimiter{}, slog.New(slog.NewTextHandler(io.Discard, nil)), policy))
	engine.POST("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})
	engine.ServeHTTP(recorder, request)
	return recorder, ctx
}

func TestRateLimitFailClosedRejectsOnRedisError(t *testing.T) {
	recorder, _ := setupRateLimitTest(true)

	if recorder.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected status %d, got %d", http.StatusServiceUnavailable, recorder.Code)
	}
	var body struct {
		Success bool `json:"success"`
		Error   struct {
			Code    string `json:"code"`
			Message string `json:"message"`
		} `json:"error"`
	}
	if err := json.Unmarshal(recorder.Body.Bytes(), &body); err != nil {
		t.Fatalf("failed to parse response body: %v", err)
	}
	if body.Success {
		t.Fatal("expected success=false in error response")
	}
	if body.Error.Code != "dependency_unavailable" {
		t.Fatalf("expected error code 'dependency_unavailable', got %q", body.Error.Code)
	}
}

func TestRateLimitFailOpenContinuesOnRedisError(t *testing.T) {
	recorder, _ := setupRateLimitTest(false)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status %d (fail-open), got %d", http.StatusOK, recorder.Code)
	}
}

func TestRateLimitNilLimiterPassesThrough(t *testing.T) {
	gin.SetMode(gin.TestMode)
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/test", nil)
	_, engine := gin.CreateTestContext(recorder)

	policy := RateLimitPolicy{
		Scope:      "test",
		FailClosed: true,
		LimitFunc:  func() (int, time.Duration) { return 10, time.Minute },
	}
	engine.Use(RateLimit(nil, nil, policy))
	engine.POST("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})
	engine.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status %d when limiter is nil, got %d", http.StatusOK, recorder.Code)
	}
}

func TestRateLimitZeroLimitPassesThrough(t *testing.T) {
	gin.SetMode(gin.TestMode)
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/test", nil)
	_, engine := gin.CreateTestContext(recorder)

	policy := RateLimitPolicy{
		Scope:      "test",
		FailClosed: true,
		LimitFunc:  func() (int, time.Duration) { return 0, 0 },
	}
	engine.Use(RateLimit(errorLimiter{}, nil, policy))
	engine.POST("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})
	engine.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status %d when limit is zero, got %d", http.StatusOK, recorder.Code)
	}
}
