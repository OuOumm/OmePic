package auth

import (
	"context"
	"strconv"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
)

func TestRevocationChecker_IsRevoked_NoRecord(t *testing.T) {
	mr := miniredis.RunT(t)
	client := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	rc := NewRevocationChecker(client)

	iat := time.Now().UTC().Add(-1 * time.Hour)
	revoked, err := rc.IsRevoked(context.Background(), iat)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if revoked {
		t.Fatal("token should not be revoked when no revocation record exists")
	}
}

func TestRevocationChecker_IsRevoked_RevokedBeforeLater(t *testing.T) {
	mr := miniredis.RunT(t)
	client := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	rc := NewRevocationChecker(client)

	// Token issued at T=100, revocation at T=200 → token is revoked.
	iat := time.Unix(100, 0).UTC()
	mr.Set(adminRevokedBeforeKey, "200")
	revoked, err := rc.IsRevoked(context.Background(), iat)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !revoked {
		t.Fatal("token issued before revocation timestamp should be revoked")
	}
}

func TestRevocationChecker_IsRevoked_RevokedBeforeEarlier(t *testing.T) {
	mr := miniredis.RunT(t)
	client := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	rc := NewRevocationChecker(client)

	// Token issued at T=200, revocation at T=100 → token is NOT revoked.
	iat := time.Unix(200, 0).UTC()
	mr.Set(adminRevokedBeforeKey, "100")
	revoked, err := rc.IsRevoked(context.Background(), iat)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if revoked {
		t.Fatal("token issued after revocation timestamp should not be revoked")
	}
}

func TestRevocationChecker_IsRevoked_CorruptValue(t *testing.T) {
	mr := miniredis.RunT(t)
	client := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	rc := NewRevocationChecker(client)

	mr.Set(adminRevokedBeforeKey, "not-a-number")
	revoked, err := rc.IsRevoked(context.Background(), time.Now())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if revoked {
		t.Fatal("corrupt revocation value should not cause revocation")
	}
}

func TestRevocationChecker_IsRevoked_RedisErrorFailOpen(t *testing.T) {
	mr := miniredis.RunT(t)
	client := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	rc := NewRevocationChecker(client)

	// Set a revocation record first.
	mr.Set(adminRevokedBeforeKey, "200")

	// Close miniredis to simulate Redis failure.
	mr.Close()

	iat := time.Unix(100, 0).UTC()
	revoked, err := rc.IsRevoked(context.Background(), iat)
	if err == nil {
		t.Fatal("expected error from Redis, got nil")
	}
	if revoked {
		t.Fatal("Redis error should fail-open: token must not be considered revoked")
	}
}

func TestRevocationChecker_RevokeAllBefore(t *testing.T) {
	mr := miniredis.RunT(t)
	client := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	rc := NewRevocationChecker(client)

	ctx := context.Background()
	if err := rc.RevokeAllBefore(ctx); err != nil {
		t.Fatalf("RevokeAllBefore returned error: %v", err)
	}

	val, err := client.Get(ctx, adminRevokedBeforeKey).Result()
	if err != nil {
		t.Fatalf("redis Get returned error: %v", err)
	}
	ts, err := strconv.ParseInt(val, 10, 64)
	if err != nil {
		t.Fatalf("stored value is not an integer: %s", val)
	}
	now := time.Now().UTC().Unix()
	if ts < now-2 || ts > now+2 {
		t.Fatalf("revocation timestamp %d is not close to now %d", ts, now)
	}
}