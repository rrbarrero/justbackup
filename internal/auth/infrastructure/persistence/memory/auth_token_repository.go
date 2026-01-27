package memory

import (
	"context"
	"sync"

	"github.com/rrbarrero/justbackup/internal/auth/domain/entities"
	shared "github.com/rrbarrero/justbackup/internal/shared/domain"
)

type AuthTokenRepositoryMemory struct {
	mu    sync.Mutex
	token *entities.AuthToken
}

func NewAuthTokenRepositoryMemory() *AuthTokenRepositoryMemory {
	return &AuthTokenRepositoryMemory{}
}

func (r *AuthTokenRepositoryMemory) Save(ctx context.Context, token *entities.AuthToken) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.token = token
	return nil
}

func (r *AuthTokenRepositoryMemory) Get(ctx context.Context) (*entities.AuthToken, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.token == nil {
		return nil, shared.ErrNotFound
	}
	return r.token, nil
}

func (r *AuthTokenRepositoryMemory) Delete(ctx context.Context) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.token = nil
	return nil
}
