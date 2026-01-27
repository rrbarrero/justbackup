package events

import (
	"time"
)

const (
	BackupCompletedEvent = "backup.completed"
	BackupFailedEvent    = "backup.failed"
)

type BackupCompleted struct {
	BackupID   string    `json:"backup_id"`
	HostID     string    `json:"host_id"`
	HostName   string    `json:"host_name"`
	SourcePath string    `json:"source_path"`
	OccurredAt time.Time `json:"occurred_at"`
	Size       string    `json:"size"`
}

func (e BackupCompleted) Name() string {
	return BackupCompletedEvent
}

func (e BackupCompleted) OccurredOn() time.Time {
	return e.OccurredAt
}

type BackupFailed struct {
	BackupID     string    `json:"backup_id"`
	HostID       string    `json:"host_id"`
	HostName     string    `json:"host_name"`
	SourcePath   string    `json:"source_path"`
	OccurredAt   time.Time `json:"occurred_at"`
	ErrorMessage string    `json:"error_message"`
}

func (e BackupFailed) Name() string {
	return BackupFailedEvent
}

func (e BackupFailed) OccurredOn() time.Time {
	return e.OccurredAt
}

func NewBackupCompleted(backupID string, hostID string, hostName string, sourcePath string, size string) BackupCompleted {
	return BackupCompleted{
		BackupID:   backupID,
		HostID:     hostID,
		HostName:   hostName,
		SourcePath: sourcePath,
		OccurredAt: time.Now(),
		Size:       size,
	}
}

func NewBackupFailed(backupID string, hostID string, hostName string, sourcePath string, errorMessage string) BackupFailed {
	return BackupFailed{
		BackupID:     backupID,
		HostID:       hostID,
		HostName:     hostName,
		SourcePath:   sourcePath,
		OccurredAt:   time.Now(),
		ErrorMessage: errorMessage,
	}
}
