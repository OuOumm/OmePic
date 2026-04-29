package config

import "testing"

func TestLoadUsesExplicitUIDEncryptionKey(t *testing.T) {
	t.Setenv("UID_PREFIX", "custom_")
	t.Setenv("UID_ENCRYPTION_KEY", "uid-secret")
	t.Setenv("JWT_SECRET", "jwt-secret")

	cfg := Load()

	if cfg.UIDPrefix != "custom_" {
		t.Fatalf("expected UIDPrefix custom_, got %q", cfg.UIDPrefix)
	}
	if cfg.UIDEncryptionKey != "uid-secret" {
		t.Fatalf("expected explicit UID encryption key, got %q", cfg.UIDEncryptionKey)
	}
}

func TestLoadFallsBackToJWTSecretWhenUIDEncryptionKeyIsUnset(t *testing.T) {
	t.Setenv("UID_ENCRYPTION_KEY", "")
	t.Setenv("JWT_SECRET", "jwt-secret")

	cfg := Load()

	if cfg.UIDEncryptionKey != "jwt-secret" {
		t.Fatalf("expected JWT secret fallback, got %q", cfg.UIDEncryptionKey)
	}
}
