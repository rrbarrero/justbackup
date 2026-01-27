package application

import (
	"context"

	"github.com/rrbarrero/justbackup/internal/backup/application/assembler"
	"github.com/rrbarrero/justbackup/internal/backup/application/dto"
	"github.com/rrbarrero/justbackup/internal/backup/domain/entities"
	"github.com/rrbarrero/justbackup/internal/backup/domain/interfaces"
	"github.com/rrbarrero/justbackup/internal/backup/domain/valueobjects"
)

type BackupQueryService struct {
	repo            interfaces.BackupRepository
	hostService     *HostService
	backupErrorRepo interfaces.BackupErrorRepository
	assembler       *assembler.BackupAssembler
}

func NewBackupQueryService(
	repo interfaces.BackupRepository,
	hostService *HostService,
	backupErrorRepo interfaces.BackupErrorRepository,
	assembler *assembler.BackupAssembler,
) *BackupQueryService {
	return &BackupQueryService{
		repo:            repo,
		hostService:     hostService,
		backupErrorRepo: backupErrorRepo,
		assembler:       assembler,
	}
}

func (s *BackupQueryService) ListBackups(ctx context.Context, hostID string, status string) ([]*dto.BackupResponse, error) {
	backups, err := s.fetchBackups(ctx, hostID)
	if err != nil {
		return nil, err
	}

	filteredBackups := s.filterBackupsByStatus(backups, status)

	return s.enrichAndAssembleBackups(ctx, filteredBackups)
}

func (s *BackupQueryService) fetchBackups(ctx context.Context, hostID string) ([]*entities.Backup, error) {
	if hostID != "" {
		id, err := entities.NewHostIDFromString(hostID)
		if err != nil {
			return nil, err
		}
		return s.repo.FindByHostID(ctx, id)
	}
	return s.repo.FindAll(ctx)
}

func (s *BackupQueryService) filterBackupsByStatus(backups []*entities.Backup, status string) []*entities.Backup {
	if status == "" {
		return backups
	}
	var filtered []*entities.Backup
	for _, b := range backups {
		if b.Status().String() == status {
			filtered = append(filtered, b)
		}
	}
	return filtered
}

func (s *BackupQueryService) enrichAndAssembleBackups(ctx context.Context, backups []*entities.Backup) ([]*dto.BackupResponse, error) {
	hostIDs := make([]entities.HostID, 0)
	seenHosts := make(map[string]bool)

	for _, b := range backups {
		hid := b.HostID().String()
		if !seenHosts[hid] {
			hostIDs = append(hostIDs, b.HostID())
			seenHosts[hid] = true
		}
	}

	hostMap, err := s.hostService.GetHostsByIDs(ctx, hostIDs)
	if err != nil {
		return nil, err
	}

	responses := make([]*dto.BackupResponse, len(backups))
	for i, b := range backups {
		hostName := "Unknown"
		hostAddress := "Unknown"
		if h, ok := hostMap[b.HostID().String()]; ok {
			hostName = h.Name
			hostAddress = h.Hostname
		}
		responses[i] = s.assembler.ToBackupResponse(b, hostName, hostAddress)
	}

	return responses, nil
}

func (s *BackupQueryService) GetBackupByID(ctx context.Context, id string) (*dto.BackupResponse, error) {
	bid, err := valueobjects.NewBackupIDFromString(id)
	if err != nil {
		return nil, err
	}

	backup, err := s.repo.FindByID(ctx, bid)
	if err != nil {
		return nil, err
	}

	hostResp, err := s.hostService.GetHost(ctx, backup.HostID().String())
	if err != nil {
		return nil, err
	}

	return s.assembler.ToBackupResponse(backup, hostResp.Name, hostResp.Hostname), nil
}

func (s *BackupQueryService) GetBackupErrors(ctx context.Context, backupID string) ([]*dto.BackupErrorResponse, error) {
	bid, err := valueobjects.NewBackupIDFromString(backupID)
	if err != nil {
		return nil, err
	}

	errors, err := s.backupErrorRepo.FindByBackupID(ctx, bid)
	if err != nil {
		return nil, err
	}

	return s.assembler.ToBackupErrorResponses(errors), nil
}

func (s *BackupQueryService) DeleteBackupErrors(ctx context.Context, backupID string) error {
	bid, err := valueobjects.NewBackupIDFromString(backupID)
	if err != nil {
		return err
	}

	return s.backupErrorRepo.DeleteByBackupID(ctx, bid)
}
