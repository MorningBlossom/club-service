package redis

import (
	"context"
	"fmt"
	"time"

	goredis "github.com/redis/go-redis/v9"
)

type RateLimiter struct {
	client *goredis.Client
}

func NewRateLimiter(client *goredis.Client) *RateLimiter {
	return &RateLimiter{client: client}
}

// Allow permits at most limit requests per window.
// The counter and its expiry are updated atomically.
func (r *RateLimiter) Allow(
	ctx context.Context,
	key string,
	limit int64,
	window time.Duration,
) (bool, error) {
	if r == nil || r.client == nil {
		return false, fmt.Errorf("redis client is not initialized")
	}
	if key == "" || limit <= 0 || window <= 0 {
		return false, fmt.Errorf("invalid rate limit configuration")
	}

	script := goredis.NewScript(`
		local count = redis.call("INCR", KEYS[1])
		if count == 1 then
			redis.call("PEXPIRE", KEYS[1], ARGV[1])
		end
		return count
	`)

	count, err := script.Run(
		ctx,
		r.client,
		[]string{key},
		window.Milliseconds(),
	).Int64()

	if err != nil {
		return false, fmt.Errorf("check rate limit: %w", err)
	}

	return count <= limit, nil
}
