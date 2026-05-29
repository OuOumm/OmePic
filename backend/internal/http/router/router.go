package router

import (
	"log/slog"
	"time"

	"github.com/gin-gonic/gin"

	"omepic/backend/internal/auth"
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
	ImageService        *service.ImageService
	RateLimiter         ratelimit.Limiter
	IPResolver          *clientip.Resolver
	JWTSecret           string
	RevChecker          *auth.RevocationChecker
	FrontendDir         string
}

func New(deps Dependencies) *gin.Engine {
	engine := gin.New()
	engine.Use(gin.Recovery())
	engine.Use(middleware.SecurityHeaders())
	engine.Use(middleware.RequestLogger(deps.Logger))
	apiLimiter := middleware.RateLimit(deps.RateLimiter, deps.Logger, middleware.RateLimitPolicy{
		Scope:      "api",
		IPResolver: deps.IPResolver,
		FailClosed: true,
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
		FailClosed: true,
		LimitFunc: func() (int, time.Duration) {
			if deps.Settings == nil {
				return 0, 0
			}
			limit, minutes := deps.Settings.Current().UploadRateLimitPolicy()
			return limit, time.Duration(minutes) * time.Minute
		},
	})

	publicCORS := middleware.PublicCORS(deps.Settings)

	engine.GET(healthRouteSpec.Path, publicCORS, deps.HealthHandler.Health)
	engine.GET(runtimeSettingsRouteSpec.Path, publicCORS, apiLimiter, deps.ImageHandler.RuntimeSettings)
	engine.GET(publicAnnouncementsSpec.Path, publicCORS, apiLimiter, deps.AnnouncementHandler.PublicList)
	engine.POST(imageUploadRouteSpec.Path, publicCORS, middleware.BodyLimit(deps.ImageService.MaxUploadSizeBytes), uploadLimiter, deps.ImageHandler.Upload)
	engine.DELETE(imageRouteSpec.Path, publicCORS, apiLimiter, deps.ImageHandler.Delete)
	engine.GET(imageRouteSpec.Path, publicCORS, deps.ImageHandler.Serve)
	engine.POST(adminLoginRouteSpec.Path, publicCORS, apiLimiter, deps.AdminHandler.Login)

	admin := engine.Group("/admin")
	admin.Use(apiLimiter)
	admin.Use(middleware.AdminAuth(deps.JWTSecret, deps.RevChecker))
	admin.PUT(adminPath(adminPasswordRouteSpec.Path), deps.AdminHandler.ChangePassword)
	admin.GET(adminPath(adminStatusRouteSpec.Path), deps.AdminHandler.Status)
	admin.GET(adminPath(adminImagesRouteSpec.Path), deps.AdminHandler.Images)
	admin.DELETE(adminPath(adminImagesRouteSpec.Path), deps.AdminHandler.DeleteImages)
	admin.POST(adminPath(adminCloudflarePurgeSpec.Path), deps.AdminHandler.PurgeCloudflareImageCache)
	admin.GET(adminPath(adminIPBansRouteSpec.Path), deps.AdminHandler.IPBans)
	admin.POST(adminPath(adminIPBansRouteSpec.Path), deps.AdminHandler.CreateIPBan)
	admin.DELETE(adminPath(adminIPBanByIDRouteSpec.Path), deps.AdminHandler.DeleteIPBan)
	admin.DELETE(adminPath(adminIPBanImagesRouteSpec.Path), deps.AdminHandler.DeleteIPBanImages)
	admin.GET(adminPath(adminAbuseOverviewRouteSpec.Path), deps.AdminHandler.AbuseOverview)
	admin.GET(adminPath(adminAbuseIPRouteSpec.Path), deps.AdminHandler.AbuseIPDetail)
	admin.GET(adminPath(adminConfigRouteSpec.Path), deps.AdminHandler.GetConfig)
	admin.POST(adminPath(adminConfigRouteSpec.Path), deps.AdminHandler.UpdateConfig)
	admin.GET(adminPath(adminStorageHealthSpec.Path), deps.AdminHandler.StorageHealth)
	admin.GET(adminPath(adminStorageHealthHistorySpec.Path), deps.AdminHandler.StorageHealthHistory)
	admin.POST(adminPath(adminStorageHealthKeySpec.Path), deps.AdminHandler.CheckStorageHealth)
	admin.POST(adminPath(adminStorageHealthAllSpec.Path), deps.AdminHandler.CheckAllStorageHealth)
	admin.POST(adminPath(adminStorageInstancesSpec.Path), deps.AdminHandler.CreateStorageConfig)
	admin.PUT(adminPath(adminStorageInstanceSpec.Path), deps.AdminHandler.UpdateStorageConfig)
	admin.DELETE(adminPath(adminStorageInstanceSpec.Path), deps.AdminHandler.DeleteStorageConfig)
	admin.POST(adminPath(adminConfigDefaultSpec.Path), deps.AdminHandler.SetDefaultStorageConfig)
	admin.GET(adminPath(adminSystemSettingsSpec.Path), deps.AdminHandler.GetSystemSettings)
	admin.PUT(adminPath(adminSystemSettingsSpec.Path), deps.AdminHandler.UpdateSystemSettings)
	admin.GET(adminPath(adminAnnouncementsSpec.Path), deps.AnnouncementHandler.AdminList)
	admin.POST(adminPath(adminAnnouncementsSpec.Path), deps.AnnouncementHandler.Create)
	admin.PUT(adminPath(adminAnnouncementSpec.Path), deps.AnnouncementHandler.Update)
	admin.DELETE(adminPath(adminAnnouncementSpec.Path), deps.AnnouncementHandler.Delete)
	admin.POST(adminPath(adminAnnouncementArchiveSpec.Path), deps.AnnouncementHandler.Archive)

	registerFrontendRoutes(engine, deps.FrontendDir, deps.Logger)

	return engine
}


