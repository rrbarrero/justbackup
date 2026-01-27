package entities

import (
	"time"

	"github.com/rrbarrero/justbackup/internal/backup/domain/valueobjects"
)

type BackupError struct {
	ID           string
	JobID        string
	BackupID     valueobjects.BackupID
	OccurredAt   time.Time
	ErrorMessage string
}

func NewBackupError(jobID string, backupID valueobjects.BackupID, errorMessage string) *BackupError {
	return &BackupError{
		JobID:        jobID,
		BackupID:     backupID,
		OccurredAt:   time.Now(),
		ErrorMessage: errorMessage,
	}
}
