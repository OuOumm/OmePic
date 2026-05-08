package model

import "time"

const (
	AnnouncementStatusDraft     = "draft"
	AnnouncementStatusPublished = "published"
	AnnouncementStatusArchived  = "archived"

	AnnouncementPriorityNormal    = "normal"
	AnnouncementPriorityImportant = "important"
	AnnouncementPriorityUrgent    = "urgent"
)

type Announcement struct {
	ID        int64      `json:"id"`
	Title     string     `json:"title"`
	Content   string     `json:"content"`
	Status    string     `json:"status"`
	Priority  string     `json:"priority"`
	StartsAt  *time.Time `json:"starts_at"`
	EndsAt    *time.Time `json:"ends_at"`
	SortOrder int        `json:"sort_order"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
}
