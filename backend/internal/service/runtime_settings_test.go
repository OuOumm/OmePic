package service

import (
	"context"
	"path/filepath"
	"testing"

	"omepic/backend/internal/repository"
)

func TestRuntimeSettingsLoadPersistsMissingDefaultsWithoutOverwritingExistingValues(t *testing.T) {
	ctx := context.Background()
	repo, err := repository.New(filepath.Join(t.TempDir(), "test.sqlite"))
	if err != nil {
		t.Fatalf("repository.New returned error: %v", err)
	}
	t.Cleanup(func() {
		_ = repo.Close()
	})
	if err := repo.Migrate(ctx); err != nil {
		t.Fatalf("Migrate returned error: %v", err)
	}
	if err := repo.SetConfigValue(ctx, "site_name", "Custom Site"); err != nil {
		t.Fatalf("SetConfigValue returned error: %v", err)
	}

	manager := NewRuntimeSettingsManager()
	if err := manager.Load(ctx, repo); err != nil {
		t.Fatalf("Load returned error: %v", err)
	}

	values, err := repo.GetAllConfig(ctx)
	if err != nil {
		t.Fatalf("GetAllConfig returned error: %v", err)
	}
	for key := range RuntimeSettingsToConfigValues(defaultRuntimeSettings()) {
		if _, ok := values[key]; !ok {
			t.Fatalf("expected default runtime key %q to be persisted", key)
		}
	}
	if values["site_name"] != "Custom Site" {
		t.Fatalf("expected existing site_name to remain unchanged, got %q", values["site_name"])
	}
	if manager.Current().SiteName != "Custom Site" {
		t.Fatalf("expected manager to load existing site name, got %q", manager.Current().SiteName)
	}

	if err := repo.SetConfigValue(ctx, "site_tagline", "Custom Tagline"); err != nil {
		t.Fatalf("SetConfigValue tagline returned error: %v", err)
	}
	if err := manager.Load(ctx, repo); err != nil {
		t.Fatalf("second Load returned error: %v", err)
	}
	values, err = repo.GetAllConfig(ctx)
	if err != nil {
		t.Fatalf("GetAllConfig after reload returned error: %v", err)
	}
	if values["site_tagline"] != "Custom Tagline" {
		t.Fatalf("expected existing site_tagline to remain unchanged, got %q", values["site_tagline"])
	}
}
