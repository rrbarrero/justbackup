package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
	"github.com/rrbarrero/justbackup/internal/backup/domain/entities"
	"github.com/rrbarrero/justbackup/internal/backup/domain/valueobjects"
	shared "github.com/rrbarrero/justbackup/internal/shared/domain"
	"github.com/rrbarrero/justbackup/internal/shared/infrastructure/crypto"
)

type BackupRepositoryPostgres struct {
	db                *sql.DB
	encryptionService crypto.EncryptionService
}

func NewBackupRepositoryPostgres(db *sql.DB, encryptionService crypto.EncryptionService) *BackupRepositoryPostgres {
	return &BackupRepositoryPostgres{
		db:                db,
		encryptionService: encryptionService,
	}
}

func (r *BackupRepositoryPostgres) Save(ctx context.Context, backup *entities.Backup) error {
	query := `
		INSERT INTO backups (id, host_id, path, destination, status, schedule, created_at, updated_at, last_run, next_run_at, excludes, enabled, incremental, size, retention, encrypted)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16)
		ON CONFLICT (id) DO UPDATE SET
			host_id = EXCLUDED.host_id,
			path = EXCLUDED.path,
			destination = EXCLUDED.destination,
			status = EXCLUDED.status,
			schedule = EXCLUDED.schedule,
			updated_at = EXCLUDED.updated_at,
			last_run = EXCLUDED.last_run,
			next_run_at = EXCLUDED.next_run_at,
			excludes = EXCLUDED.excludes,
			enabled = EXCLUDED.enabled,
			incremental = EXCLUDED.incremental,
			size = EXCLUDED.size,
			retention = EXCLUDED.retention,
			encrypted = EXCLUDED.encrypted
	`

	var lastRun *time.Time
	if !backup.Schedule().LastRun.IsZero() {
		t := backup.Schedule().LastRun
		lastRun = &t
	}

	_, err := r.db.ExecContext(ctx, query,
		backup.ID().String(),
		backup.HostID().String(),
		backup.Path(),
		backup.Destination(),
		backup.Status().String(),
		backup.Schedule().CronExpression,
		backup.CreatedAt(),
		backup.UpdatedAt(),
		lastRun,
		backup.NextRunAt(),
		pq.Array(backup.Excludes()),
		backup.Enabled(),
		backup.Incremental(),
		backup.Size(),
		backup.Retention(),
		backup.Encrypted(),
	)
	if err != nil {
		return err
	}

	return r.saveHooks(ctx, backup)
}

func (r *BackupRepositoryPostgres) saveHooks(ctx context.Context, backup *entities.Backup) error {
	const deleteQuery = `DELETE FROM backup_hooks WHERE backup_id = $1`
	if _, err := r.db.ExecContext(ctx, deleteQuery, backup.ID().String()); err != nil {
		return err
	}

	if len(backup.Hooks()) == 0 {
		return nil
	}

	const insertQuery = `
		INSERT INTO backup_hooks (id, backup_id, name, phase, enabled, params, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`

	for _, hook := range backup.Hooks() {
		paramsJSON, err := json.Marshal(hook.Params)
		if err != nil {
			return fmt.Errorf("failed to marshal parameters for hook %s: %w", hook.Name, err)
		}

		encrypted, err := r.encryptionService.Encrypt(paramsJSON)
		if err != nil {
			return err
		}

		encryptedJSON, err := json.Marshal(encrypted)
		if err != nil {
			return fmt.Errorf("failed to marshal encrypted parameters for hook %s: %w", hook.Name, err)
		}

		if _, err := r.db.ExecContext(ctx, insertQuery,
			hook.ID.String(),
			hook.BackupID.String(),
			hook.Name,
			string(hook.Phase),
			hook.Enabled,
			string(encryptedJSON),
			hook.CreatedAt,
			hook.UpdatedAt,
		); err != nil {
			return err
		}
	}
	return nil
}

func (r *BackupRepositoryPostgres) FindByID(ctx context.Context, id valueobjects.BackupID) (*entities.Backup, error) {
	query := `
		SELECT id, host_id, path, destination, status, schedule, created_at, updated_at, last_run, next_run_at, excludes, enabled, incremental, size, retention, encrypted
		FROM backups WHERE id = $1
	`
	backup, err := r.scanBackup(r.db.QueryRowContext(ctx, query, id.String()))
	if err != nil {
		return nil, err
	}
	if err := r.loadHooks(ctx, backup); err != nil {
		return nil, err
	}
	return backup, nil
}

func (r *BackupRepositoryPostgres) loadHooks(ctx context.Context, backup *entities.Backup) error {
	const query = `SELECT id, backup_id, name, phase, enabled, params, created_at, updated_at FROM backup_hooks WHERE backup_id = $1`
	rows, err := r.db.QueryContext(ctx, query, backup.ID().String())
	if err != nil {
		return err
	}
	defer func() { _ = rows.Close() }()

	var hooks []*entities.BackupHook
	for rows.Next() {
		var h entities.BackupHook
		var paramsJSON []byte
		var idStr, bidStr string

		if err := rows.Scan(&idStr, &bidStr, &h.Name, &h.Phase, &h.Enabled, &paramsJSON, &h.CreatedAt, &h.UpdatedAt); err != nil {
			return err
		}

		h.ID, err = uuid.Parse(idStr)
		if err != nil {
			return fmt.Errorf("failed to parse hook ID '%s': %w", idStr, err)
		}

		h.BackupID, err = uuid.Parse(bidStr)
		if err != nil {
			return fmt.Errorf("failed to parse backup ID '%s' for hook %s: %w", bidStr, h.Name, err)
		}

		var encryptedData []byte
		if err := json.Unmarshal(paramsJSON, &encryptedData); err != nil {
			return fmt.Errorf("failed to unmarshal encrypted hook parameters for hook %s: %w", h.Name, err)
		}

		decrypted, err := r.encryptionService.Decrypt(encryptedData)
		if err != nil {
			return fmt.Errorf("failed to decrypt hook parameters for hook %s: %w", h.Name, err)
		}

		if err := json.Unmarshal(decrypted, &h.Params); err != nil {
			return fmt.Errorf("failed to unmarshal decrypted parameters for hook %s: %w", h.Name, err)
		}

		hooks = append(hooks, &h)
	}
	backup.SetHooks(hooks)
	return nil
}

func (r *BackupRepositoryPostgres) FindByHostID(ctx context.Context, hostID entities.HostID) ([]*entities.Backup, error) {
	query := `
		SELECT id, host_id, path, destination, status, schedule, created_at, updated_at, last_run, next_run_at, excludes, enabled, incremental, size, retention, encrypted
		FROM backups WHERE host_id = $1
	`
	rows, err := r.db.QueryContext(ctx, query, hostID.String())
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	backups, err := r.scanBackups(rows)
	if err != nil {
		return nil, err
	}
	for _, b := range backups {
		if err := r.loadHooks(ctx, b); err != nil {
			return nil, err
		}
	}
	return backups, nil
}

func (r *BackupRepositoryPostgres) FindAll(ctx context.Context) ([]*entities.Backup, error) {
	query := `
		SELECT id, host_id, path, destination, status, schedule, created_at, updated_at, last_run, next_run_at, excludes, enabled, incremental, size, retention, encrypted
		FROM backups
	`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	backups, err := r.scanBackups(rows)
	if err != nil {
		return nil, err
	}
	for _, b := range backups {
		if err := r.loadHooks(ctx, b); err != nil {
			return nil, err
		}
	}
	return backups, nil
}

func (r *BackupRepositoryPostgres) FindDueBackups(ctx context.Context) ([]*entities.Backup, error) {
	query := `
		SELECT id, host_id, path, destination, status, schedule, created_at, updated_at, last_run, next_run_at, excludes, enabled, incremental, size, retention, encrypted
		FROM backups
		WHERE enabled = TRUE AND next_run_at <= NOW()
	`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	backups, err := r.scanBackups(rows)
	if err != nil {
		return nil, err
	}
	for _, b := range backups {
		if err := r.loadHooks(ctx, b); err != nil {
			return nil, err
		}
	}
	return backups, nil
}

func (r *BackupRepositoryPostgres) Delete(ctx context.Context, id valueobjects.BackupID) error {
	query := `DELETE FROM backups WHERE id = $1`
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

func (r *BackupRepositoryPostgres) GetStats(ctx context.Context) (*entities.BackupStats, error) {
	query := `
		SELECT
			COUNT(*) AS total,
			COUNT(*) FILTER (WHERE status = 'pending') AS pending,
			COUNT(*) FILTER (WHERE status = 'completed') AS completed,
			COUNT(*) FILTER (WHERE status = 'failed') AS failed
		FROM backups
	`

	stats := &entities.BackupStats{}
	err := r.db.QueryRowContext(ctx, query).Scan(
		&stats.Total,
		&stats.Pending,
		&stats.Completed,
		&stats.Failed,
	)
	if err != nil {
		return nil, err
	}

	sizeQuery := `SELECT size FROM backups WHERE status = 'completed' AND size IS NOT NULL`
	rows, err := r.db.QueryContext(ctx, sizeQuery)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	var totalSize int64
	for rows.Next() {
		var sizeStr string
		if err := rows.Scan(&sizeStr); err != nil {
			continue
		}
		s, _ := shared.ParseSize(sizeStr)
		totalSize += s
	}

	stats.TotalSize = shared.FormatSize(totalSize)

	return stats, nil
}

func (r *BackupRepositoryPostgres) GetFailedCountsByHost(ctx context.Context) (map[string]int, error) {
	query := `
		SELECT host_id, COUNT(*)
		FROM backups
		WHERE status = 'failed'
		GROUP BY host_id
	`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	counts := make(map[string]int)
	for rows.Next() {
		var hostID string
		var count int
		if err := rows.Scan(&hostID, &count); err != nil {
			return nil, err
		}
		counts[hostID] = count
	}
	return counts, nil
}

func (r *BackupRepositoryPostgres) scanBackup(row *sql.Row) (*entities.Backup, error) {
	var idStr, hostIDStr, path, destination, statusStr, scheduleCron string
	var createdAt, updatedAt time.Time
	var lastRun, nextRunAt *time.Time
	var excludes []string
	var enabled, incremental, encrypted bool
	var size sql.NullString
	var retention sql.NullInt64

	err := row.Scan(&idStr, &hostIDStr, &path, &destination, &statusStr, &scheduleCron, &createdAt, &updatedAt, &lastRun, &nextRunAt, pq.Array(&excludes), &enabled, &incremental, &size, &retention, &encrypted)
	if err == sql.ErrNoRows {
		return nil, shared.ErrNotFound
	}
	if err != nil {
		return nil, err
	}

	return r.mapToEntity(idStr, hostIDStr, path, destination, statusStr, scheduleCron, createdAt, updatedAt, lastRun, nextRunAt, excludes, enabled, incremental, size.String, int(retention.Int64), encrypted)
}

func (r *BackupRepositoryPostgres) scanBackups(rows *sql.Rows) ([]*entities.Backup, error) {
	var backups []*entities.Backup
	for rows.Next() {
		var idStr, hostIDStr, path, destination, statusStr, scheduleCron string
		var createdAt, updatedAt time.Time
		var lastRun, nextRunAt *time.Time
		var excludes []string
		var enabled, incremental, encrypted bool
		var size sql.NullString
		var retention sql.NullInt64

		if err := rows.Scan(&idStr, &hostIDStr, &path, &destination, &statusStr, &scheduleCron, &createdAt, &updatedAt, &lastRun, &nextRunAt, pq.Array(&excludes), &enabled, &incremental, &size, &retention, &encrypted); err != nil {
			return nil, err
		}

		backup, err := r.mapToEntity(idStr, hostIDStr, path, destination, statusStr, scheduleCron, createdAt, updatedAt, lastRun, nextRunAt, excludes, enabled, incremental, size.String, int(retention.Int64), encrypted)
		if err != nil {
			return nil, err
		}
		backups = append(backups, backup)
	}
	return backups, nil
}

func (r *BackupRepositoryPostgres) mapToEntity(idStr, hostIDStr, path, destination, statusStr, scheduleCron string, createdAt, updatedAt time.Time, lastRun, nextRunAt *time.Time, excludes []string, enabled, incremental bool, size string, retention int, encrypted bool) (*entities.Backup, error) {
	bid, err := valueobjects.NewBackupIDFromString(idStr)
	if err != nil {
		return nil, err
	}
	hid, err := entities.NewHostIDFromString(hostIDStr)
	if err != nil {
		return nil, err
	}

	schedule := entities.NewBackupSchedule(scheduleCron)
	if lastRun != nil {
		schedule.LastRun = *lastRun
	}

	return entities.RestoreBackup(
		bid,
		hid,
		path,
		destination,
		valueobjects.BackupStatus(statusStr),
		schedule,
		createdAt,
		updatedAt,
		nextRunAt,
		excludes,
		enabled,
		incremental,
		size,
		retention,
		encrypted,
	), nil
}
