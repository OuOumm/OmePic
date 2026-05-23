package model

import "time"

const (
	StorageHealthHealthy     = "healthy"
	StorageHealthUnavailable = "unavailable"
)

type StorageHealthCheck struct {
	ID                  int64     `json:"id"`
	StorageKey          string    `json:"storage_key"`
	Status              string    `json:"status"`
	LastCheckAt         time.Time `json:"last_check_at"`
	LatencyMS           int64     `json:"latency_ms"`
	ErrorMessage        string    `json:"error_message"`
	ConsecutiveFailures int       `json:"consecutive_failures"`
	CreatedAt           time.Time `json:"created_at"`
	UpdatedAt           time.Time `json:"updated_at"`
}
