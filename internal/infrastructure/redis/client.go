package redis

import (
	"context"
	"fmt"
	"os"
	"time"

	goredis "github.com/redis/go-redis/v9"
)

type Client struct {
	Client *goredis.Client
}

func NewClient() (*Client, error) {
	redisURL := os.Getenv("REDIS_URL")

	if redisURL == "" {
		redisURL = "redis://localhost:6379"
	}

	options, err := goredis.ParseURL(redisURL)
	if err != nil {
		return nil, fmt.Errorf("invalid Redis URL: %w", err)
	}

	// Configure connection and operation timeouts.
	options.DialTimeout = 3 * time.Second
	options.ReadTimeout = 2 * time.Second
	options.WriteTimeout = 2 * time.Second
	options.PoolSize = 20

	client := goredis.NewClient(options)

	// Avoid waiting indefinitely if Redis is unavailable.
	ctx, cancel := context.WithTimeout(
		context.Background(),
		3*time.Second,
	)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		_ = client.Close()
		return nil, fmt.Errorf("redis connection failed: %w", err)
	}

	return &Client{
		Client: client,
	}, nil
}

// Close releases the Redis client's connections.
func (c *Client) Close() error {
	if c == nil || c.Client == nil {
		return nil
	}

	return c.Client.Close()
}
