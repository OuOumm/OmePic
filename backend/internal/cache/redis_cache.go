package cache

import (
	"context"
	"encoding/json"
	"time"

	"github.com/redis/go-redis/v9"

	"omepic/backend/internal/model"
)

// ImageLookupCache is the UID lookup/write seam used by serve, upload, and delete paths.
// Callers should not depend on MD5 mapping, preheat, or health methods through this interface.
type ImageLookupCache interface {
	GetImage(ctx context.Context, uid string) (*model.CachedImage, error)
	SetImage(ctx context.Context, record model.ImageRecord) error
	DeleteImage(ctx context.Context, uid string) error
}

// ImagePreheatCache is the batch UID-cache write seam used only by startup preheat.
type ImagePreheatCache interface {
	SetImages(ctx context.Context, records []model.ImageRecord) error
}

// MD5MappingCache is the scoped original-byte MD5 mapping seam owned by md5MappingFlow.
type MD5MappingCache interface {
	GetMD5(ctx context.Context, key model.MD5MappingKey) (string, error)
	SetMD5(ctx context.Context, key model.MD5MappingKey, uid string) error
	SetMD5IfAbsent(ctx context.Context, key model.MD5MappingKey, uid string) error
	DeleteMD5(ctx context.Context, key model.MD5MappingKey) error
}

// MD5MappingPreheatCache is the batch MD5 mapping write seam used only by startup preheat.
type MD5MappingPreheatCache interface {
	SetMD5Mappings(ctx context.Context, mappings []model.MD5Mapping) error
}

// HealthCache is the health-check seam. Runtime upload/serve code must not depend on Ping.
type HealthCache interface {
	Ping(ctx context.Context) error
}

// ImageCache is a compatibility aggregate for Redis adapters that satisfy all image cache seams.
// Prefer accepting the narrower interfaces above at call sites.
type ImageCache interface {
	ImageLookupCache
	ImagePreheatCache
	MD5MappingCache
	MD5MappingPreheatCache
}

const (
	defaultRedisDialTimeout  = 3 * time.Second
	defaultRedisReadTimeout  = 2 * time.Second
	defaultRedisWriteTimeout = 2 * time.Second
	defaultRedisPoolSize     = 16
	defaultRedisPoolTimeout  = 3 * time.Second
	defaultRedisMinIdleConns = 2
)

type RedisCache struct {
	client *redis.Client
}

func New(redisURL string) (*RedisCache, error) {
	client, err := NewClient(redisURL)
	if err != nil {
		return nil, err
	}
	return NewWithClient(client), nil
}

func NewClient(redisURL string) (*redis.Client, error) {
	options, err := redis.ParseURL(redisURL)
	if err != nil {
		return nil, err
	}
	applyClientDefaults(options)
	return redis.NewClient(options), nil
}

func applyClientDefaults(options *redis.Options) {
	if options.DialTimeout == 0 {
		options.DialTimeout = defaultRedisDialTimeout
	}
	if options.ReadTimeout == 0 {
		options.ReadTimeout = defaultRedisReadTimeout
	}
	if options.WriteTimeout == 0 {
		options.WriteTimeout = defaultRedisWriteTimeout
	}
	if options.PoolSize == 0 {
		options.PoolSize = defaultRedisPoolSize
	}
	if options.PoolTimeout == 0 {
		options.PoolTimeout = defaultRedisPoolTimeout
	}
	if options.MinIdleConns == 0 {
		options.MinIdleConns = defaultRedisMinIdleConns
	}
	options.ContextTimeoutEnabled = true
}

func NewWithClient(client *redis.Client) *RedisCache {
	return &RedisCache{client: client}
}

func (c *RedisCache) Close() error {
	return c.client.Close()
}

func (c *RedisCache) Ping(ctx context.Context) error {
	return c.client.Ping(ctx).Err()
}

func (c *RedisCache) GetImage(ctx context.Context, uid string) (*model.CachedImage, error) {
	value, err := c.client.Get(ctx, uidKey(uid)).Result()
	if err == redis.Nil {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var cached model.CachedImage
	if err := json.Unmarshal([]byte(value), &cached); err != nil {
		return nil, err
	}
	return &cached, nil
}

func (c *RedisCache) SetImage(ctx context.Context, record model.ImageRecord) error {
	payload, err := json.Marshal(model.CachedImageFromRecord(record))
	if err != nil {
		return err
	}
	return c.client.Set(ctx, uidKey(record.UID), payload, 0).Err()
}

func (c *RedisCache) SetImages(ctx context.Context, records []model.ImageRecord) error {
	if len(records) == 0 {
		return nil
	}

	pipe := c.client.Pipeline()
	for _, record := range records {
		payload, err := json.Marshal(model.CachedImageFromRecord(record))
		if err != nil {
			return err
		}
		pipe.Set(ctx, uidKey(record.UID), payload, 0)
	}
	_, err := pipe.Exec(ctx)
	return err
}

func (c *RedisCache) DeleteImage(ctx context.Context, uid string) error {
	return c.client.Del(ctx, uidKey(uid)).Err()
}

func (c *RedisCache) GetMD5(ctx context.Context, key model.MD5MappingKey) (string, error) {
	value, err := c.client.Get(ctx, md5Key(key)).Result()
	if err == redis.Nil {
		return "", nil
	}
	if err != nil {
		return "", err
	}
	return value, nil
}

func (c *RedisCache) SetMD5(ctx context.Context, key model.MD5MappingKey, uid string) error {
	return c.client.Set(ctx, md5Key(key), uid, 0).Err()
}

func (c *RedisCache) SetMD5Mappings(ctx context.Context, mappings []model.MD5Mapping) error {
	if len(mappings) == 0 {
		return nil
	}

	pipe := c.client.Pipeline()
	for _, mapping := range mappings {
		pipe.Set(ctx, md5Key(mapping.Key), mapping.UID, 0)
	}
	_, err := pipe.Exec(ctx)
	return err
}

func (c *RedisCache) SetMD5IfAbsent(ctx context.Context, key model.MD5MappingKey, uid string) error {
	return c.client.SetNX(ctx, md5Key(key), uid, 0).Err()
}

func (c *RedisCache) DeleteMD5(ctx context.Context, key model.MD5MappingKey) error {
	return c.client.Del(ctx, md5Key(key)).Err()
}

func uidKey(uid string) string {
	return "uid:" + uid
}

func md5Key(key model.MD5MappingKey) string {
	return "md5:" + key.CacheScope()
}
