package main

import (
	"context"
	"go-redis/config"
	"log"
	"time"
)

const (
	CACHE_NAME  = "food"
	HCACHE_NAME = "food2"
)

func main() {

	cfg, err := config.LoadConfig("./")
	if err != nil {
		log.Fatal("failed to load env", err.Error())
	}

	ctx := context.Background()
	rdb, err := config.InitRedisConn(cfg.RedisUrl, cfg.RedisPort, cfg.RedisPass, ctx)
	if err != nil {
		log.Fatalln("Redis connection was refused!", err)
	}

	err = rdb.Set(ctx, CACHE_NAME, "Test value", 10*time.Minute).Err()
	if err != nil {
		panic(err)
	}

	rdb.HSet(ctx, HCACHE_NAME, "Bakso", 10*time.Minute)

	res, err := rdb.Get(ctx, CACHE_NAME).Result()
	if err != nil {
		panic(err)
	}
	log.Println(res)

	defer rdb.Close()

}
