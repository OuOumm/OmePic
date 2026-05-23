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
	if check.LastCheckAt.IsZero() {
		check.LastCheckAt = time.Now().UTC()
	}
	if check.CreatedAt.IsZero() {
		check.CreatedAt = check.LastCheckAt
	}
	check.UpdatedAt = check.LastCheckAt

	_, err = r.db.ExecContext(ctx, `INSERT INTO storage_health_checks(
		storage_key, status, last_check_at, latency_ms, error_message, consecutive_failures, created_at, updated_at
	) VALUES(?, ?, ?, ?, ?, ?, ?, ?)
	ON CONFLICT(storage_key) DO UPDATE SET
		status = excluded.status,
		last_check_at = excluded.last_check_at,
		latency_ms = excluded.latency_ms,
		error_message = excluded.error_message,
		consecutive_failures = excluded.consecutive_failures,
		updated_at = excluded.updated_at`,
		check.StorageKey,
		check.Status,
		formatTime(check.LastCheckAt),
		check.LatencyMS,
		check.ErrorMessage,
		check.ConsecutiveFailures,
		formatTime(check.CreatedAt),
		formatTime(check.UpdatedAt),
	)
	if err != nil {
		return model.StorageHealthCheck{}, err
	}
	return r.GetStorageHealthCheck(ctx, check.StorageKey)
}

func (r *Repository) GetStorageHealthCheck(ctx context.Context, storageKey string) (model.StorageHealthCheck, error) {
	row := r.db.QueryRowContext(ctx, `SELECT id, storage_key, status, last_check_at, latency_ms, error_message, consecutive_failures, created_at, updated_at FROM storage_health_checks WHERE storage_key = ?`, storageKey)
	return scanStorageHealthCheck(row)
}

func (r *Repository) ListStorageHealthChecks(ctx context.Context) ([]model.StorageHealthCheck, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT id, storage_key, status, last_check_at, latency_ms, error_message, consecutive_failures, created_at, updated_at FROM storage_health_checks ORDER BY storage_key ASC`)
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
	var lastCheckAt, createdAt, updatedAt string
	if err := scanner.Scan(&check.ID, &check.StorageKey, &check.Status, &lastCheckAt, &check.LatencyMS, &check.ErrorMessage, &check.ConsecutiveFailures, &createdAt, &updatedAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.StorageHealthCheck{}, sql.ErrNoRows
		}
		return model.StorageHealthCheck{}, err
	}
	check.LastCheckAt = parseTime(lastCheckAt)
	check.CreatedAt = parseTime(createdAt)
	check.UpdatedAt = parseTime(updatedAt)
	return check, nil
}

func formatTime(value time.Time) string {
	return value.UTC().Format(time.RFC3339Nano)
}
