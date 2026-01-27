package interfaces

import (
	"context"

	"github.com/rrbarrero/justbackup/internal/backup/domain/entities"
	"github.com/rrbarrero/justbackup/internal/backup/domain/valueobjects"
)

type BackupRepository interface {
	Save(ctx context.Context, backup *entities.Backup) error
	FindByID(ctx context.Context, id valueobjects.BackupID) (*entities.Backup, error)
	FindByHostID(ctx context.Context, hostID entities.HostID) ([]*entities.Backup, error)
	FindAll(ctx context.Context) ([]*entities.Backup, error)
	FindDueBackups(ctx context.Context) ([]*entities.Backup, error)
	Delete(ctx context.Context, id valueobjects.BackupID) error
	GetStats(ctx context.Context) (*entities.BackupStats, error)
	GetFailedCountsByHost(ctx context.Context) (map[string]int, error)
}

type TaskPublisher interface {
	PublishMeasureTask(ctx context.Context, hostID string, path string) (string, error)
	PublishSearchTask(ctx context.Context, pattern string) (string, error)
	Publish(ctx context.Context, backup *entities.Backup) error
	PublishRestoreTask(ctx context.Context, backup *entities.Backup, path string, restoreAddr string, restoreToken string) (string, error)
	PublishListFilesTask(ctx context.Context, path string) (string, error)
	PublishRemoteRestoreTask(ctx context.Context, backup *entities.Backup, path string, targetHost *entities.Host, targetPath string) (string, error)
}

type ResultStore interface {
	GetTaskResult(ctx context.Context, taskID string) (string, error)
}
