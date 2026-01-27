package main

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/redis/go-redis/v9"
	"github.com/rrbarrero/justbackup/internal/worker/application"
	workerDto "github.com/rrbarrero/justbackup/internal/worker/dto"
	"github.com/stretchr/testify/assert"
)

func TestNormalizePath(t *testing.T) {
	tests := []struct {
		name       string
		dest       string
		backupRoot string
		hostPath   string
		expected   string
	}{
		{
			name:       "basic path without host path",
			dest:       "mybackup",
			backupRoot: "/mnt/backups",
			hostPath:   "",
			expected:   "/mnt/backups/mybackup",
		},
		{
			name:       "path with trailing slashes",
			dest:       "mybackup/",
			backupRoot: "/mnt/backups/",
			hostPath:   "",
			expected:   "/mnt/backups/mybackup",
		},
		{
			name:       "path with host path",
			dest:       "mybackup",
			backupRoot: "/mnt/backups",
			hostPath:   "host1",
			expected:   "/mnt/backups/host1/mybackup",
		},
		{
			name:       "path with host path and trailing slashes",
			dest:       "/mybackup/",
			backupRoot: "/mnt/backups/",
			hostPath:   "/host1/",
			expected:   "/mnt/backups/host1/mybackup",
		},
		{
			name:       "nested destination path",
			dest:       "dir1/dir2/backup",
			backupRoot: "/mnt/backups",
			hostPath:   "host1",
			expected:   "/mnt/backups/host1/dir1/dir2/backup",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := application.NormalizePath(tt.dest, tt.backupRoot, tt.hostPath)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestHandleMeasureSizeTask_DevEnvironment(t *testing.T) {
	// Set ENVIRONMENT to dev to simulate measure size
	t.Setenv("ENVIRONMENT", "dev")

	redisClient := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})

	task := workerDto.WorkerTask{
		Type:   workerDto.TaskTypeMeasureSize,
		TaskID: "test-task-1",
		JobID:  "test-job-1",
		Host:   "localhost",
		User:   "user",
		Port:   22,
		Path:   "/test/path",
	}

	ctx := context.Background()

	// In dev mode, this should simulate success
	// We can't easily verify the Redis push without a mock or test Redis instance
	// but we can ensure it doesn't crash
	application.HandleMeasureSizeTask(ctx, task, redisClient, "test_queue")
}

func TestHandleBackupTask_DevEnvironment(t *testing.T) {
	// Set ENVIRONMENT to dev to simulate backup
	t.Setenv("ENVIRONMENT", "dev")
	t.Setenv("BACKUP_ROOT", "/tmp/test-backups")

	redisClient := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})

	task := workerDto.WorkerTask{
		Type:        workerDto.TaskTypeBackup,
		TaskID:      "test-task-1",
		JobID:       "test-job-1",
		Host:        "localhost",
		User:        "user",
		Port:        22,
		Path:        "/test/path",
		Destination: "test-dest",
		HostPath:    "test-host",
		Excludes:    []string{"*.tmp"},
		Incremental: false,
	}

	ctx := context.Background()

	// In dev mode, this should simulate success
	// We can't easily verify the Redis push without a mock or test Redis instance
	// but we can ensure it doesn't crash
	application.HandleBackupTask(ctx, task, redisClient, "test_queue")
}

func TestWorkerTaskSerialization(t *testing.T) {
	task := workerDto.WorkerTask{
		Type:        workerDto.TaskTypeBackup,
		TaskID:      "test-task-1",
		JobID:       "test-job-1",
		Host:        "localhost",
		User:        "user",
		Port:        22,
		Path:        "/test/path",
		Destination: "test-dest",
		HostPath:    "test-host",
		Excludes:    []string{"*.tmp", "*.log"},
		Incremental: true,
	}

	// Serialize
	data, err := json.Marshal(task)
	assert.NoError(t, err)

	// Deserialize
	var decoded workerDto.WorkerTask
	err = json.Unmarshal(data, &decoded)
	assert.NoError(t, err)

	// Verify
	assert.Equal(t, task.Type, decoded.Type)
	assert.Equal(t, task.TaskID, decoded.TaskID)
	assert.Equal(t, task.JobID, decoded.JobID)
	assert.Equal(t, task.Host, decoded.Host)
	assert.Equal(t, task.User, decoded.User)
	assert.Equal(t, task.Port, decoded.Port)
	assert.Equal(t, task.Path, decoded.Path)
	assert.Equal(t, task.Destination, decoded.Destination)
	assert.Equal(t, task.HostPath, decoded.HostPath)
	assert.Equal(t, task.Excludes, decoded.Excludes)
	assert.Equal(t, task.Incremental, decoded.Incremental)
}
