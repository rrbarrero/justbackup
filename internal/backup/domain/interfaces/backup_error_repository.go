package interfaces

import (
	"context"

	"github.com/rrbarrero/justbackup/internal/backup/domain/entities"
	"github.com/rrbarrero/justbackup/internal/backup/domain/valueobjects"
)

type BackupErrorRepository interface {
	Save(ctx context.Context, backupError *entities.BackupError) error
	FindByBackupID(ctx context.Context, backupID valueobjects.BackupID) ([]*entities.BackupError, error)
	DeleteByBackupID(ctx context.Context, backupID valueobjects.BackupID) error
}
