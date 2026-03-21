package contracts

import (
	"context"
	"time"
)

// CacheService provides a key-value cache for sessions, rate limiting, and hot data.
// Required by the auth module (session storage) and jobs module (queue backend).
type CacheService interface {
	// Get retrieves the value for a key.
	// Returns ErrNotFound if the key does not exist or has expired.
	Get(ctx context.Context, key string) ([]byte, error)

	// Set stores a value with an optional TTL.
	// If ttl is 0, the key does not expire.
	Set(ctx context.Context, key string, value []byte, ttl time.Duration) error

	// Delete removes a key. Returns nil if the key did not exist.
	Delete(ctx context.Context, key string) error

	// Exists checks whether a key exists (without retrieving its value).
	Exists(ctx context.Context, key string) (bool, error)

	// Increment atomically increments an integer value by delta.
	// Creates the key with value delta if it does not exist.
	// Returns the new value after incrementing.
	Increment(ctx context.Context, key string, delta int64) (int64, error)

	// SetNX sets a key only if it does not already exist (atomic).
	// Returns true if the key was set, false if it already existed.
	// Use for distributed locks and deduplication.
	SetNX(ctx context.Context, key string, value []byte, ttl time.Duration) (bool, error)

	// Expire updates the TTL for an existing key.
	// Returns ErrNotFound if the key does not exist.
	Expire(ctx context.Context, key string, ttl time.Duration) error

	// FlushPattern deletes all keys matching the given glob pattern.
	// Use with care — can be slow on large keyspaces.
	FlushPattern(ctx context.Context, pattern string) error
}
