package application

import (
	"context"

	"github.com/rrbarrero/justbackup/internal/backup/domain/entities"
	"github.com/rrbarrero/justbackup/internal/backup/domain/interfaces"
	workerStatsApp "github.com/rrbarrero/justbackup/internal/workerstats/application"
)

type DashboardService struct {
	backupRepo         interfaces.BackupRepository
	hostRepo           interfaces.HostRepository
	workerStatsService *workerStatsApp.WorkerStatsService
}

func NewDashboardService(backupRepo interfaces.BackupRepository, hostRepo interfaces.HostRepository, workerStatsService *workerStatsApp.WorkerStatsService) *DashboardService {
	return &DashboardService{
		backupRepo:         backupRepo,
		hostRepo:           hostRepo,
		workerStatsService: workerStatsService,
	}
}

type DashboardStats struct {
	TotalHosts    int64                 `json:"total_hosts"`
	TotalBackups  int                   `json:"total_backups"`
	ActiveWorkers int                   `json:"active_workers"`
	BackupStats   *entities.BackupStats `json:"backup_stats"`
}

func (s *DashboardService) GetStats(ctx context.Context) (*DashboardStats, error) {
	hostCount, err := s.hostRepo.Count(ctx)
	if err != nil {
		return nil, err
	}

	backupStats, err := s.backupRepo.GetStats(ctx)
	if err != nil {
		return nil, err
	}

	workerStats, err := s.workerStatsService.GetStats(ctx)
	activeWorkers := 0
	if err == nil {
		for _, w := range workerStats {
			if w.WorkerID != "backend" {
				activeWorkers++
			}
		}
	}

	return &DashboardStats{
		TotalHosts:    hostCount,
		TotalBackups:  backupStats.Total,
		ActiveWorkers: activeWorkers,
		BackupStats:   backupStats,
	}, nil
}
