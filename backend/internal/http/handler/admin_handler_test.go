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
	"testing"

	"github.com/gin-gonic/gin"

	"omepic/backend/internal/config"
	"omepic/backend/internal/repository"
	"omepic/backend/internal/service"
	"omepic/backend/internal/storage"
)

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
	request := httptest.NewRequest(http.MethodPut, "/admin/password", bytes.NewBufferString(`{"old_password":"`+service.DefaultAdminPassword+`","new_password":"nosymbol1"}`))
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

func TestAdminUpdateSystemSettingsRejectsInvalidAVIFSettings(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := NewAdminHandler(newTestAdminService(t), slog.New(slog.NewTextHandler(io.Discard, nil)))

	for _, raw := range []string{
		`{"site_name":"OmePic","site_tagline":"Tagline","public_base_url":"","max_upload_size_mb":20,"allowed_mime_types":["image/png"],"avif_quality":101,"avif_speed":8,"allow_storage_selection":true,"maintenance_mode":false,"maintenance_message":"","rate_limit_window_minutes":1,"rate_limit_max_requests":120,"upload_rate_limit_window_minutes":10,"upload_rate_limit_max_requests":20}`,
		`{"site_name":"OmePic","site_tagline":"Tagline","public_base_url":"","max_upload_size_mb":20,"allowed_mime_types":["image/png"],"avif_quality":60,"avif_speed":11,"allow_storage_selection":true,"maintenance_mode":false,"maintenance_message":"","rate_limit_window_minutes":1,"rate_limit_max_requests":120,"upload_rate_limit_window_minutes":10,"upload_rate_limit_max_requests":20}`,
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

func TestAdminUpdateSystemSettingsSuccessIncludesAVIFFields(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := NewAdminHandler(newTestAdminService(t), slog.New(slog.NewTextHandler(io.Discard, nil)))

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPut, "/admin/system-settings", bytes.NewBufferString(`{"site_name":"OmePic","site_tagline":"Tagline","public_base_url":"","max_upload_size_mb":20,"allowed_mime_types":["image/jpeg","image/png"],"avif_quality":77,"avif_speed":5,"allow_storage_selection":true,"maintenance_mode":false,"maintenance_message":"","rate_limit_window_minutes":1,"rate_limit_max_requests":120,"upload_rate_limit_window_minutes":10,"upload_rate_limit_max_requests":20}`))
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
				AvifQuality int `json:"avif_quality"`
				AvifSpeed   int `json:"avif_speed"`
			} `json:"runtime"`
		} `json:"data"`
	}
	if err := json.Unmarshal(recorder.Body.Bytes(), &body); err != nil {
		t.Fatalf("invalid json response: %v", err)
	}
	if !body.Success {
		t.Fatalf("expected success response, got %+v", body)
	}
	if body.Data.Runtime.AvifQuality != 77 || body.Data.Runtime.AvifSpeed != 5 {
		t.Fatalf("expected avif fields in response, got %+v", body.Data.Runtime)
	}
}

func newTestAdminService(t *testing.T) *service.AdminService {
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
	return service.NewAdminService(repo, manager, settings, nil, "test-secret", service.AdminEnvMetadata{})
}
