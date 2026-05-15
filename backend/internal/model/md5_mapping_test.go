package model

import "testing"

func TestMD5MappingKeyScopesCacheAndMutexByStorage(t *testing.T) {
	primary := NewMD5MappingKey(" local-primary ", " ABCDEF ")
	secondary := NewMD5MappingKey("local-secondary", "ABCDEF")

	if primary.StorageKey != "local-primary" {
		t.Fatalf("expected trimmed storage key, got %q", primary.StorageKey)
	}
	if primary.MD5Hash != "abcdef" {
		t.Fatalf("expected normalized lowercase md5 hash, got %q", primary.MD5Hash)
	}
	if primary.CacheScope() != "local-primary:abcdef" {
		t.Fatalf("unexpected cache scope %q", primary.CacheScope())
	}
	if primary.MutexScope() != "local-primary\x00abcdef" {
		t.Fatalf("unexpected mutex scope %q", primary.MutexScope())
	}
	if primary.CacheScope() == secondary.CacheScope() {
		t.Fatalf("expected same md5 in different storage keys to use different cache scopes")
	}
}

func TestParseMD5MappingCacheScope(t *testing.T) {
	key, ok := ParseMD5MappingCacheScope(" local-primary:ABCDEF ")
	if !ok {
		t.Fatalf("expected valid scoped cache key")
	}
	if key != NewMD5MappingKey("local-primary", "abcdef") {
		t.Fatalf("unexpected parsed key %+v", key)
	}

	for _, raw := range []string{"", "missing-separator", ":hash", "storage:"} {
		if _, ok := ParseMD5MappingCacheScope(raw); ok {
			t.Fatalf("expected %q to be rejected", raw)
		}
	}
}
