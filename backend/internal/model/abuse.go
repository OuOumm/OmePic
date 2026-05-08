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
	IPAddress       string    `json:"ip_address"`
	IPAddressMasked string    `json:"ip_address_masked"`
	UploadCount     int64     `json:"upload_count"`
	TotalSize       int64     `json:"total_size"`
	LatestUploadAt  time.Time `json:"latest_upload_at"`
	IsBanned        bool      `json:"is_banned"`
	BanID           int64     `json:"ban_id,omitempty"`
}

type AbuseTokenRankItem struct {
	Token          string    `json:"token"`
	TokenPreview   string    `json:"token_preview"`
	UploadCount    int64     `json:"upload_count"`
	TotalSize      int64     `json:"total_size"`
	LatestUploadAt time.Time `json:"latest_upload_at"`
}

type AbuseIPDetail struct {
	IPAddress       string `json:"ip_address"`
	IPAddressMasked string `json:"ip_address_masked"`
	UploadCount     int64  `json:"upload_count"`
	TotalSize       int64  `json:"total_size"`
	IsBanned        bool   `json:"is_banned"`
	Ban             *IPBan `json:"ban"`
}
