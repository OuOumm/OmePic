package service

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"omepic/backend/internal/config"
	"omepic/backend/internal/model"
	"omepic/backend/internal/repository"
)

const (
	configAuditScopeRuntime = "runtime"
	configAuditScopeStorage = "storage"
)

type auditContextKey string

const auditActorContextKey auditContextKey = "audit_actor"

type AuditActor struct {
	Name string
	IP   string
}

type AdminConfigAuditLogList struct {
	Items    []model.ConfigAuditLog `json:"items"`
	Page     int                    `json:"page"`
	PageSize int                    `json:"page_size"`
	Total    int64                  `json:"total"`
}

type AdminConfigAuditLogFilter struct {
	Scope    string
	Page     int
	PageSize int
}

func WithAuditActor(ctx context.Context, actor AuditActor) context.Context {
	return context.WithValue(ctx, auditActorContextKey, actor)
}

func auditActorFromContext(ctx context.Context) AuditActor {
	actor, ok := ctx.Value(auditActorContextKey).(AuditActor)
	if !ok {
		return AuditActor{Name: "admin"}
	}
	actor.Name = strings.TrimSpace(actor.Name)
	actor.IP = strings.TrimSpace(actor.IP)
	if actor.Name == "" {
		actor.Name = "admin"
	}
	return actor
}

func (s *AdminService) ListConfigAuditLogs(ctx context.Context, filter AdminConfigAuditLogFilter) (AdminConfigAuditLogList, error) {
	scope := strings.TrimSpace(filter.Scope)
	if scope != "" && scope != configAuditScopeRuntime && scope != configAuditScopeStorage {
		return AdminConfigAuditLogList{}, WithUserMessage(ErrInvalidInput, "unsupported audit log scope")
	}
	logs, err := s.repo.ListConfigAuditLogs(ctx, repository.ConfigAuditLogFilter{
		Scope:    scope,
		Page:     filter.Page,
		PageSize: filter.PageSize,
	})
	if err != nil {
		return AdminConfigAuditLogList{}, fmt.Errorf("%w: audit log query failed", ErrDependencyUnavailable)
	}
	return AdminConfigAuditLogList{Items: logs.Items, Page: logs.Page, PageSize: logs.PageSize, Total: logs.Total}, nil
}

func (s *AdminService) recordConfigAuditLog(ctx context.Context, scope string, before any, after any) error {
	beforeSnapshot, err := auditSnapshotJSON(before)
	if err != nil {
		return fmt.Errorf("%w: audit before snapshot failed", ErrDependencyUnavailable)
	}
	afterSnapshot, err := auditSnapshotJSON(after)
	if err != nil {
		return fmt.Errorf("%w: audit after snapshot failed", ErrDependencyUnavailable)
	}
	if beforeSnapshot == afterSnapshot {
		return nil
	}
	actor := auditActorFromContext(ctx)
	if err := s.repo.CreateConfigAuditLog(ctx, model.ConfigAuditLog{
		Actor:          actor.Name,
		ActorIP:        actor.IP,
		ConfigScope:    scope,
		BeforeSnapshot: beforeSnapshot,
		AfterSnapshot:  afterSnapshot,
	}); err != nil {
		return fmt.Errorf("%w: audit log save failed", ErrDependencyUnavailable)
	}
	return nil
}

func auditSnapshotJSON(value any) (string, error) {
	masked := maskAuditSecrets(value)
	payload, err := json.Marshal(masked)
	if err != nil {
		return "", err
	}
	return string(payload), nil
}

func maskRuntimeSettingsForAudit(settings RuntimeSettings) RuntimeSettings {
	return maskRuntimeSettings(settings)
}

func maskStorageConfigForAudit(cfg config.RuntimeStorageConfig) AdminStorageConfigView {
	return maskStorageConfig(cfg)
}

func maskAuditSecrets(value any) any {
	payload, err := json.Marshal(value)
	if err != nil {
		return value
	}
	var decoded any
	if err := json.Unmarshal(payload, &decoded); err != nil {
		return value
	}
	return maskAuditValue(decoded, "")
}

func maskAuditValue(value any, key string) any {
	switch typed := value.(type) {
	case map[string]any:
		result := make(map[string]any, len(typed))
		for childKey, childValue := range typed {
			result[childKey] = maskAuditValue(childValue, childKey)
		}
		return result
	case []any:
		result := make([]any, len(typed))
		for index, childValue := range typed {
			result[index] = maskAuditValue(childValue, key)
		}
		return result
	case string:
		if auditKeyIsSecret(key) {
			return maskSecret(typed)
		}
		return typed
	default:
		return typed
	}
}

func auditKeyIsSecret(key string) bool {
	normalized := strings.ToLower(strings.TrimSpace(key))
	if normalized == "" {
		return false
	}
	secretMarkers := []string{"secret", "password", "token", "access_key", "api_key", "webdav_pass"}
	for _, marker := range secretMarkers {
		if strings.Contains(normalized, marker) {
			return true
		}
	}
	return false
}
