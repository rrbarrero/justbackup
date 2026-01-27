package application

import (
	"context"
	"errors"
	"testing"

	"github.com/rrbarrero/justbackup/internal/backup/application/dto"
	"github.com/rrbarrero/justbackup/internal/backup/domain/entities"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestHostService_CreateHost(t *testing.T) {
	mockRepo := new(MockHostRepository)
	mockBackupRepo := new(MockBackupRepository)
	service := NewHostService(mockRepo, mockBackupRepo)
	ctx := context.Background()

	req := dto.CreateHostRequest{
		Name:     "test-host",
		Hostname: "localhost",
		User:     "user",
		Port:     22,
	}

	t.Run("success", func(t *testing.T) {
		// We use mock.AnythingOfType because the ID is generated inside NewHost and we can't know it in advance
		mockRepo.On("Save", ctx, mock.AnythingOfType("*entities.Host")).Return(nil).Once()

		res, err := service.CreateHost(ctx, req)

		assert.NoError(t, err)
		assert.NotNil(t, res)
		assert.NotEmpty(t, res.ID)
		assert.Equal(t, req.Name, res.Name)
		assert.Equal(t, req.Hostname, res.Hostname)
		assert.Equal(t, req.User, res.User)
		assert.Equal(t, req.Port, res.Port)

		mockRepo.AssertExpectations(t)
	})

	t.Run("repository error on save", func(t *testing.T) {
		expectedErr := errors.New("db error")
		mockRepo.On("Save", ctx, mock.AnythingOfType("*entities.Host")).Return(expectedErr).Once()

		res, err := service.CreateHost(ctx, req)

		assert.Error(t, err)
		assert.Equal(t, expectedErr, err)
		assert.Nil(t, res)

		mockRepo.AssertExpectations(t)
	})
}

func TestHostService_ListHosts(t *testing.T) {
	mockRepo := new(MockHostRepository)
	mockBackupRepo := new(MockBackupRepository)
	service := NewHostService(mockRepo, mockBackupRepo)
	ctx := context.Background()

	host1 := entities.NewHost("host1", "localhost1", "user1", 22, "host1", false)
	host2 := entities.NewHost("host2", "localhost2", "user2", 22, "host2", false)

	t.Run("success listing multiple hosts", func(t *testing.T) {
		hosts := []*entities.Host{host1, host2}
		mockRepo.On("List", ctx).Return(hosts, nil).Once()
		mockBackupRepo.On("GetFailedCountsByHost", ctx).Return(make(map[string]int), nil).Once()

		res, err := service.ListHosts(ctx)

		assert.NoError(t, err)
		assert.Len(t, res, 2)
		assert.Equal(t, "host1", res[0].Name)
		assert.Equal(t, "localhost1", res[0].Hostname)
		assert.Equal(t, "host2", res[1].Name)
		assert.Equal(t, "localhost2", res[1].Hostname)

		mockRepo.AssertExpectations(t)
	})

	t.Run("success listing zero hosts", func(t *testing.T) {
		mockRepo.On("List", ctx).Return([]*entities.Host{}, nil).Once()
		mockBackupRepo.On("GetFailedCountsByHost", ctx).Return(make(map[string]int), nil).Once()

		res, err := service.ListHosts(ctx)

		assert.NoError(t, err)
		assert.Len(t, res, 0)

		mockRepo.AssertExpectations(t)
	})

	t.Run("repository error on find all", func(t *testing.T) {
		expectedErr := errors.New("db error")
		mockRepo.On("List", ctx).Return(nil, expectedErr).Once()
		// Even if List fails, HostService might not call GetFailedCountsByHost
		// because of the "if err != nil { return nil, err }" in ListHosts.
		// Let's check the code:
		// func (s *HostService) ListHosts(ctx context.Context) ([]*dto.HostResponse, error) {
		// 	hosts, err := s.repo.List(ctx)
		// 	if err != nil {
		// 		return nil, err
		// 	}
		// So it shouldn't be called here.

		res, err := service.ListHosts(ctx)

		assert.Error(t, err)
		assert.Equal(t, expectedErr, err)
		assert.Nil(t, res)

		mockRepo.AssertExpectations(t)
	})
}
