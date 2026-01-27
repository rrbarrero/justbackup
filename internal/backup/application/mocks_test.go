package application

import (
	"context"

	"github.com/rrbarrero/justbackup/internal/backup/domain/entities"
	"github.com/rrbarrero/justbackup/internal/backup/domain/valueobjects"
	workerDto "github.com/rrbarrero/justbackup/internal/worker/dto"
	"github.com/stretchr/testify/mock"
)

// MockHostRepository is a mock of domain.HostRepository
type MockHostRepository struct {
	mock.Mock
}

func (m *MockHostRepository) Save(ctx context.Context, host *entities.Host) error {
	args := m.Called(ctx, host)
	return args.Error(0)
}

func (m *MockHostRepository) List(ctx context.Context) ([]*entities.Host, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entities.Host), args.Error(1)
}

func (m *MockHostRepository) Get(ctx context.Context, id entities.HostID) (*entities.Host, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.Host), args.Error(1)
}

func (m *MockHostRepository) GetByIDs(ctx context.Context, ids []entities.HostID) ([]*entities.Host, error) {
	args := m.Called(ctx, ids)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entities.Host), args.Error(1)
}

func (m *MockHostRepository) Update(ctx context.Context, host *entities.Host) error {
	args := m.Called(ctx, host)
	return args.Error(0)
}

func (m *MockHostRepository) Delete(ctx context.Context, id entities.HostID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockHostRepository) Count(ctx context.Context) (int64, error) {
	args := m.Called(ctx)
	return args.Get(0).(int64), args.Error(1)
}

// MockBackupRepository is a mock of domain.BackupRepository
type MockBackupRepository struct {
	mock.Mock
}

func (m *MockBackupRepository) Save(ctx context.Context, backup *entities.Backup) error {
	args := m.Called(ctx, backup)
	return args.Error(0)
}

func (m *MockBackupRepository) FindByID(ctx context.Context, id valueobjects.BackupID) (*entities.Backup, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.Backup), args.Error(1)
}

func (m *MockBackupRepository) FindByHostID(ctx context.Context, hostID entities.HostID) ([]*entities.Backup, error) {
	args := m.Called(ctx, hostID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entities.Backup), args.Error(1)
}

func (m *MockBackupRepository) FindAll(ctx context.Context) ([]*entities.Backup, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entities.Backup), args.Error(1)
}

func (m *MockBackupRepository) FindDueBackups(ctx context.Context) ([]*entities.Backup, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entities.Backup), args.Error(1)
}

func (m *MockBackupRepository) Delete(ctx context.Context, id valueobjects.BackupID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockBackupRepository) GetStats(ctx context.Context) (*entities.BackupStats, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.BackupStats), args.Error(1)
}

func (m *MockBackupRepository) GetFailedCountsByHost(ctx context.Context) (map[string]int, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(map[string]int), args.Error(1)
}

type MockTaskPublisher struct {
	mock.Mock
}

func (m *MockTaskPublisher) PublishMeasureTask(ctx context.Context, hostID string, path string) (string, error) {
	args := m.Called(ctx, hostID, path)
	return args.String(0), args.Error(1)
}

func (m *MockTaskPublisher) PublishSearchTask(ctx context.Context, pattern string) (string, error) {
	args := m.Called(ctx, pattern)
	return args.String(0), args.Error(1)
}

func (m *MockTaskPublisher) Publish(ctx context.Context, backup *entities.Backup) error {
	args := m.Called(ctx, backup)
	return args.Error(0)
}

func (m *MockTaskPublisher) PublishRestoreTask(ctx context.Context, backup *entities.Backup, path string, restoreAddr string, restoreToken string) (string, error) {
	args := m.Called(ctx, backup, path, restoreAddr, restoreToken)
	return args.String(0), args.Error(1)
}

func (m *MockTaskPublisher) PublishListFilesTask(ctx context.Context, path string) (string, error) {
	args := m.Called(ctx, path)
	return args.String(0), args.Error(1)
}

func (m *MockTaskPublisher) PublishRemoteRestoreTask(ctx context.Context, backup *entities.Backup, path string, targetHost *entities.Host, targetPath string) (string, error) {
	args := m.Called(ctx, backup, path, targetHost, targetPath)
	return args.String(0), args.Error(1)
}

type MockResultStore struct {
	mock.Mock
}

func (m *MockResultStore) GetTaskResult(ctx context.Context, taskID string) (string, error) {
	args := m.Called(ctx, taskID)
	return args.String(0), args.Error(1)
}

// MockBackupErrorRepository is a mock of BackupErrorRepository
type MockBackupErrorRepository struct {
	mock.Mock
}

func (m *MockBackupErrorRepository) Save(ctx context.Context, backupError *entities.BackupError) error {
	args := m.Called(ctx, backupError)
	return args.Error(0)
}

func (m *MockBackupErrorRepository) FindByBackupID(ctx context.Context, backupID valueobjects.BackupID) ([]*entities.BackupError, error) {
	args := m.Called(ctx, backupID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entities.BackupError), args.Error(1)
}

func (m *MockBackupErrorRepository) DeleteByBackupID(ctx context.Context, backupID valueobjects.BackupID) error {
	args := m.Called(ctx, backupID)
	return args.Error(0)
}

// MockWorkerQueryBus is a mock of WorkerQueryBus
type MockWorkerQueryBus struct {
	mock.Mock
}

func (m *MockWorkerQueryBus) SearchFiles(ctx context.Context, pattern string) (workerDto.SearchFilesResult, error) {
	args := m.Called(ctx, pattern)
	return args.Get(0).(workerDto.SearchFilesResult), args.Error(1)
}

func (m *MockWorkerQueryBus) ListFiles(ctx context.Context, path string) (workerDto.ListFilesResult, error) {
	args := m.Called(ctx, path)
	return args.Get(0).(workerDto.ListFilesResult), args.Error(1)
}
