package application

import (
	"context"

	"github.com/google/uuid"
	"github.com/rrbarrero/justbackup/internal/backup/application/assembler"
	"github.com/rrbarrero/justbackup/internal/backup/application/dto"
	"github.com/rrbarrero/justbackup/internal/backup/domain/entities"
	"github.com/rrbarrero/justbackup/internal/backup/domain/interfaces"
	"github.com/rrbarrero/justbackup/internal/backup/domain/valueobjects"
)

type BackupLifecycleService struct {
	repo        interfaces.BackupRepository
	hostService *HostService
	publisher   interfaces.TaskPublisher
	assembler   *assembler.BackupAssembler
}

func NewBackupLifecycleService(
	repo interfaces.BackupRepository,
	hostService *HostService,
	publisher interfaces.TaskPublisher,
	assembler *assembler.BackupAssembler,
) *BackupLifecycleService {
	return &BackupLifecycleService{
		repo:        repo,
		hostService: hostService,
		publisher:   publisher,
		assembler:   assembler,
	}
}

func (s *BackupLifecycleService) CreateBackup(ctx context.Context, req dto.CreateBackupRequest) (*dto.BackupResponse, error) {
	hostID, err := entities.NewHostIDFromString(req.HostID)
	if err != nil {
		return nil, err
	}

	hostResp, err := s.hostService.GetHost(ctx, req.HostID)
	if err != nil {
		return nil, err
	}

	schedule := entities.NewBackupSchedule(req.Schedule)
	backup, err := entities.NewBackup(hostID, req.Path, req.Destination, schedule, req.Excludes, req.Incremental, req.Retention, req.Encrypted)
	if err != nil {
		return nil, err
	}

	// Handle embedded hooks
	for _, h := range req.Hooks {
		hookUUID, _ := uuid.Parse(backup.ID().String())
		hook := entities.NewBackupHook(hookUUID, h.Name, entities.HookPhase(h.Phase), h.Params)
		hook.Enabled = h.Enabled
		backup.AddHook(hook)
	}

	if err := s.repo.Save(ctx, backup); err != nil {
		return nil, err
	}

	return s.assembler.ToBackupResponse(backup, hostResp.Name, hostResp.Hostname), nil
}

func (s *BackupLifecycleService) UpdateBackup(ctx context.Context, req dto.UpdateBackupRequest) (*dto.BackupResponse, error) {
	id, err := valueobjects.NewBackupIDFromString(req.ID)
	if err != nil {
		return nil, err
	}

	backup, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	hostResp, err := s.hostService.GetHost(ctx, backup.HostID().String())
	if err != nil {
		return nil, err
	}

	schedule := entities.NewBackupSchedule(req.Schedule)
	if err := backup.Update(req.Path, req.Destination, schedule, req.Excludes, req.Incremental, req.Retention, req.Encrypted); err != nil {
		return nil, err
	}

	// Handle embedded hooks (replace all)
	newHooks := make([]*entities.BackupHook, 0, len(req.Hooks))
	for _, h := range req.Hooks {
		hookUUID, _ := uuid.Parse(backup.ID().String())
		hook := entities.NewBackupHook(hookUUID, h.Name, entities.HookPhase(h.Phase), h.Params)
		hook.Enabled = h.Enabled
		newHooks = append(newHooks, hook)
	}
	backup.SetHooks(newHooks)

	if err := s.repo.Save(ctx, backup); err != nil {
		return nil, err
	}

	return s.assembler.ToBackupResponse(backup, hostResp.Name, hostResp.Hostname), nil
}

func (s *BackupLifecycleService) DeleteBackup(ctx context.Context, id string) error {
	bid, err := valueobjects.NewBackupIDFromString(id)
	if err != nil {
		return err
	}

	return s.repo.Delete(ctx, bid)
}

func (s *BackupLifecycleService) RunBackup(ctx context.Context, id string) (string, error) {
	bid, err := valueobjects.NewBackupIDFromString(id)
	if err != nil {
		return "", err
	}

	backup, err := s.repo.FindByID(ctx, bid)
	if err != nil {
		return "", err
	}

	if err := s.publisher.Publish(ctx, backup); err != nil {
		return "", err
	}

	return backup.ID().String(), nil
}

func (s *BackupLifecycleService) RunHostBackups(ctx context.Context, hostID string) ([]string, error) {
	hid, err := entities.NewHostIDFromString(hostID)
	if err != nil {
		return nil, err
	}

	backups, err := s.repo.FindByHostID(ctx, hid)
	if err != nil {
		return nil, err
	}

	var taskIDs []string
	for _, b := range backups {
		if err := s.publisher.Publish(ctx, b); err != nil {
			return nil, err
		}
		taskIDs = append(taskIDs, b.ID().String())
	}

	return taskIDs, nil
}
