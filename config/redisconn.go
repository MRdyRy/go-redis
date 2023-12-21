package config

import (
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"
)

func InitRedisConn(url, port, pass string, ctx context.Context) (*redis.Client, error) {

	cache := redis.NewClient(&redis.Options{
		Addr:     url,
		Password: pass,
	})

	status, err := cache.Ping(ctx).Result()
	if err != nil {
		return nil, err
	}

	fmt.Println(status)

	return cache, nil
}
