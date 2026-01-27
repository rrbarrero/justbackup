package scheduler

import (
	"testing"
	"time"

	"github.com/rrbarrero/justbackup/internal/backup/domain/entities"
	"github.com/rrbarrero/justbackup/internal/backup/domain/valueobjects"
	workerDto "github.com/rrbarrero/justbackup/internal/worker/dto"
	"github.com/stretchr/testify/assert"
)

func TestRedisPublisher_createWorkerTask(t *testing.T) {
	hostID := entities.NewHostID()
	host := entities.NewHost("Test Host", "example.com", "user", 22, "host/path", false)

	backupID := valueobjects.NewBackupID()
	// Using RestoreBackup to create a backup with specific state
	backup := entities.RestoreBackup(
		backupID,
		hostID,
		"/source/path",
		"dest/path",
		valueobjects.BackupStatusPending,
		entities.BackupSchedule{CronExpression: "0 0 * * *"},
		time.Now(),
		time.Now(),
		nil,
		[]string{"*.tmp"},
		true,
		false,
		"",
		0,
		false,
	)

	publisher := &RedisPublisher{}

	task := publisher.createWorkerTask(backup, host)

	assert.Equal(t, workerDto.TaskTypeBackup, task.Type)
	assert.Equal(t, backupID.String(), task.TaskID)
	assert.Equal(t, "example.com", task.Host)
	assert.Equal(t, "user", task.User)
	assert.Equal(t, 22, task.Port)
	assert.Equal(t, "/source/path", task.Path)
	assert.Equal(t, "dest/path", task.Destination)
	assert.Equal(t, []string{"*.tmp"}, task.Excludes)
	assert.Equal(t, "host/path", task.HostPath)
}
