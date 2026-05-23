package handler

import (
	"context"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"omepic/backend/internal/response"
	"omepic/backend/internal/service"
)

type AdminHandler struct {
	service *service.AdminService
	logger  *slog.Logger
}

func NewAdminHandler(adminService *service.AdminService, logger *slog.Logger) *AdminHandler {
	return &AdminHandler{service: adminService, logger: logger}
}

func (h *AdminHandler) Login(c *gin.Context) {
	var payload struct {
		Password string `json:"password"`
	}
	if err := c.ShouldBindJSON(&payload); err != nil {
		response.Error(c, http.StatusBadRequest, "invalid_input", "password is required")
		return
	}
	token, err := h.service.Login(c.Request.Context(), payload.Password)
	if err != nil {
		switch err {
		case service.ErrInvalidInput:
			response.Error(c, http.StatusBadRequest, "invalid_input", "password is required")
		case service.ErrForbidden:
			response.Error(c, http.StatusForbidden, "forbidden", "invalid admin password")
		default:
			h.logger.Error("admin login failed", "error", err.Error())
			response.Error(c, http.StatusServiceUnavailable, "dependency_unavailable", "dependency unavailable")
		}
		return
	}
	response.Success(c, http.StatusOK, gin.H{"token": token})
}

func (h *AdminHandler) ChangePassword(c *gin.Context) {
	var payload struct {
		OldPassword string `json:"old_password"`
		NewPassword string `json:"new_password"`
	}
	if err := c.ShouldBindJSON(&payload); err != nil {
		response.Error(c, http.StatusBadRequest, "invalid_input", "invalid password payload")
		return
	}
	if err := h.service.ChangePassword(c.Request.Context(), payload.OldPassword, payload.NewPassword); err != nil {
		writeServiceError(c, h.logger, err, "admin password dependency failure", "admin password handler failure", map[error]serviceErrorMapping{
			service.ErrForbidden: {
				Status:  http.StatusForbidden,
				Code:    "forbidden",
				Message: service.UserMessage(err, "current password is incorrect"),
			},
		})
		return
	}
	response.Success(c, http.StatusOK, gin.H{})
}

func (h *AdminHandler) Status(c *gin.Context) {
	status, err := h.service.Status(c.Request.Context())
	if err != nil {
		h.mapError(c, err)
		return
	}
	response.Success(c, http.StatusOK, status)
}

func (h *AdminHandler) Images(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))
	search := strings.TrimSpace(c.Query("search"))
	result, err := h.service.Images(c.Request.Context(), page, pageSize, search)
	if err != nil {
		h.mapError(c, err)
		return
	}
	response.Success(c, http.StatusOK, result)
}

func (h *AdminHandler) DeleteImages(c *gin.Context) {
	var payload struct {
		UIDs []string `json:"uids"`
	}
	if err := c.ShouldBindJSON(&payload); err != nil {
		response.Error(c, http.StatusBadRequest, "invalid_input", "uids are required")
		return
	}
	if err := h.service.DeleteImages(c.Request.Context(), payload.UIDs); err != nil {
		h.mapError(c, err)
		return
	}
	response.Success(c, http.StatusOK, gin.H{})
}

func (h *AdminHandler) PurgeCloudflareImageCache(c *gin.Context) {
	var payload struct {
		URL string `json:"url"`
	}
	if err := c.ShouldBindJSON(&payload); err != nil {
		response.Error(c, http.StatusBadRequest, "invalid_input", "url is required")
		return
	}
	result, err := h.service.PurgeCloudflareImageCache(c.Request.Context(), payload.URL)
	if err != nil {
		h.mapError(c, err)
		return
	}
	response.Success(c, http.StatusOK, result)
}

func (h *AdminHandler) CreateIPBan(c *gin.Context) {
	var payload service.AdminIPBanCreateInput
	if err := c.ShouldBindJSON(&payload); err != nil {
		response.Error(c, http.StatusBadRequest, "invalid_input", "invalid ip ban payload")
		return
	}
	result, err := h.service.CreateIPBan(c.Request.Context(), payload)
	if err != nil {
		h.mapError(c, err)
		return
	}
	response.Success(c, http.StatusOK, result)
}

func (h *AdminHandler) IPBans(c *gin.Context) {
	bans, err := h.service.IPBans(c.Request.Context())
	if err != nil {
		h.mapError(c, err)
		return
	}
	response.Success(c, http.StatusOK, bans)
}

func (h *AdminHandler) AbuseOverview(c *gin.Context) {
	from, err := parseOptionalTime(c.Query("from"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "invalid_input", "invalid from time")
		return
	}
	to, err := parseOptionalTime(c.Query("to"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "invalid_input", "invalid to time")
		return
	}
	overview, err := h.service.AbuseOverview(c.Request.Context(), service.AdminAbuseOverviewInput{From: from, To: to})
	if err != nil {
		h.mapError(c, err)
		return
	}
	response.Success(c, http.StatusOK, overview)
}

func (h *AdminHandler) AbuseIPDetail(c *gin.Context) {
	detail, err := h.service.AbuseIPDetail(c.Request.Context(), c.Query("ip"))
	if err != nil {
		h.mapError(c, err)
		return
	}
	response.Success(c, http.StatusOK, detail)
}

func (h *AdminHandler) Tokens(c *gin.Context) {
	result, err := h.service.Tokens(c.Request.Context())
	if err != nil {
		h.mapError(c, err)
		return
	}
	response.Success(c, http.StatusOK, result)
}

func (h *AdminHandler) DisableToken(c *gin.Context) {
	var payload service.TokenDisableInput
	if c.Request.ContentLength != 0 {
		if err := c.ShouldBindJSON(&payload); err != nil {
			response.Error(c, http.StatusBadRequest, "invalid_input", "invalid token payload")
			return
		}
	}
	if err := h.service.DisableToken(c.Request.Context(), c.Param("token_hash"), payload.Reason); err != nil {
		h.mapError(c, err)
		return
	}
	response.Success(c, http.StatusOK, gin.H{})
}

func (h *AdminHandler) EnableToken(c *gin.Context) {
	if err := h.service.EnableToken(c.Request.Context(), c.Param("token_hash")); err != nil {
		h.mapError(c, err)
		return
	}
	response.Success(c, http.StatusOK, gin.H{})
}

func (h *AdminHandler) DeleteIPBan(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	if err := h.service.DeleteIPBan(c.Request.Context(), id); err != nil {
		h.mapError(c, err)
		return
	}
	response.Success(c, http.StatusOK, gin.H{})
}

func (h *AdminHandler) DeleteIPBanImages(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	result, err := h.service.DeleteImagesByIPBan(c.Request.Context(), id)
	if err != nil {
		h.mapError(c, err)
		return
	}
	response.Success(c, http.StatusOK, result)
}

func (h *AdminHandler) GetConfig(c *gin.Context) {
	view, err := h.service.GetConfig(c.Request.Context())
	if err != nil {
		h.mapError(c, err)
		return
	}
	response.Success(c, http.StatusOK, view)
}

func (h *AdminHandler) AuditLogs(c *gin.Context) {
	page, err := parsePositiveIntQuery(c.Query("page"), 1)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "invalid_input", "invalid page")
		return
	}
	pageSize, err := parsePositiveIntQuery(c.Query("page_size"), 20)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "invalid_input", "invalid page_size")
		return
	}
	logs, err := h.service.ListConfigAuditLogs(c.Request.Context(), service.AdminConfigAuditLogFilter{
		Scope:    c.Query("scope"),
		Page:     page,
		PageSize: pageSize,
	})
	if err != nil {
		h.mapError(c, err)
		return
	}
	response.Success(c, http.StatusOK, logs)
}

func (h *AdminHandler) UpdateConfig(c *gin.Context) {
	var payload service.AdminConfigUpdateInput
	if err := c.ShouldBindJSON(&payload); err != nil {
		response.Error(c, http.StatusBadRequest, "invalid_input", "invalid config payload")
		return
	}
	view, err := h.service.UpdateConfig(auditRequestContext(c), payload)
	if err != nil {
		h.mapError(c, err)
		return
	}
	response.Success(c, http.StatusOK, view)
}

func (h *AdminHandler) CreateStorageConfig(c *gin.Context) {
	var payload service.AdminStorageConfigCreateInput
	if err := c.ShouldBindJSON(&payload); err != nil {
		response.Error(c, http.StatusBadRequest, "invalid_input", "invalid config payload")
		return
	}
	view, err := h.service.CreateStorageConfig(auditRequestContext(c), payload)
	if err != nil {
		h.mapError(c, err)
		return
	}
	response.Success(c, http.StatusOK, view)
}

func (h *AdminHandler) UpdateStorageConfig(c *gin.Context) {
	var payload service.AdminStorageConfigUpdateInput
	if err := c.ShouldBindJSON(&payload); err != nil {
		response.Error(c, http.StatusBadRequest, "invalid_input", "invalid config payload")
		return
	}
	view, err := h.service.UpdateStorageConfig(auditRequestContext(c), c.Param("storageKey"), payload)
	if err != nil {
		h.mapError(c, err)
		return
	}
	response.Success(c, http.StatusOK, view)
}

func (h *AdminHandler) DeleteStorageConfig(c *gin.Context) {
	view, err := h.service.DeleteStorageConfig(auditRequestContext(c), c.Param("storageKey"))
	if err != nil {
		h.mapError(c, err)
		return
	}
	response.Success(c, http.StatusOK, view)
}

func (h *AdminHandler) SetDefaultStorageConfig(c *gin.Context) {
	var payload service.AdminSetDefaultStorageInput
	if err := c.ShouldBindJSON(&payload); err != nil {
		response.Error(c, http.StatusBadRequest, "invalid_input", "invalid config payload")
		return
	}
	view, err := h.service.SetDefaultStorageConfig(auditRequestContext(c), payload.StorageKey)
	if err != nil {
		h.mapError(c, err)
		return
	}
	response.Success(c, http.StatusOK, view)
}

func (h *AdminHandler) GetSystemSettings(c *gin.Context) {
	settings, err := h.service.GetSystemSettings(c.Request.Context())
	if err != nil {
		h.mapError(c, err)
		return
	}
	response.Success(c, http.StatusOK, settings)
}

func (h *AdminHandler) UpdateSystemSettings(c *gin.Context) {
	var payload service.RuntimeSettingsUpdateInput
	if err := c.ShouldBindJSON(&payload); err != nil {
		response.Error(c, http.StatusBadRequest, "invalid_input", "invalid config payload")
		return
	}
	settings, err := h.service.UpdateSystemSettings(auditRequestContext(c), payload)
	if err != nil {
		h.mapError(c, err)
		return
	}
	response.Success(c, http.StatusOK, settings)
}

func (h *AdminHandler) mapError(c *gin.Context, err error) {
	writeServiceError(c, h.logger, err, "admin dependency failure", "admin handler failure", map[error]serviceErrorMapping{
		service.ErrNotFound: {
			Status:  http.StatusNotFound,
			Code:    "not_found",
			Message: "resource not found",
		},
		service.ErrForbidden: {
			Status:  http.StatusForbidden,
			Code:    "forbidden",
			Message: service.UserMessage(err, "forbidden"),
		},
	})
}

func auditRequestContext(c *gin.Context) context.Context {
	return service.WithAuditActor(c.Request.Context(), service.AuditActor{
		Name: "admin",
		IP:   c.ClientIP(),
	})
}

func parsePositiveIntQuery(value string, fallback int) (int, error) {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return fallback, nil
	}
	parsed, err := strconv.Atoi(trimmed)
	if err != nil {
		return 0, err
	}
	if parsed < 1 {
		return 0, strconv.ErrSyntax
	}
	return parsed, nil
}

func parseOptionalTime(value string) (time.Time, error) {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return time.Time{}, nil
	}
	parsed, err := time.Parse(time.RFC3339, trimmed)
	if err != nil {
		return time.Time{}, err
	}
	return parsed, nil
}
