package scheduler

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/rrbarrero/justbackup/internal/backup/application"
	"github.com/rrbarrero/justbackup/internal/backup/domain/entities"
	"github.com/rrbarrero/justbackup/internal/backup/domain/events"
	"github.com/rrbarrero/justbackup/internal/backup/domain/interfaces"
	"github.com/rrbarrero/justbackup/internal/backup/domain/valueobjects"
	"github.com/rrbarrero/justbackup/internal/shared/infrastructure/event"
	"github.com/rrbarrero/justbackup/internal/shared/infrastructure/websocket"
	workerDto "github.com/rrbarrero/justbackup/internal/worker/dto"
)

type ResultConsumer struct {
	client          *redis.Client
	queue           string
	backupRepo      interfaces.BackupRepository
	hostService     *application.HostService
	backupErrorRepo interfaces.BackupErrorRepository
	hub             *websocket.Hub
	eventBus        *event.RedisEventBus
}

func NewResultConsumer(client *redis.Client, queue string, backupRepo interfaces.BackupRepository, hostService *application.HostService, backupErrorRepo interfaces.BackupErrorRepository, hub *websocket.Hub, eventBus *event.RedisEventBus) *ResultConsumer {
	return &ResultConsumer{
		client:          client,
		queue:           queue,
		backupRepo:      backupRepo,
		hostService:     hostService,
		backupErrorRepo: backupErrorRepo,
		hub:             hub,
		eventBus:        eventBus,
	}
}

func (c *ResultConsumer) Start(ctx context.Context) {
	log.Printf("ResultConsumer started, listening on %s", c.queue)
	for {
		select {
		case <-ctx.Done():
			log.Println("ResultConsumer stopped")
			return
		default:
			// BLPOP blocks until a result is available
			redisResult, err := c.client.BLPop(ctx, 1*time.Second, c.queue).Result()
			if err != nil {
				if err != redis.Nil {
					log.Printf("Redis BLPop error: %v", err)
				}
				continue
			}

			payload := redisResult[1]
			log.Printf("Received result: %s", payload)

			var result workerDto.WorkerResult
			if err := json.Unmarshal([]byte(payload), &result); err != nil {
				log.Printf("Failed to unmarshal result: %v", err)
				continue
			}

			if err := c.processResult(ctx, result); err != nil {
				log.Printf("Failed to process result for task %s: %v", result.TaskID, err)
			}
		}
	}
}

func (c *ResultConsumer) processResult(ctx context.Context, result workerDto.WorkerResult) error {
	switch result.Type {
	case workerDto.TaskTypeBackup:
		return c.processBackupResult(ctx, result)
	case workerDto.TaskTypeMeasureSize, workerDto.TaskTypeRestoreRemote, workerDto.TaskTypeRestoreLocal, workerDto.TaskTypeListFiles, workerDto.TaskTypePurge:
		// These types only need to be stored in Redis for the requester to pick up,
		// or they were already handled by another mechanism.
		// Storing them avoids the "unknown result" error.
		return c.processGenericResult(ctx, result)
	default:
		return fmt.Errorf("unknown result type: %s", result.Type)
	}
}

func (c *ResultConsumer) processGenericResult(ctx context.Context, result workerDto.WorkerResult) error {
	// Store result in Redis with a TTL so the caller can check status
	key := fmt.Sprintf("task_result:%s", result.TaskID)
	data, err := json.Marshal(result)
	if err != nil {
		return err
	}
	log.Printf("Storing generic result for task %s in Redis", result.TaskID)
	return c.client.Set(ctx, key, data, 10*time.Minute).Err()
}

func (c *ResultConsumer) processBackupResult(ctx context.Context, result workerDto.WorkerResult) error {
	backupID, err := valueobjects.NewBackupIDFromString(result.TaskID)
	if err != nil {
		return fmt.Errorf("invalid backup ID: %w", err)
	}

	backup, err := c.backupRepo.FindByID(ctx, backupID)
	if err != nil {
		return fmt.Errorf("failed to find backup: %w", err)
	}

	hostName := "Unknown"
	hostResp, err := c.hostService.GetHost(ctx, backup.HostID().String())
	if err == nil {
		hostName = hostResp.Name
	} else {
		log.Printf("Failed to fetch host info for backup %s: %v", backup.ID(), err)
	}

	if result.Status == "completed" {
		if err := backup.Complete(); err != nil {
			log.Printf("Failed to complete backup %s: %v", backup.ID(), err)
		}
		size := ""
		if dataMap, ok := result.Data.(map[string]interface{}); ok {
			if s, ok := dataMap["size"].(string); ok {
				backup.SetSize(s)
				size = s
			}
		}
		// Publish BackupCompleted event
		event := events.NewBackupCompleted(backup.ID().String(), backup.HostID().String(), hostName, backup.Path(), size)
		if err := c.eventBus.Publish(ctx, event); err != nil {
			log.Printf("Failed to publish BackupCompleted event: %v", err)
		}
	} else {
		backup.Fail()
		// Save error
		backupError := entities.NewBackupError(result.JobID, backup.ID(), result.Message)
		if err := c.backupErrorRepo.Save(ctx, backupError); err != nil {
			log.Printf("Failed to save backup error: %v", err)
		}
		// Publish BackupFailed event
		event := events.NewBackupFailed(backup.ID().String(), backup.HostID().String(), hostName, backup.Path(), result.Message)
		if err := c.eventBus.Publish(ctx, event); err != nil {
			log.Printf("Failed to publish BackupFailed event: %v", err)
		}
	}

	if err := c.backupRepo.Save(ctx, backup); err != nil {
		return fmt.Errorf("failed to save backup: %w", err)
	}

	// Broadcast result
	msg := map[string]string{
		"type":      "backup_completed",
		"backup_id": backup.ID().String(),
		"status":    string(backup.Status()),
	}
	if result.Status != "completed" {
		msg["type"] = "backup_failed"
	}

	data, err := json.Marshal(msg)
	if err == nil {
		c.hub.Broadcast(data)
	}

	return nil
}
