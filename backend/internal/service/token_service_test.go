package service

import (
	"context"
	"errors"
	"path/filepath"
	"testing"
	"time"

	"omepic/backend/internal/repository"
)

func TestTokenHashAndPreviewDoNotExposePlaintext(t *testing.T) {
	token := "secret-token-123"
	hash := TokenHash(token)
	if len(hash) != 64 || hash == token {
		t.Fatalf("unexpected hash %q", hash)
	}
	if got := TokenPreview(token); got != "************-123" {
		t.Fatalf("unexpected preview %q", got)
	}
}

func TestTokenServiceRecordDisableEnable(t *testing.T) {
	ctx := context.Background()
	repo := newTokenTestRepo(t)
	service := NewTokenService(repo)
	token := "token-service-secret"
	hash := TokenHash(token)

	if err := service.RecordUpload(ctx, token, 123, "203.0.113.9", time.Date(2026, 5, 24, 1, 2, 3, 0, time.UTC)); err != nil {
		t.Fatalf("RecordUpload returned error: %v", err)
	}
	list, err := service.List(ctx)
	if err != nil {
		t.Fatalf("List returned error: %v", err)
	}
	if len(list.Items) != 1 || list.Items[0].TokenHash != hash || list.Items[0].UploadCount != 1 || list.Items[0].TotalBytes != 123 || list.Items[0].LastIP != "203.0.113.9" {
		t.Fatalf("unexpected token list: %+v", list.Items)
	}

	if err := service.Disable(ctx, hash, "abuse"); err != nil {
		t.Fatalf("Disable returned error: %v", err)
	}
	if err := service.EnsureTokenAllowed(ctx, token); !errors.Is(err, ErrForbidden) {
		t.Fatalf("expected disabled token forbidden, got %v", err)
	}
	if err := service.Enable(ctx, hash); err != nil {
		t.Fatalf("Enable returned error: %v", err)
	}
	if err := service.EnsureTokenAllowed(ctx, token); err != nil {
		t.Fatalf("expected token allowed after enable, got %v", err)
	}
}

func newTokenTestRepo(t *testing.T) *repository.Repository {
	t.Helper()
	repo, err := repository.New(filepath.Join(t.TempDir(), "test.sqlite"))
	if err != nil {
		t.Fatalf("repository.New returned error: %v", err)
	}
	t.Cleanup(func() { _ = repo.Close() })
	if err := repo.Migrate(context.Background()); err != nil {
		t.Fatalf("Migrate returned error: %v", err)
	}
	return repo
}
