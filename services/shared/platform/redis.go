package platform

import (
	"context"

	"github.com/redis/go-redis/v9"
)

func NewRedis(ctx context.Context, addr string) (*redis.Client, error) {
	client := redis.NewClient(&redis.Options{Addr: addr})
	return client, client.Ping(ctx).Err()
}
