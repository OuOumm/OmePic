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
