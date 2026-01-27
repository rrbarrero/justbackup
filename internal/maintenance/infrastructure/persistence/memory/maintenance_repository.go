package memory

import (
	"context"
	"sync"
	"time"

	"github.com/rrbarrero/justbackup/internal/maintenance/domain/entities"
)

type MaintenanceRepositoryMemory struct {
	mu    sync.RWMutex
	tasks map[string]*entities.MaintenanceTask
}

func NewMaintenanceRepositoryMemory() *MaintenanceRepositoryMemory {
	return &MaintenanceRepositoryMemory{
		tasks: make(map[string]*entities.MaintenanceTask),
	}
}

func (r *MaintenanceRepositoryMemory) Save(ctx context.Context, task *entities.MaintenanceTask) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.tasks[task.ID().String()] = task
	return nil
}

func (r *MaintenanceRepositoryMemory) FindAll(ctx context.Context) ([]*entities.MaintenanceTask, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var tasks []*entities.MaintenanceTask
	for _, t := range r.tasks {
		tasks = append(tasks, t)
	}
	return tasks, nil
}

func (r *MaintenanceRepositoryMemory) FindDueTasks(ctx context.Context) ([]*entities.MaintenanceTask, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var tasks []*entities.MaintenanceTask
	now := time.Now()
	for _, t := range r.tasks {
		if t.Enabled() && t.NextRunAt() != nil && !t.NextRunAt().After(now) {
			tasks = append(tasks, t)
		}
	}
	return tasks, nil
}
