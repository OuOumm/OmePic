package middleware

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"omepic/backend/internal/http/clientip"
	"omepic/backend/internal/ratelimit"
	"omepic/backend/internal/response"
)

type RateLimitPolicy struct {
	Scope      string
	LimitFunc  func() (int, time.Duration)
	IPResolver *clientip.Resolver
}

func RateLimit(limiter ratelimit.Limiter, logger *slog.Logger, policy RateLimitPolicy) gin.HandlerFunc {
	return func(c *gin.Context) {
		if limiter == nil || policy.LimitFunc == nil {
			c.Next()
			return
		}
		limit, window := policy.LimitFunc()
		if limit <= 0 || window <= 0 {
			c.Next()
			return
		}
		key := rateLimitKey(policy.Scope, clientIP(c, policy.IPResolver))
		result, err := limiter.Allow(c.Request.Context(), key, limit, window)
		if err != nil {
			if logger != nil {
				logger.WarnContext(context.Background(), "rate limiter unavailable", "scope", policy.Scope, "error", err.Error())
			}
			c.Next()
			return
		}
		c.Header("X-RateLimit-Limit", strconv.Itoa(result.Limit))
		c.Header("X-RateLimit-Remaining", strconv.Itoa(result.Remaining))
		if !result.Allowed {
			if policy.Scope == "upload" && c.Request.Body != nil {
				_, _ = io.Copy(io.Discard, c.Request.Body)
			}
			retryAfter := int(result.RetryAfter.Seconds())
			if retryAfter < 1 {
				retryAfter = 1
			}
			c.Header("Retry-After", strconv.Itoa(retryAfter))
			response.Error(c, http.StatusTooManyRequests, "rate_limited", "too many requests")
			c.Abort()
			return
		}
		c.Next()
	}
}

func clientIP(c *gin.Context, resolver *clientip.Resolver) string {
	if resolver == nil {
		return c.ClientIP()
	}
	return resolver.Resolve(c.Request)
}

func rateLimitKey(scope string, ip string) string {
	normalizedScope := strings.TrimSpace(scope)
	if normalizedScope == "" {
		normalizedScope = "api"
	}
	sum := sha256.Sum256([]byte(strings.TrimSpace(ip)))
	return fmt.Sprintf("ratelimit:%s:ip:%s", normalizedScope, hex.EncodeToString(sum[:]))
}
