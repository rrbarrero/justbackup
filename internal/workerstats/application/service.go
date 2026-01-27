package application

import (
	"context"
	"encoding/json"
	"time"

	"github.com/rrbarrero/justbackup/internal/shared/infrastructure/websocket"
	"github.com/rrbarrero/justbackup/internal/workerstats/domain/entities"
	"github.com/rrbarrero/justbackup/internal/workerstats/domain/interfaces"
)

// WorkerStatsReportDTO represents the incoming data from a worker
type WorkerStatsReportDTO struct {
	WorkerID      string    `json:"worker_id"`
	CPUUsage      float64   `json:"cpu_usage"`
	MemoryTotal   uint64    `json:"memory_total"`
	MemoryUsed    uint64    `json:"memory_used"`
	MemoryPercent float64   `json:"memory_percent"`
	DiskTotal     uint64    `json:"disk_total"`
	DiskUsed      uint64    `json:"disk_used"`
	DiskPercent   float64   `json:"disk_percent"`
	Timestamp     time.Time `json:"timestamp"`
}

// WorkerStatsService orchestrates worker stats reporting
type WorkerStatsService struct {
	repo interfaces.WorkerStatsRepository
	hub  *websocket.Hub
}

// NewWorkerStatsService creates a new stats service
func NewWorkerStatsService(repo interfaces.WorkerStatsRepository, hub *websocket.Hub) *WorkerStatsService {
	return &WorkerStatsService{
		repo: repo,
		hub:  hub,
	}
}

// PostStats handles a new report from a worker
func (s *WorkerStatsService) PostStats(ctx context.Context, dto WorkerStatsReportDTO) error {
	report := entities.WorkerStatsReport{
		CPUUsage:      dto.CPUUsage,
		MemoryTotal:   dto.MemoryTotal,
		MemoryUsed:    dto.MemoryUsed,
		MemoryPercent: dto.MemoryPercent,
		DiskTotal:     dto.DiskTotal,
		DiskUsed:      dto.DiskUsed,
		DiskPercent:   dto.DiskPercent,
		Timestamp:     dto.Timestamp,
	}

	if report.Timestamp.IsZero() {
		report.Timestamp = time.Now()
	}

	if err := s.repo.SaveReport(ctx, dto.WorkerID, report); err != nil {
		return err
	}

	// Notify frontend
	msg := map[string]string{"type": "worker_stats_updated"}
	if data, err := json.Marshal(msg); err == nil {
		s.hub.Broadcast(data)
	}

	return nil
}

// GetStats retrieves stats for all workers
func (s *WorkerStatsService) GetStats(ctx context.Context) ([]*entities.WorkerStatsWindow, error) {
	return s.repo.GetAllStats(ctx)
}
