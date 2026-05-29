package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"
)

const (
	// CSP without unsafe-inline; inline scripts/styles will be blocked.
	// If the frontend needs inline scripts (e.g. SvelteKit hydrate), add nonce support later.
	frontendCSP = "default-src 'self'; script-src 'self'; style-src 'self'; img-src 'self' data: blob: https:; font-src 'self' data:; connect-src 'self' http: https:; object-src 'none'; base-uri 'self'; frame-ancestors 'none'; form-action 'self'"
)

// SecurityHeaders sets response headers based on route type:
//   - All responses:    X-Content-Type-Options, Referrer-Policy
//   - Frontend HTML:    CSP (without unsafe-inline), X-Frame-Options: DENY
//   - API (/v1/*):      Cache-Control: no-store
//   - Images (/i/*):    no Cache-Control override (images have their own caching)
func SecurityHeaders() gin.HandlerFunc {
	return func(c *gin.Context) {
		path := c.Request.URL.Path

		// Base headers for all responses
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")

		if strings.HasPrefix(path, "/v1/") {
			// API JSON responses: no caching
			c.Header("Cache-Control", "no-store")
		} else if strings.HasPrefix(path, "/i/") {
			// Image responses: keep existing long Cache-Control strategy
			// (set by image handler; don't override)
		} else {
			// Frontend HTML pages: CSP + X-Frame-Options
			c.Header("Content-Security-Policy", frontendCSP)
			c.Header("X-Frame-Options", "DENY")
		}

		c.Next()
	}
}