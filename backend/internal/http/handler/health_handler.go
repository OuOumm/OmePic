package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"omepic/backend/internal/cache"
	"omepic/backend/internal/repository"
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
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"success": false,
			"error":   gin.H{"code": "dependency_unavailable", "message": "sqlite unavailable"},
		})
		return
	}
	if err := h.cache.Ping(c.Request.Context()); err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"success": false,
			"error":   gin.H{"code": "dependency_unavailable", "message": "redis unavailable"},
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    gin.H{"status": "ok"},
	})
}
