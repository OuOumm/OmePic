package handler

import (
	"log/slog"
	"net/http"
	"strconv"

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
	writeServiceError(c, h.logger, err, "announcement dependency failure", "announcement handler failure", map[error]serviceErrorMapping{
		service.ErrNotFound: {
			Status:  http.StatusNotFound,
			Code:    "not_found",
			Message: "announcement not found",
		},
	})
}

func parseID(value string) (int64, error) {
	id, err := strconv.ParseInt(value, 10, 64)
	if err != nil || id <= 0 {
		return 0, strconv.ErrSyntax
	}
	return id, nil
}
