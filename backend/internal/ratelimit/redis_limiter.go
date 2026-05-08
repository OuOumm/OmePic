package ratelimit

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

type Result struct {
	Allowed    bool
	Limit      int
	Remaining  int
	RetryAfter time.Duration
}

type Limiter interface {
	Allow(ctx context.Context, key string, limit int, window time.Duration) (Result, error)
}

type RedisLimiter struct {
	client *redis.Client
}

var fixedWindowScript = redis.NewScript(`
local current = redis.call("INCR", KEYS[1])
if current == 1 then
  redis.call("PEXPIRE", KEYS[1], ARGV[1])
end
local ttl = redis.call("PTTL", KEYS[1])
return {current, ttl}
`)

func NewRedisLimiter(client *redis.Client) *RedisLimiter {
	return &RedisLimiter{client: client}
}

func (l *RedisLimiter) Allow(ctx context.Context, key string, limit int, window time.Duration) (Result, error) {
	if limit <= 0 || window <= 0 {
		return Result{Allowed: true, Limit: limit, Remaining: limit}, nil
	}
	values, err := fixedWindowScript.Run(ctx, l.client, []string{key}, window.Milliseconds()).Slice()
	if err != nil {
		return Result{}, err
	}
	count := toInt64(values[0])
	ttl := toInt64(values[1])
	remaining := limit - int(count)
	if remaining < 0 {
		remaining = 0
	}
	if ttl < 0 {
		ttl = int64(window / time.Millisecond)
	}
	return Result{
		Allowed:    count <= int64(limit),
		Limit:      limit,
		Remaining:  remaining,
		RetryAfter: time.Duration(ttl) * time.Millisecond,
	}, nil
}

func toInt64(value any) int64 {
	switch v := value.(type) {
	case int64:
		return v
	case int:
		return int64(v)
	case uint64:
		return int64(v)
	default:
		return 0
	}
}
