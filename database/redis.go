package database

import (
	"context"
	"github.com/go-redis/redis/v9"
	"time"
)

type RedisClientParams struct {
	RedisClient *redis.Client
}

func GetRedisClient(address string, password string, db int) (*redis.Client, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:     address,
		Password: password,
		DB:       db,
	})

	pong := rdb.Ping(context.Background())
	if pong.Err() != nil {
		return nil, pong.Err()
	}

	return rdb, nil
}

// SetCacheValue accepts Redis Client Params, a string key, and a value
// to insert in to the cache
func SetCacheValue[K any](params RedisClientParams, key string, value K, ttl int) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	insertResult := params.RedisClient.Set(ctx, key, value, time.Duration(time.Duration(ttl)*time.Minute))
	return insertResult.Result()
}

// GetCacheValue queries and returns the value corresponding to the provided
// key string
func GetCacheValue(params RedisClientParams, key string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	result := params.RedisClient.Get(ctx, key)
	return result.Result()
}
