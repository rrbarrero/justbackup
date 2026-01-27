package interfaces

import (
	"context"

	"github.com/rrbarrero/justbackup/internal/user/domain/entities"
)

type UserRepository interface {
	Save(ctx context.Context, user *entities.User) error
	FindByUsername(ctx context.Context, username string) (*entities.User, error)
	Count(ctx context.Context) (int, error)
}
