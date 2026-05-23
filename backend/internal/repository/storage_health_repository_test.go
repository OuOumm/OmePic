package repository

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"omepic/backend/internal/model"
)

func TestMigrateCreatesStorageHealthChecksIdempotently(t *testing.T) {
	ctx := context.Background()
	repo, err := New(filepath.Join(t.TempDir(), "test.sqlite"))
	if err != nil {
		t.Fatalf("New returned error: %v", err)
	}
	t.Cleanup(func() { _ = repo.Close() })

	if err := repo.Migrate(ctx); err != nil {
		t.Fatalf("first Migrate returned error: %v", err)
	}
	if err := repo.Migrate(ctx); err != nil {
		t.Fatalf("second Migrate returned error: %v", err)
	}
	for _, column := range []string{"id", "storage_key", "status", "last_check_at", "latency_ms", "error_message", "consecutive_failures", "created_at", "updated_at"} {
		exists, err := testTableColumnExists(ctx, repo.db, "storage_health_checks", column)
		if err != nil {
			t.Fatalf("testTableColumnExists(%s) returned error: %v", column, err)
		}
		if !exists {
			t.Fatalf("expected storage_health_checks.%s to exist", column)
		}
	}
}

func TestUpsertStorageHealthCheckIncrementsAndResetsFailures(t *testing.T) {
	ctx := context.Background()
	repo, err := New(filepath.Join(t.TempDir(), "test.sqlite"))
	if err != nil {
		t.Fatalf("New returned error: %v", err)
	}
	t.Cleanup(func() { _ = repo.Close() })
	if err := repo.Migrate(ctx); err != nil {
		t.Fatalf("Migrate returned error: %v", err)
	}

	first, err := repo.UpsertStorageHealthCheck(ctx, model.StorageHealthCheck{StorageKey: "local-default", Status: model.StorageHealthUnavailable, LastCheckAt: time.Now(), ErrorMessage: "boom"})
	if err != nil {
		t.Fatalf("first upsert returned error: %v", err)
	}
	second, err := repo.UpsertStorageHealthCheck(ctx, model.StorageHealthCheck{StorageKey: "local-default", Status: model.StorageHealthUnavailable, LastCheckAt: time.Now(), ErrorMessage: "boom"})
	if err != nil {
		t.Fatalf("second upsert returned error: %v", err)
	}
	healthy, err := repo.UpsertStorageHealthCheck(ctx, model.StorageHealthCheck{StorageKey: "local-default", Status: model.StorageHealthHealthy, LastCheckAt: time.Now()})
	if err != nil {
		t.Fatalf("healthy upsert returned error: %v", err)
	}
	if first.ConsecutiveFailures != 1 || second.ConsecutiveFailures != 2 || healthy.ConsecutiveFailures != 0 {
		t.Fatalf("unexpected failure counters: first=%d second=%d healthy=%d", first.ConsecutiveFailures, second.ConsecutiveFailures, healthy.ConsecutiveFailures)
	}
}
