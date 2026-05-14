package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"omepic/backend/internal/response"
	"omepic/backend/internal/service"
)

type HealthHandler struct {
	service *service.HealthService
}

func NewHealthHandler(healthService *service.HealthService) *HealthHandler {
	return &HealthHandler{service: healthService}
}

func (h *HealthHandler) Health(c *gin.Context) {
	status, err := h.service.Check(c.Request.Context())
	if err != nil {
		response.Error(c, http.StatusServiceUnavailable, "dependency_unavailable", "dependency unavailable")
		return
	}
	response.Success(c, http.StatusOK, gin.H{"status": status.Status})
}
