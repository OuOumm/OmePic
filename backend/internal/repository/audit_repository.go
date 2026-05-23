package repository

import (
	"context"
	"fmt"
	"time"

	"omepic/backend/internal/model"
)

type ConfigAuditLogFilter struct {
	Scope    string
	Page     int
	PageSize int
}

type ConfigAuditLogList struct {
	Items    []model.ConfigAuditLog
	Page     int
	PageSize int
	Total    int64
}

func (r *Repository) CreateConfigAuditLog(ctx context.Context, log model.ConfigAuditLog) error {
	createdAt := log.CreatedAt.UTC()
	if createdAt.IsZero() {
		createdAt = time.Now().UTC()
	}
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO config_audit_logs(actor, actor_ip, config_scope, before_snapshot, after_snapshot, created_at)
		 VALUES (?, ?, ?, ?, ?, ?)`,
		log.Actor,
		log.ActorIP,
		log.ConfigScope,
		log.BeforeSnapshot,
		log.AfterSnapshot,
		createdAt.Format(time.RFC3339),
	)
	return err
}

func (r *Repository) ListConfigAuditLogs(ctx context.Context, filter ConfigAuditLogFilter) (ConfigAuditLogList, error) {
	page := normalizePage(filter.Page)
	pageSize := normalizePageSize(filter.PageSize)
	where := ""
	args := []any{}
	if filter.Scope != "" {
		where = " WHERE config_scope = ?"
		args = append(args, filter.Scope)
	}

	var total int64
	if err := r.db.QueryRowContext(ctx, `SELECT COUNT(1) FROM config_audit_logs`+where, args...).Scan(&total); err != nil {
		return ConfigAuditLogList{}, err
	}

	queryArgs := append([]any{}, args...)
	queryArgs = append(queryArgs, pageSize, (page-1)*pageSize)
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, actor, actor_ip, config_scope, before_snapshot, after_snapshot, created_at
		 FROM config_audit_logs`+where+`
		 ORDER BY created_at DESC, id DESC
		 LIMIT ? OFFSET ?`,
		queryArgs...,
	)
	if err != nil {
		return ConfigAuditLogList{}, err
	}
	defer rows.Close()

	items := make([]model.ConfigAuditLog, 0)
	for rows.Next() {
		item, err := scanConfigAuditLog(rows)
		if err != nil {
			return ConfigAuditLogList{}, err
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return ConfigAuditLogList{}, err
	}

	return ConfigAuditLogList{Items: items, Page: page, PageSize: pageSize, Total: total}, nil
}

func scanConfigAuditLog(scanner interface{ Scan(dest ...any) error }) (model.ConfigAuditLog, error) {
	var item model.ConfigAuditLog
	var createdAt string
	if err := scanner.Scan(
		&item.ID,
		&item.Actor,
		&item.ActorIP,
		&item.ConfigScope,
		&item.BeforeSnapshot,
		&item.AfterSnapshot,
		&createdAt,
	); err != nil {
		return model.ConfigAuditLog{}, err
	}
	parsed, err := time.Parse(time.RFC3339, createdAt)
	if err != nil {
		parsed, err = time.Parse("2006-01-02 15:04:05", createdAt)
		if err != nil {
			return model.ConfigAuditLog{}, fmt.Errorf("parse audit created_at: %w", err)
		}
	}
	item.CreatedAt = parsed
	return item, nil
}

func normalizePage(page int) int {
	if page < 1 {
		return 1
	}
	return page
}

func normalizePageSize(pageSize int) int {
	if pageSize < 1 {
		return 20
	}
	if pageSize > 100 {
		return 100
	}
	return pageSize
}
