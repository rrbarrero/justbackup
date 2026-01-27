package interfaces

import (
	"context"

	"github.com/rrbarrero/justbackup/internal/workerstats/domain/entities"
)

// WorkerStatsRepository defines the interface for storing and retrieving worker stats
type WorkerStatsRepository interface {
	SaveReport(ctx context.Context, workerID string, report entities.WorkerStatsReport) error
	GetStats(ctx context.Context, workerID string) (*entities.WorkerStatsWindow, error)
	GetAllStats(ctx context.Context) ([]*entities.WorkerStatsWindow, error)
}
