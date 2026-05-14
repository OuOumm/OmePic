package handler

import (
	"errors"
	"log/slog"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"

	"omepic/backend/internal/http/clientip"
	"omepic/backend/internal/response"
	"omepic/backend/internal/service"
)

type ImageHandler struct {
	service    *service.ImageService
	logger     *slog.Logger
	ipResolver *clientip.Resolver
}

func NewImageHandler(imageService *service.ImageService, logger *slog.Logger, ipResolver *clientip.Resolver) *ImageHandler {
	return &ImageHandler{
		service:    imageService,
		logger:     logger,
		ipResolver: ipResolver,
	}
}

func (h *ImageHandler) Upload(c *gin.Context) {
	limit := h.service.MaxUploadSizeBytes()
	fileHeader, err := c.FormFile("file")
	if err != nil {
		response.Error(c, http.StatusBadRequest, "invalid_input", "file is required")
		return
	}
	if limit > 0 && fileHeader.Size > limit {
		response.Error(c, http.StatusBadRequest, "invalid_input", "file exceeds the configured upload size limit")
		return
	}

	file, err := fileHeader.Open()
	if err != nil {
		response.Error(c, http.StatusBadRequest, "invalid_input", "failed to read uploaded file")
		return
	}
	defer file.Close()

	contentType := fileHeader.Header.Get("Content-Type")
	if contentType == "" {
		contentType = detectContentType(fileHeader.Filename)
	}

	result, err := h.service.Upload(c.Request.Context(), service.UploadInput{
		Token:            c.GetHeader("X-Token"),
		OriginalFilename: fileHeader.Filename,
		MIMEType:         contentType,
		IPAddress:        h.clientIP(c),
		Source:           file,
		DeclaredSize:     fileHeader.Size,
		BaseURL:          h.service.EffectivePublicBaseURL(h.requestBaseURL(c)),
		StorageKey:       c.PostForm("storage_key"),
	})
	if err != nil {
		h.mapJSONError(c, err)
		return
	}

	h.logger.Info("image uploaded", "uid", result.UID, "size", result.Size, "mime_type", result.MIMEType, "duplicate", result.Duplicate, "storage_key", result.StorageKey, "storage_backend", result.StorageBackend)
	response.Success(c, http.StatusOK, result)
}

func (h *ImageHandler) RuntimeSettings(c *gin.Context) {
	settings, err := h.service.PublicRuntimeSettings(c.Request.Context())
	if err != nil {
		h.mapJSONError(c, err)
		return
	}
	response.Success(c, http.StatusOK, settings)
}

func (h *ImageHandler) Delete(c *gin.Context) {
	err := h.service.Delete(c.Request.Context(), c.Param("uid"), c.GetHeader("X-Token"), false, h.clientIP(c))
	if err != nil {
		h.mapJSONError(c, err)
		return
	}
	response.Success(c, http.StatusOK, gin.H{})
}

func (h *ImageHandler) Serve(c *gin.Context) {
	result, err := h.service.Open(c.Request.Context(), c.Param("uid"))
	if err != nil {
		switch {
		case errors.Is(err, service.ErrNotFound):
			c.Status(http.StatusNotFound)
		default:
			h.logger.Error("image open failed", "error", err.Error(), "uid", c.Param("uid"))
			c.Status(http.StatusServiceUnavailable)
		}
		return
	}
	defer result.Reader.Close()

	c.Header("Content-Type", result.MIMEType)
	c.Header("Cache-Control", "public, max-age=31536000, immutable")
	c.Header("Content-Disposition", result.ContentDisposition)
	c.DataFromReader(http.StatusOK, result.Size, result.MIMEType, result.Reader, nil)
}

func (h *ImageHandler) mapJSONError(c *gin.Context, err error) {
	if errors.Is(err, service.ErrNotFound) {
		response.Error(c, http.StatusNotFound, "not_found", service.UserMessage(err, "image not found"))
		return
	}
	writeServiceError(c, h.logger, err, "dependency failure", "unexpected image handler error", map[error]serviceErrorMapping{
		service.ErrMissingToken: {
			Status:  http.StatusUnauthorized,
			Code:    "missing_token",
			Message: "X-Token is required",
		},
		service.ErrIPBanned: {
			Status:  http.StatusForbidden,
			Code:    "ip_banned",
			Message: "current network is not allowed to upload or delete images",
		},
		service.ErrForbidden: {
			Status:  http.StatusForbidden,
			Code:    "forbidden",
			Message: "token does not own this image",
		},
	})
}

func (h *ImageHandler) requestBaseURL(c *gin.Context) string {
	scheme := "http"
	if c.Request.TLS != nil {
		scheme = "https"
	}
	return scheme + "://" + c.Request.Host
}

func (h *ImageHandler) clientIP(c *gin.Context) string {
	if h.ipResolver == nil {
		return c.ClientIP()
	}
	return h.ipResolver.Resolve(c.Request)
}

func detectContentType(filename string) string {
	switch strings.ToLower(filepath.Ext(filename)) {
	case ".avif":
		return "image/avif"
	case ".png":
		return "image/png"
	case ".jpg", ".jpeg":
		return "image/jpeg"
	case ".gif":
		return "image/gif"
	case ".webp":
		return "image/webp"
	case ".bmp":
		return "image/bmp"
	default:
		return "application/octet-stream"
	}
}
