package model

import "time"

type ImageRecord struct {
	ID             int64     `json:"id"`
	UID            string    `json:"uid"`
	Token          string    `json:"token"`
	StorageKey     string    `json:"storage_key"`
	StorageBackend string    `json:"storage_backend"`
	FilePath       string    `json:"file_path"`
	MIMEType       string    `json:"mime_type"`
	Size           int64     `json:"size"`
	MD5Hash        string    `json:"md5_hash"`
	IPAddress      string    `json:"ip_address"`
	CreatedAt      time.Time `json:"created_at"`
}

type CachedImage struct {
	UID            string    `json:"uid"`
	Token          string    `json:"token"`
	StorageKey     string    `json:"storage_key"`
	StorageBackend string    `json:"storage_backend"`
	FilePath       string    `json:"file_path"`
	MIMEType       string    `json:"mime_type"`
	Size           int64     `json:"size"`
	MD5Hash        string    `json:"md5_hash"`
	CreatedAt      time.Time `json:"created_at"`
}

func CachedImageFromRecord(record ImageRecord) CachedImage {
	return CachedImage{
		UID:            record.UID,
		Token:          record.Token,
		StorageKey:     record.StorageKey,
		StorageBackend: record.StorageBackend,
		FilePath:       record.FilePath,
		MIMEType:       record.MIMEType,
		Size:           record.Size,
		MD5Hash:        record.MD5Hash,
		CreatedAt:      record.CreatedAt,
	}
}

type AdminStatus struct {
	TotalImages      int64 `json:"total_images"`
	TotalStorageSize int64 `json:"total_storage_size"`
	TodayUploads     int64 `json:"today_uploads"`
	UniqueTokens     int64 `json:"unique_tokens"`
}
