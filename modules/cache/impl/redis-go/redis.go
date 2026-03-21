// Package redis implements CacheService using github.com/redis/go-redis/v9.
package redis

import (
	"context"
	"errors"
	"fmt"
	"time"

	goredis "github.com/redis/go-redis/v9"
)

// Config holds the configuration for the Redis cache provider.
type Config struct {
	// URL is the Redis connection URL (redis://localhost:6379 or rediss:// for TLS).
	URL string

	// MaxConnections is the maximum pool size. Default: 10.
	MaxConnections int
}

// Service implements contracts.CacheService using Redis.
type Service struct {
	client *goredis.Client
}

// New creates and validates a Redis cache service. Returns an error if the
// connection URL is invalid or the server is unreachable.
func New(cfg Config) (*Service, error) {
	if cfg.MaxConnections == 0 {
		cfg.MaxConnections = 10
	}

	opts, err := goredis.ParseURL(cfg.URL)
	if err != nil {
		return nil, fmt.Errorf("redis: invalid URL: %w", err)
	}
	opts.PoolSize = cfg.MaxConnections

	client := goredis.NewClient(opts)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("redis: ping failed: %w", err)
	}

	return &Service{client: client}, nil
}

func (s *Service) Get(ctx context.Context, key string) ([]byte, error) {
	val, err := s.client.Get(ctx, key).Bytes()
	if errors.Is(err, goredis.Nil) {
		return nil, nil
	}
	return val, err
}

func (s *Service) Set(ctx context.Context, key string, value []byte, ttl time.Duration) error {
	return s.client.Set(ctx, key, value, ttl).Err()
}

func (s *Service) Delete(ctx context.Context, key string) error {
	return s.client.Del(ctx, key).Err()
}

func (s *Service) Exists(ctx context.Context, key string) (bool, error) {
	n, err := s.client.Exists(ctx, key).Result()
	return n > 0, err
}

func (s *Service) Increment(ctx context.Context, key string, delta int64) (int64, error) {
	return s.client.IncrBy(ctx, key, delta).Result()
}

func (s *Service) SetNX(ctx context.Context, key string, value []byte, ttl time.Duration) (bool, error) {
	return s.client.SetNX(ctx, key, value, ttl).Result()
}

func (s *Service) Expire(ctx context.Context, key string, ttl time.Duration) error {
	return s.client.Expire(ctx, key, ttl).Err()
}

func (s *Service) FlushPattern(ctx context.Context, pattern string) error {
	var cursor uint64
	var keys []string
	for {
		batch, next, err := s.client.Scan(ctx, cursor, pattern, 100).Result()
		if err != nil {
			return err
		}
		keys = append(keys, batch...)
		cursor = next
		if cursor == 0 {
			break
		}
	}
	if len(keys) == 0 {
		return nil
	}
	return s.client.Del(ctx, keys...).Err()
}
