package cache

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

var ErrMiss = errors.New("cache miss")

const (
	KeyDirections = "cache:directions"
	KeyNews       = "cache:news"
	KeySummary    = "cache:summary"
)

type Redis struct {
	client *redis.Client
	prefix string
	logger *zap.Logger
}

func Open(ctx context.Context, redisURL, prefix string, logger *zap.Logger) (*Redis, error) {
	options, err := redis.ParseURL(redisURL)
	if err != nil {
		return nil, fmt.Errorf("parse Redis URL: %w", err)
	}
	options.MaxRetries = 2
	options.MinRetryBackoff = 20 * time.Millisecond
	options.MaxRetryBackoff = 200 * time.Millisecond
	options.DialTimeout = 3 * time.Second
	options.ReadTimeout = 2 * time.Second
	options.WriteTimeout = 2 * time.Second
	options.PoolSize = 20
	client := redis.NewClient(options)
	store := &Redis{client: client, prefix: prefix, logger: logger}
	if err = store.Ping(ctx); err != nil {
		_ = client.Close()
		return nil, fmt.Errorf("ping Redis: %w", err)
	}
	logger.Info("Redis connection established")
	return store, nil
}

func (r *Redis) GetJSON(ctx context.Context, key string, target any) error {
	data, err := r.client.Get(ctx, r.key(key)).Bytes()
	if errors.Is(err, redis.Nil) {
		return ErrMiss
	}
	if err != nil {
		return fmt.Errorf("get Redis key %s: %w", key, err)
	}
	if err = json.Unmarshal(data, target); err != nil {
		_ = r.client.Del(ctx, r.key(key)).Err()
		return fmt.Errorf("decode Redis key %s: %w", key, err)
	}
	return nil
}

func (r *Redis) SetJSON(ctx context.Context, key string, value any, ttl time.Duration) error {
	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("encode Redis key %s: %w", key, err)
	}
	if err = r.client.Set(ctx, r.key(key), data, ttl).Err(); err != nil {
		return fmt.Errorf("set Redis key %s: %w", key, err)
	}
	return nil
}

func (r *Redis) Delete(ctx context.Context, keys ...string) error {
	if len(keys) == 0 {
		return nil
	}
	prefixed := make([]string, len(keys))
	for i, key := range keys {
		prefixed[i] = r.key(key)
	}
	if err := r.client.Del(ctx, prefixed...).Err(); err != nil {
		return fmt.Errorf("delete Redis keys: %w", err)
	}
	return nil
}

func (r *Redis) Ping(ctx context.Context) error {
	return r.client.Ping(ctx).Err()
}

func (r *Redis) Close() error {
	return r.client.Close()
}

func (r *Redis) key(key string) string {
	return r.prefix + key
}

var fixedWindowScript = redis.NewScript(`
local current = redis.call('INCR', KEYS[1])
if current == 1 then
  redis.call('PEXPIRE', KEYS[1], ARGV[1])
end
local ttl = redis.call('PTTL', KEYS[1])
return {current, ttl}
`)

func (r *Redis) Allow(ctx context.Context, key string, limit int, window time.Duration) (bool, time.Duration, error) {
	result, err := fixedWindowScript.Run(ctx, r.client, []string{r.key("ratelimit:" + key)}, window.Milliseconds()).Slice()
	if err != nil {
		return false, 0, fmt.Errorf("execute rate limit: %w", err)
	}
	if len(result) != 2 {
		return false, 0, errors.New("invalid rate limit result")
	}
	current, ok := result[0].(int64)
	if !ok {
		return false, 0, errors.New("invalid rate limit counter")
	}
	ttlMillis, ok := result[1].(int64)
	if !ok {
		return false, 0, errors.New("invalid rate limit ttl")
	}
	retryAfter := time.Duration(max(ttlMillis, 0)) * time.Millisecond
	return current <= int64(limit), retryAfter, nil
}
