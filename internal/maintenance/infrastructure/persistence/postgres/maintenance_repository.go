package postgres

import (
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"
	"github.com/rrbarrero/justbackup/internal/maintenance/domain/entities"
)

type MaintenanceRepositoryPostgres struct {
	db *sql.DB
}

func NewMaintenanceRepositoryPostgres(db *sql.DB) *MaintenanceRepositoryPostgres {
	return &MaintenanceRepositoryPostgres{
		db: db,
	}
}

func (r *MaintenanceRepositoryPostgres) Save(ctx context.Context, task *entities.MaintenanceTask) error {
	query := `
		INSERT INTO maintenance_tasks (id, name, type, schedule, next_run_at, last_run_at, enabled, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		ON CONFLICT (id) DO UPDATE SET
			name = EXCLUDED.name,
			type = EXCLUDED.type,
			schedule = EXCLUDED.schedule,
			next_run_at = EXCLUDED.next_run_at,
			last_run_at = EXCLUDED.last_run_at,
			enabled = EXCLUDED.enabled,
			updated_at = EXCLUDED.updated_at
	`

	_, err := r.db.ExecContext(ctx, query,
		task.ID(),
		task.Name(),
		task.Type(),
		task.Schedule(),
		task.NextRunAt(),
		task.LastRunAt(),
		task.Enabled(),
		task.CreatedAt(),
		task.UpdatedAt(),
	)

	return err
}

func (r *MaintenanceRepositoryPostgres) FindAll(ctx context.Context) ([]*entities.MaintenanceTask, error) {
	query := `SELECT id, name, type, schedule, next_run_at, last_run_at, enabled, created_at, updated_at FROM maintenance_tasks`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	var tasks []*entities.MaintenanceTask
	for rows.Next() {
		task, err := r.scanTask(rows)
		if err != nil {
			return nil, err
		}
		tasks = append(tasks, task)
	}

	return tasks, nil
}

func (r *MaintenanceRepositoryPostgres) FindDueTasks(ctx context.Context) ([]*entities.MaintenanceTask, error) {
	query := `
		SELECT id, name, type, schedule, next_run_at, last_run_at, enabled, created_at, updated_at 
		FROM maintenance_tasks 
		WHERE enabled = TRUE AND (next_run_at IS NULL OR next_run_at <= $1)
	`
	rows, err := r.db.QueryContext(ctx, query, time.Now())
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	var tasks []*entities.MaintenanceTask
	for rows.Next() {
		task, err := r.scanTask(rows)
		if err != nil {
			return nil, err
		}
		tasks = append(tasks, task)
	}

	return tasks, nil
}

func (r *MaintenanceRepositoryPostgres) scanTask(rows *sql.Rows) (*entities.MaintenanceTask, error) {
	var id uuid.UUID
	var name string
	var taskType string
	var schedule string
	var nextRunAt *time.Time
	var lastRunAt *time.Time
	var enabled bool
	var createdAt time.Time
	var updatedAt time.Time

	if err := rows.Scan(&id, &name, &taskType, &schedule, &nextRunAt, &lastRunAt, &enabled, &createdAt, &updatedAt); err != nil {
		return nil, err
	}

	return entities.RestoreMaintenanceTask(
		id,
		name,
		entities.MaintenanceTaskType(taskType),
		schedule,
		nextRunAt,
		lastRunAt,
		enabled,
		createdAt,
		updatedAt,
	), nil
}
