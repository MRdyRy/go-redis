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

	cacheFactory, err := config.NewCacheFactory(ctx, config.Cfg{
		Url: cfg.RedisUrl, Port: cfg.RedisPort, Pass: cfg.RedisPass, Protocol: "http", User: "ryan",
	}, "redis")
	if err != nil {
		panic(err)
	}

	rdb := cacheFactory.(config.RedisClient)

	// rdb := config.NewRedisClient(ctx, cfg.RedisUrl, cfg.RedisPort, cfg.RedisPass)

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
