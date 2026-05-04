package router

import (
	"log/slog"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"

	"omepic/backend/internal/http/handler"
	"omepic/backend/internal/http/middleware"
)

type Dependencies struct {
	Logger        *slog.Logger
	ImageHandler  *handler.ImageHandler
	AdminHandler  *handler.AdminHandler
	HealthHandler *handler.HealthHandler
	JWTSecret     string
	FrontendDir   string
}

func New(deps Dependencies) *gin.Engine {
	engine := gin.New()
	engine.Use(gin.Recovery())
	engine.Use(cors.New(cors.Config{
		AllowAllOrigins: true,
		AllowMethods:    []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:    []string{"Authorization", "Content-Type", "X-Token"},
	}))
	engine.Use(middleware.RequestLogger(deps.Logger))

	engine.GET("/health", deps.HealthHandler.Health)
	engine.GET("/v1/storage-options", deps.ImageHandler.StorageOptions)
	engine.POST("/v1/image", deps.ImageHandler.Upload)
	engine.DELETE("/i/:uid", deps.ImageHandler.Delete)
	engine.GET("/i/:uid", deps.ImageHandler.Serve)
	engine.POST("/admin/login", deps.AdminHandler.Login)

	admin := engine.Group("/admin")
	admin.Use(middleware.AdminAuth(deps.JWTSecret))
	admin.GET("/status", deps.AdminHandler.Status)
	admin.GET("/images", deps.AdminHandler.Images)
	admin.DELETE("/images", deps.AdminHandler.DeleteImages)
	admin.GET("/config", deps.AdminHandler.GetConfig)
	admin.POST("/config", deps.AdminHandler.UpdateConfig)
	admin.POST("/config/storage-instances", deps.AdminHandler.CreateStorageConfig)
	admin.PUT("/config/storage-instances/:storageKey", deps.AdminHandler.UpdateStorageConfig)
	admin.DELETE("/config/storage-instances/:storageKey", deps.AdminHandler.DeleteStorageConfig)
	admin.POST("/config/default", deps.AdminHandler.SetDefaultStorageConfig)

	registerFrontendRoutes(engine, deps.FrontendDir, deps.Logger)

	return engine
}
