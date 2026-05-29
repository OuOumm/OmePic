package middleware

import (
	"strings"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"

	"omepic/backend/internal/service"
)

// PublicCORS returns a gin middleware that applies CORS headers for public API
// routes only. Allowed origins are resolved from runtime settings on each
// request, supporting hot-updated configuration.
//
// Admin routes must NOT use this middleware — they enforce same-origin by
// default (browser blocks cross-origin fetch when no CORS headers are present).
func PublicCORS(settings *service.RuntimeSettingsManager) gin.HandlerFunc {
	config := cors.Config{
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Authorization", "Content-Type", "X-Token"},
		ExposeHeaders:    []string{"Retry-After", "X-RateLimit-Limit", "X-RateLimit-Remaining"},
		MaxAge:           12 * time.Hour,
		AllowCredentials: false,
	}

	// Dynamically resolve allowed origins from runtime settings on each request.
	// This ensures hot-updated PublicBaseURL takes effect immediately.
	config.AllowOriginFunc = func(origin string) bool {
		if settings == nil {
			// Development fallback: allow all origins when no settings manager.
			return true
		}
		s := settings.Current()
		if s.PublicBaseURL == "" {
			// No public URL configured: allow all origins (development mode).
			return true
		}
		// In production, only allow the configured public base URL.
		allowed := strings.TrimRight(s.PublicBaseURL, "/")
		return origin == allowed
	}

	return cors.New(config)
}