package repository

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"time"

	"omepic/backend/internal/iputil"
)

func (r *Repository) countByQuery(ctx context.Context, query string, args ...any) (int64, error) {
	var count int64
	if err := r.db.QueryRowContext(ctx, query, args...).Scan(&count); err != nil {
		return 0, err
	}
	return count, nil
}

func ensureRowsAffected(result sql.Result) error {
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func parseTime(value string) time.Time {
	if value == "" {
		return time.Time{}
	}
	layouts := []string{time.RFC3339Nano, time.RFC3339, "2006-01-02 15:04:05"}
	for _, layout := range layouts {
		if parsed, err := time.Parse(layout, value); err == nil {
			return parsed.UTC()
		}
	}
	return time.Time{}
}

func parseNullableTime(value sql.NullString) *time.Time {
	if !value.Valid || strings.TrimSpace(value.String) == "" {
		return nil
	}
	parsed := parseTime(value.String)
	if parsed.IsZero() {
		return nil
	}
	return &parsed
}

func nullableTimeString(value *time.Time) any {
	if value == nil || value.IsZero() {
		return nil
	}
	return value.UTC().Format(time.RFC3339)
}

func ipHashValue(ipAddress string) string {
	return iputil.Hash(ipAddress)
}

func maskIPValue(ipAddress string) string {
	return iputil.Mask(ipAddress)
}

func previewValue(value string, max int) string {
	trimmed := strings.TrimSpace(value)
	if max < 1 || len(trimmed) <= max {
		return trimmed
	}
	return trimmed[:max] + "..."
}

func boolString(value bool) string {
	if value {
		return "true"
	}
	return "false"
}

func boolInt(value bool) int {
	if value {
		return 1
	}
	return 0
}

func parseBool(value string) bool {
	return value == "true" || value == "1" || value == "yes"
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}
	return ""
}

func IsNotFound(err error) bool {
	return errors.Is(err, sql.ErrNoRows)
}
