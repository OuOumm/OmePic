package repository

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"omepic/backend/internal/model"
)

func (r *Repository) UpsertStorageHealthCheck(ctx context.Context, check model.StorageHealthCheck) (model.StorageHealthCheck, error) {
	previous, err := r.GetStorageHealthCheck(ctx, check.StorageKey)
	if err != nil && !IsNotFound(err) {
		return model.StorageHealthCheck{}, err
	}
	if check.ConsecutiveFailures == 0 {
		if check.Status == model.StorageHealthHealthy {
			check.ConsecutiveFailures = 0
		} else if !IsNotFound(err) {
			check.ConsecutiveFailures = previous.ConsecutiveFailures + 1
		} else {
			check.ConsecutiveFailures = 1
		}
	}
	now := time.Now().UTC()
	if check.CreatedAt.IsZero() {
		check.CreatedAt = now
	}
	if check.UpdatedAt.IsZero() {
		check.UpdatedAt = now
	}

	result, err := r.db.ExecContext(ctx, `INSERT INTO storage_health_checks(
		storage_key, status, latency_ms, error_message, consecutive_failures, created_at, updated_at
	) VALUES(?, ?, ?, ?, ?, ?, ?)`,
		check.StorageKey,
		check.Status,
		check.LatencyMS,
		check.ErrorMessage,
		check.ConsecutiveFailures,
		formatTime(check.CreatedAt),
		formatTime(check.UpdatedAt),
	)
	if err != nil {
		return model.StorageHealthCheck{}, err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return model.StorageHealthCheck{}, err
	}
	return r.GetStorageHealthCheckByID(ctx, id)
}

func (r *Repository) GetStorageHealthCheck(ctx context.Context, storageKey string) (model.StorageHealthCheck, error) {
	row := r.db.QueryRowContext(ctx, `SELECT id, storage_key, status, latency_ms, error_message, consecutive_failures, created_at, updated_at FROM storage_health_checks WHERE storage_key = ? ORDER BY updated_at DESC, id DESC LIMIT 1`, storageKey)
	return scanStorageHealthCheck(row)
}

func (r *Repository) GetStorageHealthCheckByID(ctx context.Context, id int64) (model.StorageHealthCheck, error) {
	row := r.db.QueryRowContext(ctx, `SELECT id, storage_key, status, latency_ms, error_message, consecutive_failures, created_at, updated_at FROM storage_health_checks WHERE id = ?`, id)
	return scanStorageHealthCheck(row)
}

func (r *Repository) ListStorageHealthChecks(ctx context.Context) ([]model.StorageHealthCheck, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT h.id, h.storage_key, h.status, h.latency_ms, h.error_message, h.consecutive_failures, h.created_at, h.updated_at
		FROM storage_health_checks h
		WHERE h.id IN (SELECT MAX(id) FROM storage_health_checks GROUP BY storage_key)
		ORDER BY h.storage_key ASC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	checks := make([]model.StorageHealthCheck, 0)
	for rows.Next() {
		check, err := scanStorageHealthCheck(rows)
		if err != nil {
			return nil, err
		}
		checks = append(checks, check)
	}
	return checks, rows.Err()
}

func (r *Repository) ListStorageHealthHistory(ctx context.Context, storageKey string, since time.Time) ([]model.StorageHealthCheck, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT id, storage_key, status, latency_ms, error_message, consecutive_failures, created_at, updated_at
		FROM storage_health_checks
		WHERE storage_key = ? AND updated_at >= ?
		ORDER BY updated_at ASC, id ASC`, storageKey, formatTime(since))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	checks := make([]model.StorageHealthCheck, 0)
	for rows.Next() {
		check, err := scanStorageHealthCheck(rows)
		if err != nil {
			return nil, err
		}
		checks = append(checks, check)
	}
	return checks, rows.Err()
}

func scanStorageHealthCheck(scanner interface{ Scan(dest ...any) error }) (model.StorageHealthCheck, error) {
	var check model.StorageHealthCheck
	var createdAt, updatedAt string
	if err := scanner.Scan(&check.ID, &check.StorageKey, &check.Status, &check.LatencyMS, &check.ErrorMessage, &check.ConsecutiveFailures, &createdAt, &updatedAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.StorageHealthCheck{}, sql.ErrNoRows
		}
		return model.StorageHealthCheck{}, err
	}
	check.CreatedAt = parseTime(createdAt)
	check.UpdatedAt = parseTime(updatedAt)
	return check, nil
}

func formatTime(value time.Time) string {
	return value.UTC().Format(time.RFC3339Nano)
}
