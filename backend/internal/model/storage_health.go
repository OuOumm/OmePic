package model

import "time"

const (
	StorageHealthHealthy     = 1
	StorageHealthUnavailable = 0
)

type StorageHealthCheck struct {
	ID                  int64     `json:"id"`
	StorageKey          string    `json:"storage_key"`
	Status              int       `json:"status"`
	LatencyMS           int64     `json:"latency_ms"`
	ErrorMessage        string    `json:"error_message"`
	ConsecutiveFailures int       `json:"consecutive_failures"`
	CreatedAt           time.Time `json:"created_at"`
	UpdatedAt           time.Time `json:"updated_at"`
}
