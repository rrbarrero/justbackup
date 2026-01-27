package application

import (
	"context"
	"testing"

	"github.com/rrbarrero/justbackup/internal/backup/application/assembler"
	"github.com/rrbarrero/justbackup/internal/backup/domain/entities"
	"github.com/rrbarrero/justbackup/internal/backup/domain/valueobjects"
	workerDto "github.com/rrbarrero/justbackup/internal/worker/dto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestBackupSearchService_SearchFiles(t *testing.T) {
	mockRepo := new(MockBackupRepository)
	mockHostRepo := new(MockHostRepository)
	mockQueryBus := new(MockWorkerQueryBus)
	hostService := NewHostService(mockHostRepo, mockRepo)
	backupAssembler := assembler.NewBackupAssembler()
	service := NewBackupSearchService(mockRepo, hostService, mockQueryBus, backupAssembler)
	ctx := context.Background()

	// Setup data
	hostID := entities.NewHostID()
	host := entities.NewHostWithID(hostID, "Host1", "host1", "user", 22, "host_path", false)

	backupID := valueobjects.NewBackupID()
	backup, _ := entities.NewBackupWithID(backupID, hostID, "/src", "backup_dest", entities.NewBackupSchedule("@daily"), nil, false, 0, false)

	t.Run("success finding files associated with backup", func(t *testing.T) {
		// Mock search result from worker
		filesFound := []string{
			"/mnt/backups/host_path/backup_dest/file1.txt",
			"/mnt/backups/host_path/backup_dest/subdir/file2.txt",
			"/mnt/backups/other/file3.txt", // Should not match
		}

		searchResult := workerDto.SearchFilesResult{Files: filesFound}

		mockQueryBus.On("SearchFiles", ctx, "*.txt").Return(searchResult, nil)

		mockRepo.On("FindAll", ctx).Return([]*entities.Backup{backup}, nil)
		mockHostRepo.On("GetByIDs", ctx, mock.Anything).Return([]*entities.Host{host}, nil)

		results, err := service.SearchFiles(ctx, "*.txt")

		assert.NoError(t, err)
		assert.Len(t, results, 3)

		// Check matches
		assert.Equal(t, "/mnt/backups/host_path/backup_dest/file1.txt", results[0].Path)
		assert.NotNil(t, results[0].Backup)
		assert.Equal(t, backupID.String(), results[0].Backup.ID)

		assert.Equal(t, "/mnt/backups/host_path/backup_dest/subdir/file2.txt", results[1].Path)
		assert.NotNil(t, results[1].Backup)

		assert.Equal(t, "/mnt/backups/other/file3.txt", results[2].Path)
		assert.Nil(t, results[2].Backup)
	})
}
