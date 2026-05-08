package router

import (
	"log/slog"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
)

const defaultFrontendDir = "web"

func registerFrontendRoutes(engine *gin.Engine, frontendDir string, logger *slog.Logger) {
	root := strings.TrimSpace(frontendDir)
	if root == "" {
		root = defaultFrontendDir
	}

	indexPath := filepath.Join(root, "index.html")
	if _, err := os.Stat(indexPath); err != nil {
		if logger != nil {
			logger.Info("frontend static build not found; serving API only", "frontend_dir", root)
		}
		return
	}

	if logger != nil {
		logger.Info("serving frontend static build", "frontend_dir", root)
	}

	engine.NoRoute(func(c *gin.Context) {
		requestPath := c.Request.URL.Path
		if shouldKeepAsAPI404(c.Request.Method, requestPath) {
			c.JSON(http.StatusNotFound, gin.H{
				"success": false,
				"error": gin.H{
					"code":    "not_found",
					"message": "route not found",
				},
			})
			return
		}

		if c.Request.Method != http.MethodGet && c.Request.Method != http.MethodHead {
			c.Status(http.StatusNotFound)
			return
		}

		if serveFrontendFile(c, root, requestPath) {
			return
		}
		if strings.Contains(path.Base(path.Clean("/"+requestPath)), ".") {
			c.Status(http.StatusNotFound)
			return
		}

		c.File(indexPath)
	})
}

func serveFrontendFile(c *gin.Context, root string, requestPath string) bool {
	cleanPath := path.Clean("/" + requestPath)
	relativePath := strings.TrimPrefix(cleanPath, "/")
	if relativePath == "" || strings.HasPrefix(relativePath, "..") {
		return false
	}

	candidatePath := filepath.Join(root, filepath.FromSlash(relativePath))
	info, err := os.Stat(candidatePath)
	if err == nil && !info.IsDir() {
		c.File(candidatePath)
		return true
	}

	if strings.Contains(path.Base(cleanPath), ".") {
		return false
	}

	htmlPath := candidatePath + ".html"
	if info, err = os.Stat(htmlPath); err == nil && !info.IsDir() {
		c.File(htmlPath)
		return true
	}

	indexPath := filepath.Join(candidatePath, "index.html")
	if info, err = os.Stat(indexPath); err == nil && !info.IsDir() {
		c.File(indexPath)
		return true
	}

	return false
}

func shouldKeepAsAPI404(method string, requestPath string) bool {
	switch {
	case isReadMethod(method) && requestPath == "/health":
		return true
	case strings.HasPrefix(requestPath, "/v1/"):
		return true
	case strings.HasPrefix(requestPath, "/i/"):
		return true
	case method == http.MethodPost && requestPath == "/admin/login":
		return true
	case isReadMethod(method) && requestPath == "/admin/status":
		return true
	case (isReadMethod(method) || method == http.MethodDelete) && requestPath == "/admin/images":
		return true
	case (isReadMethod(method) || method == http.MethodPost) && requestPath == "/admin/config":
		return true
	case (isReadMethod(method) || method == http.MethodPut) && requestPath == "/admin/system-settings":
		return true
	case isAdminAnnouncementRoute(method, requestPath):
		return true
	case isAdminConfigMutation(method, requestPath):
		return true
	default:
		return false
	}
}

func isReadMethod(method string) bool {
	return method == http.MethodGet || method == http.MethodHead
}

func isAdminConfigMutation(method string, requestPath string) bool {
	switch {
	case method == http.MethodPost && requestPath == "/admin/config/default":
		return true
	case method == http.MethodPost && requestPath == "/admin/config/storage-instances":
		return true
	case (method == http.MethodPut || method == http.MethodDelete) && strings.HasPrefix(requestPath, "/admin/config/storage-instances/"):
		return true
	default:
		return false
	}
}

func isAdminAnnouncementRoute(method string, requestPath string) bool {
	if requestPath == "/admin/announcements" {
		return isReadMethod(method) || method == http.MethodPost
	}
	if !strings.HasPrefix(requestPath, "/admin/announcements/") {
		return false
	}
	return method == http.MethodPut || method == http.MethodDelete || method == http.MethodPost
}
