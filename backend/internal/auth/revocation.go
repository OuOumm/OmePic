package auth

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
)

const adminRevokedBeforeKey = "admin:revoked_before"

// RevocationChecker checks whether a JWT's issued-at time predates a
// revocation event stored in Redis. If the JWT was issued before the
// revocation timestamp, it is considered invalid.
type RevocationChecker struct {
	client *redis.Client
}

func NewRevocationChecker(client *redis.Client) *RevocationChecker {
	return &RevocationChecker{client: client}
}

// IsRevoked reports whether a JWT with the given iat (issued-at) timestamp
// has been revoked. Returns false (not revoked) when Redis is unavailable
// so that transient Redis failures do not lock out the admin.
func (rc *RevocationChecker) IsRevoked(ctx context.Context, iat time.Time) (bool, error) {
	val, err := rc.client.Get(ctx, adminRevokedBeforeKey).Result()
	if err == redis.Nil {
		// No revocation record exists → token is valid.
		return false, nil
	}
	if err != nil {
		// Redis error → fail-open to avoid locking out admin on Redis outage.
		return false, fmt.Errorf("revocation check: %w", err)
	}
	revokedBefore, err := strconv.ParseInt(val, 10, 64)
	if err != nil {
		// Corrupt value → treat as no revocation.
		return false, nil
	}
	return iat.Unix() < revokedBefore, nil
}

// RevokeAllBefore marks all JWTs issued before now as revoked.
func (rc *RevocationChecker) RevokeAllBefore(ctx context.Context) error {
	now := time.Now().UTC().Unix()
	return rc.client.Set(ctx, adminRevokedBeforeKey, strconv.FormatInt(now, 10), 0).Err()
}