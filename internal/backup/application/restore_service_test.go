package application

import (
	"context"
	"testing"

	"github.com/rrbarrero/justbackup/internal/backup/application/dto"
	"github.com/rrbarrero/justbackup/internal/backup/domain/entities"
	"github.com/rrbarrero/justbackup/internal/backup/domain/valueobjects"
	"github.com/stretchr/testify/assert"
)

func TestBackupRestoreService_Restore_Local_Simple_Success(t *testing.T) {
	ctx := context.Background()

	mockBackupRepo := new(MockBackupRepository)
	mockHostRepo := new(MockHostRepository)
	mockPublisher := new(MockTaskPublisher)

	hostService := NewHostService(mockHostRepo, mockBackupRepo)
	service := NewBackupRestoreService(mockBackupRepo, hostService, mockPublisher)

	// Setup data
	validHostID := entities.NewHostID()
	host := entities.NewHostWithID(validHostID, "Test Host", "localhost", "user", 22, "/backups", false)

	backupID := valueobjects.NewBackupID()
	backup, _ := entities.NewBackupWithID(
		backupID,
		validHostID,
		"/source/data",
		"dest_folder",
		entities.NewBackupSchedule("@daily"),
		[]string{},
		false,
		5,
		false,
	)
	// Backup is created with pending status, etc.

	req := dto.RestoreRequest{
		BackupID:     backupID.String(),
		RestoreType:  "local",
		Path:         "", // Full restore
		RestoreAddr:  "127.0.0.1:8080",
		RestoreToken: "token123",
	}

	// Mock expectations
	mockBackupRepo.On("FindByID", ctx, backupID).Return(backup, nil)

	// HostService.GetHost logic will be triggered
	mockHostRepo.On("Get", ctx, validHostID).Return(host, nil)
	mockBackupRepo.On("FindByHostID", ctx, validHostID).Return([]*entities.Backup{}, nil)

	expectedPath := "/mnt/backups/backups/dest_folder"

	mockPublisher.On("PublishRestoreTask", ctx, backup, expectedPath, "127.0.0.1:8080", "token123").Return("task-id-1", nil)

	// Execute
	taskID, err := service.Restore(ctx, req)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, "task-id-1", taskID)
	mockBackupRepo.AssertExpectations(t)
	mockHostRepo.AssertExpectations(t)
	mockPublisher.AssertExpectations(t)
}

func TestBackupRestoreService_Restore_Local_Encrypted_File_Success(t *testing.T) {
	ctx := context.Background()

	mockBackupRepo := new(MockBackupRepository)
	mockHostRepo := new(MockHostRepository)
	mockPublisher := new(MockTaskPublisher)

	hostService := NewHostService(mockHostRepo, mockBackupRepo)
	service := NewBackupRestoreService(mockBackupRepo, hostService, mockPublisher)

	validHostID := entities.NewHostID()
	host := entities.NewHostWithID(validHostID, "Test Host", "localhost", "user", 22, "/backups", false)

	backupID := valueobjects.NewBackupID()
	backup, _ := entities.NewBackupWithID(
		backupID,
		validHostID,
		"/source/data",
		"dest_folder",
		entities.NewBackupSchedule("@daily"),
		[]string{},
		false,
		5,
		false,
	)

	req := dto.RestoreRequest{
		BackupID:     backupID.String(),
		RestoreType:  "local",
		Path:         "specific/file.txt",
		RestoreAddr:  "",
		RestoreToken: "",
	}

	mockBackupRepo.On("FindByID", ctx, backupID).Return(backup, nil)
	mockHostRepo.On("Get", ctx, validHostID).Return(host, nil)
	mockBackupRepo.On("FindByHostID", ctx, validHostID).Return([]*entities.Backup{}, nil)

	expectedPath := "/mnt/backups/backups/dest_folder/specific/file.txt"

	mockPublisher.On("PublishRestoreTask", ctx, backup, expectedPath, "", "").Return("task-id-2", nil)

	taskID, err := service.Restore(ctx, req)

	assert.NoError(t, err)
	assert.Equal(t, "task-id-2", taskID)
}

func TestBackupRestoreService_Restore_Remote_Success(t *testing.T) {
	ctx := context.Background()

	mockBackupRepo := new(MockBackupRepository)
	mockHostRepo := new(MockHostRepository)
	mockPublisher := new(MockTaskPublisher)

	hostService := NewHostService(mockHostRepo, mockBackupRepo)
	service := NewBackupRestoreService(mockBackupRepo, hostService, mockPublisher)

	// Source Host
	sourceHostID := entities.NewHostID()
	sourceHost := entities.NewHostWithID(sourceHostID, "Source", "1.1.1.1", "user", 22, "/src", false)

	// Target Host
	targetHostID := entities.NewHostID()
	targetHost := entities.NewHostWithID(targetHostID, "Target", "2.2.2.2", "user", 22, "/dst", false)

	backupID := valueobjects.NewBackupID()
	backup, _ := entities.NewBackupWithID(
		backupID,
		sourceHostID,
		"/data",
		"backup-dest",
		entities.NewBackupSchedule("@daily"),
		[]string{},
		false,
		5,
		false,
	)

	req := dto.RestoreRequest{
		BackupID:     backupID.String(),
		RestoreType:  "remote",
		Path:         "",
		TargetHostID: targetHostID.String(), // Remote restore target
		TargetPath:   "/tmp/restore",
	}

	mockBackupRepo.On("FindByID", ctx, backupID).Return(backup, nil)

	// Mock target host fetch
	mockHostRepo.On("Get", ctx, targetHostID).Return(targetHost, nil)

	// Mock source host fetch (GetHost calls Get and FindByHostID)
	mockHostRepo.On("Get", ctx, sourceHostID).Return(sourceHost, nil)
	mockBackupRepo.On("FindByHostID", ctx, sourceHostID).Return([]*entities.Backup{}, nil)

	// Path calculation: /mnt/backups + /src + backup-dest = /mnt/backups/src/backup-dest
	expectedMsg := "/mnt/backups/src/backup-dest"

	mockPublisher.On("PublishRemoteRestoreTask", ctx, backup, expectedMsg, targetHost, "/tmp/restore").Return("task-remote-1", nil)

	taskID, err := service.Restore(ctx, req)

	assert.NoError(t, err)
	assert.Equal(t, "task-remote-1", taskID)
}

func TestBackupRestoreService_Restore_Remote_MissingTargetPath(t *testing.T) {
	ctx := context.Background()

	mockBackupRepo := new(MockBackupRepository)
	mockHostRepo := new(MockHostRepository)
	mockPublisher := new(MockTaskPublisher)

	hostService := NewHostService(mockHostRepo, mockBackupRepo)
	service := NewBackupRestoreService(mockBackupRepo, hostService, mockPublisher)

	sourceHostID := entities.NewHostID()
	sourceHost := entities.NewHostWithID(sourceHostID, "Source", "1.1.1.1", "user", 22, "/src", false)

	// Target same as source for this test case (optional)
	targetHostID := sourceHostID

	backupID := valueobjects.NewBackupID()
	backup, _ := entities.NewBackupWithID(
		backupID,
		sourceHostID,
		"/data",
		"backup-dest",
		entities.NewBackupSchedule("@daily"),
		[]string{},
		false,
		5,
		false,
	)

	req := dto.RestoreRequest{
		BackupID:     backupID.String(),
		RestoreType:  "remote",
		TargetHostID: targetHostID.String(),
		TargetPath:   "", // Missing!
	}

	mockBackupRepo.On("FindByID", ctx, backupID).Return(backup, nil)
	mockHostRepo.On("Get", ctx, sourceHostID).Return(sourceHost, nil).Once()

	// Source host is fetched first to calculate path
	mockHostRepo.On("Get", ctx, sourceHostID).Return(sourceHost, nil)
	mockBackupRepo.On("FindByHostID", ctx, sourceHostID).Return([]*entities.Backup{}, nil)

	// Then inside handleRemoteRestore, it fetches target host
	mockHostRepo.On("Get", ctx, targetHostID).Return(sourceHost, nil) // Return same host for simplicity

	// Then checks path -> Error

	_, err := service.Restore(ctx, req)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "target_path is required")
}

func TestBackupRestoreService_Restore_UnsupportedType(t *testing.T) {
	ctx := context.Background()
	mockBackupRepo := new(MockBackupRepository)
	mockHostRepo := new(MockHostRepository)
	mockPublisher := new(MockTaskPublisher)

	hostService := NewHostService(mockHostRepo, mockBackupRepo)
	service := NewBackupRestoreService(mockBackupRepo, hostService, mockPublisher)

	validHostID := entities.NewHostID()
	host := entities.NewHostWithID(validHostID, "h", "h", "u", 22, "p", false)

	backupID := valueobjects.NewBackupID()
	backup, _ := entities.NewBackupWithID(
		backupID,
		validHostID,
		"/source/data",
		"dest_folder",
		entities.NewBackupSchedule("@daily"),
		[]string{},
		false,
		5,
		false,
	)

	req := dto.RestoreRequest{
		BackupID:    backupID.String(),
		RestoreType: "cloud", // Unsupported
	}

	mockBackupRepo.On("FindByID", ctx, backupID).Return(backup, nil)

	mockHostRepo.On("Get", ctx, validHostID).Return(host, nil)
	mockBackupRepo.On("FindByHostID", ctx, validHostID).Return([]*entities.Backup{}, nil)

	_, err := service.Restore(ctx, req)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "restore type cloud not supported")
}
