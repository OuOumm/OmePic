package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"omepic/backend/internal/cache"
	"omepic/backend/internal/repository"
	"omepic/backend/internal/response"
)

type HealthHandler struct {
	repo  *repository.Repository
	cache cache.ImageCache
}

func NewHealthHandler(repo *repository.Repository, imageCache cache.ImageCache) *HealthHandler {
	return &HealthHandler{repo: repo, cache: imageCache}
}

func (h *HealthHandler) Health(c *gin.Context) {
	if err := h.repo.Ping(c.Request.Context()); err != nil {
		response.Error(c, http.StatusServiceUnavailable, "dependency_unavailable", "sqlite unavailable")
		return
	}
	if err := h.cache.Ping(c.Request.Context()); err != nil {
		response.Error(c, http.StatusServiceUnavailable, "dependency_unavailable", "redis unavailable")
		return
	}
	response.Success(c, http.StatusOK, gin.H{"status": "ok"})
}
