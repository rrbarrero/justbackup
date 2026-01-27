package memory

import (
	"context"
	"sync"
	"time"

	"github.com/rrbarrero/justbackup/internal/backup/domain/entities"
	shared "github.com/rrbarrero/justbackup/internal/shared/domain"
)

type HostRepositoryMemory struct {
	mu    sync.RWMutex
	hosts map[string]*entities.Host
}

func NewHostRepositoryMemory() *HostRepositoryMemory {
	hosts := map[string]*entities.Host{}

	// Helper to create ID ignoring error for example data
	mustID := func(id string) entities.HostID {
		hid, _ := entities.NewHostIDFromString(id)
		return hid
	}

	// Add 5 example hosts with deterministic IDs
	// Add 5 example hosts with deterministic IDs
	host1 := entities.RestoreHost(mustID("a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11"), "Localhost Dev", "localhost", "dev", 22, "localhost-dev", false, time.Now())
	host2 := entities.RestoreHost(mustID("b0eebc99-9c0b-4ef8-bb6d-6bb9bd380a22"), "Staging Server", "staging.example.com", "ubuntu", 2222, "staging-server", false, time.Now())
	host3 := entities.RestoreHost(mustID("c0eebc99-9c0b-4ef8-bb6d-6bb9bd380a33"), "Production Web", "web.prod.example.com", "root", 22, "production-web", false, time.Now())
	host4 := entities.RestoreHost(mustID("d0eebc99-9c0b-4ef8-bb6d-6bb9bd380a44"), "Backup Storage", "backup.example.com", "backupuser", 22, "backup-storage", false, time.Now())
	host5 := entities.RestoreHost(mustID("e0eebc99-9c0b-4ef8-bb6d-6bb9bd380a55"), "Another Dev Machine", "devmachine.local", "user", 22, "another-dev-machine", false, time.Now())

	hosts[host1.ID().String()] = host1
	hosts[host2.ID().String()] = host2
	hosts[host3.ID().String()] = host3
	hosts[host4.ID().String()] = host4
	hosts[host5.ID().String()] = host5

	return &HostRepositoryMemory{
		hosts: hosts,
	}
}

func (r *HostRepositoryMemory) Save(ctx context.Context, host *entities.Host) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.hosts[host.ID().String()] = host
	return nil
}

func (r *HostRepositoryMemory) Get(ctx context.Context, id entities.HostID) (*entities.Host, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if host, exists := r.hosts[id.String()]; exists {
		return host, nil
	}
	return nil, shared.ErrNotFound
}

func (r *HostRepositoryMemory) GetByIDs(ctx context.Context, ids []entities.HostID) ([]*entities.Host, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var hosts []*entities.Host
	for _, id := range ids {
		if host, exists := r.hosts[id.String()]; exists {
			hosts = append(hosts, host)
		}
	}
	return hosts, nil
}

func (r *HostRepositoryMemory) List(ctx context.Context) ([]*entities.Host, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	hosts := make([]*entities.Host, 0, len(r.hosts))
	for _, host := range r.hosts {
		hosts = append(hosts, host)
	}
	return hosts, nil
}

func (r *HostRepositoryMemory) Count(ctx context.Context) (int64, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return int64(len(r.hosts)), nil
}

func (r *HostRepositoryMemory) Update(ctx context.Context, host *entities.Host) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.hosts[host.ID().String()]; !exists {
		return shared.ErrNotFound
	}

	r.hosts[host.ID().String()] = host
	return nil
}

func (r *HostRepositoryMemory) Delete(ctx context.Context, id entities.HostID) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.hosts[id.String()]; !exists {
		return shared.ErrNotFound
	}

	delete(r.hosts, id.String())
	return nil
}
