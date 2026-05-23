package repository

import (
	"context"
	"path/filepath"
	"testing"

	"omepic/backend/internal/model"
)

func TestConfigAuditLogMigrationIsIdempotentAndQueryPaginates(t *testing.T) {
	ctx := context.Background()
	repo, err := New(filepath.Join(t.TempDir(), "test.sqlite"))
	if err != nil {
		t.Fatalf("New returned error: %v", err)
	}
	defer repo.Close()

	if err := repo.Migrate(ctx); err != nil {
		t.Fatalf("first Migrate returned error: %v", err)
	}
	if err := repo.Migrate(ctx); err != nil {
		t.Fatalf("second Migrate returned error: %v", err)
	}

	logs := []model.ConfigAuditLog{
		{Actor: "admin", ActorIP: "127.0.0.1", ConfigScope: "runtime", BeforeSnapshot: `{"a":1}`, AfterSnapshot: `{"a":2}`},
		{Actor: "admin", ActorIP: "127.0.0.1", ConfigScope: "storage", BeforeSnapshot: `{"b":1}`, AfterSnapshot: `{"b":2}`},
		{Actor: "admin", ActorIP: "127.0.0.1", ConfigScope: "storage", BeforeSnapshot: `{"c":1}`, AfterSnapshot: `{"c":2}`},
	}
	for _, log := range logs {
		if err := repo.CreateConfigAuditLog(ctx, log); err != nil {
			t.Fatalf("CreateConfigAuditLog returned error: %v", err)
		}
	}

	page, err := repo.ListConfigAuditLogs(ctx, ConfigAuditLogFilter{Scope: "storage", Page: 1, PageSize: 1})
	if err != nil {
		t.Fatalf("ListConfigAuditLogs returned error: %v", err)
	}
	if page.Total != 2 || page.Page != 1 || page.PageSize != 1 || len(page.Items) != 1 {
		t.Fatalf("unexpected paged result: %+v", page)
	}
	if page.Items[0].ConfigScope != "storage" {
		t.Fatalf("expected storage scope, got %q", page.Items[0].ConfigScope)
	}
}
