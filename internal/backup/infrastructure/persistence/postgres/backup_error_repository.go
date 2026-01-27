package postgres

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/rrbarrero/justbackup/internal/backup/domain/entities"
	"github.com/rrbarrero/justbackup/internal/backup/domain/valueobjects"
)

type BackupErrorRepositoryPostgres struct {
	db *sql.DB
}

func NewBackupErrorRepositoryPostgres(db *sql.DB) *BackupErrorRepositoryPostgres {
	return &BackupErrorRepositoryPostgres{db: db}
}

func (r *BackupErrorRepositoryPostgres) Save(ctx context.Context, backupError *entities.BackupError) error {
	query := `INSERT INTO backup_errors (job_id, backup_id, occurred_at, error_message) VALUES ($1, $2, $3, $4)`
	_, err := r.db.ExecContext(ctx, query, backupError.JobID, backupError.BackupID.String(), backupError.OccurredAt, backupError.ErrorMessage)
	if err != nil {
		return fmt.Errorf("failed to save backup error: %w", err)
	}
	return nil
}

func (r *BackupErrorRepositoryPostgres) FindByBackupID(ctx context.Context, backupID valueobjects.BackupID) ([]*entities.BackupError, error) {
	query := `SELECT id, job_id, backup_id, occurred_at, error_message FROM backup_errors WHERE backup_id = $1 ORDER BY occurred_at DESC`
	rows, err := r.db.QueryContext(ctx, query, backupID.String())
	if err != nil {
		return nil, fmt.Errorf("failed to query backup errors: %w", err)
	}
	defer rows.Close()

	errors := make([]*entities.BackupError, 0)
	for rows.Next() {
		var e entities.BackupError
		var backupIDStr string
		if err := rows.Scan(&e.ID, &e.JobID, &backupIDStr, &e.OccurredAt, &e.ErrorMessage); err != nil {
			return nil, fmt.Errorf("failed to scan backup error: %w", err)
		}
		bid, err := valueobjects.NewBackupIDFromString(backupIDStr)
		if err != nil {
			return nil, fmt.Errorf("failed to parse backup ID: %w", err)
		}
		e.BackupID = bid
		errors = append(errors, &e)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating backup errors: %w", err)
	}

	return errors, nil
}

func (r *BackupErrorRepositoryPostgres) DeleteByBackupID(ctx context.Context, backupID valueobjects.BackupID) error {
	query := `DELETE FROM backup_errors WHERE backup_id = $1`
	_, err := r.db.ExecContext(ctx, query, backupID.String())
	if err != nil {
		return fmt.Errorf("failed to delete backup errors: %w", err)
	}
	return nil
}
