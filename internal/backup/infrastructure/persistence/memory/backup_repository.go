package memory

import (
	"context"
	"sync"
	"time"

	"github.com/rrbarrero/justbackup/internal/backup/domain/entities"
	"github.com/rrbarrero/justbackup/internal/backup/domain/valueobjects"
	shared "github.com/rrbarrero/justbackup/internal/shared/domain"
)

type BackupRepositoryMemory struct {
	mu      sync.RWMutex
	backups map[string]*entities.Backup
}

// NewBackupRepositoryMemoryEmpty creates an empty backup repository for testing
func NewBackupRepositoryMemoryEmpty() *BackupRepositoryMemory {
	return &BackupRepositoryMemory{
		backups: make(map[string]*entities.Backup),
	}
}

func NewBackupRepositoryMemory() *BackupRepositoryMemory {
	backups := map[string]*entities.Backup{}

	mustHostID := func(id string) entities.HostID {
		hid, _ := entities.NewHostIDFromString(id)
		return hid
	}
	mustBackupID := func(id string) valueobjects.BackupID {
		bid, _ := valueobjects.NewBackupIDFromString(id)
		return bid
	}

	// Host 1: a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11
	host1ID := mustHostID("a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11")

	// Backup 1 for Host 1 (Completed)
	b1 := entities.RestoreBackup(
		mustBackupID("10000000-0000-0000-0000-000000000001"),
		host1ID,
		"/var/www/html",
		"s3://backups/html",
		valueobjects.BackupStatusCompleted,
		entities.NewBackupSchedule("0 0 * * *"),
		time.Now().Add(-24*time.Hour),
		time.Now().Add(-23*time.Hour),
		nil,
		nil,
		true,
		false,
		"1.2GB",
		0,
		false,
	)
	backups[b1.ID().String()] = b1

	// Backup 2 for Host 1 (Pending)
	b2 := entities.RestoreBackup(
		mustBackupID("10000000-0000-0000-0000-000000000002"),
		host1ID,
		"/etc/nginx",
		"s3://backups/nginx",
		valueobjects.BackupStatusPending,
		entities.NewBackupSchedule("0 2 * * *"),
		time.Now().Add(-1*time.Hour),
		time.Now().Add(-1*time.Hour),
		nil,
		[]string{"tmp/*"},
		true,
		true,
		"",
		4,
		false,
	)
	backups[b2.ID().String()] = b2

	// Host 2: b0eebc99-9c0b-4ef8-bb6d-6bb9bd380a22
	host2ID := mustHostID("b0eebc99-9c0b-4ef8-bb6d-6bb9bd380a22")

	// Backup 3 for Host 2 (Failed)
	b3 := entities.RestoreBackup(
		mustBackupID("20000000-0000-0000-0000-000000000001"),
		host2ID,
		"/var/lib/mysql",
		"s3://backups/mysql",
		valueobjects.BackupStatusFailed,
		entities.NewBackupSchedule("0 4 * * *"),
		time.Now().Add(-48*time.Hour),
		time.Now().Add(-47*time.Hour),
		nil,
		nil,
		true,
		false,
		"",
		0,
		false,
	)
	backups[b3.ID().String()] = b3

	return &BackupRepositoryMemory{
		backups: backups,
	}
}

func (r *BackupRepositoryMemory) Save(ctx context.Context, backup *entities.Backup) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.backups[backup.ID().String()] = backup
	return nil
}

func (r *BackupRepositoryMemory) FindByID(ctx context.Context, id valueobjects.BackupID) (*entities.Backup, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	backup, ok := r.backups[id.String()]
	if !ok {
		return nil, shared.ErrNotFound
	}
	return backup, nil
}

func (r *BackupRepositoryMemory) FindAll(ctx context.Context) ([]*entities.Backup, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	backups := make([]*entities.Backup, 0, len(r.backups))
	for _, b := range r.backups {
		backups = append(backups, b)
	}
	return backups, nil
}

func (r *BackupRepositoryMemory) FindByHostID(ctx context.Context, hostID entities.HostID) ([]*entities.Backup, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var backups []*entities.Backup
	for _, b := range r.backups {
		if b.HostID() == hostID {
			backups = append(backups, b)
		}
	}
	return backups, nil
}

func (r *BackupRepositoryMemory) FindDueBackups(ctx context.Context) ([]*entities.Backup, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var backups []*entities.Backup
	now := time.Now()
	for _, b := range r.backups {
		if b.Enabled() && b.NextRunAt() != nil && !b.NextRunAt().After(now) {
			backups = append(backups, b)
		}
	}
	return backups, nil
}

func (r *BackupRepositoryMemory) Delete(ctx context.Context, id valueobjects.BackupID) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.backups[id.String()]; !ok {
		return shared.ErrNotFound
	}

	delete(r.backups, id.String())
	return nil
}

func (r *BackupRepositoryMemory) GetStats(ctx context.Context) (*entities.BackupStats, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	stats := &entities.BackupStats{
		Total: len(r.backups),
	}

	var totalSize int64
	for _, b := range r.backups {
		switch b.Status() {
		case valueobjects.BackupStatusPending:
			stats.Pending++
		case valueobjects.BackupStatusCompleted:
			stats.Completed++
			s, _ := shared.ParseSize(b.Size())
			totalSize += s
		case valueobjects.BackupStatusFailed:
			stats.Failed++
		}
	}

	stats.TotalSize = shared.FormatSize(totalSize)

	return stats, nil
}

func (r *BackupRepositoryMemory) GetFailedCountsByHost(ctx context.Context) (map[string]int, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	counts := make(map[string]int)
	for _, b := range r.backups {
		if b.Status() == valueobjects.BackupStatusFailed {
			counts[b.HostID().String()]++
		}
	}
	return counts, nil
}
