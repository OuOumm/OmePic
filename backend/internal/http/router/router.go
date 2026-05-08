package router

import (
	"log/slog"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"

	"omepic/backend/internal/http/clientip"
	"omepic/backend/internal/http/handler"
	"omepic/backend/internal/http/middleware"
	"omepic/backend/internal/ratelimit"
	"omepic/backend/internal/service"
)

type Dependencies struct {
	Logger              *slog.Logger
	ImageHandler        *handler.ImageHandler
	AdminHandler        *handler.AdminHandler
	AnnouncementHandler *handler.AnnouncementHandler
	HealthHandler       *handler.HealthHandler
	Settings            *service.RuntimeSettingsManager
	RateLimiter         ratelimit.Limiter
	IPResolver          *clientip.Resolver
	JWTSecret           string
	FrontendDir         string
}

func New(deps Dependencies) *gin.Engine {
	engine := gin.New()
	engine.Use(gin.Recovery())
	engine.Use(cors.New(cors.Config{
		AllowAllOrigins: true,
		AllowMethods:    []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:    []string{"Authorization", "Content-Type", "X-Token"},
		ExposeHeaders:   []string{"Retry-After", "X-RateLimit-Limit", "X-RateLimit-Remaining"},
	}))
	engine.Use(middleware.RequestLogger(deps.Logger))
	apiLimiter := middleware.RateLimit(deps.RateLimiter, deps.Logger, middleware.RateLimitPolicy{
		Scope:      "api",
		IPResolver: deps.IPResolver,
		LimitFunc: func() (int, time.Duration) {
			if deps.Settings == nil {
				return 0, 0
			}
			limit, minutes := deps.Settings.Current().RateLimitPolicy()
			return limit, time.Duration(minutes) * time.Minute
		},
	})
	uploadLimiter := middleware.RateLimit(deps.RateLimiter, deps.Logger, middleware.RateLimitPolicy{
		Scope:      "upload",
		IPResolver: deps.IPResolver,
		LimitFunc: func() (int, time.Duration) {
			if deps.Settings == nil {
				return 0, 0
			}
			limit, minutes := deps.Settings.Current().UploadRateLimitPolicy()
			return limit, time.Duration(minutes) * time.Minute
		},
	})

	engine.GET("/health", deps.HealthHandler.Health)
	engine.GET("/v1/runtime-settings", apiLimiter, deps.ImageHandler.RuntimeSettings)
	engine.GET("/v1/announcements", apiLimiter, deps.AnnouncementHandler.PublicList)
	engine.POST("/v1/image", uploadLimiter, deps.ImageHandler.Upload)
	engine.DELETE("/i/:uid", apiLimiter, deps.ImageHandler.Delete)
	engine.GET("/i/:uid", deps.ImageHandler.Serve)
	engine.POST("/admin/login", apiLimiter, deps.AdminHandler.Login)

	admin := engine.Group("/admin")
	admin.Use(apiLimiter)
	admin.Use(middleware.AdminAuth(deps.JWTSecret))
	admin.GET("/status", deps.AdminHandler.Status)
	admin.GET("/images", deps.AdminHandler.Images)
	admin.DELETE("/images", deps.AdminHandler.DeleteImages)
	admin.GET("/ip-bans", deps.AdminHandler.IPBans)
	admin.POST("/ip-bans", deps.AdminHandler.CreateIPBan)
	admin.DELETE("/ip-bans/:id", deps.AdminHandler.DeleteIPBan)
	admin.DELETE("/ip-bans/:id/images", deps.AdminHandler.DeleteIPBanImages)
	admin.GET("/abuse/overview", deps.AdminHandler.AbuseOverview)
	admin.GET("/abuse/ip", deps.AdminHandler.AbuseIPDetail)
	admin.GET("/config", deps.AdminHandler.GetConfig)
	admin.POST("/config", deps.AdminHandler.UpdateConfig)
	admin.POST("/config/storage-instances", deps.AdminHandler.CreateStorageConfig)
	admin.PUT("/config/storage-instances/:storageKey", deps.AdminHandler.UpdateStorageConfig)
	admin.DELETE("/config/storage-instances/:storageKey", deps.AdminHandler.DeleteStorageConfig)
	admin.POST("/config/default", deps.AdminHandler.SetDefaultStorageConfig)
	admin.GET("/system-settings", deps.AdminHandler.GetSystemSettings)
	admin.PUT("/system-settings", deps.AdminHandler.UpdateSystemSettings)
	admin.GET("/announcements", deps.AnnouncementHandler.AdminList)
	admin.POST("/announcements", deps.AnnouncementHandler.Create)
	admin.PUT("/announcements/:id", deps.AnnouncementHandler.Update)
	admin.DELETE("/announcements/:id", deps.AnnouncementHandler.Delete)
	admin.POST("/announcements/:id/archive", deps.AnnouncementHandler.Archive)

	registerFrontendRoutes(engine, deps.FrontendDir, deps.Logger)

	return engine
}
