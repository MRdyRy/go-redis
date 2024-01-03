package main

import (
	"context"
	"go-redis/config"
	"log"

	"github.com/redis/go-redis/v9"
)

const (
	CACHE_NAME  = "food"
	HCACHE_NAME = "food2"
)

type Cache interface{}

func main() {

	cfg, err := config.LoadConfig(".")
	if err != nil {
		log.Fatal("failed to load env", err.Error())
	}

	ctx := context.Background()

	rdb := config.NewRedisClient(ctx, cfg.RedisUrl, cfg.RedisPort, cfg.RedisPass)

	err = rdb.PutCacheTtl(ctx, "2024", "New Year", redis.KeepTTL)
	if err != nil {
		log.Println(err)
	}

	res, err := rdb.GetCache(ctx, "2024")
	if err != nil {
		log.Println(err)
	}
	log.Println(res)

}
