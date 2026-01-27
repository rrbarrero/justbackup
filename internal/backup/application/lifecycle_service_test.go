package application

import (
	"context"
	"errors"
	"testing"

	"github.com/rrbarrero/justbackup/internal/backup/application/assembler"
	"github.com/rrbarrero/justbackup/internal/backup/application/dto"
	"github.com/rrbarrero/justbackup/internal/backup/domain/entities"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestBackupLifecycleService_CreateBackup(t *testing.T) {
	mockRepo := new(MockBackupRepository)
	mockHostRepo := new(MockHostRepository)
	mockPublisher := new(MockTaskPublisher)
	hostService := NewHostService(mockHostRepo, mockRepo)
	backupAssembler := assembler.NewBackupAssembler()
	service := NewBackupLifecycleService(mockRepo, hostService, mockPublisher, backupAssembler)
	ctx := context.Background()

	validHostID := "d85f812d-7c2a-4c2f-b8d9-2e0f4f9f7d2f" // Example valid UUID
	req := dto.CreateBackupRequest{
		HostID:      validHostID,
		Path:        "/data/backup",
		Destination: "/documents",
		Schedule:    "0 0 * * *",
		Excludes:    []string{"*.tmp", "vendor/"},
	}

	host := entities.NewHost("Test Host", "test.example.com", "user", 22, "path", false)

	t.Run("success", func(t *testing.T) {
		mockHostRepo.On("Get", ctx, mock.AnythingOfType("entities.HostID")).Return(host, nil).Once()
		mockRepo.On("FindByHostID", ctx, mock.AnythingOfType("entities.HostID")).Return([]*entities.Backup{}, nil).Once()
		mockRepo.On("Save", ctx, mock.AnythingOfType("*entities.Backup")).Return(nil).Once()

		res, err := service.CreateBackup(ctx, req)

		assert.NoError(t, err)
		assert.NotNil(t, res)
		assert.NotEmpty(t, res.ID)
		assert.Equal(t, validHostID, res.HostID)
		assert.Equal(t, host.Name(), res.HostName)
		assert.Equal(t, host.Hostname(), res.HostAddress)
		assert.Equal(t, req.Path, res.Path)
		assert.Equal(t, req.Destination, res.Destination)
		assert.Equal(t, "pending", res.Status)
		assert.Equal(t, req.Schedule, res.Schedule)
		assert.Equal(t, req.Excludes, res.Excludes)

		mockRepo.AssertExpectations(t)
		mockHostRepo.AssertExpectations(t)
	})

	t.Run("invalid host ID", func(t *testing.T) {
		invalidReq := req
		invalidReq.HostID = "invalid-uuid"

		res, err := service.CreateBackup(ctx, invalidReq)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid id")
		assert.Nil(t, res)

		mockRepo.AssertNotCalled(t, "Save")
	})

	t.Run("repository error on save", func(t *testing.T) {
		expectedErr := errors.New("failed to save backup")
		mockHostRepo.On("Get", ctx, mock.AnythingOfType("entities.HostID")).Return(host, nil).Once()
		mockRepo.On("FindByHostID", ctx, mock.AnythingOfType("entities.HostID")).Return([]*entities.Backup{}, nil).Once()
		mockRepo.On("Save", ctx, mock.AnythingOfType("*entities.Backup")).Return(expectedErr).Once()

		res, err := service.CreateBackup(ctx, req)

		assert.Error(t, err)
		assert.Equal(t, expectedErr, err)
		assert.Nil(t, res)

		mockRepo.AssertExpectations(t)
	})
}
