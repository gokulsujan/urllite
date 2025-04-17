package cache

import (
	"context"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
)

type RedisClient interface {
	Set(key string, value string, expiration time.Duration) error
	Get(key string) (string, error)
	Exists(key string) (bool, error)
}

type redisClient struct {
	Client  *redis.Client
	Context context.Context
}

func InitRedis(ctx context.Context) RedisClient {
	db, err := strconv.Atoi(os.Getenv("REDIS_DB"))
	if err != nil {
		log.Panic("Failed to get redis db from env: %v", err)
		return &redisClient{}
	}
	client := redis.NewClient(&redis.Options{
		Addr:     os.Getenv("REDIS_ADDR"),
		Password: os.Getenv("REDIS_PASSWORD"),
		DB:       db, 
	})

	// Ping to test connection
	_, err = client.Ping(ctx).Result()
	if err != nil {
		log.Panic("Failed to connect to Redis: %v", err)
		return &redisClient{}
	}

	return &redisClient{Client: client, Context: ctx}
}

func (rc *redisClient) Set(key string, value string, expiration time.Duration) error {
	return rc.Client.Set(rc.Context, key, value, expiration).Err()
}

func (rc *redisClient) Get(key string) (string, error) {
	return rc.Client.Get(rc.Context, key).Result()
}

func (rc *redisClient) Exists(key string) (bool, error) {
	count, err := rc.Client.Exists(rc.Context, key).Result()
	return count > 0, err
}
