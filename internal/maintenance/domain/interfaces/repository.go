package interfaces

import (
	"context"

	"github.com/rrbarrero/justbackup/internal/maintenance/domain/entities"
)

type MaintenanceTaskRepository interface {
	Save(ctx context.Context, task *entities.MaintenanceTask) error
	FindAll(ctx context.Context) ([]*entities.MaintenanceTask, error)
	FindDueTasks(ctx context.Context) ([]*entities.MaintenanceTask, error)
}
