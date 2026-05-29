package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"

	"omepic/backend/internal/config"
	"omepic/backend/internal/model"
	"omepic/backend/internal/repository"
	"omepic/backend/internal/service"
	"omepic/backend/internal/storage"
)

const testAdminPassword = "Admin-start!"

func TestAdminChangePasswordWrongOldPasswordReturnsClearMessage(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := NewAdminHandler(newTestAdminService(t), slog.New(slog.NewTextHandler(io.Discard, nil)))

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPut, "/admin/password", bytes.NewBufferString(`{"old_password":"wrong-password","new_password":"New-secret!"}`))
	request.Header.Set("Content-Type", "application/json")
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Request = request

	handler.ChangePassword(ctx)

	if recorder.Code != http.StatusForbidden {
		t.Fatalf("expected status %d, got %d", http.StatusForbidden, recorder.Code)
	}
	var body struct {
		Success bool `json:"success"`
		Error   struct {
			Code    string `json:"code"`
			Message string `json:"message"`
		} `json:"error"`
	}
	if err := json.Unmarshal(recorder.Body.Bytes(), &body); err != nil {
		t.Fatalf("invalid json response: %v", err)
	}
	if body.Success || body.Error.Code != "forbidden" || body.Error.Message != "current password is incorrect" {
		t.Fatalf("expected clear password error, got %+v", body)
	}
}

func TestAdminChangePasswordWeakNewPasswordReturnsInvalidInput(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := NewAdminHandler(newTestAdminService(t), slog.New(slog.NewTextHandler(io.Discard, nil)))

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPut, "/admin/password", bytes.NewBufferString(`{"old_password":"`+testAdminPassword+`","new_password":"nosymbol1"}`))
	request.Header.Set("Content-Type", "application/json")
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Request = request

	handler.ChangePassword(ctx)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, recorder.Code)
	}
	var body struct {
		Success bool `json:"success"`
		Error   struct {
			Code    string `json:"code"`
			Message string `json:"message"`
		} `json:"error"`
	}
	if err := json.Unmarshal(recorder.Body.Bytes(), &body); err != nil {
		t.Fatalf("invalid json response: %v", err)
	}
	if body.Success || body.Error.Code != "invalid_input" || body.Error.Message != "new password must be at least 8 characters and include uppercase, lowercase, and symbol characters" {
		t.Fatalf("expected password strength error, got %+v", body)
	}
}

func TestAdminUpdateSystemSettingsRejectsInvalidRateLimitSettings(t *testing.T) {
	gin.SetMode(gin.TestMode)
	adminService := newTestAdminService(t)
	if err := adminService.ChangePassword(context.Background(), testAdminPassword, "New-secret!"); err != nil {
		t.Fatalf("ChangePassword returned error: %v", err)
	}
	handler := NewAdminHandler(adminService, slog.New(slog.NewTextHandler(io.Discard, nil)))

	for _, raw := range []string{
		`{"site_name":"OmePic","site_tagline":"Tagline","public_base_url":"","allow_storage_selection":true,"maintenance_mode":false,"maintenance_message":"","rate_limit_window_minutes":-1,"rate_limit_max_requests":120,"upload_rate_limit_window_minutes":10,"upload_rate_limit_max_requests":20}`,
		`{"site_name":"OmePic","site_tagline":"Tagline","public_base_url":"","allow_storage_selection":true,"maintenance_mode":false,"maintenance_message":"","rate_limit_window_minutes":1,"rate_limit_max_requests":-5,"upload_rate_limit_window_minutes":10,"upload_rate_limit_max_requests":20}`,
	} {
		recorder := httptest.NewRecorder()
		request := httptest.NewRequest(http.MethodPut, "/admin/system-settings", bytes.NewBufferString(raw))
		request.Header.Set("Content-Type", "application/json")
		ctx, _ := gin.CreateTestContext(recorder)
		ctx.Request = request

		handler.UpdateSystemSettings(ctx)

		if recorder.Code != http.StatusBadRequest {
			t.Fatalf("expected status %d, got %d body=%s", http.StatusBadRequest, recorder.Code, recorder.Body.String())
		}
		var body struct {
			Success bool `json:"success"`
			Error   struct {
				Code string `json:"code"`
			} `json:"error"`
		}
		if err := json.Unmarshal(recorder.Body.Bytes(), &body); err != nil {
			t.Fatalf("invalid json response: %v", err)
		}
		if body.Success || body.Error.Code != "invalid_input" {
			t.Fatalf("expected invalid_input error, got %+v", body)
		}
	}
}

func TestAdminUpdateSystemSettingsSuccessIncludesRuntimeFields(t *testing.T) {
	gin.SetMode(gin.TestMode)
	adminService := newTestAdminService(t)
	if err := adminService.ChangePassword(context.Background(), testAdminPassword, "New-secret!"); err != nil {
		t.Fatalf("ChangePassword returned error: %v", err)
	}
	handler := NewAdminHandler(adminService, slog.New(slog.NewTextHandler(io.Discard, nil)))

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPut, "/admin/system-settings", bytes.NewBufferString(`{"site_name":"OmePic","site_tagline":"Tagline","public_base_url":"","allow_storage_selection":true,"maintenance_mode":false,"maintenance_message":"","rate_limit_window_minutes":1,"rate_limit_max_requests":120,"upload_rate_limit_window_minutes":10,"upload_rate_limit_max_requests":20}`))
	request.Header.Set("Content-Type", "application/json")
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Request = request

	handler.UpdateSystemSettings(ctx)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d body=%s", http.StatusOK, recorder.Code, recorder.Body.String())
	}
	var body struct {
		Success bool `json:"success"`
		Data    struct {
			Runtime struct {
				SiteName           string `json:"site_name"`
				AllowStorageSelect bool   `json:"allow_storage_selection"`
				MaintenanceMode    bool   `json:"maintenance_mode"`
				RateLimitWindowMinutes int `json:"rate_limit_window_minutes"`
				RateLimitMaxRequests   int `json:"rate_limit_max_requests"`
			} `json:"runtime"`
		} `json:"data"`
	}
	if err := json.Unmarshal(recorder.Body.Bytes(), &body); err != nil {
		t.Fatalf("invalid json response: %v", err)
	}
	if !body.Success {
		t.Fatalf("expected success response, got %+v", body)
	}
	if body.Data.Runtime.SiteName != "OmePic" || !body.Data.Runtime.AllowStorageSelect {
		t.Fatalf("expected runtime fields in response, got %+v", body.Data.Runtime)
	}
}

func TestAdminStorageHealthManualCheckAndList(t *testing.T) {
	gin.SetMode(gin.TestMode)
	adminService := newTestAdminService(t)
	handler := NewAdminHandler(adminService, slog.New(slog.NewTextHandler(io.Discard, nil)))

	checkRecorder := httptest.NewRecorder()
	checkCtx, _ := gin.CreateTestContext(checkRecorder)
	checkCtx.Params = gin.Params{{Key: "key", Value: "local-default"}}
	checkCtx.Request = httptest.NewRequest(http.MethodPost, "/admin/storage/local-default/health-check", nil)
	handler.CheckStorageHealth(checkCtx)
	if checkRecorder.Code != http.StatusOK {
		t.Fatalf("expected manual check status 200, got %d body=%s", checkRecorder.Code, checkRecorder.Body.String())
	}
	var checkBody struct {
		Success bool                     `json:"success"`
		Data    model.StorageHealthCheck `json:"data"`
	}
	if err := json.Unmarshal(checkRecorder.Body.Bytes(), &checkBody); err != nil {
		t.Fatalf("invalid manual check json: %v", err)
	}
	if !checkBody.Success || checkBody.Data.StorageKey != "local-default" || checkBody.Data.Status != model.StorageHealthHealthy {
		t.Fatalf("unexpected manual check response: %+v", checkBody)
	}

	listRecorder := httptest.NewRecorder()
	listCtx, _ := gin.CreateTestContext(listRecorder)
	listCtx.Request = httptest.NewRequest(http.MethodGet, "/admin/storage/health", nil)
	handler.StorageHealth(listCtx)
	if listRecorder.Code != http.StatusOK {
		t.Fatalf("expected list status 200, got %d body=%s", listRecorder.Code, listRecorder.Body.String())
	}
	var listBody struct {
		Success bool                       `json:"success"`
		Data    []model.StorageHealthCheck `json:"data"`
	}
	if err := json.Unmarshal(listRecorder.Body.Bytes(), &listBody); err != nil {
		t.Fatalf("invalid list json: %v", err)
	}
	if !listBody.Success || len(listBody.Data) != 1 || listBody.Data[0].StorageKey != "local-default" {
		t.Fatalf("unexpected list response: %+v", listBody)
	}
}

func TestAdminStorageHealthCheckAll(t *testing.T) {
	gin.SetMode(gin.TestMode)
	adminService := newTestAdminService(t)
	handler := NewAdminHandler(adminService, slog.New(slog.NewTextHandler(io.Discard, nil)))

	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Request = httptest.NewRequest(http.MethodPost, "/admin/storage/health-check-all", nil)
	handler.CheckAllStorageHealth(ctx)
	if recorder.Code != http.StatusOK {
		t.Fatalf("expected check all status 200, got %d body=%s", recorder.Code, recorder.Body.String())
	}
	if !strings.Contains(recorder.Body.String(), "local-default") || !strings.Contains(recorder.Body.String(), "\"status\":1") {
		t.Fatalf("unexpected check all body: %s", recorder.Body.String())
	}
}

func newTestAdminService(t *testing.T) *service.AdminService {
	t.Helper()
	adminService, _ := newTestAdminServiceWithRepo(t)
	return adminService
}

func newTestAdminServiceWithRepo(t *testing.T) (*service.AdminService, *repository.Repository) {
	t.Helper()

	dir := t.TempDir()
	repo, err := repository.New(filepath.Join(dir, "test.sqlite"))
	if err != nil {
		t.Fatalf("repository.New returned error: %v", err)
	}
	t.Cleanup(func() { _ = repo.Close() })

	ctx := context.Background()
	if err := repo.Migrate(ctx); err != nil {
		t.Fatalf("Migrate returned error: %v", err)
	}
	catalog, err := repo.InitializeStorageCatalog(ctx, config.RuntimeStorageConfig{
		StorageKey:       "local-default",
		Name:             "Default Local Storage",
		IsDefault:        true,
		Backend:          config.StorageBackendLocal,
		LocalStoragePath: filepath.Join(dir, "images"),
	})
	if err != nil {
		t.Fatalf("InitializeStorageCatalog returned error: %v", err)
	}
	manager, err := storage.NewManager(catalog.StorageConfigs)
	if err != nil {
		t.Fatalf("storage.NewManager returned error: %v", err)
	}
	settings := service.NewRuntimeSettingsManager()
	return service.NewAdminService(repo, manager, settings, nil, "test-secret", service.AdminEnvMetadata{AdminPassword: testAdminPassword}), repo
}
