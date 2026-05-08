package cache

import (
	"context"
	"encoding/json"

	"github.com/redis/go-redis/v9"

	"omepic/backend/internal/model"
)

type ImageCache interface {
	GetImage(ctx context.Context, uid string) (*model.CachedImage, error)
	SetImage(ctx context.Context, record model.ImageRecord) error
	DeleteImage(ctx context.Context, uid string) error
	GetMD5(ctx context.Context, md5Hash string) (string, error)
	SetMD5(ctx context.Context, md5Hash string, uid string) error
	SetMD5IfAbsent(ctx context.Context, md5Hash string, uid string) error
	DeleteMD5(ctx context.Context, md5Hash string) error
	Ping(ctx context.Context) error
}

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
	return redis.NewClient(options), nil
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

func (c *RedisCache) DeleteImage(ctx context.Context, uid string) error {
	return c.client.Del(ctx, uidKey(uid)).Err()
}

func (c *RedisCache) GetMD5(ctx context.Context, md5Hash string) (string, error) {
	value, err := c.client.Get(ctx, md5Key(md5Hash)).Result()
	if err == redis.Nil {
		return "", nil
	}
	if err != nil {
		return "", err
	}
	return value, nil
}

func (c *RedisCache) SetMD5(ctx context.Context, md5Hash string, uid string) error {
	return c.client.Set(ctx, md5Key(md5Hash), uid, 0).Err()
}

func (c *RedisCache) SetMD5IfAbsent(ctx context.Context, md5Hash string, uid string) error {
	return c.client.SetNX(ctx, md5Key(md5Hash), uid, 0).Err()
}

func (c *RedisCache) DeleteMD5(ctx context.Context, md5Hash string) error {
	return c.client.Del(ctx, md5Key(md5Hash)).Err()
}

func uidKey(uid string) string {
	return "uid:" + uid
}

func md5Key(hash string) string {
	return "md5:" + hash
}
