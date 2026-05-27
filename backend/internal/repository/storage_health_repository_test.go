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
	for _, column := range []string{"id", "storage_key", "status", "latency_ms", "error_message", "consecutive_failures", "created_at", "updated_at"} {
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

	first, err := repo.UpsertStorageHealthCheck(ctx, model.StorageHealthCheck{StorageKey: "local-default", Status: model.StorageHealthUnavailable, ErrorMessage: "boom"})
	if err != nil {
		t.Fatalf("first upsert returned error: %v", err)
	}
	second, err := repo.UpsertStorageHealthCheck(ctx, model.StorageHealthCheck{StorageKey: "local-default", Status: model.StorageHealthUnavailable, ErrorMessage: "boom"})
	if err != nil {
		t.Fatalf("second upsert returned error: %v", err)
	}
	healthy, err := repo.UpsertStorageHealthCheck(ctx, model.StorageHealthCheck{StorageKey: "local-default", Status: model.StorageHealthHealthy})
	if err != nil {
		t.Fatalf("healthy upsert returned error: %v", err)
	}
	if first.ConsecutiveFailures != 1 || second.ConsecutiveFailures != 2 || healthy.ConsecutiveFailures != 0 {
		t.Fatalf("unexpected failure counters: first=%d second=%d healthy=%d", first.ConsecutiveFailures, second.ConsecutiveFailures, healthy.ConsecutiveFailures)
	}
}

func TestStorageHealthHistoryKeepsRecordsAndLatestList(t *testing.T) {
	ctx := context.Background()
	repo, err := New(filepath.Join(t.TempDir(), "test.sqlite"))
	if err != nil {
		t.Fatalf("New returned error: %v", err)
	}
	t.Cleanup(func() { _ = repo.Close() })
	if err := repo.Migrate(ctx); err != nil {
		t.Fatalf("Migrate returned error: %v", err)
	}

	now := time.Now().UTC()
	oldCheck, err := repo.UpsertStorageHealthCheck(ctx, model.StorageHealthCheck{StorageKey: "local-default", Status: model.StorageHealthHealthy, CreatedAt: now.Add(-25 * time.Hour), UpdatedAt: now.Add(-25 * time.Hour), LatencyMS: 10})
	if err != nil {
		t.Fatalf("old upsert returned error: %v", err)
	}
	recentCheck, err := repo.UpsertStorageHealthCheck(ctx, model.StorageHealthCheck{StorageKey: "local-default", Status: model.StorageHealthUnavailable, CreatedAt: now.Add(-1 * time.Hour), LatencyMS: 20, ErrorMessage: "boom"})
	if err != nil {
		t.Fatalf("recent upsert returned error: %v", err)
	}
	otherCheck, err := repo.UpsertStorageHealthCheck(ctx, model.StorageHealthCheck{StorageKey: "archive", Status: model.StorageHealthHealthy, CreatedAt: now.Add(-30 * time.Minute), LatencyMS: 5})
	if err != nil {
		t.Fatalf("other upsert returned error: %v", err)
	}
	if oldCheck.ID == recentCheck.ID || recentCheck.ID == otherCheck.ID {
		t.Fatalf("expected append-only health records, got ids old=%d recent=%d other=%d", oldCheck.ID, recentCheck.ID, otherCheck.ID)
	}

	history, err := repo.ListStorageHealthHistory(ctx, "local-default", now.Add(-24*time.Hour))
	if err != nil {
		t.Fatalf("ListStorageHealthHistory returned error: %v", err)
	}
	if len(history) != 1 || history[0].ID != recentCheck.ID {
		t.Fatalf("expected only recent local-default record, got %+v", history)
	}

	latest, err := repo.ListStorageHealthChecks(ctx)
	if err != nil {
		t.Fatalf("ListStorageHealthChecks returned error: %v", err)
	}
	if len(latest) != 2 {
		t.Fatalf("expected latest record per storage key, got %+v", latest)
	}
}
