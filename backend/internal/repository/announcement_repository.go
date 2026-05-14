package repository

import (
	"context"
	"database/sql"
	"time"

	"omepic/backend/internal/model"
)

func (r *Repository) ListPublicAnnouncements(ctx context.Context, limit int) ([]model.Announcement, error) {
	if limit < 1 {
		limit = 10
	}
	if limit > 50 {
		limit = 50
	}
	now := time.Now().UTC().Format(time.RFC3339)
	rows, err := r.db.QueryContext(
		ctx,
		`SELECT id, title, content, status, priority, starts_at, ends_at, sort_order, created_at, updated_at
		 FROM announcements
		 WHERE status = ?
		   AND (starts_at IS NULL OR starts_at = '' OR starts_at <= ?)
		   AND (ends_at IS NULL OR ends_at = '' OR ends_at > ?)
		 ORDER BY CASE priority WHEN 'urgent' THEN 3 WHEN 'important' THEN 2 ELSE 1 END DESC, sort_order DESC, created_at DESC, id DESC
		 LIMIT ?`,
		model.AnnouncementStatusPublished,
		now,
		now,
		limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanAnnouncements(rows)
}

func (r *Repository) ListAnnouncements(ctx context.Context) ([]model.Announcement, error) {
	rows, err := r.db.QueryContext(
		ctx,
		`SELECT id, title, content, status, priority, starts_at, ends_at, sort_order, created_at, updated_at
		 FROM announcements
		 ORDER BY updated_at DESC, id DESC`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanAnnouncements(rows)
}

func (r *Repository) GetAnnouncement(ctx context.Context, id int64) (model.Announcement, error) {
	row := r.db.QueryRowContext(
		ctx,
		`SELECT id, title, content, status, priority, starts_at, ends_at, sort_order, created_at, updated_at
		 FROM announcements
		 WHERE id = ?`,
		id,
	)
	return scanAnnouncement(row)
}

func (r *Repository) CreateAnnouncement(ctx context.Context, announcement model.Announcement) (model.Announcement, error) {
	now := time.Now().UTC()
	if announcement.CreatedAt.IsZero() {
		announcement.CreatedAt = now
	}
	announcement.UpdatedAt = now
	result, err := r.db.ExecContext(
		ctx,
		`INSERT INTO announcements(title, content, status, priority, starts_at, ends_at, sort_order, created_at, updated_at)
		 VALUES(?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		announcement.Title,
		announcement.Content,
		announcement.Status,
		announcement.Priority,
		nullableTimeString(announcement.StartsAt),
		nullableTimeString(announcement.EndsAt),
		announcement.SortOrder,
		announcement.CreatedAt.Format(time.RFC3339),
		announcement.UpdatedAt.Format(time.RFC3339),
	)
	if err != nil {
		return model.Announcement{}, err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return model.Announcement{}, err
	}
	return r.GetAnnouncement(ctx, id)
}

func (r *Repository) UpdateAnnouncement(ctx context.Context, announcement model.Announcement) (model.Announcement, error) {
	announcement.UpdatedAt = time.Now().UTC()
	result, err := r.db.ExecContext(
		ctx,
		`UPDATE announcements
		 SET title = ?, content = ?, status = ?, priority = ?, starts_at = ?, ends_at = ?, sort_order = ?, updated_at = ?
		 WHERE id = ?`,
		announcement.Title,
		announcement.Content,
		announcement.Status,
		announcement.Priority,
		nullableTimeString(announcement.StartsAt),
		nullableTimeString(announcement.EndsAt),
		announcement.SortOrder,
		announcement.UpdatedAt.Format(time.RFC3339),
		announcement.ID,
	)
	if err != nil {
		return model.Announcement{}, err
	}
	if err := ensureRowsAffected(result); err != nil {
		return model.Announcement{}, err
	}
	return r.GetAnnouncement(ctx, announcement.ID)
}

func (r *Repository) DeleteAnnouncement(ctx context.Context, id int64) error {
	result, err := r.db.ExecContext(ctx, `DELETE FROM announcements WHERE id = ?`, id)
	if err != nil {
		return err
	}
	if err := ensureRowsAffected(result); err != nil {
		return err
	}
	return nil
}

func (r *Repository) ArchiveAnnouncement(ctx context.Context, id int64) (model.Announcement, error) {
	result, err := r.db.ExecContext(
		ctx,
		`UPDATE announcements SET status = ?, updated_at = ? WHERE id = ?`,
		model.AnnouncementStatusArchived,
		time.Now().UTC().Format(time.RFC3339),
		id,
	)
	if err != nil {
		return model.Announcement{}, err
	}
	if err := ensureRowsAffected(result); err != nil {
		return model.Announcement{}, err
	}
	return r.GetAnnouncement(ctx, id)
}

func scanAnnouncement(scanner interface{ Scan(dest ...any) error }) (model.Announcement, error) {
	var record model.Announcement
	var startsAt sql.NullString
	var endsAt sql.NullString
	var createdAt string
	var updatedAt string
	err := scanner.Scan(
		&record.ID,
		&record.Title,
		&record.Content,
		&record.Status,
		&record.Priority,
		&startsAt,
		&endsAt,
		&record.SortOrder,
		&createdAt,
		&updatedAt,
	)
	if err != nil {
		return model.Announcement{}, err
	}
	record.StartsAt = parseNullableTime(startsAt)
	record.EndsAt = parseNullableTime(endsAt)
	record.CreatedAt = parseTime(createdAt)
	record.UpdatedAt = parseTime(updatedAt)
	return record, nil
}

func scanAnnouncements(rows *sql.Rows) ([]model.Announcement, error) {
	var records []model.Announcement
	for rows.Next() {
		record, err := scanAnnouncement(rows)
		if err != nil {
			return nil, err
		}
		records = append(records, record)
	}
	return records, rows.Err()
}
