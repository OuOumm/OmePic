package repository

import (
	"context"
	"database/sql"
	"time"

	"omepic/backend/internal/model"
)

func (r *Repository) CreateIPBan(ctx context.Context, ban model.IPBan) (model.IPBan, error) {
	now := time.Now().UTC()
	if ban.CreatedAt.IsZero() {
		ban.CreatedAt = now
	}
	ban.UpdatedAt = now
	result, err := r.db.ExecContext(
		ctx,
		`INSERT INTO ip_bans(ip_hash, ip_address, ip_address_masked, reason, expires_at, created_at, updated_at)
		 VALUES(?, ?, ?, ?, ?, ?, ?)`,
		ban.IPHash,
		ban.IPAddress,
		ban.IPAddressMasked,
		ban.Reason,
		nullableTimeString(ban.ExpiresAt),
		ban.CreatedAt.Format(time.RFC3339),
		ban.UpdatedAt.Format(time.RFC3339),
	)
	if err != nil {
		return model.IPBan{}, err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return model.IPBan{}, err
	}
	return r.GetIPBan(ctx, id)
}

func (r *Repository) ListIPBans(ctx context.Context) ([]model.IPBan, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT id, ip_hash, ip_address, ip_address_masked, reason, expires_at, created_at, updated_at FROM ip_bans ORDER BY id DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanIPBans(rows)
}

func (r *Repository) GetIPBan(ctx context.Context, id int64) (model.IPBan, error) {
	row := r.db.QueryRowContext(ctx, `SELECT id, ip_hash, ip_address, ip_address_masked, reason, expires_at, created_at, updated_at FROM ip_bans WHERE id = ?`, id)
	return scanIPBan(row)
}

func (r *Repository) FindActiveIPBanByHash(ctx context.Context, ipHash string) (model.IPBan, error) {
	now := time.Now().UTC().Format(time.RFC3339)
	row := r.db.QueryRowContext(ctx, `SELECT id, ip_hash, ip_address, ip_address_masked, reason, expires_at, created_at, updated_at FROM ip_bans WHERE ip_hash = ? AND (expires_at IS NULL OR expires_at = '' OR expires_at > ?) ORDER BY id DESC LIMIT 1`, ipHash, now)
	return scanIPBan(row)
}

func (r *Repository) DeleteIPBan(ctx context.Context, id int64) error {
	result, err := r.db.ExecContext(ctx, `DELETE FROM ip_bans WHERE id = ?`, id)
	if err != nil {
		return err
	}
	if err := ensureRowsAffected(result); err != nil {
		return err
	}
	return nil
}

func (r *Repository) FindActiveIPBanByIP(ctx context.Context, ipAddress string) (model.IPBan, error) {
	return r.FindActiveIPBanByHash(ctx, ipHashValue(ipAddress))
}

func (r *Repository) ActiveIPBansByHash(ctx context.Context) (map[string]model.IPBan, error) {
	now := time.Now().UTC().Format(time.RFC3339)
	rows, err := r.db.QueryContext(ctx, `SELECT id, ip_hash, ip_address, ip_address_masked, reason, expires_at, created_at, updated_at FROM ip_bans WHERE expires_at IS NULL OR expires_at = '' OR expires_at > ? ORDER BY id DESC`, now)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	bans, err := scanIPBans(rows)
	if err != nil {
		return nil, err
	}
	result := make(map[string]model.IPBan, len(bans))
	for _, ban := range bans {
		if _, exists := result[ban.IPHash]; !exists {
			result[ban.IPHash] = ban
		}
	}
	return result, nil
}

func (r *Repository) CountActiveIPBans(ctx context.Context) (int64, error) {
	now := time.Now().UTC().Format(time.RFC3339)
	return r.countByQuery(ctx, `SELECT COUNT(1) FROM ip_bans WHERE expires_at IS NULL OR expires_at = '' OR expires_at > ?`, now)
}

func (r *Repository) AbuseOverviewTotals(ctx context.Context, from time.Time, to time.Time) (int64, int64, error) {
	var count int64
	var totalSize int64
	err := r.db.QueryRowContext(ctx, `SELECT COUNT(1), COALESCE(SUM(size), 0) FROM images WHERE created_at >= ? AND created_at <= ?`, from.UTC().Format(time.RFC3339), to.UTC().Format(time.RFC3339)).Scan(&count, &totalSize)
	return count, totalSize, err
}

func (r *Repository) TopAbuseIPs(ctx context.Context, from time.Time, to time.Time, limit int) ([]model.AbuseIPRankItem, error) {
	if limit < 1 {
		limit = 10
	}
	if limit > 50 {
		limit = 50
	}
	activeBans, err := r.ActiveIPBansByHash(ctx)
	if err != nil {
		return nil, err
	}
	rows, err := r.db.QueryContext(
		ctx,
		`SELECT ip_address, COUNT(1), COALESCE(SUM(size), 0), MAX(created_at)
		 FROM images
		 WHERE created_at >= ? AND created_at <= ? AND ip_address IS NOT NULL AND TRIM(ip_address) != ''
		 GROUP BY ip_address
		 ORDER BY COUNT(1) DESC, COALESCE(SUM(size), 0) DESC, MAX(created_at) DESC
		 LIMIT ?`,
		from.UTC().Format(time.RFC3339),
		to.UTC().Format(time.RFC3339),
		limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := make([]model.AbuseIPRankItem, 0)
	for rows.Next() {
		var item model.AbuseIPRankItem
		var latest string
		if err := rows.Scan(&item.IPAddress, &item.UploadCount, &item.TotalSize, &latest); err != nil {
			return nil, err
		}
		item.IPAddressMasked = maskIPValue(item.IPAddress)
		item.LatestUploadAt = parseTime(latest)
		if ban, exists := activeBans[ipHashValue(item.IPAddress)]; exists {
			item.IsBanned = true
			item.BanID = ban.ID
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *Repository) TopAbuseTokens(ctx context.Context, from time.Time, to time.Time, limit int) ([]model.AbuseTokenRankItem, error) {
	if limit < 1 {
		limit = 10
	}
	if limit > 50 {
		limit = 50
	}
	rows, err := r.db.QueryContext(
		ctx,
		`SELECT token, COUNT(1), COALESCE(SUM(size), 0), MAX(created_at)
		 FROM images
		 WHERE created_at >= ? AND created_at <= ? AND token IS NOT NULL AND TRIM(token) != ''
		 GROUP BY token
		 ORDER BY COUNT(1) DESC, COALESCE(SUM(size), 0) DESC, MAX(created_at) DESC
		 LIMIT ?`,
		from.UTC().Format(time.RFC3339),
		to.UTC().Format(time.RFC3339),
		limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := make([]model.AbuseTokenRankItem, 0)
	for rows.Next() {
		var item model.AbuseTokenRankItem
		var latest string
		if err := rows.Scan(&item.Token, &item.UploadCount, &item.TotalSize, &latest); err != nil {
			return nil, err
		}
		item.TokenPreview = previewValue(item.Token, 8)
		item.LatestUploadAt = parseTime(latest)
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *Repository) IPDetail(ctx context.Context, ipAddress string) (model.AbuseIPDetail, error) {
	summary, err := r.ImageSummaryByIP(ctx, ipAddress)
	if err != nil {
		return model.AbuseIPDetail{}, err
	}
	detail := model.AbuseIPDetail{
		IPAddress:       ipAddress,
		IPAddressMasked: maskIPValue(ipAddress),
		UploadCount:     summary.Count,
		TotalSize:       summary.TotalSize,
	}
	ban, err := r.FindActiveIPBanByIP(ctx, ipAddress)
	if err == nil {
		detail.IsBanned = true
		detail.Ban = &ban
		return detail, nil
	}
	if IsNotFound(err) {
		return detail, nil
	}
	return model.AbuseIPDetail{}, err
}

func scanIPBan(scanner interface{ Scan(dest ...any) error }) (model.IPBan, error) {
	var ban model.IPBan
	var expiresAt sql.NullString
	var createdAt string
	var updatedAt string
	err := scanner.Scan(
		&ban.ID,
		&ban.IPHash,
		&ban.IPAddress,
		&ban.IPAddressMasked,
		&ban.Reason,
		&expiresAt,
		&createdAt,
		&updatedAt,
	)
	if err != nil {
		return model.IPBan{}, err
	}
	ban.ExpiresAt = parseNullableTime(expiresAt)
	ban.CreatedAt = parseTime(createdAt)
	ban.UpdatedAt = parseTime(updatedAt)
	return ban, nil
}

func scanIPBans(rows *sql.Rows) ([]model.IPBan, error) {
	var bans []model.IPBan
	for rows.Next() {
		ban, err := scanIPBan(rows)
		if err != nil {
			return nil, err
		}
		bans = append(bans, ban)
	}
	return bans, rows.Err()
}
