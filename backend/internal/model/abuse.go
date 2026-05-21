package model

import "time"

type AbuseOverview struct {
	From             time.Time            `json:"from"`
	To               time.Time            `json:"to"`
	UploadCount      int64                `json:"upload_count"`
	UploadSize       int64                `json:"upload_size"`
	ActiveIPBanCount int64                `json:"active_ip_ban_count"`
	TopIPs           []AbuseIPRankItem    `json:"top_ips"`
	TopTokens        []AbuseTokenRankItem `json:"top_tokens"`
}

type AbuseIPRankItem struct {
	IPAddress      string    `json:"ip_address"`
	UploadCount    int64     `json:"upload_count"`
	TotalSize      int64     `json:"total_size"`
	LatestUploadAt time.Time `json:"latest_upload_at"`
	IsBanned       bool      `json:"is_banned"`
	BanID          int64     `json:"ban_id,omitempty"`
}

type AbuseTokenRankItem struct {
	Token          string    `json:"token"`
	TokenPreview   string    `json:"token_preview"`
	UploadCount    int64     `json:"upload_count"`
	TotalSize      int64     `json:"total_size"`
	LatestUploadAt time.Time `json:"latest_upload_at"`
}

// AbuseIPAggregate is a repository-level upload aggregation fact.
// Service-layer security analysis adds active-ban annotations.
type AbuseIPAggregate struct {
	IPAddress      string
	UploadCount    int64
	TotalSize      int64
	LatestUploadAt time.Time
}

// AbuseTokenAggregate is a repository-level token aggregation fact.
// Service-layer security analysis adds presentation fields such as token previews.
type AbuseTokenAggregate struct {
	Token          string
	UploadCount    int64
	TotalSize      int64
	LatestUploadAt time.Time
}

type AbuseIPDetail struct {
	IPAddress   string `json:"ip_address"`
	UploadCount int64  `json:"upload_count"`
	TotalSize   int64  `json:"total_size"`
	IsBanned    bool   `json:"is_banned"`
	Ban         *IPBan `json:"ban"`
}
