package service

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"omepic/backend/internal/model"
	"omepic/backend/internal/repository"
)

const publicAnnouncementLimit = 10

type AnnouncementView struct {
	ID        int64      `json:"id"`
	Title     string     `json:"title"`
	Content   string     `json:"content"`
	Status    string     `json:"status,omitempty"`
	Priority  string     `json:"priority"`
	StartsAt  *time.Time `json:"starts_at"`
	EndsAt    *time.Time `json:"ends_at"`
	SortOrder int        `json:"sort_order,omitempty"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
}

type AnnouncementListView struct {
	Items []AnnouncementView `json:"items"`
}

type AnnouncementInput struct {
	Title     string  `json:"title"`
	Content   string  `json:"content"`
	Status    string  `json:"status"`
	Priority  string  `json:"priority"`
	StartsAt  *string `json:"starts_at"`
	EndsAt    *string `json:"ends_at"`
	SortOrder int     `json:"sort_order"`
}

type AnnouncementService struct {
	repo *repository.Repository
}

func NewAnnouncementService(repo *repository.Repository) *AnnouncementService {
	return &AnnouncementService{repo: repo}
}

func (s *AnnouncementService) PublicAnnouncements(ctx context.Context) (AnnouncementListView, error) {
	items, err := s.repo.ListPublicAnnouncements(ctx, publicAnnouncementLimit)
	if err != nil {
		return AnnouncementListView{}, fmt.Errorf("%w: announcement query failed", ErrDependencyUnavailable)
	}
	return AnnouncementListView{Items: announcementViews(items, false)}, nil
}

func (s *AnnouncementService) AdminAnnouncements(ctx context.Context) (AnnouncementListView, error) {
	items, err := s.repo.ListAnnouncements(ctx)
	if err != nil {
		return AnnouncementListView{}, fmt.Errorf("%w: announcement query failed", ErrDependencyUnavailable)
	}
	return AnnouncementListView{Items: announcementViews(items, true)}, nil
}

func (s *AnnouncementService) CreateAnnouncement(ctx context.Context, input AnnouncementInput) (AnnouncementView, error) {
	announcement, err := announcementFromInput(input)
	if err != nil {
		return AnnouncementView{}, err
	}
	created, err := s.repo.CreateAnnouncement(ctx, announcement)
	if err != nil {
		return AnnouncementView{}, fmt.Errorf("%w: announcement create failed", ErrDependencyUnavailable)
	}
	return announcementView(created, true), nil
}

func (s *AnnouncementService) UpdateAnnouncement(ctx context.Context, id int64, input AnnouncementInput) (AnnouncementView, error) {
	if id <= 0 {
		return AnnouncementView{}, WithUserMessage(ErrInvalidInput, "announcement id is required")
	}
	announcement, err := announcementFromInput(input)
	if err != nil {
		return AnnouncementView{}, err
	}
	announcement.ID = id
	updated, err := s.repo.UpdateAnnouncement(ctx, announcement)
	if err != nil {
		if err == sql.ErrNoRows {
			return AnnouncementView{}, ErrNotFound
		}
		return AnnouncementView{}, fmt.Errorf("%w: announcement update failed", ErrDependencyUnavailable)
	}
	return announcementView(updated, true), nil
}

func (s *AnnouncementService) DeleteAnnouncement(ctx context.Context, id int64) error {
	if id <= 0 {
		return WithUserMessage(ErrInvalidInput, "announcement id is required")
	}
	if err := s.repo.DeleteAnnouncement(ctx, id); err != nil {
		if err == sql.ErrNoRows {
			return ErrNotFound
		}
		return fmt.Errorf("%w: announcement delete failed", ErrDependencyUnavailable)
	}
	return nil
}

func (s *AnnouncementService) ArchiveAnnouncement(ctx context.Context, id int64) (AnnouncementView, error) {
	if id <= 0 {
		return AnnouncementView{}, WithUserMessage(ErrInvalidInput, "announcement id is required")
	}
	archived, err := s.repo.ArchiveAnnouncement(ctx, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return AnnouncementView{}, ErrNotFound
		}
		return AnnouncementView{}, fmt.Errorf("%w: announcement archive failed", ErrDependencyUnavailable)
	}
	return announcementView(archived, true), nil
}

func announcementFromInput(input AnnouncementInput) (model.Announcement, error) {
	title := strings.TrimSpace(input.Title)
	content := strings.TrimSpace(input.Content)
	status := normalizeAnnouncementStatus(input.Status)
	priority := normalizeAnnouncementPriority(input.Priority)
	if title == "" {
		return model.Announcement{}, WithUserMessage(ErrInvalidInput, "title is required")
	}
	if content == "" {
		return model.Announcement{}, WithUserMessage(ErrInvalidInput, "content is required")
	}
	startsAt, err := parseAnnouncementTime(input.StartsAt)
	if err != nil {
		return model.Announcement{}, err
	}
	endsAt, err := parseAnnouncementTime(input.EndsAt)
	if err != nil {
		return model.Announcement{}, err
	}
	if startsAt != nil && endsAt != nil && !endsAt.After(*startsAt) {
		return model.Announcement{}, WithUserMessage(ErrInvalidInput, "end time must be after start time")
	}
	return model.Announcement{
		Title:     title,
		Content:   content,
		Status:    status,
		Priority:  priority,
		StartsAt:  startsAt,
		EndsAt:    endsAt,
		SortOrder: input.SortOrder,
	}, nil
}

func normalizeAnnouncementStatus(value string) string {
	switch strings.TrimSpace(strings.ToLower(value)) {
	case model.AnnouncementStatusPublished:
		return model.AnnouncementStatusPublished
	case model.AnnouncementStatusArchived:
		return model.AnnouncementStatusArchived
	default:
		return model.AnnouncementStatusDraft
	}
}

func normalizeAnnouncementPriority(value string) string {
	switch strings.TrimSpace(strings.ToLower(value)) {
	case model.AnnouncementPriorityImportant:
		return model.AnnouncementPriorityImportant
	case model.AnnouncementPriorityUrgent:
		return model.AnnouncementPriorityUrgent
	default:
		return model.AnnouncementPriorityNormal
	}
}

func parseAnnouncementTime(value *string) (*time.Time, error) {
	if value == nil || strings.TrimSpace(*value) == "" {
		return nil, nil
	}
	raw := strings.TrimSpace(*value)
	layouts := []string{time.RFC3339Nano, time.RFC3339, "2006-01-02T15:04", "2006-01-02 15:04:05"}
	for _, layout := range layouts {
		if parsed, err := time.Parse(layout, raw); err == nil {
			utc := parsed.UTC()
			return &utc, nil
		}
	}
	return nil, WithUserMessage(ErrInvalidInput, "announcement time is invalid")
}

func announcementViews(items []model.Announcement, includeStatus bool) []AnnouncementView {
	views := make([]AnnouncementView, 0, len(items))
	for _, item := range items {
		views = append(views, announcementView(item, includeStatus))
	}
	return views
}

func announcementView(item model.Announcement, includeStatus bool) AnnouncementView {
	view := AnnouncementView{
		ID:        item.ID,
		Title:     item.Title,
		Content:   item.Content,
		Priority:  item.Priority,
		StartsAt:  item.StartsAt,
		EndsAt:    item.EndsAt,
		SortOrder: item.SortOrder,
		CreatedAt: item.CreatedAt,
		UpdatedAt: item.UpdatedAt,
	}
	if includeStatus {
		view.Status = item.Status
	}
	return view
}
