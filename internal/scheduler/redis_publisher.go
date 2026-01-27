package scheduler

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/rrbarrero/justbackup/internal/backup/domain/entities"
	"github.com/rrbarrero/justbackup/internal/backup/domain/interfaces"
	workerDto "github.com/rrbarrero/justbackup/internal/worker/dto"
)

type RedisPublisher struct {
	client   *redis.Client
	queue    string
	hostRepo interfaces.HostRepository
}

func NewRedisPublisher(client *redis.Client, queue string, hostRepo interfaces.HostRepository) *RedisPublisher {
	return &RedisPublisher{
		client:   client,
		queue:    queue,
		hostRepo: hostRepo,
	}
}

func (p *RedisPublisher) Publish(ctx context.Context, backup *entities.Backup) error {
	host, err := p.hostRepo.Get(ctx, backup.HostID())
	if err != nil {
		return fmt.Errorf("failed to get host: %w", err)
	}

	task := p.createWorkerTask(backup, host)

	data, err := json.Marshal(task)
	if err != nil {
		return fmt.Errorf("failed to marshal backup task: %w", err)
	}

	if err := p.client.RPush(ctx, p.queue, data).Err(); err != nil {
		return fmt.Errorf("failed to publish backup task to redis: %w", err)
	}

	return nil
}

func (p *RedisPublisher) PublishMeasureTask(ctx context.Context, hostID string, path string) (string, error) {
	// Generate a random TaskID
	taskID := fmt.Sprintf("measure-%d", time.Now().UnixNano())

	hostIDObj, err := entities.NewHostIDFromString(hostID)
	if err != nil {
		return "", err
	}

	host, err := p.hostRepo.Get(ctx, hostIDObj)
	if err != nil {
		return "", fmt.Errorf("failed to find host: %w", err)
	}

	task := workerDto.WorkerTask{
		Type:   workerDto.TaskTypeMeasureSize,
		TaskID: taskID,
		JobID:  uuid.New().String(),
		Host:   host.Hostname(),
		User:   host.User(),
		Port:   host.Port(),
		Path:   path,
	}

	data, err := json.Marshal(task)
	if err != nil {
		return "", fmt.Errorf("failed to marshal task: %w", err)
	}

	if err := p.client.RPush(ctx, p.queue, data).Err(); err != nil {
		return "", err
	}

	return taskID, nil
}

func (p *RedisPublisher) PublishSearchTask(ctx context.Context, pattern string) (string, error) {
	taskID := uuid.New().String()

	task := workerDto.WorkerTask{
		Type:          workerDto.TaskTypeSearchFiles,
		TaskID:        taskID,
		SearchPattern: pattern,
	}

	data, err := json.Marshal(task)
	if err != nil {
		return "", fmt.Errorf("failed to marshal search task: %w", err)
	}

	if err := p.client.RPush(ctx, p.queue, data).Err(); err != nil {
		return "", err
	}

	return taskID, nil
}

func (p *RedisPublisher) PublishRestoreTask(ctx context.Context, backup *entities.Backup, path string, restoreAddr string, restoreToken string) (string, error) {
	taskID := uuid.New().String()

	host, err := p.hostRepo.Get(ctx, backup.HostID())
	if err != nil {
		return "", fmt.Errorf("failed to get host: %w", err)
	}

	task := workerDto.WorkerTask{
		Type:         workerDto.TaskTypeRestoreLocal,
		TaskID:       taskID,
		BackupID:     backup.ID().String(),
		JobID:        uuid.New().String(),
		Host:         host.Hostname(),
		User:         host.User(),
		Port:         host.Port(),
		Path:         path,
		RestoreAddr:  restoreAddr,
		RestoreToken: restoreToken,
		Encrypted:    backup.Encrypted(),
	}

	data, err := json.Marshal(task)
	if err != nil {
		return "", fmt.Errorf("failed to marshal restore task: %w", err)
	}

	if err := p.client.RPush(ctx, p.queue, data).Err(); err != nil {
		return "", fmt.Errorf("failed to publish restore task to redis: %w", err)
	}

	return taskID, nil
}

func (p *RedisPublisher) PublishListFilesTask(ctx context.Context, path string) (string, error) {
	taskID := uuid.New().String()

	task := workerDto.WorkerTask{
		Type:   workerDto.TaskTypeListFiles,
		TaskID: taskID,
		Path:   path,
	}

	data, err := json.Marshal(task)
	if err != nil {
		return "", fmt.Errorf("failed to marshal list files task: %w", err)
	}

	if err := p.client.RPush(ctx, p.queue, data).Err(); err != nil {
		return "", fmt.Errorf("failed to publish list files task to redis: %w", err)
	}

	return taskID, nil
}

func (p *RedisPublisher) PublishRemoteRestoreTask(ctx context.Context, backup *entities.Backup, path string, targetHost *entities.Host, targetPath string) (string, error) {
	taskID := uuid.New().String()

	// Get host where backup is stored (the physical files)
	// The worker must have access to the backup storage path.

	task := workerDto.WorkerTask{
		Type:       workerDto.TaskTypeRestoreRemote,
		TaskID:     taskID,
		BackupID:   backup.ID().String(),
		JobID:      uuid.New().String(),
		Path:       path, // This should be the path in the worker filesystem
		TargetHost: targetHost.Hostname(),
		TargetUser: targetHost.User(),
		TargetPort: targetHost.Port(),
		TargetPath: targetPath,
		Encrypted:  backup.Encrypted(),
	}

	data, err := json.Marshal(task)
	if err != nil {
		return "", fmt.Errorf("failed to marshal remote restore task: %w", err)
	}

	if err := p.client.RPush(ctx, p.queue, data).Err(); err != nil {
		return "", fmt.Errorf("failed to publish remote restore task to redis: %w", err)
	}

	return taskID, nil
}

func (p *RedisPublisher) PublishPurgeTask(ctx context.Context, backup *entities.Backup) error {
	host, err := p.hostRepo.Get(ctx, backup.HostID())
	if err != nil {
		return fmt.Errorf("failed to get host: %w", err)
	}

	task := workerDto.WorkerTask{
		Type:        workerDto.TaskTypePurge,
		TaskID:      uuid.New().String(),
		JobID:       uuid.New().String(),
		Host:        host.Hostname(),
		User:        host.User(),
		Port:        host.Port(),
		Path:        backup.Path(),
		Destination: backup.Destination(),
		HostPath:    host.Path(),
		Incremental: backup.Incremental(),
		Retention:   backup.Retention(),
	}

	data, err := json.Marshal(task)
	if err != nil {
		return fmt.Errorf("failed to marshal purge task: %w", err)
	}

	if err := p.client.RPush(ctx, p.queue, data).Err(); err != nil {
		return fmt.Errorf("failed to publish purge task to redis: %w", err)
	}

	return nil
}

func (p *RedisPublisher) createWorkerTask(backup *entities.Backup, host *entities.Host) workerDto.WorkerTask {
	hooks := make([]workerDto.HookTask, 0, len(backup.Hooks()))
	for _, h := range backup.Hooks() {
		hooks = append(hooks, workerDto.HookTask{
			Name:    h.Name,
			Phase:   string(h.Phase),
			Params:  h.Params,
			Enabled: h.Enabled,
		})
	}

	return workerDto.WorkerTask{
		Type:        workerDto.TaskTypeBackup,
		TaskID:      backup.ID().String(),
		JobID:       uuid.New().String(),
		Host:        host.Hostname(),
		User:        host.User(),
		Port:        host.Port(),
		Path:        backup.Path(),
		Destination: backup.Destination(),
		Excludes:    backup.Excludes(),
		HostPath:    host.Path(),
		Incremental: backup.Incremental(),
		Retention:   backup.Retention(),
		Encrypted:   backup.Encrypted(),
		Hooks:       hooks,
	}
}
