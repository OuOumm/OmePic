package router

import (
	"net/http"
	"strings"
)

const (
	healthRoute                = "/health"
	runtimeSettingsRoute       = "/v1/runtime-settings"
	publicAnnouncementsRoute   = "/v1/announcements"
	imageUploadRoute           = "/v1/image"
	imageRoutePrefix           = "/i/"
	adminLoginRoute            = "/admin/login"
	adminPasswordRoute         = "/admin/password"
	adminStatusRoute           = "/admin/status"
	adminImagesRoute           = "/admin/images"
	adminIPBansRoute           = "/admin/ip-bans"
	adminAbuseOverviewRoute    = "/admin/abuse/overview"
	adminAbuseIPRoute          = "/admin/abuse/ip"
	adminConfigRoute           = "/admin/config"
	adminStorageInstancesRoute = "/admin/config/storage-instances"
	adminConfigDefaultRoute    = "/admin/config/default"
	adminSystemSettingsRoute   = "/admin/system-settings"
	adminAnnouncementsRoute    = "/admin/announcements"
)

type routeSpec struct {
	Path    string
	Methods []string
}

func methods(values ...string) []string {
	return values
}

var (
	healthRouteSpec              = routeSpec{Path: healthRoute, Methods: methods(http.MethodGet, http.MethodHead)}
	runtimeSettingsRouteSpec     = routeSpec{Path: runtimeSettingsRoute, Methods: methods(http.MethodGet, http.MethodHead)}
	publicAnnouncementsSpec      = routeSpec{Path: publicAnnouncementsRoute, Methods: methods(http.MethodGet, http.MethodHead)}
	imageUploadRouteSpec         = routeSpec{Path: imageUploadRoute, Methods: methods(http.MethodPost)}
	imageRouteSpec               = routeSpec{Path: imageRoutePrefix + ":uid", Methods: methods(http.MethodGet, http.MethodHead, http.MethodDelete)}
	adminLoginRouteSpec          = routeSpec{Path: adminLoginRoute, Methods: methods(http.MethodPost)}
	adminPasswordRouteSpec       = routeSpec{Path: adminPasswordRoute, Methods: methods(http.MethodPut)}
	adminStatusRouteSpec         = routeSpec{Path: adminStatusRoute, Methods: methods(http.MethodGet, http.MethodHead)}
	adminImagesRouteSpec         = routeSpec{Path: adminImagesRoute, Methods: methods(http.MethodGet, http.MethodHead, http.MethodDelete)}
	adminIPBansRouteSpec         = routeSpec{Path: adminIPBansRoute, Methods: methods(http.MethodGet, http.MethodHead, http.MethodPost)}
	adminIPBanByIDRouteSpec      = routeSpec{Path: adminIPBansRoute + "/:id", Methods: methods(http.MethodDelete)}
	adminIPBanImagesRouteSpec    = routeSpec{Path: adminIPBansRoute + "/:id/images", Methods: methods(http.MethodDelete)}
	adminAbuseOverviewRouteSpec  = routeSpec{Path: adminAbuseOverviewRoute, Methods: methods(http.MethodGet, http.MethodHead)}
	adminAbuseIPRouteSpec        = routeSpec{Path: adminAbuseIPRoute, Methods: methods(http.MethodGet, http.MethodHead)}
	adminConfigRouteSpec         = routeSpec{Path: adminConfigRoute, Methods: methods(http.MethodGet, http.MethodHead, http.MethodPost)}
	adminStorageInstancesSpec    = routeSpec{Path: adminStorageInstancesRoute, Methods: methods(http.MethodPost)}
	adminStorageInstanceSpec     = routeSpec{Path: adminStorageInstancesRoute + "/:storageKey", Methods: methods(http.MethodPut, http.MethodDelete)}
	adminConfigDefaultSpec       = routeSpec{Path: adminConfigDefaultRoute, Methods: methods(http.MethodPost)}
	adminSystemSettingsSpec      = routeSpec{Path: adminSystemSettingsRoute, Methods: methods(http.MethodGet, http.MethodHead, http.MethodPut)}
	adminAnnouncementsSpec       = routeSpec{Path: adminAnnouncementsRoute, Methods: methods(http.MethodGet, http.MethodHead, http.MethodPost)}
	adminAnnouncementSpec        = routeSpec{Path: adminAnnouncementsRoute + "/:id", Methods: methods(http.MethodPut, http.MethodDelete)}
	adminAnnouncementArchiveSpec = routeSpec{Path: adminAnnouncementsRoute + "/:id/archive", Methods: methods(http.MethodPost)}
)

var publicRouteSpecs = []routeSpec{
	healthRouteSpec,
	runtimeSettingsRouteSpec,
	publicAnnouncementsSpec,
	imageUploadRouteSpec,
	imageRouteSpec,
	adminLoginRouteSpec,
}

var adminRouteSpecs = []routeSpec{
	adminPasswordRouteSpec,
	adminStatusRouteSpec,
	adminImagesRouteSpec,
	adminIPBansRouteSpec,
	adminIPBanByIDRouteSpec,
	adminIPBanImagesRouteSpec,
	adminAbuseOverviewRouteSpec,
	adminAbuseIPRouteSpec,
	adminConfigRouteSpec,
	adminStorageInstancesSpec,
	adminStorageInstanceSpec,
	adminConfigDefaultSpec,
	adminSystemSettingsSpec,
	adminAnnouncementsSpec,
	adminAnnouncementSpec,
	adminAnnouncementArchiveSpec,
}

func shouldKeepAsAPI404(method string, requestPath string) bool {
	if strings.HasPrefix(requestPath, "/v1/") || strings.HasPrefix(requestPath, imageRoutePrefix) {
		return true
	}
	for _, spec := range publicRouteSpecs {
		if spec.matches(method, requestPath) {
			return true
		}
	}
	for _, spec := range adminRouteSpecs {
		if spec.matches(method, requestPath) {
			return true
		}
	}
	return false
}

func (spec routeSpec) matches(method string, requestPath string) bool {
	if !methodAllowed(method, spec.Methods...) {
		return false
	}
	patternParts := strings.Split(strings.Trim(spec.Path, "/"), "/")
	requestParts := strings.Split(strings.Trim(requestPath, "/"), "/")
	if len(patternParts) != len(requestParts) {
		return false
	}
	for index, patternPart := range patternParts {
		if strings.HasPrefix(patternPart, ":") {
			if requestParts[index] == "" {
				return false
			}
			continue
		}
		if patternPart != requestParts[index] {
			return false
		}
	}
	return true
}

func methodAllowed(method string, methods ...string) bool {
	for _, allowed := range methods {
		if method == allowed {
			return true
		}
	}
	return false
}

func adminPath(routePath string) string {
	return strings.TrimPrefix(routePath, "/admin")
}
