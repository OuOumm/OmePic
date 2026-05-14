package service

import (
	"context"
	"fmt"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"sync"

	"omepic/backend/internal/repository"
)

const (
	DefaultSiteName                     = "OmePic"
	DefaultSiteTagline                  = "上传、分享和管理图片"
	DefaultMaintenanceMessage           = "系统维护中，请稍后再试"
	DefaultRateLimitWindowMinutes       = 1
	DefaultRateLimitMaxRequests         = 120
	DefaultUploadRateLimitWindowMinutes = 10
	DefaultUploadRateLimitMaxRequests   = 20
	DefaultAVIFQuality                  = 60
	DefaultAVIFSpeed                    = 8
	bytesPerMB                          = 1024 * 1024
)

var defaultAllowedMIMETypes = []string{
	"image/jpeg",
	"image/png",
	"image/gif",
	"image/webp",
	"image/avif",
}

var defaultAllowedMIMETypesCSV = strings.Join(defaultAllowedMIMETypes, ",")

type RuntimeSettings struct {
	SiteName                     string   `json:"site_name"`
	SiteTagline                  string   `json:"site_tagline"`
	PublicBaseURL                string   `json:"public_base_url"`
	MaxUploadSizeMB              int      `json:"max_upload_size_mb"`
	AllowedMIMETypes             []string `json:"allowed_mime_types"`
	AvifQuality                  int      `json:"avif_quality"`
	AvifSpeed                    int      `json:"avif_speed"`
	AllowStorageSelect           bool     `json:"allow_storage_selection"`
	MaintenanceMode              bool     `json:"maintenance_mode"`
	MaintenanceMessage           string   `json:"maintenance_message"`
	RateLimitWindowMinutes       int      `json:"rate_limit_window_minutes"`
	RateLimitMaxRequests         int      `json:"rate_limit_max_requests"`
	UploadRateLimitWindowMinutes int      `json:"upload_rate_limit_window_minutes"`
	UploadRateLimitMaxRequests   int      `json:"upload_rate_limit_max_requests"`
}

type PublicRuntimeSettingsView struct {
	Site     PublicSiteSettingsView    `json:"site"`
	Access   PublicAccessSettingsView  `json:"access"`
	Upload   PublicUploadSettingsView  `json:"upload"`
	Features PublicFeatureSettingsView `json:"features"`
	Storage  PublicStorageSettingsView `json:"storage"`
}

type PublicSiteSettingsView struct {
	Name    string `json:"name"`
	Tagline string `json:"tagline"`
}

type PublicAccessSettingsView struct {
	PublicBaseURL string `json:"public_base_url"`
}

type PublicUploadSettingsView struct {
	MaxUploadSizeMB           int      `json:"max_upload_size_mb"`
	AllowedMIMETypes          []string `json:"allowed_mime_types"`
	EffectiveAllowedMIMETypes []string `json:"effective_allowed_mime_types"`
}

type PublicFeatureSettingsView struct {
	AllowStorageSelection bool   `json:"allow_storage_selection"`
	MaintenanceMode       bool   `json:"maintenance_mode"`
	MaintenanceMessage    string `json:"maintenance_message"`
}

type PublicStorageSettingsView struct {
	Options []PublicStorageOption `json:"options"`
}

type AdminSystemSettingsView struct {
	Runtime  RuntimeSettings       `json:"runtime"`
	Readonly AdminReadonlySettings `json:"readonly"`
}

type AdminReadonlySettings struct {
	Environment AdminEnvironmentStatus `json:"environment"`
	Security    AdminSecurityStatus    `json:"security"`
	Storage     AdminStorageStatus     `json:"storage"`
	Service     AdminServiceStatus     `json:"service"`
}

type AdminEnvironmentStatus struct {
	HTTPAddr                string `json:"http_addr"`
	DatabasePath            string `json:"database_path"`
	RedisConfigured         bool   `json:"redis_configured"`
	PublicBaseURLSource     string `json:"public_base_url_source"`
	RuntimePublicBaseURLSet bool   `json:"runtime_public_base_url_set"`
}

type SecretStatus struct {
	Configured   bool `json:"configured"`
	UsingDefault bool `json:"using_default"`
}

type AdminSecurityStatus struct {
	JWTSecret        SecretStatus `json:"jwt_secret"`
	AdminPassword    SecretStatus `json:"admin_password"`
	UIDEncryptionKey SecretStatus `json:"uid_encryption_key"`
}

type AdminStorageStatus struct {
	DefaultStorageKey     string `json:"default_storage_key"`
	StorageConfigCount    int    `json:"storage_config_count"`
	AllowStorageSelection bool   `json:"allow_storage_selection"`
}

type AdminServiceStatus struct {
	Health          string `json:"health"`
	MaintenanceMode bool   `json:"maintenance_mode"`
}

type RuntimeSettingsUpdateInput struct {
	SiteName                     string   `json:"site_name"`
	SiteTagline                  string   `json:"site_tagline"`
	PublicBaseURL                string   `json:"public_base_url"`
	MaxUploadSizeMB              int      `json:"max_upload_size_mb"`
	AllowedMIMETypes             []string `json:"allowed_mime_types"`
	AvifQuality                  int      `json:"avif_quality"`
	AvifSpeed                    int      `json:"avif_speed"`
	AllowStorageSelect           bool     `json:"allow_storage_selection"`
	MaintenanceMode              bool     `json:"maintenance_mode"`
	MaintenanceMessage           string   `json:"maintenance_message"`
	RateLimitWindowMinutes       int      `json:"rate_limit_window_minutes"`
	RateLimitMaxRequests         int      `json:"rate_limit_max_requests"`
	UploadRateLimitWindowMinutes int      `json:"upload_rate_limit_window_minutes"`
	UploadRateLimitMaxRequests   int      `json:"upload_rate_limit_max_requests"`
}

type RuntimeSettingsManager struct {
	mu       sync.RWMutex
	settings RuntimeSettings
}

func NewRuntimeSettingsManager() *RuntimeSettingsManager {
	return &RuntimeSettingsManager{
		settings: defaultRuntimeSettings(),
	}
}

func (m *RuntimeSettingsManager) Load(ctx context.Context, repo *repository.Repository) error {
	defaults := RuntimeSettingsToConfigValues(defaultRuntimeSettings())
	if err := repo.InsertMissingConfigValues(ctx, defaults); err != nil {
		return fmt.Errorf("%w: default settings save failed", ErrDependencyUnavailable)
	}

	values, err := repo.GetAllConfig(ctx)
	if err != nil {
		return fmt.Errorf("%w: settings query failed", ErrDependencyUnavailable)
	}
	settings, err := runtimeSettingsFromValues(values)
	if err != nil {
		return err
	}
	m.Reconfigure(settings)
	return nil
}

func (m *RuntimeSettingsManager) Current() RuntimeSettings {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return cloneRuntimeSettings(m.settings)
}

func (m *RuntimeSettingsManager) Reconfigure(settings RuntimeSettings) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.settings = normalizeRuntimeSettings(settings)
}

func (m *RuntimeSettingsManager) EffectivePublicBaseURL(requestBase string) string {
	settings := m.Current()
	if settings.PublicBaseURL != "" {
		return strings.TrimRight(settings.PublicBaseURL, "/")
	}
	return strings.TrimRight(requestBase, "/")
}

func (m *RuntimeSettingsManager) PublicBaseURLSource() string {
	settings := m.Current()
	if settings.PublicBaseURL != "" {
		return "runtime"
	}
	return "request_host"
}

func (s RuntimeSettings) EffectiveMaintenanceMessage() string {
	message := strings.TrimSpace(s.MaintenanceMessage)
	if message == "" {
		return DefaultMaintenanceMessage
	}
	return message
}

func (s RuntimeSettings) EffectiveAllowedMIMETypes() []string {
	return append([]string(nil), s.AllowedMIMETypes...)
}

func (s RuntimeSettings) MaxUploadSizeBytes() int64 {
	if s.MaxUploadSizeMB <= 0 {
		return 0
	}
	return int64(s.MaxUploadSizeMB) * bytesPerMB
}

func DefaultAllowedMIMETypes() []string {
	return append([]string(nil), defaultAllowedMIMETypes...)
}

func ValidateRuntimeSettingsInput(input RuntimeSettingsUpdateInput) (RuntimeSettings, error) {
	settings := RuntimeSettings{
		SiteName:                     strings.TrimSpace(input.SiteName),
		SiteTagline:                  strings.TrimSpace(input.SiteTagline),
		PublicBaseURL:                strings.TrimSpace(input.PublicBaseURL),
		MaxUploadSizeMB:              input.MaxUploadSizeMB,
		AllowedMIMETypes:             input.AllowedMIMETypes,
		AvifQuality:                  input.AvifQuality,
		AvifSpeed:                    input.AvifSpeed,
		AllowStorageSelect:           input.AllowStorageSelect,
		MaintenanceMode:              input.MaintenanceMode,
		MaintenanceMessage:           strings.TrimSpace(input.MaintenanceMessage),
		RateLimitWindowMinutes:       input.RateLimitWindowMinutes,
		RateLimitMaxRequests:         input.RateLimitMaxRequests,
		UploadRateLimitWindowMinutes: input.UploadRateLimitWindowMinutes,
		UploadRateLimitMaxRequests:   input.UploadRateLimitMaxRequests,
	}
	if settings.PublicBaseURL != "" {
		parsed, err := url.Parse(settings.PublicBaseURL)
		if err != nil || parsed.Scheme == "" || parsed.Host == "" || (parsed.Scheme != "http" && parsed.Scheme != "https") {
			return RuntimeSettings{}, WithUserMessage(ErrInvalidInput, "public base url must be an http or https URL")
		}
	}
	if settings.MaxUploadSizeMB < 0 {
		return RuntimeSettings{}, WithUserMessage(ErrInvalidInput, "max upload size must be zero or greater")
	}
	if settings.RateLimitWindowMinutes < 0 || settings.RateLimitMaxRequests < 0 || settings.UploadRateLimitWindowMinutes < 0 || settings.UploadRateLimitMaxRequests < 0 {
		return RuntimeSettings{}, WithUserMessage(ErrInvalidInput, "rate limit values must be zero or greater")
	}
	if settings.AvifQuality < 0 || settings.AvifQuality > 100 {
		return RuntimeSettings{}, WithUserMessage(ErrInvalidInput, "avif quality must be between 0 and 100")
	}
	if settings.AvifSpeed < 0 || settings.AvifSpeed > 10 {
		return RuntimeSettings{}, WithUserMessage(ErrInvalidInput, "avif speed must be between 0 and 10")
	}
	allowed, err := normalizeMIMETypes(settings.AllowedMIMETypes)
	if err != nil {
		return RuntimeSettings{}, err
	}
	settings.AllowedMIMETypes = allowed
	return normalizeRuntimeSettings(settings), nil
}

func RuntimeSettingsToConfigValues(settings RuntimeSettings) map[string]string {
	settings = normalizeRuntimeSettings(settings)
	return map[string]string{
		"site_name":                        settings.SiteName,
		"site_tagline":                     settings.SiteTagline,
		"public_base_url":                  settings.PublicBaseURL,
		"max_upload_size_mb":               strconv.Itoa(settings.MaxUploadSizeMB),
		"allowed_mime_types":               strings.Join(settings.AllowedMIMETypes, ","),
		"avif_quality":                     strconv.Itoa(settings.AvifQuality),
		"avif_speed":                       strconv.Itoa(settings.AvifSpeed),
		"allow_storage_selection":          boolStringValue(settings.AllowStorageSelect),
		"maintenance_mode":                 boolStringValue(settings.MaintenanceMode),
		"maintenance_message":              settings.MaintenanceMessage,
		"rate_limit_window_minutes":        strconv.Itoa(settings.RateLimitWindowMinutes),
		"rate_limit_max_requests":          strconv.Itoa(settings.RateLimitMaxRequests),
		"upload_rate_limit_window_minutes": strconv.Itoa(settings.UploadRateLimitWindowMinutes),
		"upload_rate_limit_max_requests":   strconv.Itoa(settings.UploadRateLimitMaxRequests),
	}
}

func runtimeSettingsFromValues(values map[string]string) (RuntimeSettings, error) {
	settings := defaultRuntimeSettings()
	if value, ok := values["site_name"]; ok {
		settings.SiteName = strings.TrimSpace(value)
	}
	if value, ok := values["site_tagline"]; ok {
		settings.SiteTagline = strings.TrimSpace(value)
	}
	if value, ok := values["public_base_url"]; ok {
		settings.PublicBaseURL = strings.TrimSpace(value)
	}
	if value, ok := values["max_upload_size_mb"]; ok && strings.TrimSpace(value) != "" {
		parsed, err := strconv.Atoi(strings.TrimSpace(value))
		if err != nil {
			return RuntimeSettings{}, WithUserMessage(ErrInvalidInput, "max upload size is invalid")
		}
		settings.MaxUploadSizeMB = parsed
	}
	if value, ok := values["allowed_mime_types"]; ok {
		allowed, err := normalizeMIMETypes(splitCSV(value))
		if err != nil {
			return RuntimeSettings{}, err
		}
		settings.AllowedMIMETypes = allowed
	} else {
		values["allowed_mime_types"] = defaultAllowedMIMETypesCSV
	}
	if value, ok := values["avif_quality"]; ok && strings.TrimSpace(value) != "" {
		parsed, err := strconv.Atoi(strings.TrimSpace(value))
		if err != nil {
			return RuntimeSettings{}, WithUserMessage(ErrInvalidInput, "avif quality is invalid")
		}
		settings.AvifQuality = parsed
	}
	if value, ok := values["avif_speed"]; ok && strings.TrimSpace(value) != "" {
		parsed, err := strconv.Atoi(strings.TrimSpace(value))
		if err != nil {
			return RuntimeSettings{}, WithUserMessage(ErrInvalidInput, "avif speed is invalid")
		}
		settings.AvifSpeed = parsed
	}
	if value, ok := values["allow_storage_selection"]; ok && strings.TrimSpace(value) != "" {
		settings.AllowStorageSelect = parseBoolValue(value)
	}
	if value, ok := values["maintenance_mode"]; ok && strings.TrimSpace(value) != "" {
		settings.MaintenanceMode = parseBoolValue(value)
	}
	if value, ok := values["maintenance_message"]; ok {
		settings.MaintenanceMessage = strings.TrimSpace(value)
	}
	if value, ok := values["rate_limit_window_minutes"]; ok && strings.TrimSpace(value) != "" {
		parsed, err := strconv.Atoi(strings.TrimSpace(value))
		if err != nil {
			return RuntimeSettings{}, WithUserMessage(ErrInvalidInput, "rate limit window is invalid")
		}
		settings.RateLimitWindowMinutes = parsed
	}
	if value, ok := values["rate_limit_max_requests"]; ok && strings.TrimSpace(value) != "" {
		parsed, err := strconv.Atoi(strings.TrimSpace(value))
		if err != nil {
			return RuntimeSettings{}, WithUserMessage(ErrInvalidInput, "rate limit max requests is invalid")
		}
		settings.RateLimitMaxRequests = parsed
	}
	if value, ok := values["upload_rate_limit_window_minutes"]; ok && strings.TrimSpace(value) != "" {
		parsed, err := strconv.Atoi(strings.TrimSpace(value))
		if err != nil {
			return RuntimeSettings{}, WithUserMessage(ErrInvalidInput, "upload rate limit window is invalid")
		}
		settings.UploadRateLimitWindowMinutes = parsed
	}
	if value, ok := values["upload_rate_limit_max_requests"]; ok && strings.TrimSpace(value) != "" {
		parsed, err := strconv.Atoi(strings.TrimSpace(value))
		if err != nil {
			return RuntimeSettings{}, WithUserMessage(ErrInvalidInput, "upload rate limit max requests is invalid")
		}
		settings.UploadRateLimitMaxRequests = parsed
	}
	return ValidateRuntimeSettingsInput(RuntimeSettingsUpdateInput(settings))
}

func defaultRuntimeSettings() RuntimeSettings {
	return RuntimeSettings{
		SiteName:                     DefaultSiteName,
		SiteTagline:                  DefaultSiteTagline,
		MaxUploadSizeMB:              20,
		AllowedMIMETypes:             DefaultAllowedMIMETypes(),
		AvifQuality:                  DefaultAVIFQuality,
		AvifSpeed:                    DefaultAVIFSpeed,
		AllowStorageSelect:           true,
		RateLimitWindowMinutes:       DefaultRateLimitWindowMinutes,
		RateLimitMaxRequests:         DefaultRateLimitMaxRequests,
		UploadRateLimitWindowMinutes: DefaultUploadRateLimitWindowMinutes,
		UploadRateLimitMaxRequests:   DefaultUploadRateLimitMaxRequests,
	}
}

func normalizeRuntimeSettings(settings RuntimeSettings) RuntimeSettings {
	settings.SiteName = strings.TrimSpace(settings.SiteName)
	if settings.SiteName == "" {
		settings.SiteName = DefaultSiteName
	}
	settings.SiteTagline = strings.TrimSpace(settings.SiteTagline)
	if settings.SiteTagline == "" {
		settings.SiteTagline = DefaultSiteTagline
	}
	settings.PublicBaseURL = strings.TrimRight(strings.TrimSpace(settings.PublicBaseURL), "/")
	settings.MaintenanceMessage = strings.TrimSpace(settings.MaintenanceMessage)
	allowed, _ := normalizeMIMETypes(settings.AllowedMIMETypes)
	if allowed == nil {
		allowed = []string{}
	}
	settings.AllowedMIMETypes = allowed
	return settings
}

func (s RuntimeSettings) RateLimitPolicy() (int, int) {
	return s.RateLimitMaxRequests, s.RateLimitWindowMinutes
}

func (s RuntimeSettings) UploadRateLimitPolicy() (int, int) {
	return s.UploadRateLimitMaxRequests, s.UploadRateLimitWindowMinutes
}

func cloneRuntimeSettings(settings RuntimeSettings) RuntimeSettings {
	settings.AllowedMIMETypes = append([]string(nil), settings.AllowedMIMETypes...)
	return settings
}

func normalizeMIMETypes(values []string) ([]string, error) {
	seen := make(map[string]struct{}, len(values))
	result := make([]string, 0, len(values))
	for _, value := range values {
		mimeType := strings.TrimSpace(strings.ToLower(value))
		if mimeType == "" {
			continue
		}
		if mimeType == "image/jpg" {
			mimeType = "image/jpeg"
		}
		if !strings.HasPrefix(mimeType, "image/") || strings.ContainsAny(mimeType, " ;") {
			return nil, WithUserMessage(ErrInvalidInput, "allowed mime types must be image MIME values")
		}
		if mimeType == "image/svg+xml" {
			return nil, WithUserMessage(ErrInvalidInput, "svg uploads are not allowed")
		}
		if _, ok := seen[mimeType]; ok {
			continue
		}
		seen[mimeType] = struct{}{}
		result = append(result, mimeType)
	}
	sort.Strings(result)
	if result == nil {
		return []string{}, nil
	}
	return result, nil
}

func splitCSV(value string) []string {
	parts := strings.Split(value, ",")
	result := make([]string, 0, len(parts))
	for _, part := range parts {
		result = append(result, strings.TrimSpace(part))
	}
	return result
}

func parseBoolValue(value string) bool {
	switch strings.TrimSpace(strings.ToLower(value)) {
	case "1", "true", "yes", "on":
		return true
	default:
		return false
	}
}

func boolStringValue(value bool) string {
	if value {
		return "true"
	}
	return "false"
}
