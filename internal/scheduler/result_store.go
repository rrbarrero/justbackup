package scheduler

import (
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"
)

type RedisResultStore struct {
	client *redis.Client
}

func NewRedisResultStore(client *redis.Client) *RedisResultStore {
	return &RedisResultStore{
		client: client,
	}
}

func (s *RedisResultStore) GetTaskResult(ctx context.Context, taskID string) (string, error) {
	key := fmt.Sprintf("task_result:%s", taskID)
	return s.client.Get(ctx, key).Result()
}
