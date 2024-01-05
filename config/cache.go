package config

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
)

/*
author @Rudy Ryanto (RR59782X)
this class provide redis connection
*/

// error message
const (
	ERROR_CONNECTION         = "Failed to retrieve connection, host : "
	ERROR_PUT                = "Failed to put cache key : "
	ERROR_GET                = "Failed to get cache key : "
	ERROR_DELETE             = "Failed to delete cache key : "
	DATAGRID_ENDPOINT        = "/rest/v2"
	DATAGRID_ALLKEY_ENDPOINT = "?action=keys"
	ERROR_GENERAL            = "Request Failed with status code %d"
	UNKNOWN_ERROR            = "Error caused : "
)

// same as object in Java
type Cache interface{}

type Cfg struct {
	Url      string
	Port     string
	Pass     string
	Protocol string
	User     string
}

// redis client
type RedisClient struct {
	conn *redis.Client
}

// type DatagridConfig struct {
// 	Protocol string
// 	Host     string
// 	Port     string
// 	User     string
// 	Password string
// }

// datagrid client
type DatagridClient struct {
	dataCfg Cfg
}

var (
	body      []byte
	headerReq map[string]string
)

/*
=========================Cache Factory==================
*/

// func return any struct that implement redis or datagrid client
// params contex, config struct, mode (redis / datagrid)
func NewCacheFactory(ctx context.Context, config Cfg, mode string) (interface{}, error) {
	var cacheClient any
	if strings.ToLower(mode) == "redis" {
		cacheClient = newRedisClient(ctx, config.Url, config.Port, config.Pass)
	}
	if strings.ToLower(mode) == "datagrid" {
		cacheClient = newDatagridClient(config.Protocol, config.Url, config.Port, config.User, config.Pass)
	}
	return cacheClient, nil
}

/*
=========================REDIS SECTION===================
*/

// init redis
func newRedisClient(ctx context.Context, url, port, pass string) RedisClient {
	return RedisClient{
		initRedisConnection(ctx, GetRedisCfg(url, port, pass)),
	}
}

func GetRedisCfg(url, port, pass string) Cfg {
	return Cfg{
		Url:  url,
		Port: port,
		Pass: pass,
	}

}

// to initialize redis connection,
// return redis client and error
func initRedisConnection(ctx context.Context, rd Cfg) *redis.Client {

	host := rd.Url + ":" + rd.Port
	conn := redis.NewClient(&redis.Options{
		Addr:     host,
		Password: rd.Pass,
	})

	checkConnectionRedis(ctx, conn)

	return conn

}

// put simple cache
// param ctx, cachename string, value any, ttl time.Duration
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
// param ctx, cachename string, value any, ttl time.Duration
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

/*
========================END REDIS SECTION===================

=========================DATAGRID SECTION===================
*/
func newDatagridClient(protocol, host, port, user, password string) DatagridClient {
	return DatagridClient{GetCacheConfig(protocol, host, port, user, password)}
}

func initHeaderReq(req *http.Request) *http.Request {
	headerReq = make(map[string]string)
	headerReq["Accept"] = "*/*"
	headerReq["content-type"] = "application/json"

	for k, v := range headerReq {
		req.Header.Set(k, v)
	}
	return req
}

func GetCacheConfig(protocol, host, port, user, password string) Cfg {
	cacheConfig := Cfg{
		Protocol: protocol,
		Url:      host,
		Port:     port,
		User:     user,
		Pass:     password,
	}
	return cacheConfig
}

// private func to provide baseurl
func (dg *DatagridClient) baseUrl() string {
	return dg.dataCfg.Protocol + "://" + dg.dataCfg.Url + ":" + dg.dataCfg.Port + DATAGRID_ENDPOINT
}

// private func, provide http client to call datagrid server
func buildHttpClient() *http.Client {
	transport := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: transport}

	return client
}

// get all cache keys
func (dg *DatagridClient) GetAllKeysFromCache(cacheName string) ([]string, error) {
	req, err := http.NewRequest("GET", dg.baseUrl()+"/caches/"+cacheName+DATAGRID_ALLKEY_ENDPOINT, nil)
	if err != nil {
		return nil, err
	}
	client := buildHttpClient()
	initHeaderReq(req)

	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	defer res.Body.Close()
	var keys []string

	err = json.NewDecoder(res.Body).Decode(&keys)
	if err != nil {
		return nil, err
	}

	return keys, nil
}

// check key if exist, return map, error
// param slice string
func (dg *DatagridClient) CheckExistKey(cacheName string, keys ...string) (map[string]bool, error) {
	data := make(map[string]bool)
	for _, key := range keys {
		req, err := http.NewRequest("HEAD", dg.baseUrl()+"/caches/"+cacheName+"/"+key, nil)
		if err != nil {
			fmt.Println(UNKNOWN_ERROR + dg.baseUrl())
		}
		client := buildHttpClient()
		res, err := client.Do(req)
		if err != nil {
			fmt.Println(UNKNOWN_ERROR + err.Error())
		}

		data[key] = res.StatusCode == http.StatusOK
	}
	return data, nil
}

// get data from datagrid
// param cachename, key
// return any, error
func (dg *DatagridClient) GetDataFromCache(cacheName, key string) (Cache, error) {
	req, err := http.NewRequest("GET", dg.baseUrl()+"/caches/"+cacheName+"/"+key, nil)
	if err != nil {
		return nil, err
	}
	initHeaderReq(req)
	client := buildHttpClient()
	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	defer res.Body.Close()

	return res.Body, nil
}

// func to add cache
func (dg *DatagridClient) AddToCache(cacheName, key string, value Cache) error {
	jsonData, err := json.Marshal(value)
	if err != nil {
		return err
	}

	exists, err := dg.CheckExistKey(cacheName)
	if err != nil {
		return err
	}

	client := buildHttpClient()

	if exists[key] {
		//do update data

		req, err := http.NewRequest("PUT", dg.baseUrl()+"/caches/"+cacheName+"/"+key, bytes.NewBuffer(jsonData))
		if err != nil {
			return err
		}

		initHeaderReq(req)

		res, err := client.Do(req)
		if err != nil {
			return err
		}

		if res.StatusCode != http.StatusNoContent {
			return fmt.Errorf(fmt.Sprintf(ERROR_GENERAL, res.StatusCode))
		}

	} else {
		// create
		req, err := http.NewRequest("POST", dg.baseUrl()+"/caches/"+cacheName+"/"+key, bytes.NewBuffer(jsonData))
		if err != nil {
			return err
		}

		initHeaderReq(req)

		res, err := client.Do(req)
		if err != nil {
			return err
		}

		if res.StatusCode != http.StatusNoContent {
			return fmt.Errorf(fmt.Sprintf(ERROR_GENERAL, res.StatusCode))
		}

	}

	return nil
}

// delete cache by key from datagrid
func (dg *DatagridClient) DeleteFromDG(cacheName, key string) error {
	req, err := http.NewRequest("DELETE", dg.baseUrl()+"/caches/"+cacheName+"/"+key, nil)
	if err != nil {
		return err
	}
	initHeaderReq(req)
	client := buildHttpClient()
	res, err := client.Do(req)
	if err != nil {
		return err
	}

	defer res.Body.Close()

	return nil
}

// create new cache
func (dg *DatagridClient) CreateNewNameCache(cacheName string) error {
	body, _ = json.Marshal(GenerateTemplate("JBOSS"))
	req, err := http.NewRequest("POST", dg.baseUrl()+"/caches/"+cacheName, bytes.NewBuffer(body))
	if err != nil {
		return err
	}

	client := buildHttpClient()

	initHeaderReq(req)

	res, err := client.Do(req)
	if err != nil {
		return err
	}

	if res.StatusCode != http.StatusNoContent {
		return fmt.Errorf(fmt.Sprintf(ERROR_GENERAL, res.StatusCode))
	}

	return nil
}

/*
===========================END DATAGRID SECTION=================
*/
