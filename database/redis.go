package database

import (
    "github.com/redis/go-redis/v9"
)

func InitRedis(url string) *redis.Client {
	opts, err := redis.ParseURL(url)
	
	if err != nil {
		panic(err)
	}

	return redis.NewClient(opts)
}
