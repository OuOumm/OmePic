package middleware

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"omepic/backend/internal/response"
	"omepic/backend/internal/service"
)

// BodyLimit wraps the request body with http.MaxBytesReader so oversized uploads
// are rejected early at the HTTP layer, before the handler reads the body.
// The limit is MaxUploadSizeBytes + 1 MiB to account for multipart overhead.
func BodyLimit(settings *service.RuntimeSettingsManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		if settings == nil {
			c.Next()
			return
		}
		max := settings.Current().MaxUploadSizeBytes()
		if max > 0 {
			limit := max + (1 << 20) // +1 MiB for multipart framing
			c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, limit)
		}
		c.Next()
	}
}

// RespondMaxBytesError writes a 413 response if err is a *http.MaxBytesError.
// Returns true when the error was handled.  Handlers can call this after
// FormFile / ReadBody failures to surface a precise status code.
func RespondMaxBytesError(c *gin.Context, err error) bool {
	var maxErr *http.MaxBytesError
	if errors.As(err, &maxErr) {
		response.Error(c, http.StatusRequestEntityTooLarge, "file_too_large", "uploaded file exceeds the maximum allowed size")
		c.Abort()
		return true
	}
	return false
}
