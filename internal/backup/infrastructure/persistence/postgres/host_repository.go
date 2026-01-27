package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/lib/pq"
	"github.com/rrbarrero/justbackup/internal/backup/domain/entities"
	shared "github.com/rrbarrero/justbackup/internal/shared/domain"
)

type HostRepositoryPostgres struct {
	db *sql.DB
}

func NewHostRepositoryPostgres(db *sql.DB) *HostRepositoryPostgres {
	return &HostRepositoryPostgres{
		db: db,
	}
}

func (r *HostRepositoryPostgres) Save(ctx context.Context, host *entities.Host) error {
	query := `
		INSERT INTO hosts (id, name, hostname, "user", port, host_path, is_workstation, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		ON CONFLICT (id) DO UPDATE SET
			name = EXCLUDED.name,
			hostname = EXCLUDED.hostname,
			"user" = EXCLUDED."user",
			port = EXCLUDED.port,
			host_path = EXCLUDED.host_path,
			is_workstation = EXCLUDED.is_workstation
	`

	_, err := r.db.ExecContext(ctx, query,
		host.ID().String(),
		host.Name(),
		host.Hostname(),
		host.User(),
		host.Port(),
		host.Path(),
		host.IsWorkstation(),
		host.CreatedAt(),
	)
	return err
}

func (r *HostRepositoryPostgres) Get(ctx context.Context, id entities.HostID) (*entities.Host, error) {
	query := `SELECT id, name, hostname, "user", port, host_path, is_workstation, created_at FROM hosts WHERE id = $1`

	var hostIDStr string
	var name, hostname, user, path string
	var port int
	var isWorkstation bool
	var createdAt time.Time

	err := r.db.QueryRowContext(ctx, query, id.String()).Scan(
		&hostIDStr, &name, &hostname, &user, &port, &path, &isWorkstation, &createdAt,
	)

	if err == sql.ErrNoRows {
		return nil, shared.ErrNotFound
	}
	if err != nil {
		return nil, err
	}

	// Reconstruct entity
	hid, err := entities.NewHostIDFromString(hostIDStr)
	if err != nil {
		return nil, fmt.Errorf("database corruption: invalid host id %s: %w", hostIDStr, err)
	}
	return entities.RestoreHost(hid, name, hostname, user, port, path, isWorkstation, createdAt), nil
}

func (r *HostRepositoryPostgres) GetByIDs(ctx context.Context, ids []entities.HostID) ([]*entities.Host, error) {
	if len(ids) == 0 {
		return []*entities.Host{}, nil
	}

	query := `SELECT id, name, hostname, "user", port, host_path, is_workstation, created_at FROM hosts WHERE id = ANY($1)`

	// Convert IDs to string slice for postgres array
	idStrings := make([]string, len(ids))
	for i, id := range ids {
		idStrings[i] = id.String()
	}

	rows, err := r.db.QueryContext(ctx, query, pq.Array(idStrings))
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	var hosts []*entities.Host
	for rows.Next() {
		var hostIDStr string
		var name, hostname, user, path string
		var port int
		var isWorkstation bool
		var createdAt time.Time

		if err := rows.Scan(&hostIDStr, &name, &hostname, &user, &port, &path, &isWorkstation, &createdAt); err != nil {
			return nil, err
		}

		hid, err := entities.NewHostIDFromString(hostIDStr)
		if err != nil {
			return nil, fmt.Errorf("database corruption: invalid host id %s: %w", hostIDStr, err)
		}
		hosts = append(hosts, entities.RestoreHost(hid, name, hostname, user, port, path, isWorkstation, createdAt))
	}

	return hosts, nil
}

func (r *HostRepositoryPostgres) List(ctx context.Context) ([]*entities.Host, error) {
	query := `SELECT id, name, hostname, "user", port, host_path, is_workstation, created_at FROM hosts`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	var hosts []*entities.Host
	for rows.Next() {
		var hostIDStr string
		var name, hostname, user, path string
		var port int
		var isWorkstation bool
		var createdAt time.Time

		if err := rows.Scan(&hostIDStr, &name, &hostname, &user, &port, &path, &isWorkstation, &createdAt); err != nil {
			return nil, err
		}

		hid, err := entities.NewHostIDFromString(hostIDStr)
		if err != nil {
			return nil, fmt.Errorf("database corruption: invalid host id %s: %w", hostIDStr, err)
		}
		hosts = append(hosts, entities.RestoreHost(hid, name, hostname, user, port, path, isWorkstation, createdAt))
	}

	return hosts, nil
}

func (r *HostRepositoryPostgres) Update(ctx context.Context, host *entities.Host) error {
	query := `
		UPDATE hosts
		SET name = $2, hostname = $3, "user" = $4, port = $5, host_path = $6, is_workstation = $7
		WHERE id = $1
	`

	result, err := r.db.ExecContext(ctx, query,
		host.ID().String(),
		host.Name(),
		host.Hostname(),
		host.User(),
		host.Port(),
		host.Path(),
		host.IsWorkstation(),
	)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return shared.ErrNotFound
	}

	return nil
}

func (r *HostRepositoryPostgres) Delete(ctx context.Context, id entities.HostID) error {
	query := `DELETE FROM hosts WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query, id.String())
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return shared.ErrNotFound
	}

	return nil
}

func (r *HostRepositoryPostgres) Count(ctx context.Context) (int64, error) {
	query := `SELECT COUNT(*) FROM hosts`
	var count int64
	err := r.db.QueryRowContext(ctx, query).Scan(&count)
	if err != nil {
		return 0, err
	}
	return count, nil
}
