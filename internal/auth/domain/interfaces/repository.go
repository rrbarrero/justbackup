package interfaces

import (
	"context"

	"github.com/rrbarrero/justbackup/internal/auth/domain/entities"
)

type AuthTokenRepository interface {
	Save(ctx context.Context, token *entities.AuthToken) error
	Get(ctx context.Context) (*entities.AuthToken, error)
	Delete(ctx context.Context) error
}
