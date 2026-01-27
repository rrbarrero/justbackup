package application

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/redis/go-redis/v9"
	"github.com/rrbarrero/justbackup/internal/shared/infrastructure/config"
	workerDto "github.com/rrbarrero/justbackup/internal/worker/dto"
)

func HandlePurgeTask(ctx context.Context, task workerDto.WorkerTask, redisClient *redis.Client, resultQueue string) {
	log.Printf("Purging backups for %s, path %s (retention: %d)", task.Host, task.Path, task.Retention)

	cfg, err := config.LoadWorkerConfig()
	if err != nil {
		log.Printf("CRITICAL: Failed to load worker config: %v", err)
		PublishResult(ctx, redisClient, resultQueue, workerDto.WorkerResult{
			Type:    workerDto.TaskTypePurge,
			TaskID:  task.TaskID,
			JobID:   task.JobID,
			Status:  "failed",
			Message: fmt.Sprintf("Worker configuration error: %v", err),
		})
		return
	}

	backupDir := NormalizePath(task.Destination, cfg.ContainerBackupRoot, task.HostPath)
	entries, err := os.ReadDir(backupDir)
	if err != nil {
		log.Printf("Failed to read backup dir %s: %v", backupDir, err)
		PublishResult(ctx, redisClient, resultQueue, workerDto.WorkerResult{
			Type:    workerDto.TaskTypePurge,
			TaskID:  task.TaskID,
			JobID:   task.JobID,
			Status:  "failed",
			Message: fmt.Sprintf("Failed to read backup dir: %v", err),
		})
		return
	}

	var backups []string
	for _, entry := range entries {
		if entry.IsDir() && entry.Name() != "latest" {
			// Basic check for timestamp format YYYY-MM-DD_HH-MM-SS
			if len(entry.Name()) == 19 && strings.Contains(entry.Name(), "_") {
				backups = append(backups, entry.Name())
			}
		}
	}

	// Sort backups ascending (oldest first)
	// Alphabetical sort works for YYYY-MM-DD_HH-MM-SS
	SortStrings(backups)

	if len(backups) <= task.Retention {
		log.Printf("Retention threshold not reached (%d/%d). Nothing to purge.", len(backups), task.Retention)
		PublishResult(ctx, redisClient, resultQueue, workerDto.WorkerResult{
			Type:    workerDto.TaskTypePurge,
			TaskID:  task.TaskID,
			JobID:   task.JobID,
			Status:  "completed",
			Message: "Nothing to purge",
		})
		return
	}

	toDelete := SelectBackupsToPurge(backups, task.Retention)
	log.Printf("Purging %d old backups", len(toDelete))

	var deletedCount int
	for _, b := range toDelete {
		path := filepath.Join(backupDir, b)
		log.Printf("Deleting old backup: %s", path)
		if err := os.RemoveAll(path); err != nil {
			log.Printf("Failed to delete %s: %v", path, err)
			continue
		}
		deletedCount++
	}

	PublishResult(ctx, redisClient, resultQueue, workerDto.WorkerResult{
		Type:    workerDto.TaskTypePurge,
		TaskID:  task.TaskID,
		JobID:   task.JobID,
		Status:  "completed",
		Message: fmt.Sprintf("Successfully purged %d backups", deletedCount),
	})
}

// SelectBackupsToPurge returns the list of backup directories that should be deleted
// based on the retention policy.
// It expects the 'backups' slice to be sorted chronologically (oldest first).
func SelectBackupsToPurge(backups []string, retention int) []string {
	if len(backups) <= retention {
		return []string{}
	}
	return backups[:len(backups)-retention]
}
