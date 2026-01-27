package memory

import (
	"context"
	"sort"
	"sync"
	"time"

	"github.com/rrbarrero/justbackup/internal/workerstats/domain/entities"
)

// WorkerStatsRepositoryMemory implements WorkerStatsRepository in memory
type WorkerStatsRepositoryMemory struct {
	mu    sync.RWMutex
	stats map[string]*entities.WorkerStatsWindow
}

// NewWorkerStatsRepositoryMemory creates a new in-memory repository with a cleanup routine
func NewWorkerStatsRepositoryMemory() *WorkerStatsRepositoryMemory {
	repo := &WorkerStatsRepositoryMemory{
		stats: make(map[string]*entities.WorkerStatsWindow),
	}

	// Periodically purge stale workers (inactive for > 2 minutes)
	go func() {
		ticker := time.NewTicker(1 * time.Minute)
		for range ticker.C {
			repo.mu.Lock()
			now := time.Now()
			for id, window := range repo.stats {
				if len(window.Reports) == 0 {
					delete(repo.stats, id)
					continue
				}

				lastReport := window.Reports[len(window.Reports)-1]
				if now.Sub(lastReport.Timestamp) > 2*time.Minute {
					delete(repo.stats, id)
				}
			}
			repo.mu.Unlock()
		}
	}()

	return repo
}

// SaveReport stores a new report for a worker
func (r *WorkerStatsRepositoryMemory) SaveReport(ctx context.Context, workerID string, report entities.WorkerStatsReport) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	window, ok := r.stats[workerID]
	if !ok {
		// Limit total number of workers to prevent memory leaks (safety measure)
		if len(r.stats) >= 1000 {
			return nil
		}
		window = entities.NewWorkerStatsWindow(workerID)
		r.stats[workerID] = window
	}

	window.AddReport(report)
	return nil
}

// GetStats retrieves stats for a specific worker
func (r *WorkerStatsRepositoryMemory) GetStats(ctx context.Context, workerID string) (*entities.WorkerStatsWindow, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	window, ok := r.stats[workerID]
	if !ok {
		return nil, nil
	}

	return window, nil
}

// GetAllStats retrieves stats for all workers
func (r *WorkerStatsRepositoryMemory) GetAllStats(ctx context.Context) ([]*entities.WorkerStatsWindow, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	all := make([]*entities.WorkerStatsWindow, 0, len(r.stats))
	for _, window := range r.stats {
		all = append(all, window)
	}

	sort.Slice(all, func(i, j int) bool {
		return all[i].WorkerID < all[j].WorkerID
	})

	return all, nil
}
