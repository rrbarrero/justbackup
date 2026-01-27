package entities_test

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/rrbarrero/justbackup/internal/backup/domain/entities"
	"github.com/rrbarrero/justbackup/internal/backup/domain/valueobjects"
	"github.com/stretchr/testify/assert"
)

func TestNewBackupError(t *testing.T) {
	jobID := uuid.New().String()
	backupID := valueobjects.NewBackupID()
	errorMessage := "something went wrong"

	backupError := entities.NewBackupError(jobID, backupID, errorMessage)

	assert.NotNil(t, backupError)
	assert.Equal(t, jobID, backupError.JobID)
	assert.Equal(t, backupID, backupError.BackupID)
	assert.Equal(t, errorMessage, backupError.ErrorMessage)
	assert.WithinDuration(t, time.Now(), backupError.OccurredAt, 1*time.Second)
}
