package handler

import (
	"log/slog"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	"omepic/backend/internal/response"
	"omepic/backend/internal/service"
)

type AnnouncementHandler struct {
	service *service.AnnouncementService
	logger  *slog.Logger
}

func NewAnnouncementHandler(announcementService *service.AnnouncementService, logger *slog.Logger) *AnnouncementHandler {
	return &AnnouncementHandler{service: announcementService, logger: logger}
}

func (h *AnnouncementHandler) PublicList(c *gin.Context) {
	items, err := h.service.PublicAnnouncements(c.Request.Context())
	if err != nil {
		h.mapError(c, err)
		return
	}
	response.Success(c, http.StatusOK, items)
}

func (h *AnnouncementHandler) AdminList(c *gin.Context) {
	items, err := h.service.AdminAnnouncements(c.Request.Context())
	if err != nil {
		h.mapError(c, err)
		return
	}
	response.Success(c, http.StatusOK, items)
}

func (h *AnnouncementHandler) Create(c *gin.Context) {
	var payload service.AnnouncementInput
	if err := c.ShouldBindJSON(&payload); err != nil {
		response.Error(c, http.StatusBadRequest, "invalid_input", "invalid announcement payload")
		return
	}
	item, err := h.service.CreateAnnouncement(c.Request.Context(), payload)
	if err != nil {
		h.mapError(c, err)
		return
	}
	response.Success(c, http.StatusCreated, item)
}

func (h *AnnouncementHandler) Update(c *gin.Context) {
	id, err := parseID(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "invalid_input", "invalid announcement id")
		return
	}
	var payload service.AnnouncementInput
	if err := c.ShouldBindJSON(&payload); err != nil {
		response.Error(c, http.StatusBadRequest, "invalid_input", "invalid announcement payload")
		return
	}
	item, err := h.service.UpdateAnnouncement(c.Request.Context(), id, payload)
	if err != nil {
		h.mapError(c, err)
		return
	}
	response.Success(c, http.StatusOK, item)
}

func (h *AnnouncementHandler) Delete(c *gin.Context) {
	id, err := parseID(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "invalid_input", "invalid announcement id")
		return
	}
	if err := h.service.DeleteAnnouncement(c.Request.Context(), id); err != nil {
		h.mapError(c, err)
		return
	}
	response.Success(c, http.StatusOK, gin.H{})
}

func (h *AnnouncementHandler) Archive(c *gin.Context) {
	id, err := parseID(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "invalid_input", "invalid announcement id")
		return
	}
	item, err := h.service.ArchiveAnnouncement(c.Request.Context(), id)
	if err != nil {
		h.mapError(c, err)
		return
	}
	response.Success(c, http.StatusOK, item)
}

func (h *AnnouncementHandler) mapError(c *gin.Context, err error) {
	switch {
	case err == service.ErrInvalidInput || strings.Contains(err.Error(), service.ErrInvalidInput.Error()):
		response.Error(c, http.StatusBadRequest, "invalid_input", sanitizeMessage(err))
	case err == service.ErrNotFound:
		response.Error(c, http.StatusNotFound, "not_found", "announcement not found")
	case strings.Contains(err.Error(), service.ErrDependencyUnavailable.Error()):
		h.logger.Error("announcement dependency failure", "error", err.Error())
		response.Error(c, http.StatusServiceUnavailable, "dependency_unavailable", "dependency unavailable")
	default:
		h.logger.Error("announcement handler failure", "error", err.Error())
		response.Error(c, http.StatusInternalServerError, "internal_error", "internal server error")
	}
}

func parseID(value string) (int64, error) {
	id, err := strconv.ParseInt(value, 10, 64)
	if err != nil || id <= 0 {
		return 0, strconv.ErrSyntax
	}
	return id, nil
}
