package handler

import (
	"io"
	"log/slog"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"

	"omepic/backend/internal/response"
	"omepic/backend/internal/service"
	"omepic/backend/internal/storage"
)

type ImageHandler struct {
	service    *service.ImageService
	storage    *storage.Manager
	logger     *slog.Logger
	publicBase string
}

func NewImageHandler(imageService *service.ImageService, storageManager *storage.Manager, logger *slog.Logger, publicBase string) *ImageHandler {
	return &ImageHandler{
		service:    imageService,
		storage:    storageManager,
		logger:     logger,
		publicBase: publicBase,
	}
}

func (h *ImageHandler) Upload(c *gin.Context) {
	fileHeader, err := c.FormFile("file")
	if err != nil {
		response.Error(c, http.StatusBadRequest, "invalid_input", "file is required")
		return
	}
	if fileHeader.Size > service.MaxUploadSizeBytes() {
		response.Error(c, http.StatusBadRequest, "invalid_input", "file must be 20 MB or smaller")
		return
	}

	file, err := fileHeader.Open()
	if err != nil {
		response.Error(c, http.StatusBadRequest, "invalid_input", "failed to read uploaded file")
		return
	}
	defer file.Close()

	payload, err := io.ReadAll(io.LimitReader(file, service.MaxUploadSizeBytes()+1))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "invalid_input", "failed to read uploaded file")
		return
	}
	if int64(len(payload)) > service.MaxUploadSizeBytes() {
		response.Error(c, http.StatusBadRequest, "invalid_input", "file must be 20 MB or smaller")
		return
	}

	contentType := fileHeader.Header.Get("Content-Type")
	if contentType == "" {
		contentType = detectContentType(fileHeader.Filename)
	}

	result, err := h.service.Upload(c.Request.Context(), service.UploadInput{
		Token:            c.GetHeader("X-Token"),
		OriginalFilename: fileHeader.Filename,
		MIMEType:         contentType,
		IPAddress:        c.ClientIP(),
		Bytes:            payload,
		BaseURL:          h.publicURL(c),
		StorageKey:       c.PostForm("storage_key"),
	})
	if err != nil {
		h.mapJSONError(c, err)
		return
	}

	h.logger.Info("image uploaded", "uid", result.UID, "size", result.Size, "mime_type", result.MIMEType, "duplicate", result.Duplicate, "storage_key", result.StorageKey, "storage_backend", result.StorageBackend)
	response.Success(c, http.StatusOK, result)
}

func (h *ImageHandler) StorageOptions(c *gin.Context) {
	options, err := h.service.PublicStorageOptions(c.Request.Context())
	if err != nil {
		h.mapJSONError(c, err)
		return
	}
	response.Success(c, http.StatusOK, gin.H{"items": options})
}

func (h *ImageHandler) Delete(c *gin.Context) {
	err := h.service.Delete(c.Request.Context(), c.Param("uid"), c.GetHeader("X-Token"), false)
	if err != nil {
		h.mapJSONError(c, err)
		return
	}
	response.Success(c, http.StatusOK, gin.H{})
}

func (h *ImageHandler) Serve(c *gin.Context) {
	result, err := h.service.Resolve(c.Request.Context(), c.Param("uid"))
	if err != nil {
		switch {
		case err == service.ErrNotFound:
			c.Status(http.StatusNotFound)
		default:
			h.logger.Error("image resolve failed", "error", err.Error(), "uid", c.Param("uid"))
			c.Status(http.StatusServiceUnavailable)
		}
		return
	}

	resolved, providerErr := h.storage.ForKey(result.Record.StorageKey)
	if providerErr != nil {
		h.logger.Error("storage backend resolution failed", "error", providerErr.Error(), "uid", result.Record.UID, "storage_key", result.Record.StorageKey, "storage_backend", result.Record.StorageBackend)
		c.Status(http.StatusServiceUnavailable)
		return
	}

	file, err := resolved.Provider.Open(c.Request.Context(), result.Record.FilePath)
	if err != nil {
		h.logger.Error("image open failed", "error", err.Error(), "uid", result.Record.UID, "storage_key", result.Record.StorageKey, "storage_backend", result.Record.StorageBackend)
		c.Status(http.StatusServiceUnavailable)
		return
	}
	defer file.Reader.Close()

	c.Header("Content-Type", result.Record.MIMEType)
	c.Header("Cache-Control", "public, max-age=31536000, immutable")
	c.Header("Content-Disposition", "inline; filename=\""+filepath.Base(result.Record.FilePath)+"\"")
	c.DataFromReader(http.StatusOK, file.Size, result.Record.MIMEType, file.Reader, nil)
}

func (h *ImageHandler) mapJSONError(c *gin.Context, err error) {
	switch {
	case err == service.ErrMissingToken:
		response.Error(c, http.StatusUnauthorized, "missing_token", "X-Token is required")
	case strings.Contains(err.Error(), service.ErrInvalidInput.Error()):
		response.Error(c, http.StatusBadRequest, "invalid_input", sanitizeMessage(err))
	case err == service.ErrForbidden:
		response.Error(c, http.StatusForbidden, "forbidden", "token does not own this image")
	case err == service.ErrNotFound:
		response.Error(c, http.StatusNotFound, "not_found", "image not found")
	case strings.Contains(err.Error(), service.ErrNotFound.Error()):
		response.Error(c, http.StatusNotFound, "not_found", sanitizeMessage(err))
	case strings.Contains(err.Error(), service.ErrDependencyUnavailable.Error()):
		h.logger.Error("dependency failure", "error", err.Error())
		response.Error(c, http.StatusServiceUnavailable, "dependency_unavailable", "dependency unavailable")
	default:
		h.logger.Error("unexpected image handler error", "error", err.Error())
		response.Error(c, http.StatusInternalServerError, "internal_error", "internal server error")
	}
}

func (h *ImageHandler) publicURL(c *gin.Context) string {
	if h.publicBase != "" {
		return strings.TrimRight(h.publicBase, "/")
	}
	scheme := "http"
	if c.Request.TLS != nil {
		scheme = "https"
	}
	return scheme + "://" + c.Request.Host
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

func sanitizeMessage(err error) string {
	message := err.Error()
	parts := strings.SplitN(message, ": ", 2)
	if len(parts) == 2 {
		return parts[1]
	}
	return message
}
