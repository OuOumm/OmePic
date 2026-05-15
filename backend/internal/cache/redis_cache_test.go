package cache

import (
	"testing"
	"time"

	"omepic/backend/internal/model"
)

func TestNewClientAppliesRedisTimeoutAndPoolDefaults(t *testing.T) {
	client, err := NewClient("redis://localhost:6379/0")
	if err != nil {
		t.Fatalf("NewClient returned error: %v", err)
	}
	defer client.Close()

	options := client.Options()
	if options.DialTimeout != defaultRedisDialTimeout {
		t.Fatalf("DialTimeout = %s, want %s", options.DialTimeout, defaultRedisDialTimeout)
	}
	if options.ReadTimeout != defaultRedisReadTimeout {
		t.Fatalf("ReadTimeout = %s, want %s", options.ReadTimeout, defaultRedisReadTimeout)
	}
	if options.WriteTimeout != defaultRedisWriteTimeout {
		t.Fatalf("WriteTimeout = %s, want %s", options.WriteTimeout, defaultRedisWriteTimeout)
	}
	if options.PoolSize != defaultRedisPoolSize {
		t.Fatalf("PoolSize = %d, want %d", options.PoolSize, defaultRedisPoolSize)
	}
	if options.PoolTimeout != defaultRedisPoolTimeout {
		t.Fatalf("PoolTimeout = %s, want %s", options.PoolTimeout, defaultRedisPoolTimeout)
	}
	if options.MinIdleConns != defaultRedisMinIdleConns {
		t.Fatalf("MinIdleConns = %d, want %d", options.MinIdleConns, defaultRedisMinIdleConns)
	}
	if !options.ContextTimeoutEnabled {
		t.Fatal("expected context timeout support to be enabled")
	}
}

func TestMD5RedisKeyKeepsExistingScopedShape(t *testing.T) {
	key := model.NewMD5MappingKey("local-primary", "abcdef")
	if got := md5Key(key); got != "md5:local-primary:abcdef" {
		t.Fatalf("expected Redis md5 key shape to remain compatible, got %q", got)
	}
}

func TestNewClientPreservesRedisURLOverrides(t *testing.T) {
	client, err := NewClient("redis://localhost:6379/0?dial_timeout=7s&read_timeout=8s&write_timeout=9s&pool_size=32&pool_timeout=11s&min_idle_conns=4")
	if err != nil {
		t.Fatalf("NewClient returned error: %v", err)
	}
	defer client.Close()

	options := client.Options()
	if options.DialTimeout != 7*time.Second {
		t.Fatalf("DialTimeout = %s, want 7s", options.DialTimeout)
	}
	if options.ReadTimeout != 8*time.Second {
		t.Fatalf("ReadTimeout = %s, want 8s", options.ReadTimeout)
	}
	if options.WriteTimeout != 9*time.Second {
		t.Fatalf("WriteTimeout = %s, want 9s", options.WriteTimeout)
	}
	if options.PoolSize != 32 {
		t.Fatalf("PoolSize = %d, want 32", options.PoolSize)
	}
	if options.PoolTimeout != 11*time.Second {
		t.Fatalf("PoolTimeout = %s, want 11s", options.PoolTimeout)
	}
	if options.MinIdleConns != 4 {
		t.Fatalf("MinIdleConns = %d, want 4", options.MinIdleConns)
	}
	if !options.ContextTimeoutEnabled {
		t.Fatal("expected context timeout support to be enabled")
	}
}
