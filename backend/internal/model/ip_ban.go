package model

import "time"

type IPBan struct {
	ID              int64      `json:"id"`
	IPHash          string     `json:"ip_hash"`
	IPAddress       string     `json:"ip_address"`
	IPAddressMasked string     `json:"ip_address_masked"`
	Reason          string     `json:"reason"`
	ExpiresAt       *time.Time `json:"expires_at"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
}

type IPImageSummary struct {
	Count     int64 `json:"count"`
	TotalSize int64 `json:"total_size"`
}
