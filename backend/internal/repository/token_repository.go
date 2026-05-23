package repository

import (
	"context"
	"time"

	"omepic/backend/internal/model"
)

func (r *Repository) UpsertTokenUsage(ctx context.Context, usage model.TokenUsage) error {
	if usage.LastUsedAt.IsZero() {
		usage.LastUsedAt = time.Now().UTC()
	}
	_, err := r.db.ExecContext(ctx, `INSERT INTO token_usage(
		token_hash, token_preview, upload_count, total_bytes, last_ip, last_used_at, created_at, updated_at
	) VALUES(?, ?, 1, ?, ?, ?, ?, ?)
	ON CONFLICT(token_hash) DO UPDATE SET
		token_preview = excluded.token_preview,
		upload_count = token_usage.upload_count + 1,
		total_bytes = token_usage.total_bytes + excluded.total_bytes,
		last_ip = excluded.last_ip,
		last_used_at = excluded.last_used_at,
		updated_at = excluded.updated_at`,
		usage.TokenHash,
		usage.TokenPreview,
		usage.TotalBytes,
		usage.LastIP,
		usage.LastUsedAt.UTC().Format(time.RFC3339),
		usage.LastUsedAt.UTC().Format(time.RFC3339),
		usage.LastUsedAt.UTC().Format(time.RFC3339),
	)
	return err
}

func (r *Repository) FindTokenControl(ctx context.Context, tokenHash string) (*model.TokenControl, error) {
	row := r.db.QueryRowContext(ctx, `SELECT token_hash, token_preview, disabled, reason, disabled_at, created_at, updated_at FROM token_controls WHERE token_hash = ?`, tokenHash)
	control, err := scanTokenControl(row)
	if err != nil {
		return nil, err
	}
	return &control, nil
}

func (r *Repository) SetTokenDisabled(ctx context.Context, tokenHash string, tokenPreview string, reason string, disabled bool) error {
	now := time.Now().UTC()
	disabledAt := ""
	if disabled {
		disabledAt = now.UTC().Format(time.RFC3339)
	}
	_, err := r.db.ExecContext(ctx, `INSERT INTO token_controls(
		token_hash, token_preview, disabled, reason, disabled_at, created_at, updated_at
	) VALUES(?, ?, ?, ?, ?, ?, ?)
	ON CONFLICT(token_hash) DO UPDATE SET
		token_preview = CASE WHEN excluded.token_preview != '' THEN excluded.token_preview ELSE token_controls.token_preview END,
		disabled = excluded.disabled,
		reason = excluded.reason,
		disabled_at = excluded.disabled_at,
		updated_at = excluded.updated_at`,
		tokenHash,
		tokenPreview,
		boolToInt(disabled),
		reason,
		disabledAt,
		now.UTC().Format(time.RFC3339),
		now.UTC().Format(time.RFC3339),
	)
	return err
}

func (r *Repository) ListTokenGovernance(ctx context.Context) ([]model.TokenGovernanceEntry, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT
		COALESCE(u.token_hash, c.token_hash) AS token_hash,
		COALESCE(NULLIF(u.token_preview, ''), c.token_preview, '') AS token_preview,
		COALESCE(u.upload_count, 0) AS upload_count,
		COALESCE(u.total_bytes, 0) AS total_bytes,
		COALESCE(u.last_ip, '') AS last_ip,
		COALESCE(u.last_used_at, '') AS last_used_at,
		COALESCE(c.disabled, 0) AS disabled,
		COALESCE(c.reason, '') AS reason,
		COALESCE(c.disabled_at, '') AS disabled_at,
		COALESCE(u.created_at, c.created_at, '') AS created_at,
		COALESCE(MAX(u.updated_at, c.updated_at), u.updated_at, c.updated_at, '') AS updated_at
	FROM token_usage u
	LEFT JOIN token_controls c ON c.token_hash = u.token_hash
	UNION ALL
	SELECT
		c.token_hash,
		c.token_preview,
		0,
		0,
		'',
		'',
		c.disabled,
		c.reason,
		COALESCE(c.disabled_at, ''),
		c.created_at,
		c.updated_at
	FROM token_controls c
	LEFT JOIN token_usage u ON u.token_hash = c.token_hash
	WHERE u.token_hash IS NULL
	ORDER BY updated_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var entries []model.TokenGovernanceEntry
	for rows.Next() {
		entry, err := scanTokenGovernanceEntry(rows)
		if err != nil {
			return nil, err
		}
		entries = append(entries, entry)
	}
	return entries, rows.Err()
}

func scanTokenControl(scanner interface{ Scan(dest ...any) error }) (model.TokenControl, error) {
	var control model.TokenControl
	var disabled int
	var disabledAt, createdAt, updatedAt string
	err := scanner.Scan(&control.TokenHash, &control.TokenPreview, &disabled, &control.Reason, &disabledAt, &createdAt, &updatedAt)
	if err != nil {
		return model.TokenControl{}, err
	}
	control.Disabled = disabled != 0
	control.DisabledAt = parseTime(disabledAt)
	control.CreatedAt = parseTime(createdAt)
	control.UpdatedAt = parseTime(updatedAt)
	return control, nil
}

func scanTokenGovernanceEntry(scanner interface{ Scan(dest ...any) error }) (model.TokenGovernanceEntry, error) {
	var entry model.TokenGovernanceEntry
	var disabled int
	var lastUsedAt, disabledAt, createdAt, updatedAt string
	err := scanner.Scan(&entry.TokenHash, &entry.TokenPreview, &entry.UploadCount, &entry.TotalBytes, &entry.LastIP, &lastUsedAt, &disabled, &entry.Reason, &disabledAt, &createdAt, &updatedAt)
	if err != nil {
		return model.TokenGovernanceEntry{}, err
	}
	entry.LastUsedAt = parseTime(lastUsedAt)
	entry.Disabled = disabled != 0
	entry.DisabledAt = parseTime(disabledAt)
	entry.CreatedAt = parseTime(createdAt)
	entry.UpdatedAt = parseTime(updatedAt)
	return entry, nil
}

func boolToInt(value bool) int {
	if value {
		return 1
	}
	return 0
}
