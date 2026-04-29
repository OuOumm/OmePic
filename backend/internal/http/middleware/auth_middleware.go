package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"omepic/backend/internal/auth"
	"omepic/backend/internal/response"
)

func AdminAuth(jwtSecret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		token, err := auth.ParseBearer(header)
		if err != nil {
			response.Error(c, http.StatusUnauthorized, "invalid_admin_token", "missing or invalid admin token")
			c.Abort()
			return
		}
		if _, err := auth.ParseJWT(jwtSecret, token); err != nil {
			response.Error(c, http.StatusUnauthorized, "invalid_admin_token", "missing or invalid admin token")
			c.Abort()
			return
		}
		c.Next()
	}
}
