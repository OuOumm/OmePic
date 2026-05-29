package middleware

import (
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"

	"omepic/backend/internal/auth"
	"omepic/backend/internal/response"
)

func AdminAuth(jwtSecret string, revChecker *auth.RevocationChecker) gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		token, err := auth.ParseBearer(header)
		if err != nil {
			response.Error(c, http.StatusUnauthorized, "invalid_admin_token", "missing or invalid admin token")
			c.Abort()
			return
		}
		claims, err := auth.ParseJWT(jwtSecret, token)
		if err != nil {
			response.Error(c, http.StatusUnauthorized, "invalid_admin_token", "missing or invalid admin token")
			c.Abort()
			return
		}
		if revChecker != nil && claims.IssuedAt != nil {
			revoked, revErr := revChecker.IsRevoked(c.Request.Context(), claims.IssuedAt.Time)
			if revErr != nil {
				// Log but don't block — fail-open for Redis outage.
				// Use the default slog logger since middleware has no injected logger.
				logger := slog.Default()
				logger.Warn("JWT revocation check: Redis lookup failed", "error", revErr.Error())
			}
			if revoked {
				response.Error(c, http.StatusUnauthorized, "token_revoked", "admin token has been revoked; please log in again")
				c.Abort()
				return
			}
		}
		c.Next()
	}
}
