// Package cache provides a Redis-backed cache client with typed JSON get/set operations.
//
// The client is designed to be optional: a nil *Client is safe to use and all
// methods become no-ops, allowing callers to treat cache as an opt-in feature.
package cache

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/redis/go-redis/v9"
)

// Client wraps a Redis client with typed get/set operations.
type Client struct {
	rdb *redis.Client
}

// New creates a new cache client from a Redis URL.
// If url is empty or unparseable, returns nil (cache disabled).
func New(url string) *Client {
	if url == "" {
		return nil
	}
	opts, err := redis.ParseURL(url)
	if err != nil {
		return nil
	}
	return &Client{rdb: redis.NewClient(opts)}
}

// Available reports whether the cache is usable.
func (c *Client) Available() bool {
	return c != nil && c.rdb != nil
}

// Get retrieves a value and JSON-decodes it into dest.
// Returns false if the key does not exist.
func (c *Client) Get(ctx context.Context, key string, dest any) (bool, error) {
	if !c.Available() {
		return false, nil
	}
	val, err := c.rdb.Get(ctx, key).Result()
	if errors.Is(err, redis.Nil) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, json.Unmarshal([]byte(val), dest)
}

// Set JSON-encodes the value and stores it with the given TTL.
func (c *Client) Set(ctx context.Context, key string, value any, ttl time.Duration) error {
	if !c.Available() {
		return nil
	}
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}
	return c.rdb.Set(ctx, key, data, ttl).Err()
}

// Delete removes a key.
func (c *Client) Delete(ctx context.Context, key string) error {
	if !c.Available() {
		return nil
	}
	return c.rdb.Del(ctx, key).Err()
}

// Ping checks connectivity.
func (c *Client) Ping(ctx context.Context) error {
	if !c.Available() {
		return errors.New("cache not available")
	}
	return c.rdb.Ping(ctx).Err()
}

// Close closes the connection.
func (c *Client) Close() error {
	if !c.Available() {
		return nil
	}
	return c.rdb.Close()
}
