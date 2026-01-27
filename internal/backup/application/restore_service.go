package application

import (
	"context"
	"fmt"
	"path"
	"strings"

	"github.com/rrbarrero/justbackup/internal/backup/application/dto"
	"github.com/rrbarrero/justbackup/internal/backup/domain/entities"
	"github.com/rrbarrero/justbackup/internal/backup/domain/interfaces"
	"github.com/rrbarrero/justbackup/internal/backup/domain/valueobjects"
)

type BackupRestoreService struct {
	repo        interfaces.BackupRepository
	hostService *HostService
	publisher   interfaces.TaskPublisher
}

func NewBackupRestoreService(
	repo interfaces.BackupRepository,
	hostService *HostService,
	publisher interfaces.TaskPublisher,
) *BackupRestoreService {
	return &BackupRestoreService{
		repo:        repo,
		hostService: hostService,
		publisher:   publisher,
	}
}

func (s *BackupRestoreService) Restore(ctx context.Context, req dto.RestoreRequest) (string, error) {
	bid, err := valueobjects.NewBackupIDFromString(req.BackupID)
	if err != nil {
		return "", err
	}

	backup, err := s.repo.FindByID(ctx, bid)
	if err != nil {
		return "", err
	}

	fullPath, err := s.calculateSourcePath(ctx, backup, req.Path)
	if err != nil {
		return "", err
	}

	if req.RestoreType == "local" {
		return s.handleLocalRestore(ctx, backup, fullPath, req)
	}

	if req.RestoreType == "remote" {
		return s.handleRemoteRestore(ctx, backup, fullPath, req)
	}

	return "", fmt.Errorf("restore type %s not supported", req.RestoreType)
}

func (s *BackupRestoreService) calculateSourcePath(ctx context.Context, backup *entities.Backup, reqPath string) (string, error) {
	hostResp, err := s.hostService.GetHost(ctx, backup.HostID().String())
	if err != nil {
		return "", err
	}
	fullPath := path.Join("/mnt/backups", hostResp.Path, backup.Destination())

	cleanPath := strings.Trim(reqPath, "./")

	if backup.Encrypted() && cleanPath == "" {
		if backup.Incremental() {
			fullPath = path.Join(fullPath, "latest.tar.gz.enc")
		} else {
			fullPath = fullPath + ".tar.gz.enc"
		}
	} else if reqPath != "" {
		fullPath = path.Join(fullPath, reqPath)
	}

	return fullPath, nil
}

func (s *BackupRestoreService) handleLocalRestore(ctx context.Context, backup *entities.Backup, fullPath string, req dto.RestoreRequest) (string, error) {
	return s.publisher.PublishRestoreTask(ctx, backup, fullPath, req.RestoreAddr, req.RestoreToken)
}

func (s *BackupRestoreService) handleRemoteRestore(ctx context.Context, backup *entities.Backup, fullPath string, req dto.RestoreRequest) (string, error) {
	hostID := req.TargetHostID
	if hostID == "" {
		hostID = backup.HostID().String()
	}

	targetHost, err := s.hostService.GetHostEntity(ctx, hostID)
	if err != nil {
		return "", err
	}

	targetPath := req.TargetPath
	if targetPath == "" {
		return "", fmt.Errorf("target_path is required for remote restore")
	}

	return s.publisher.PublishRemoteRestoreTask(ctx, backup, fullPath, targetHost, targetPath)
}
