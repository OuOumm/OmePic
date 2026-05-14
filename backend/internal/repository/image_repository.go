package repository

import (
	"context"
	"database/sql"
	"time"

	"omepic/backend/internal/model"
)

const imageColumns = "id, uid, token, storage_key, storage_backend, file_path, mime_type, size, md5_hash, ip_address, created_at"

func (r *Repository) InsertImage(ctx context.Context, record model.ImageRecord) error {
	if record.CreatedAt.IsZero() {
		record.CreatedAt = time.Now().UTC()
	}
	_, err := r.db.ExecContext(
		ctx,
		`INSERT INTO images(
			uid, token, storage_key, storage_backend, file_path, mime_type, size, md5_hash, ip_address, created_at
		) VALUES(?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		record.UID,
		record.Token,
		record.StorageKey,
		record.StorageBackend,
		record.FilePath,
		record.MIMEType,
		record.Size,
		record.MD5Hash,
		record.IPAddress,
		record.CreatedAt.UTC().Format(time.RFC3339),
	)
	return err
}

func (r *Repository) FindByUID(ctx context.Context, uid string) (*model.ImageRecord, error) {
	row := r.db.QueryRowContext(ctx, `SELECT `+imageColumns+` FROM images WHERE uid = ?`, uid)
	record, err := scanImage(row)
	if err != nil {
		return nil, err
	}
	return &record, nil
}

func (r *Repository) FindByMD5(ctx context.Context, md5Hash string) (*model.ImageRecord, error) {
	row := r.db.QueryRowContext(ctx, `SELECT `+imageColumns+` FROM images WHERE md5_hash = ? ORDER BY id ASC LIMIT 1`, md5Hash)
	record, err := scanImage(row)
	if err != nil {
		return nil, err
	}
	return &record, nil
}

func (r *Repository) FindByMD5AndStorageKey(ctx context.Context, md5Hash string, storageKey string) (*model.ImageRecord, error) {
	row := r.db.QueryRowContext(ctx, `SELECT `+imageColumns+` FROM images WHERE md5_hash = ? AND storage_key = ? ORDER BY id ASC LIMIT 1`, md5Hash, storageKey)
	record, err := scanImage(row)
	if err != nil {
		return nil, err
	}
	return &record, nil
}

func (r *Repository) DeleteByUID(ctx context.Context, uid string) error {
	result, err := r.db.ExecContext(ctx, `DELETE FROM images WHERE uid = ?`, uid)
	if err != nil {
		return err
	}
	if err := ensureRowsAffected(result); err != nil {
		return err
	}
	return nil
}

func (r *Repository) CountByMD5(ctx context.Context, md5Hash string) (int64, error) {
	return r.countByQuery(ctx, `SELECT COUNT(1) FROM images WHERE md5_hash = ?`, md5Hash)
}

func (r *Repository) CountByMD5AndStorageKey(ctx context.Context, md5Hash string, storageKey string) (int64, error) {
	return r.countByQuery(ctx, `SELECT COUNT(1) FROM images WHERE md5_hash = ? AND storage_key = ?`, md5Hash, storageKey)
}

func (r *Repository) CountByStoredFile(ctx context.Context, storageKey string, filePath string) (int64, error) {
	return r.countByQuery(ctx, `SELECT COUNT(1) FROM images WHERE storage_key = ? AND file_path = ?`, storageKey, filePath)
}

func (r *Repository) ImageSummaryByIP(ctx context.Context, ipAddress string) (model.IPImageSummary, error) {
	var summary model.IPImageSummary
	err := r.db.QueryRowContext(ctx, `SELECT COUNT(1), COALESCE(SUM(size), 0) FROM images WHERE ip_address = ?`, ipAddress).Scan(&summary.Count, &summary.TotalSize)
	return summary, err
}

func (r *Repository) ListImagesByIP(ctx context.Context, ipAddress string) ([]model.ImageRecord, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT `+imageColumns+` FROM images WHERE ip_address = ? ORDER BY id ASC`, ipAddress)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanImages(rows)
}

func (r *Repository) ListAllImages(ctx context.Context) ([]model.ImageRecord, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT `+imageColumns+` FROM images ORDER BY id ASC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanImages(rows)
}

func (r *Repository) SearchImages(ctx context.Context, page int, pageSize int, search string) ([]model.ImageRecord, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100
	}

	like := "%" + search + "%"
	where := `WHERE (? = '' OR uid LIKE ? OR token LIKE ? OR ip_address LIKE ? OR md5_hash LIKE ? OR storage_key LIKE ?)`

	total, err := r.countByQuery(ctx, `SELECT COUNT(1) FROM images `+where, search, like, like, like, like, like)
	if err != nil {
		return nil, 0, err
	}

	rows, err := r.db.QueryContext(
		ctx,
		`SELECT `+imageColumns+`
		 FROM images `+where+`
		 ORDER BY id DESC
		 LIMIT ? OFFSET ?`,
		search,
		like,
		like,
		like,
		like,
		like,
		pageSize,
		(page-1)*pageSize,
	)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	records, err := scanImages(rows)
	if err != nil {
		return nil, 0, err
	}
	return records, total, nil
}

func (r *Repository) AggregateStatus(ctx context.Context) (model.AdminStatus, error) {
	var status model.AdminStatus
	if err := r.db.QueryRowContext(ctx, `SELECT COUNT(1), COALESCE(SUM(size), 0), COUNT(DISTINCT token) FROM images`).Scan(&status.TotalImages, &status.TotalStorageSize, &status.UniqueTokens); err != nil {
		return status, err
	}
	if err := r.db.QueryRowContext(ctx, `SELECT COUNT(1) FROM images WHERE DATE(created_at) = DATE('now')`).Scan(&status.TodayUploads); err != nil {
		return status, err
	}
	return status, nil
}

func scanImage(scanner interface{ Scan(dest ...any) error }) (model.ImageRecord, error) {
	var record model.ImageRecord
	var createdAt string
	err := scanner.Scan(
		&record.ID,
		&record.UID,
		&record.Token,
		&record.StorageKey,
		&record.StorageBackend,
		&record.FilePath,
		&record.MIMEType,
		&record.Size,
		&record.MD5Hash,
		&record.IPAddress,
		&createdAt,
	)
	if err != nil {
		return model.ImageRecord{}, err
	}
	record.CreatedAt = parseTime(createdAt)
	return record, nil
}

func scanImages(rows *sql.Rows) ([]model.ImageRecord, error) {
	var records []model.ImageRecord
	for rows.Next() {
		record, err := scanImage(rows)
		if err != nil {
			return nil, err
		}
		records = append(records, record)
	}
	return records, rows.Err()
}
