package middleware

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"omepic/backend/internal/http/clientip"
	"omepic/backend/internal/iputil"
	"omepic/backend/internal/ratelimit"
	"omepic/backend/internal/response"
)

type RateLimitPolicy struct {
	Scope      string
	LimitFunc  func() (int, time.Duration)
	IPResolver *clientip.Resolver
	// FailClosed rejects requests with 503 when the rate limiter backend is
	// unavailable. Use this for sensitive endpoints (upload, login) where
	// losing rate-limit protection is unacceptable. Default (false) is
	// fail-open: the request proceeds without rate limiting.
	FailClosed bool
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
			if policy.FailClosed {
				response.Error(c, http.StatusServiceUnavailable, "dependency_unavailable", "rate limiter is temporarily unavailable")
				c.Abort()
				return
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
	return fmt.Sprintf("ratelimit:%s:ip:%s", normalizedScope, iputil.Hash(ip))
}
