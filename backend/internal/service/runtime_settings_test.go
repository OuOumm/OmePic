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
	if err := repo.SetConfigValue(ctx, "avif_quality", "75"); err != nil {
		t.Fatalf("SetConfigValue avif_quality returned error: %v", err)
	}
	if err := repo.SetConfigValue(ctx, "avif_speed", "4"); err != nil {
		t.Fatalf("SetConfigValue avif_speed returned error: %v", err)
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
	if values["avif_quality"] != "75" || values["avif_speed"] != "4" {
		t.Fatalf("expected existing avif settings to remain unchanged, got quality=%q speed=%q", values["avif_quality"], values["avif_speed"])
	}
	current := manager.Current()
	if current.SiteName != "Custom Site" {
		t.Fatalf("expected manager to load existing site name, got %q", current.SiteName)
	}
	if current.AvifQuality != 75 || current.AvifSpeed != 4 {
		t.Fatalf("expected manager to load existing avif settings, got quality=%d speed=%d", current.AvifQuality, current.AvifSpeed)
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

func TestValidateRuntimeSettingsInputRejectsInvalidAVIFSettings(t *testing.T) {
	base := RuntimeSettingsUpdateInput(defaultRuntimeSettings())
	cases := []struct {
		name   string
		mutate func(*RuntimeSettingsUpdateInput)
	}{
		{name: "quality below min", mutate: func(input *RuntimeSettingsUpdateInput) { input.AvifQuality = -1 }},
		{name: "quality above max", mutate: func(input *RuntimeSettingsUpdateInput) { input.AvifQuality = 101 }},
		{name: "speed below min", mutate: func(input *RuntimeSettingsUpdateInput) { input.AvifSpeed = -1 }},
		{name: "speed above max", mutate: func(input *RuntimeSettingsUpdateInput) { input.AvifSpeed = 11 }},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			input := base
			tc.mutate(&input)
			if _, err := ValidateRuntimeSettingsInput(input); err == nil || !containsError(err, ErrInvalidInput) {
				t.Fatalf("expected ErrInvalidInput, got %v", err)
			}
		})
	}
}
