package interfaces

import (
	"context"

	"github.com/rrbarrero/justbackup/internal/backup/domain/entities"
)

type HostRepository interface {
	Save(ctx context.Context, host *entities.Host) error
	Get(ctx context.Context, id entities.HostID) (*entities.Host, error)
	GetByIDs(ctx context.Context, ids []entities.HostID) ([]*entities.Host, error)
	List(ctx context.Context) ([]*entities.Host, error)
	Update(ctx context.Context, host *entities.Host) error
	Delete(ctx context.Context, id entities.HostID) error
	Count(ctx context.Context) (int64, error)
}
