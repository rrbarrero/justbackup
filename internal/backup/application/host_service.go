package application

import (
	"context"

	"github.com/rrbarrero/justbackup/internal/backup/application/dto"
	"github.com/rrbarrero/justbackup/internal/backup/domain/entities"
	"github.com/rrbarrero/justbackup/internal/backup/domain/interfaces"
)

type HostService struct {
	repo       interfaces.HostRepository
	backupRepo interfaces.BackupRepository
}

func NewHostService(repo interfaces.HostRepository, backupRepo interfaces.BackupRepository) *HostService {
	return &HostService{
		repo:       repo,
		backupRepo: backupRepo,
	}
}

func (s *HostService) CreateHost(ctx context.Context, req dto.CreateHostRequest) (*dto.HostResponse, error) {
	host := entities.NewHost(req.Name, req.Hostname, req.User, req.Port, req.Path, req.IsWorkstation)

	if err := s.repo.Save(ctx, host); err != nil {
		return nil, err
	}

	return &dto.HostResponse{
		ID:            host.ID().String(),
		Name:          host.Name(),
		Hostname:      host.Hostname(),
		User:          host.User(),
		Port:          host.Port(),
		Path:          host.Path(),
		IsWorkstation: host.IsWorkstation(),
	}, nil
}

func (s *HostService) ListHosts(ctx context.Context) ([]*dto.HostResponse, error) {
	hosts, err := s.repo.List(ctx)
	if err != nil {
		return nil, err
	}

	failedCounts, err := s.backupRepo.GetFailedCountsByHost(ctx)
	if err != nil {
		// If calculation fails, proceed with an empty map.
		failedCounts = make(map[string]int)
	}

	var responses []*dto.HostResponse
	for _, host := range hosts {
		resp := dto.ToHostResponse(host)
		resp.FailedBackupsCount = failedCounts[host.ID().String()]
		responses = append(responses, resp)
	}

	return responses, nil
}

func (s *HostService) GetHostsByIDs(ctx context.Context, ids []entities.HostID) (map[string]*dto.HostResponse, error) {
	hosts, err := s.repo.GetByIDs(ctx, ids)
	if err != nil {
		return nil, err
	}

	hostMap := make(map[string]*dto.HostResponse)
	for _, host := range hosts {
		hostMap[host.ID().String()] = dto.ToHostResponse(host)
	}

	return hostMap, nil
}

func (s *HostService) GetHost(ctx context.Context, id string) (*dto.HostResponse, error) {
	hostID, err := entities.NewHostIDFromString(id)
	if err != nil {
		return nil, err
	}

	host, err := s.repo.Get(ctx, hostID)
	if err != nil {
		return nil, err
	}

	resp := dto.ToHostResponse(host)
	// Optionally fetch failed count for single host
	backups, err := s.backupRepo.FindByHostID(ctx, hostID)
	if err == nil {
		count := 0
		for _, b := range backups {
			if b.Status() == "failed" {
				count++
			}
		}
		resp.FailedBackupsCount = count
	}

	return resp, nil
}

func (s *HostService) GetHostEntity(ctx context.Context, id string) (*entities.Host, error) {
	hostID, err := entities.NewHostIDFromString(id)
	if err != nil {
		return nil, err
	}
	return s.repo.Get(ctx, hostID)
}

func (s *HostService) UpdateHost(ctx context.Context, req dto.UpdateHostRequest) (*dto.HostResponse, error) {
	hostID, err := entities.NewHostIDFromString(req.ID)
	if err != nil {
		return nil, err
	}

	host, err := s.repo.Get(ctx, hostID)
	if err != nil {
		return nil, err
	}

	host.Update(req.Name, req.Hostname, req.User, req.Port, req.Path, req.IsWorkstation)

	if err := s.repo.Update(ctx, host); err != nil {
		return nil, err
	}

	return dto.ToHostResponse(host), nil
}

func (s *HostService) DeleteHost(ctx context.Context, id string) error {
	hostID, err := entities.NewHostIDFromString(id)
	if err != nil {
		return err
	}

	return s.repo.Delete(ctx, hostID)
}
