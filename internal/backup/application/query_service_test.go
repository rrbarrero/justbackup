package application

import (
	"context"
	"errors"
	"testing"

	"github.com/rrbarrero/justbackup/internal/backup/application/assembler"
	"github.com/rrbarrero/justbackup/internal/backup/domain/entities"
	"github.com/rrbarrero/justbackup/internal/backup/domain/valueobjects"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestBackupQueryService_ListBackups(t *testing.T) {
	mockRepo := new(MockBackupRepository)
	mockHostRepo := new(MockHostRepository)
	mockErrorRepo := new(MockBackupErrorRepository)
	hostService := NewHostService(mockHostRepo, mockRepo)
	backupAssembler := assembler.NewBackupAssembler()
	service := NewBackupQueryService(mockRepo, hostService, mockErrorRepo, backupAssembler)
	ctx := context.Background()

	hostID1 := entities.NewHostID()
	hostID2 := entities.NewHostID()

	host1 := entities.NewHostWithID(hostID1, "Host 1", "host1.example.com", "user", 22, "path", false)

	schedule1 := entities.NewBackupSchedule("0 0 * * *")
	schedule2 := entities.NewBackupSchedule("*/5 * * * *")

	backup1, _ := entities.NewBackupWithID(valueobjects.NewBackupID(), hostID1, "/path/to/data1", "dest1", schedule1, []string{"*.log"}, false, 0, false)
	backup2, _ := entities.NewBackupWithID(valueobjects.NewBackupID(), hostID2, "/path/to/data2", "dest2", schedule2, nil, true, 4, false)

	t.Run("success listing multiple backups", func(t *testing.T) {
		backups := []*entities.Backup{backup1, backup2}
		hosts := []*entities.Host{host1}

		mockRepo.On("FindAll", ctx).Return(backups, nil).Once()
		mockHostRepo.On("GetByIDs", ctx, mock.AnythingOfType("[]entities.HostID")).Return(hosts, nil).Once()

		res, err := service.ListBackups(ctx, "", "")

		assert.NoError(t, err)
		assert.Len(t, res, 2)
		assert.Equal(t, backup1.ID().String(), res[0].ID)
		assert.Equal(t, backup1.HostID().String(), res[0].HostID)

		mockRepo.AssertExpectations(t)
		mockHostRepo.AssertExpectations(t)
	})

	t.Run("success listing zero backups", func(t *testing.T) {
		mockRepo.On("FindAll", ctx).Return([]*entities.Backup{}, nil).Once()
		mockHostRepo.On("GetByIDs", ctx, mock.AnythingOfType("[]entities.HostID")).Return([]*entities.Host{}, nil).Once()

		res, err := service.ListBackups(ctx, "", "")

		assert.NoError(t, err)
		assert.Len(t, res, 0)

		mockRepo.AssertExpectations(t)
	})

	t.Run("repository error on find all", func(t *testing.T) {
		expectedErr := errors.New("db error")
		mockRepo.On("FindAll", ctx).Return(nil, expectedErr).Once()

		res, err := service.ListBackups(ctx, "", "")

		assert.Error(t, err)
		assert.Equal(t, expectedErr, err)
		assert.Nil(t, res)

		mockRepo.AssertExpectations(t)
	})
}
