package repository

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"omepic/backend/internal/model"
)

func TestIPBanRepositoryActiveBanSemantics(t *testing.T) {
	ctx := context.Background()
	repo := newRepositoryTestHarness(t)

	now := time.Now().UTC()
	expiredAt := now.Add(-time.Hour)
	expired, err := repo.CreateIPBan(ctx, model.IPBan{
		IPHash:    "hash-expired",
		IPAddress: "192.0.2.10",
		Reason:    "expired",
		ExpiresAt: &expiredAt,
	})
	if err != nil {
		t.Fatalf("CreateIPBan expired returned error: %v", err)
	}
	active, err := repo.CreateIPBan(ctx, model.IPBan{
		IPHash:    "hash-active",
		IPAddress: "192.0.2.20",
		Reason:    "active",
	})
	if err != nil {
		t.Fatalf("CreateIPBan active returned error: %v", err)
	}

	if _, err := repo.FindActiveIPBanByHash(ctx, expired.IPHash); err == nil || !IsNotFound(err) {
		t.Fatalf("expected expired ban not found as active, got %v", err)
	}
	found, err := repo.FindActiveIPBanByHash(ctx, active.IPHash)
	if err != nil {
		t.Fatalf("FindActiveIPBanByHash active returned error: %v", err)
	}
	if found.ID != active.ID {
		t.Fatalf("expected active ban ID %d, got %d", active.ID, found.ID)
	}
	count, err := repo.CountActiveIPBans(ctx)
	if err != nil {
		t.Fatalf("CountActiveIPBans returned error: %v", err)
	}
	if count != 1 {
		t.Fatalf("expected one active ban, got %d", count)
	}
	activeByHash, err := repo.ActiveIPBansByHash(ctx)
	if err != nil {
		t.Fatalf("ActiveIPBansByHash returned error: %v", err)
	}
	if _, exists := activeByHash[expired.IPHash]; exists {
		t.Fatalf("expected expired ban to be omitted from active map")
	}
	if activeByHash[active.IPHash].ID != active.ID {
		t.Fatalf("expected active ban in map, got %+v", activeByHash)
	}
}

func TestAbuseRepositoryReturnsRawAggregatesWithoutPresentationFields(t *testing.T) {
	ctx := context.Background()
	repo := newRepositoryTestHarness(t)

	now := time.Now().UTC()
	if err := repo.InsertImage(ctx, repositoryImageRecord("uid-a", "203.0.113.10", "token-a-long-value", 10, now.Add(-2*time.Hour))); err != nil {
		t.Fatalf("InsertImage uid-a returned error: %v", err)
	}
	if err := repo.InsertImage(ctx, repositoryImageRecord("uid-b", "203.0.113.10", "token-a-long-value", 20, now.Add(-time.Hour))); err != nil {
		t.Fatalf("InsertImage uid-b returned error: %v", err)
	}
	if err := repo.InsertImage(ctx, repositoryImageRecord("uid-c", "203.0.113.20", "token-b", 5, now.Add(-30*time.Minute))); err != nil {
		t.Fatalf("InsertImage uid-c returned error: %v", err)
	}

	from := now.Add(-24 * time.Hour)
	to := now.Add(time.Hour)
	count, size, err := repo.AbuseOverviewTotals(ctx, from, to)
	if err != nil {
		t.Fatalf("AbuseOverviewTotals returned error: %v", err)
	}
	if count != 3 || size != 35 {
		t.Fatalf("unexpected totals count=%d size=%d", count, size)
	}

	ips, err := repo.TopAbuseIPAggregates(ctx, from, to, 10)
	if err != nil {
		t.Fatalf("TopAbuseIPAggregates returned error: %v", err)
	}
	if len(ips) != 2 {
		t.Fatalf("expected two IP aggregates, got %+v", ips)
	}
	if ips[0].IPAddress != "203.0.113.10" || ips[0].UploadCount != 2 || ips[0].TotalSize != 30 {
		t.Fatalf("unexpected top IP aggregate: %+v", ips[0])
	}
	if ips[0].LatestUploadAt.IsZero() {
		t.Fatalf("expected latest upload time to be parsed")
	}

	tokens, err := repo.TopAbuseTokenAggregates(ctx, from, to, 10)
	if err != nil {
		t.Fatalf("TopAbuseTokenAggregates returned error: %v", err)
	}
	if len(tokens) != 2 {
		t.Fatalf("expected two token aggregates, got %+v", tokens)
	}
	if tokens[0].Token != "token-a-long-value" || tokens[0].UploadCount != 2 || tokens[0].TotalSize != 30 {
		t.Fatalf("unexpected top token aggregate: %+v", tokens[0])
	}
}

func newRepositoryTestHarness(t *testing.T) *Repository {
	t.Helper()
	repo, err := New(filepath.Join(t.TempDir(), "test.sqlite"))
	if err != nil {
		t.Fatalf("New returned error: %v", err)
	}
	t.Cleanup(func() {
		_ = repo.Close()
	})
	if err := repo.Migrate(context.Background()); err != nil {
		t.Fatalf("Migrate returned error: %v", err)
	}
	return repo
}

func repositoryImageRecord(uid string, ipAddress string, token string, size int64, createdAt time.Time) model.ImageRecord {
	return model.ImageRecord{
		UID:            uid,
		Token:          token,
		StorageKey:     "local-default",
		StorageBackend: "local",
		FilePath:       "2026/05/" + uid + ".avif",
		MIMEType:       "image/avif",
		Size:           size,
		MD5Hash:        "hash-" + uid,
		IPAddress:      ipAddress,
		CreatedAt:      createdAt,
	}
}
