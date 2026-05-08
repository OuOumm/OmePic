package service

import (
	"context"
	"path/filepath"
	"testing"

	"omepic/backend/internal/model"
	"omepic/backend/internal/repository"
)

func TestDeleteAnnouncementAllowsPublishedPhysicalDelete(t *testing.T) {
	ctx := context.Background()
	repo := newAnnouncementServiceRepository(t)
	announcementService := NewAnnouncementService(repo)

	created, err := announcementService.CreateAnnouncement(ctx, AnnouncementInput{
		Title:    "Published Notice",
		Content:  "Published content",
		Status:   model.AnnouncementStatusPublished,
		Priority: model.AnnouncementPriorityImportant,
	})
	if err != nil {
		t.Fatalf("CreateAnnouncement returned error: %v", err)
	}

	if err := announcementService.DeleteAnnouncement(ctx, created.ID); err != nil {
		t.Fatalf("DeleteAnnouncement returned error for published announcement: %v", err)
	}

	items, err := repo.ListAnnouncements(ctx)
	if err != nil {
		t.Fatalf("ListAnnouncements returned error: %v", err)
	}
	if len(items) != 0 {
		t.Fatalf("expected published announcement to be physically deleted, got %d items", len(items))
	}
}

func newAnnouncementServiceRepository(t *testing.T) *repository.Repository {
	t.Helper()
	dir := t.TempDir()
	repo, err := repository.New(filepath.Join(dir, "announcements.db"))
	if err != nil {
		t.Fatalf("repository.New returned error: %v", err)
	}
	t.Cleanup(func() { _ = repo.Close() })
	if err := repo.Migrate(context.Background()); err != nil {
		t.Fatalf("Migrate returned error: %v", err)
	}
	return repo
}
