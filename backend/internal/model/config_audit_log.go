package model

import "time"

// ConfigAuditLog records an admin-visible audit trail for runtime and storage configuration changes.
type ConfigAuditLog struct {
	ID             int64     `json:"id"`
	Actor          string    `json:"actor"`
	ActorIP        string    `json:"actor_ip"`
	ConfigScope    string    `json:"config_scope"`
	BeforeSnapshot string    `json:"before_snapshot"`
	AfterSnapshot  string    `json:"after_snapshot"`
	CreatedAt      time.Time `json:"created_at"`
}
