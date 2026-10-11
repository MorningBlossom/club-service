package redis

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	goredis "github.com/redis/go-redis/v9"
)

var (
	ErrInvalidLock  = errors.New("invalid lock configuration")
	ErrLockNotOwned = errors.New("lock is no longer owned by this token")
)

type Locker struct {
	client *goredis.Client
	ttl    time.Duration
	retry  time.Duration
}

type Lock struct {
	client *goredis.Client
	key    string
	token  string
}

func NewLocker(client *goredis.Client) (*Locker, error) {
	if client == nil {
		return nil, ErrInvalidLock
	}

	return &Locker{
		client: client,
		ttl:    10 * time.Second,
		retry:  100 * time.Millisecond,
	}, nil
}

// Acquire waits until the lock is available or the context expires.
func (l *Locker) Acquire(
	ctx context.Context,
	key string,
) (*Lock, error) {
	if l == nil || l.client == nil || key == "" ||
		l.ttl <= 0 || l.retry <= 0 {
		return nil, ErrInvalidLock
	}

	token, err := newToken()
	if err != nil {
		return nil, fmt.Errorf("generate lock token: %w", err)
	}

	ticker := time.NewTicker(l.retry)
	defer ticker.Stop()

	for {
		acquired, err := l.client.SetNX(
			ctx,
			key,
			token,
			l.ttl,
		).Result()

		if err != nil {
			return nil, fmt.Errorf(
				"acquire Redis lock %q: %w",
				key,
				err,
			)
		}

		if acquired {
			return &Lock{
				client: l.client,
				key:    key,
				token:  token,
			}, nil
		}

		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-ticker.C:
		}
	}
}

func newToken() (string, error) {
	bytes := make([]byte, 32)

	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}

	return hex.EncodeToString(bytes), nil
}

// Release deletes the lock only if the stored token matches ours.
func (lock *Lock) Release(ctx context.Context) error {
	if lock == nil || lock.client == nil {
		return ErrInvalidLock
	}

	script := goredis.NewScript(`
		if redis.call("GET", KEYS[1]) == ARGV[1] then
			return redis.call("DEL", KEYS[1])
		end
		return 0
	`)

	result, err := script.Run(
		ctx,
		lock.client,
		[]string{lock.key},
		lock.token,
	).Int64()

	if err != nil {
		return fmt.Errorf(
			"release Redis lock %q: %w",
			lock.key,
			err,
		)
	}

	if result == 0 {
		return ErrLockNotOwned
	}

	return nil
}
