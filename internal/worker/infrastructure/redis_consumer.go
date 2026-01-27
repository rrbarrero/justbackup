package infrastructure

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/rrbarrero/justbackup/internal/worker/application"
	workerDto "github.com/rrbarrero/justbackup/internal/worker/dto"
)

type RedisTaskConsumer struct {
	client      *redis.Client
	queueName   string
	resultQueue string
}

func NewRedisTaskConsumer(redisURL string, queueName string, resultQueue string) *RedisTaskConsumer {
	client := redis.NewClient(&redis.Options{
		Addr: redisURL,
	})
	return &RedisTaskConsumer{
		client:      client,
		queueName:   queueName,
		resultQueue: resultQueue,
	}
}

func (c *RedisTaskConsumer) Start(ctx context.Context) {
	log.Printf("Worker listening on queue: %s", c.queueName)

	for {
		// BLPOP blocks until a task is available
		result, err := c.client.BLPop(ctx, 0, c.queueName).Result()
		if err != nil {
			log.Printf("Redis BLPop error: %v", err)
			time.Sleep(5 * time.Second)
			continue
		}

		// result[0] is queue name, result[1] is value
		payload := result[1]
		log.Printf("Received task: %s", payload)

		var task workerDto.WorkerTask
		if err := json.Unmarshal([]byte(payload), &task); err != nil {
			log.Printf("Failed to unmarshal task: %v", err)
			continue
		}

		c.processTask(ctx, task)
	}
}

func (c *RedisTaskConsumer) processTask(ctx context.Context, task workerDto.WorkerTask) {
	switch task.Type {
	case workerDto.TaskTypeBackup:
		application.HandleBackupTask(ctx, task, c.client, c.resultQueue)
	case workerDto.TaskTypeMeasureSize:
		application.HandleMeasureSizeTask(ctx, task, c.client, c.resultQueue)
	case workerDto.TaskTypeGetDiskUsage:
		application.HandleGetDiskUsage(ctx, task, c.client)
	case workerDto.TaskTypeSearchFiles:
		application.HandleSearchFiles(ctx, task, c.client)
	case workerDto.TaskTypeRestoreLocal:
		application.HandleRestoreLocalTask(ctx, task, c.client, c.resultQueue)
	case workerDto.TaskTypeListFiles:
		application.HandleListFiles(ctx, task, c.client)
	case workerDto.TaskTypeRestoreRemote:
		application.HandleRestoreRemoteTask(ctx, task, c.client, c.resultQueue)
	case workerDto.TaskTypePurge:
		application.HandlePurgeTask(ctx, task, c.client, c.resultQueue)
	default:
		log.Printf("Unknown task type: %s", task.Type)
	}
}
