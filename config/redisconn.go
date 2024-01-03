package config

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/redis/go-redis/v9"
)

/*
author @Rudy Ryanto (RR59782X)
this class provide redis connection
*/

// error message
const (
	ERROR_CONNECTION = "Failed to retrieve redis connection, host : "
	ERROR_PUT        = "Failed to put cache key : "
	ERROR_GET        = "Failed to get cache key : "
	ERROR_DELETE     = "Failed to delete cache key : "
)

// same as object in Java
type Cache interface{}

type RedisCfg struct {
	Url  string
	Port string
	Pass string
}

type RedisClient struct {
	rediscfg RedisCfg
	conn     *redis.Client
}

func NewRedisClient(ctx context.Context, url, port, pass string) RedisClient {
	// return RedisClient{GetRedisCfg(url, port, pass)}
	return RedisClient{
		GetRedisCfg(url, port, pass),
		InitRedisConnection(ctx, GetRedisCfg(url, port, pass)),
	}
}

func GetRedisCfg(url, port, pass string) RedisCfg {
	return RedisCfg{
		Url:  url,
		Port: port,
		Pass: pass,
	}

}

// to initialize redis connection,
// return redis client and error
func InitRedisConnection(ctx context.Context, rd RedisCfg) *redis.Client {

	host := rd.Url + ":" + rd.Port
	conn := redis.NewClient(&redis.Options{
		Addr:     host,
		Password: rd.Pass,
	})

	checkConnectionRedis(ctx, conn)

	return conn

}

// put simple cache
// param ctx, redisclient, cachename string, value any, ttl time.Duration
// e.g:
// store cache for 30 sec, 30*time.Second
// store cache for 10 minute, 10*time.Minute
// store cache no expiration time, use redis.KeepTTL
func (rd *RedisClient) PutCacheTtl(ctx context.Context, cacheName string, value Cache, ttl time.Duration) error {
	err := rd.conn.Set(ctx, cacheName, value, ttl).Err()

	if err != nil {
		return errorRedis(rd.conn.Options().Addr, "PUT", cacheName, err)
	}

	return nil

}

// put complex cache (eg: for suggestion)
// param ctx, redisclient, cachename string, value any, ttl time.Duration
// e.g:
// store cache for 30 sec, 30*time.Second
// store cache for 10 minute, 10*time.Minute
// store cache no expiration time, use redis.KeepTTL
func (rd *RedisClient) StoreComplexCache(ctx context.Context, cacheName string, value Cache, ttl time.Duration) error {
	err := rd.conn.HSet(ctx, cacheName, value, ttl).Err()
	if err != nil {
		return errorRedis(rd.conn.Options().Addr, "PUT", cacheName, err)
	}

	return nil
}

// this function for simple cache
// return any, error
func (rd *RedisClient) GetCache(ctx context.Context, cacheName string) (Cache, error) {

	res, err := rd.conn.Get(ctx, cacheName).Result()
	if err != nil {
		return nil, errorRedis(rd.conn.Options().Addr, "GET", cacheName, err)
	}
	return res, nil
}

// this function for get complex cache, return any,
// if field is empty string so it will get all cache
// if field is not empty then get using field
func (rd *RedisClient) GetHashCache(ctx context.Context, cacheName, field string) (Cache, error) {
	var res Cache
	if field == "" {
		data, err := rd.conn.HGetAll(ctx, cacheName).Result()
		if err != nil {
			return nil, errorRedis(rd.conn.Options().Addr, "GET", cacheName, err)
		}
		res = data
	} else {
		data, err := rd.conn.HGet(ctx, cacheName, field).Result()
		if err != nil {
			return nil, errorRedis(rd.conn.Options().Addr, "GET", cacheName, err)
		}
		res = data
	}
	return res, nil
}

// this function for delete cache
// return error / nil
func (rd *RedisClient) DeleteCache(ctx context.Context, keys ...string) map[string]error {
	Error := make(map[string]error)
	for _, key := range keys {
		err := rd.conn.Del(ctx, key).Err()
		if err != nil {
			Error[key] = errorRedis("", "DELETE", key, err)
			fmt.Println(err)
		}
	}
	if len(Error) != 0 {
		return Error
	}
	return nil
}

// private func
func errorRedis(addr, opr, key string, Error error) error {
	var err error
	switch opr {
	case "GET":
		err = errors.New(ERROR_GET + " " + key + ", caused : " + Error.Error())
	case "PUT":
		err = errors.New(ERROR_PUT + " " + key + ", caused : " + Error.Error())
	case "CONNECTION":
		err = errors.New(ERROR_CONNECTION + " " + addr + ", caused : " + Error.Error())
	case "DELETE":
		err = errors.New(ERROR_DELETE + " " + key + ", caused : " + Error.Error())
	default:
		err = nil
	}
	return err
}

// private func
func checkConnectionRedis(ctx context.Context, conn *redis.Client) error {

	status, err := conn.Ping(ctx).Result()
	if err != nil {
		return errorRedis(conn.Options().Addr, "CONNECTION", "", err)
	}
	log.Println(status)
	return nil
}
