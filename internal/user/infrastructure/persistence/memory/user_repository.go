package memory

import (
	"context"
	"sync"

	"github.com/rrbarrero/justbackup/internal/user/domain"
	"github.com/rrbarrero/justbackup/internal/user/domain/entities"
)

type UserRepositoryMemory struct {
	users  map[string]*entities.User
	nextID int
	mu     sync.RWMutex
}

func NewUserRepositoryMemory() *UserRepositoryMemory {
	return &UserRepositoryMemory{
		users:  make(map[string]*entities.User),
		nextID: 1,
	}
}

func (r *UserRepositoryMemory) Save(ctx context.Context, user *entities.User) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if user.ID == 0 {
		user.ID = r.nextID
		r.nextID++
	}

	r.users[user.Username] = user
	return nil
}

func (r *UserRepositoryMemory) FindByUsername(ctx context.Context, username string) (*entities.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	user, exists := r.users[username]
	if !exists {
		return nil, domain.ErrUserNotFound
	}
	return user, nil
}

func (r *UserRepositoryMemory) Count(ctx context.Context) (int, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.users), nil
}
