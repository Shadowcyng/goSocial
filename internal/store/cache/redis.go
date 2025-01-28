package cache

import (
	"github.com/go-redis/redis/v8"
)

func NewRedisClient(addr, pw string, db int) *redis.Client {
	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: pw,
		DB:       db,
	})

	return client
}
