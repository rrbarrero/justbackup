package memory

import (
	"context"
	"sync"

	"github.com/rrbarrero/justbackup/internal/backup/domain/entities"
	"github.com/rrbarrero/justbackup/internal/backup/domain/valueobjects"
)

type BackupErrorRepositoryMemory struct {
	errors []*entities.BackupError
	mu     sync.Mutex
}

func NewBackupErrorRepositoryMemory() *BackupErrorRepositoryMemory {
	return &BackupErrorRepositoryMemory{
		errors: []*entities.BackupError{},
	}
}

func (r *BackupErrorRepositoryMemory) Save(ctx context.Context, backupError *entities.BackupError) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.errors = append(r.errors, backupError)
	return nil
}

func (r *BackupErrorRepositoryMemory) FindByBackupID(ctx context.Context, backupID valueobjects.BackupID) ([]*entities.BackupError, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	var result []*entities.BackupError
	for _, e := range r.errors {
		if e.BackupID.String() == backupID.String() {
			result = append(result, e)
		}
	}
	return result, nil
}

func (r *BackupErrorRepositoryMemory) DeleteByBackupID(ctx context.Context, backupID valueobjects.BackupID) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	filtered := make([]*entities.BackupError, 0)
	for _, e := range r.errors {
		if e.BackupID.String() != backupID.String() {
			filtered = append(filtered, e)
		}
	}
	r.errors = filtered
	return nil
}

// SeedTestErrors adds predefined test errors to the repository.
// This is useful for E2E testing to ensure deterministic test data.
func (r *BackupErrorRepositoryMemory) SeedTestErrors(errors []*entities.BackupError) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.errors = append(r.errors, errors...)
}
