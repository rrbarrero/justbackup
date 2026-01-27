package scheduler

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/rrbarrero/justbackup/internal/backup/domain/interfaces"
	workerDto "github.com/rrbarrero/justbackup/internal/worker/dto"
)

type RedisWorkerQueryBus struct {
	client    *redis.Client
	publisher interfaces.TaskPublisher
}

func NewRedisWorkerQueryBus(client *redis.Client, publisher interfaces.TaskPublisher) *RedisWorkerQueryBus {
	return &RedisWorkerQueryBus{
		client:    client,
		publisher: publisher,
	}
}

func (b *RedisWorkerQueryBus) SearchFiles(ctx context.Context, pattern string) (workerDto.SearchFilesResult, error) {
	pubsub := b.client.Subscribe(ctx, "worker_sync_responses")
	defer pubsub.Close()

	if _, err := pubsub.ReceiveTimeout(ctx, 2*time.Second); err != nil {
		return workerDto.SearchFilesResult{}, fmt.Errorf("failed to subscribe to worker responses: %w", err)
	}

	taskID, err := b.publisher.PublishSearchTask(ctx, pattern)
	if err != nil {
		return workerDto.SearchFilesResult{}, err
	}

	ch := pubsub.Channel()
	timeout := time.After(30 * time.Second)

	for {
		select {
		case msg := <-ch:
			var result workerDto.WorkerResult
			if err := json.Unmarshal([]byte(msg.Payload), &result); err != nil {
				continue
			}

			if result.TaskID == taskID {
				if result.Status == "failed" {
					return workerDto.SearchFilesResult{}, fmt.Errorf("worker error: %s", result.Message)
				}

				// Improvement (Tech Debt #3): Use structured parsing
				dataJSON, err := json.Marshal(result.Data)
				if err != nil {
					return workerDto.SearchFilesResult{}, fmt.Errorf("failed to re-marshal result data: %w", err)
				}

				var searchResult workerDto.SearchFilesResult
				if err := json.Unmarshal(dataJSON, &searchResult); err != nil {
					return workerDto.SearchFilesResult{}, fmt.Errorf("failed to unmarshal search results: %w", err)
				}

				return searchResult, nil
			}
		case <-timeout:
			return workerDto.SearchFilesResult{}, fmt.Errorf("timeout waiting for worker search response (30s)")
		case <-ctx.Done():
			return workerDto.SearchFilesResult{}, ctx.Err()
		}
	}
}

func (b *RedisWorkerQueryBus) ListFiles(ctx context.Context, path string) (workerDto.ListFilesResult, error) {
	pubsub := b.client.Subscribe(ctx, "worker_sync_responses")
	defer pubsub.Close()

	if _, err := pubsub.ReceiveTimeout(ctx, 2*time.Second); err != nil {
		return workerDto.ListFilesResult{}, fmt.Errorf("failed to subscribe to worker responses: %w", err)
	}

	taskID, err := b.publisher.PublishListFilesTask(ctx, path)
	if err != nil {
		return workerDto.ListFilesResult{}, err
	}

	ch := pubsub.Channel()
	timeout := time.After(30 * time.Second)

	for {
		select {
		case msg := <-ch:
			var result workerDto.WorkerResult
			if err := json.Unmarshal([]byte(msg.Payload), &result); err != nil {
				continue
			}

			if result.TaskID == taskID {
				if result.Status == "failed" {
					return workerDto.ListFilesResult{}, fmt.Errorf("worker error: %s", result.Message)
				}

				// Improvement (Tech Debt #3): Use structured parsing
				dataJSON, err := json.Marshal(result.Data)
				if err != nil {
					return workerDto.ListFilesResult{}, fmt.Errorf("failed to re-marshal result data: %w", err)
				}

				var listResult workerDto.ListFilesResult
				if err := json.Unmarshal(dataJSON, &listResult); err != nil {
					return workerDto.ListFilesResult{}, fmt.Errorf("failed to unmarshal list results: %w", err)
				}

				return listResult, nil
			}
		case <-timeout:
			return workerDto.ListFilesResult{}, fmt.Errorf("timeout waiting for worker list response (30s)")
		case <-ctx.Done():
			return workerDto.ListFilesResult{}, ctx.Err()
		}
	}
}
