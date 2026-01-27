package http_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/rrbarrero/justbackup/internal/backup/application"
	"github.com/rrbarrero/justbackup/internal/backup/application/assembler"
	"github.com/rrbarrero/justbackup/internal/backup/application/dto"
	"github.com/rrbarrero/justbackup/internal/backup/domain/entities"
	"github.com/rrbarrero/justbackup/internal/backup/domain/valueobjects"
	"github.com/rrbarrero/justbackup/internal/backup/infrastructure/persistence/memory"
	backupHttp "github.com/rrbarrero/justbackup/internal/backup/interfaces/http"
	workerDto "github.com/rrbarrero/justbackup/internal/worker/dto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockTaskPublisher
type MockTaskPublisher struct {
	mock.Mock
}

func (m *MockTaskPublisher) Publish(ctx context.Context, backup *entities.Backup) error {
	args := m.Called(ctx, backup)
	return args.Error(0)
}

func (m *MockTaskPublisher) PublishSearchTask(ctx context.Context, pattern string) (string, error) {
	args := m.Called(ctx, pattern)
	return args.String(0), args.Error(1)
}

func (m *MockTaskPublisher) PublishMeasureTask(ctx context.Context, hostID string, path string) (string, error) {
	args := m.Called(ctx, hostID, path)
	return args.String(0), args.Error(1)
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

// MockResultStore
type MockResultStore struct {
	mock.Mock
}

func (m *MockResultStore) GetTaskResult(ctx context.Context, taskID string) (string, error) {
	args := m.Called(ctx, taskID)
	return args.String(0), args.Error(1)
}

// MockWorkerQueryBus
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

func setupBackupHandler() (*backupHttp.BackupHandler, *memory.BackupRepositoryMemory, *memory.HostRepositoryMemory, *MockTaskPublisher, *MockResultStore, *MockWorkerQueryBus) {
	// Create an empty backup repository for testing
	backupRepo := memory.NewBackupRepositoryMemoryEmpty()
	backupErrorRepo := memory.NewBackupErrorRepositoryMemory()
	hostRepo := memory.NewHostRepositoryMemory()
	publisher := new(MockTaskPublisher)
	resultStore := new(MockResultStore)
	queryBus := new(MockWorkerQueryBus)
	backupAssembler := assembler.NewBackupAssembler()

	hostService := application.NewHostService(hostRepo, backupRepo)

	lifecycleService := application.NewBackupLifecycleService(backupRepo, hostService, publisher, backupAssembler)
	queryService := application.NewBackupQueryService(backupRepo, hostService, backupErrorRepo, backupAssembler)
	searchService := application.NewBackupSearchService(backupRepo, hostService, queryBus, backupAssembler)
	restoreService := application.NewBackupRestoreService(backupRepo, hostService, publisher)
	taskService := application.NewBackupTaskService(publisher, resultStore)
	hookService := application.NewBackupHookService(backupRepo, backupAssembler)

	handler := backupHttp.NewBackupHandler(
		lifecycleService,
		queryService,
		searchService,
		restoreService,
		taskService,
		hookService,
	)

	return handler, backupRepo, hostRepo, publisher, resultStore, queryBus
}

func TestCreateBackup(t *testing.T) {
	handler, _, hostRepo, _, _, _ := setupBackupHandler()

	// Create a host first
	host := entities.NewHost("Test Host", "test.example.com", "user", 22, "path", false)
	_ = hostRepo.Save(context.Background(), host)

	reqBody := dto.CreateBackupRequest{
		HostID:      host.ID().String(),
		Path:        "/source",
		Destination: "/dest",
		Schedule:    "0 0 * * *",
		Excludes:    []string{"*.tmp"},
		Incremental: true,
	}
	body, _ := json.Marshal(reqBody)

	req, _ := http.NewRequest("POST", "/backups", bytes.NewBuffer(body))
	rr := httptest.NewRecorder()

	handler.Create(rr, req)

	assert.Equal(t, http.StatusCreated, rr.Code)

	var resp dto.BackupResponse
	err := json.Unmarshal(rr.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, reqBody.HostID, resp.HostID)
	assert.Equal(t, reqBody.Path, resp.Path)
	assert.Equal(t, reqBody.Destination, resp.Destination)
	assert.Equal(t, reqBody.Schedule, resp.Schedule)
	assert.Equal(t, reqBody.Excludes, resp.Excludes)
	assert.True(t, resp.Incremental) // Check incremental flag
}

func TestListBackups(t *testing.T) {
	handler, backupRepo, hostRepo, _, _, _ := setupBackupHandler()

	// Create a host
	host := entities.NewHost("Test Host", "test.example.com", "user", 22, "path", false)
	_ = hostRepo.Save(context.Background(), host)

	// Create a backup
	schedule := entities.NewBackupSchedule("0 0 * * *")
	backup, err := entities.NewBackup(host.ID(), "/source", "/dest", schedule, []string{}, false, 0, false)
	assert.NoError(t, err)
	_ = backupRepo.Save(context.Background(), backup)

	req, _ := http.NewRequest("GET", "/backups", nil)
	rr := httptest.NewRecorder()

	handler.List(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var backups []dto.BackupResponse
	err = json.Unmarshal(rr.Body.Bytes(), &backups)
	assert.NoError(t, err)
	// Repository is now empty, so we should have exactly 1 backup
	assert.Len(t, backups, 1)
	assert.Equal(t, backup.ID().String(), backups[0].ID)
}

func TestUpdateBackup(t *testing.T) {
	handler, backupRepo, hostRepo, _, _, _ := setupBackupHandler()

	// Create a host
	host := entities.NewHost("Test Host", "test.example.com", "user", 22, "path", false)
	_ = hostRepo.Save(context.Background(), host)

	// Create a backup
	schedule := entities.NewBackupSchedule("0 0 * * *")
	backup, err := entities.NewBackup(host.ID(), "/source", "/dest", schedule, []string{}, false, 0, false)
	assert.NoError(t, err)
	_ = backupRepo.Save(context.Background(), backup)

	// Update the backup
	updateBody := dto.UpdateBackupRequest{
		ID:          backup.ID().String(),
		Path:        "/new-source",
		Destination: "/new-dest",
		Schedule:    "0 12 * * *",
		Excludes:    []string{"*.log"},
		Incremental: true,
	}
	body, _ := json.Marshal(updateBody)
	req, _ := http.NewRequest("PUT", "/backups/"+backup.ID().String(), bytes.NewBuffer(body))
	req.SetPathValue("id", backup.ID().String())
	rr := httptest.NewRecorder()

	handler.Update(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var updatedBackup dto.BackupResponse
	err = json.Unmarshal(rr.Body.Bytes(), &updatedBackup)
	assert.NoError(t, err)
	assert.Equal(t, "/new-source", updatedBackup.Path)
	assert.Equal(t, "/new-dest", updatedBackup.Destination)
	assert.Equal(t, "0 12 * * *", updatedBackup.Schedule)
	assert.Equal(t, []string{"*.log"}, updatedBackup.Excludes)
	// Incremental field is now part of BackupResponse
	assert.True(t, updatedBackup.Incremental)
}

func TestDeleteBackup(t *testing.T) {
	handler, backupRepo, hostRepo, _, _, _ := setupBackupHandler()

	// Create a host
	host := entities.NewHost("Test Host", "test.example.com", "user", 22, "path", false)
	_ = hostRepo.Save(context.Background(), host)

	// Create a backup
	schedule := entities.NewBackupSchedule("0 0 * * *")
	backup, err := entities.NewBackup(host.ID(), "/source", "/dest", schedule, []string{}, false, 0, false)
	assert.NoError(t, err)
	_ = backupRepo.Save(context.Background(), backup)

	// Delete the backup
	req, _ := http.NewRequest("DELETE", "/backups/"+backup.ID().String(), nil)
	req.SetPathValue("id", backup.ID().String())
	rr := httptest.NewRecorder()

	handler.Delete(rr, req)

	assert.Equal(t, http.StatusNoContent, rr.Code)

	// Verify it's gone
	_, err = backupRepo.FindByID(context.Background(), backup.ID())
	assert.Error(t, err)
}

func TestRunBackup(t *testing.T) {
	handler, backupRepo, hostRepo, publisher, _, _ := setupBackupHandler()

	// Create a host
	host := entities.NewHost("Test Host", "test.example.com", "user", 22, "path", false)
	_ = hostRepo.Save(context.Background(), host)

	// Create a backup
	schedule := entities.NewBackupSchedule("0 0 * * *")
	backup, err := entities.NewBackup(host.ID(), "/source", "/dest", schedule, []string{}, false, 0, false)
	assert.NoError(t, err)
	_ = backupRepo.Save(context.Background(), backup)

	// Mock publisher
	publisher.On("Publish", mock.Anything, mock.Anything).Return(nil)

	// Run the backup - need to call Run directly since HandleBackupID expects the URL path
	req, _ := http.NewRequest("POST", "/backups/"+backup.ID().String()+"/run", nil)
	req.SetPathValue("id", backup.ID().String())
	rr := httptest.NewRecorder()

	// Call Run directly instead of going through HandleBackupID
	handler.Run(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var resp map[string]string
	err = json.Unmarshal(rr.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, backup.ID().String(), resp["task_id"])

	publisher.AssertExpectations(t)
}

func TestMeasureSize(t *testing.T) {
	handler, _, hostRepo, publisher, _, _ := setupBackupHandler()

	// Create a host
	host := entities.NewHost("Test Host", "test.example.com", "user", 22, "path", false)
	_ = hostRepo.Save(context.Background(), host)

	// Mock publisher
	expectedTaskID := "task-123"
	publisher.On("PublishMeasureTask", mock.Anything, host.ID().String(), "/path/to/measure").Return(expectedTaskID, nil)

	reqBody := map[string]string{"path": "/path/to/measure"}
	body, _ := json.Marshal(reqBody)

	req, _ := http.NewRequest("POST", "/hosts/"+host.ID().String()+"/measure", bytes.NewBuffer(body))
	req.SetPathValue("id", host.ID().String())
	rr := httptest.NewRecorder()

	handler.MeasureSize(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var resp map[string]string
	err := json.Unmarshal(rr.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, expectedTaskID, resp["task_id"])

	publisher.AssertExpectations(t)
}

func TestGetTaskResult(t *testing.T) {
	handler, _, _, _, resultStore, _ := setupBackupHandler()

	taskID := "task-123"
	expectedResult := `{"status": "completed", "size": "1.5G"}`

	// Mock result store
	resultStore.On("GetTaskResult", mock.Anything, taskID).Return(expectedResult, nil)

	req, _ := http.NewRequest("GET", "/tasks/"+taskID, nil)
	req.SetPathValue("id", taskID)
	rr := httptest.NewRecorder()

	handler.GetTaskResult(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Equal(t, expectedResult, rr.Body.String())

	resultStore.AssertExpectations(t)
}

func TestRunHostBackups(t *testing.T) {
	handler, backupRepo, hostRepo, publisher, _, _ := setupBackupHandler()

	// Create a host
	host := entities.NewHost("Test Host", "test.example.com", "user", 22, "path", false)
	_ = hostRepo.Save(context.Background(), host)

	// Create two backups for this host
	schedule := entities.NewBackupSchedule("0 0 * * *")
	backup1, err := entities.NewBackup(host.ID(), "/source1", "/dest1", schedule, []string{}, false, 0, false)
	assert.NoError(t, err)
	backup2, err := entities.NewBackup(host.ID(), "/source2", "/dest2", schedule, []string{}, false, 0, false)
	assert.NoError(t, err)
	_ = backupRepo.Save(context.Background(), backup1)
	_ = backupRepo.Save(context.Background(), backup2)

	// Mock publisher
	publisher.On("Publish", mock.Anything, mock.Anything).Return(nil)

	// Run all backups for the host
	req, _ := http.NewRequest("POST", "/hosts/"+host.ID().String()+"/run", nil)
	req.SetPathValue("id", host.ID().String())
	rr := httptest.NewRecorder()

	handler.RunHostBackups(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var resp map[string][]string
	err = json.Unmarshal(rr.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Len(t, resp["task_ids"], 2)
	assert.Contains(t, resp["task_ids"], backup1.ID().String())
	assert.Contains(t, resp["task_ids"], backup2.ID().String())

	publisher.AssertExpectations(t)
}
