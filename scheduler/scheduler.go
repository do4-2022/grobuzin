package scheduler

import (
	"context"

	"github.com/redis/go-redis/v9"
)

type Scheduler struct {
	RedisHandle *redis.Client
	Context *context.Context
}

