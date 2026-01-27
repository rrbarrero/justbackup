package entities_test

import (
	"testing"
	"time"

	"github.com/rrbarrero/justbackup/internal/backup/domain/entities"
	"github.com/rrbarrero/justbackup/internal/backup/domain/valueobjects"
	"github.com/stretchr/testify/assert"
)

func TestBackup(t *testing.T) {
	originalNowFunc := entities.NowFunc
	defer func() { entities.NowFunc = originalNowFunc }()

	t.Run("should create a new backup with correct initial values", func(t *testing.T) {
		fixedTime := time.Date(2023, time.January, 1, 12, 0, 0, 0, time.UTC)
		entities.NowFunc = func() time.Time { return fixedTime }

		hostID := entities.NewHostID()
		schedule := entities.NewBackupSchedule("0 0 * * *")
		path := "/data/source"
		destination := "/my-destination"
		excludes := []string{"*.log", "node_modules"}

		backup, err := entities.NewBackup(hostID, path, destination, schedule, excludes, false, 0, false)
		assert.NoError(t, err)

		assert.NotNil(t, backup)
		assert.NotEmpty(t, backup.ID().String())
		assert.Equal(t, hostID, backup.HostID())
		assert.Equal(t, path, backup.Path())
		assert.Equal(t, "/my-destination", backup.Destination())
		assert.Equal(t, valueobjects.BackupStatusPending, backup.Status())
		assert.Equal(t, schedule, backup.Schedule())
		assert.Equal(t, fixedTime, backup.CreatedAt())
		assert.Equal(t, fixedTime, backup.UpdatedAt())
		assert.NotNil(t, backup.NextRunAt())
		assert.True(t, backup.Enabled())
		assert.Equal(t, excludes, backup.Excludes())
	})

	t.Run("should correctly start a backup", func(t *testing.T) {
		fixedTime := time.Date(2023, time.January, 1, 12, 0, 0, 0, time.UTC)
		entities.NowFunc = func() time.Time { return fixedTime }

		backup, err := entities.NewBackup(entities.NewHostID(), "path", "dest", entities.NewBackupSchedule("0 0 * * *"), nil, false, 0, false)
		assert.NoError(t, err)

		updatedTime := time.Date(2023, time.January, 1, 12, 1, 0, 0, time.UTC)
		entities.NowFunc = func() time.Time { return updatedTime }

		err = backup.Start()

		assert.NoError(t, err)
		assert.Equal(t, valueobjects.BackupStatusRunning, backup.Status())
		assert.Equal(t, fixedTime, backup.CreatedAt()) // Should not change
		assert.Equal(t, updatedTime, backup.UpdatedAt())
	})

	t.Run("should not change state if already running", func(t *testing.T) {
		fixedTime := time.Date(2023, time.January, 1, 12, 0, 0, 0, time.UTC)
		entities.NowFunc = func() time.Time { return fixedTime }

		backup, err := entities.NewBackup(entities.NewHostID(), "path", "dest", entities.NewBackupSchedule("0 0 * * *"), nil, false, 0, false)
		assert.NoError(t, err)
		backup.Start()

		// Attempt to start again
		err = backup.Start()

		assert.NoError(t, err)
		assert.Equal(t, valueobjects.BackupStatusRunning, backup.Status())
		assert.Equal(t, fixedTime, backup.UpdatedAt()) // Should not change
	})

	t.Run("should correctly complete a backup", func(t *testing.T) {
		startTime := time.Date(2023, time.January, 1, 12, 0, 0, 0, time.UTC)
		entities.NowFunc = func() time.Time { return startTime }

		backup, err := entities.NewBackup(entities.NewHostID(), "path", "dest", entities.NewBackupSchedule("0 0 * * *"), nil, false, 0, false)
		assert.NoError(t, err)
		backup.Start()

		completeTime := time.Date(2023, time.January, 1, 12, 30, 0, 0, time.UTC)
		entities.NowFunc = func() time.Time { return completeTime }

		backup.Complete()

		assert.Equal(t, valueobjects.BackupStatusCompleted, backup.Status())
		assert.Equal(t, completeTime, backup.UpdatedAt())
	})

	t.Run("should correctly fail a backup", func(t *testing.T) {
		startTime := time.Date(2023, time.January, 1, 12, 0, 0, 0, time.UTC)
		entities.NowFunc = func() time.Time { return startTime }

		backup, err := entities.NewBackup(entities.NewHostID(), "path", "dest", entities.NewBackupSchedule("0 0 * * *"), nil, false, 0, false)
		assert.NoError(t, err)
		backup.Start()

		failTime := time.Date(2023, time.January, 1, 12, 15, 0, 0, time.UTC)
		entities.NowFunc = func() time.Time { return failTime }

		backup.Fail()

		assert.Equal(t, valueobjects.BackupStatusFailed, backup.Status())
		assert.Equal(t, failTime, backup.UpdatedAt())
	})

	t.Run("should enable and disable a backup", func(t *testing.T) {
		fixedTime := time.Date(2023, time.January, 1, 12, 0, 0, 0, time.UTC)
		entities.NowFunc = func() time.Time { return fixedTime }

		backup, err := entities.NewBackup(entities.NewHostID(), "path", "dest", entities.NewBackupSchedule("0 0 * * *"), nil, false, 0, false)
		assert.NoError(t, err)
		assert.True(t, backup.Enabled())
		assert.NotNil(t, backup.NextRunAt())

		// Disable
		disableTime := fixedTime.Add(1 * time.Minute)
		entities.NowFunc = func() time.Time { return disableTime }
		backup.Disable()

		assert.False(t, backup.Enabled())
		assert.Equal(t, disableTime, backup.UpdatedAt())
		assert.Nil(t, backup.NextRunAt())

		// Enable
		enableTime := fixedTime.Add(2 * time.Minute)
		entities.NowFunc = func() time.Time { return enableTime }
		backup.Enable()

		assert.True(t, backup.Enabled())
		assert.Equal(t, enableTime, backup.UpdatedAt())
		assert.NotNil(t, backup.NextRunAt())
	})

	t.Run("should handle different destination paths correctly", func(t *testing.T) {
		hostID := entities.NewHostID()
		schedule := entities.NewBackupSchedule("0 0 * * *")
		path := "/data/source"

		testCases := []struct {
			name         string
			destination  string
			expectedDest string
		}{
			{"empty destination", "", ""},
			{"single slash", "/", "/"},
			{"trailing slash", "data/", "data/"},
			{"leading slash", "/data", "/data"},
			{"both slashes", "/data/", "/data/"},
			{"no slashes", "data", "data"},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				backup, err := entities.NewBackup(hostID, path, tc.destination, schedule, nil, false, 0, false)
				assert.NoError(t, err)
				assert.Equal(t, tc.expectedDest, backup.Destination())
			})
		}
	})
}
