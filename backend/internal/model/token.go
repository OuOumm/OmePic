package model

import "time"

type TokenUsage struct {
	TokenHash    string    `json:"token_hash"`
	TokenPreview string    `json:"token_preview"`
	UploadCount  int64     `json:"upload_count"`
	TotalBytes   int64     `json:"total_bytes"`
	LastIP       string    `json:"last_ip"`
	LastUsedAt   time.Time `json:"last_used_at"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type TokenControl struct {
	TokenHash    string    `json:"token_hash"`
	TokenPreview string    `json:"token_preview"`
	Disabled     bool      `json:"disabled"`
	Reason       string    `json:"reason"`
	DisabledAt   time.Time `json:"disabled_at"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type TokenGovernanceEntry struct {
	TokenHash    string    `json:"token_hash"`
	TokenPreview string    `json:"token_preview"`
	UploadCount  int64     `json:"upload_count"`
	TotalBytes   int64     `json:"total_bytes"`
	LastIP       string    `json:"last_ip"`
	LastUsedAt   time.Time `json:"last_used_at"`
	Disabled     bool      `json:"disabled"`
	Reason       string    `json:"reason"`
	DisabledAt   time.Time `json:"disabled_at"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}
