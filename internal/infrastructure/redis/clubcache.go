package redis

import (
	"context"
	"fmt"
	"time"
)

func (c *Client) SetClub(
	ctx context.Context,
	clubID string,
	value string,
) error {

	key := fmt.Sprintf("club:%s", clubID)

	return c.Client.Set(
		ctx,
		key,
		value,
		10*time.Minute,
	).Err()
}

func (c *Client) GetClub(
	ctx context.Context,
	clubID string,
) (string, error) {

	key := fmt.Sprintf("club:%s", clubID)

	return c.Client.Get(ctx, key).Result()
}

func (c *Client) DeleteClub(
	ctx context.Context,
	clubID string,
) error {

	key := fmt.Sprintf("club:%s", clubID)

	return c.Client.Del(ctx, key).Err()
}