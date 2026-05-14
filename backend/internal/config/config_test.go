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

func TestLoadUsesOnlyStartupEnvironmentContract(t *testing.T) {
	t.Setenv("HTTP_ADDR", ":9090")
	t.Setenv("DATABASE_PATH", "data/custom.db")
	t.Setenv("REDIS_URL", "redis://localhost:6380/1")
	t.Setenv("UID_PREFIX", "custom_")
	t.Setenv("UID_ENCRYPTION_KEY", "")
	t.Setenv("JWT_SECRET", "jwt-secret")
	t.Setenv("TRUSTED_PROXY_CIDRS", "10.0.0.0/8")
	t.Setenv("REAL_IP_HEADER", "X-Real-IP")
	t.Setenv("PUBLIC_BASE_URL", "https://env.example.com")
	t.Setenv("ADMIN_PASSWORD", "secret")
	t.Setenv("STORAGE_BACKEND", "s3")

	cfg := Load()

	if cfg.HTTPAddr != ":9090" || cfg.DatabasePath != "data/custom.db" || cfg.RedisURL != "redis://localhost:6380/1" {
		t.Fatalf("unexpected startup config: %+v", cfg)
	}
	if cfg.UIDPrefix != "custom_" || cfg.JWTSecret != "jwt-secret" {
		t.Fatalf("unexpected security config: %+v", cfg)
	}
	if cfg.UIDEncryptionKey != "change-me-uid-secret" {
		t.Fatalf("expected UID encryption key default independent from JWT secret, got %q", cfg.UIDEncryptionKey)
	}
}
