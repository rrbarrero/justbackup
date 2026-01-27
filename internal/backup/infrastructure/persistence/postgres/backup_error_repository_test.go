package postgres_test

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/rrbarrero/justbackup/internal/backup/domain/entities"
	"github.com/rrbarrero/justbackup/internal/backup/domain/valueobjects"
	"github.com/rrbarrero/justbackup/internal/backup/infrastructure/persistence/postgres"
)

func TestBackupErrorRepositoryPostgres_Save(t *testing.T) {
	db, mockDB, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := postgres.NewBackupErrorRepositoryPostgres(db)

	t.Run("success", func(t *testing.T) {
		backupID := valueobjects.NewBackupID()
		backupError := &entities.BackupError{
			JobID:        "job-123",
			BackupID:     backupID,
			OccurredAt:   time.Now(),
			ErrorMessage: "Test error message",
		}

		mockDB.ExpectExec(`INSERT INTO backup_errors \(job_id, backup_id, occurred_at, error_message\) VALUES \(\$1, \$2, \$3, \$4\)`).
			WithArgs(
				backupError.JobID,
				backupID.String(),
				backupError.OccurredAt,
				backupError.ErrorMessage,
			).
			WillReturnResult(sqlmock.NewResult(1, 1))

		err := repo.Save(context.Background(), backupError)
		assert.NoError(t, err)
	})

	t.Run("database error", func(t *testing.T) {
		backupID := valueobjects.NewBackupID()
		backupError := &entities.BackupError{
			JobID:        "job-123",
			BackupID:     backupID,
			OccurredAt:   time.Now(),
			ErrorMessage: "Test error message",
		}

		mockDB.ExpectExec(`INSERT INTO backup_errors \(job_id, backup_id, occurred_at, error_message\) VALUES \(\$1, \$2, \$3, \$4\)`).
			WithArgs(
				backupError.JobID,
				backupID.String(),
				backupError.OccurredAt,
				backupError.ErrorMessage,
			).
			WillReturnError(sql.ErrConnDone)

		err := repo.Save(context.Background(), backupError)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to save backup error")
		assert.Contains(t, err.Error(), sql.ErrConnDone.Error())
	})
}

func TestBackupErrorRepositoryPostgres_FindByBackupID(t *testing.T) {
	db, mockDB, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := postgres.NewBackupErrorRepositoryPostgres(db)

	t.Run("success with multiple errors", func(t *testing.T) {
		backupID := valueobjects.NewBackupID()
		occurredAt1 := time.Now().Add(-2 * time.Hour)
		occurredAt2 := time.Now().Add(-1 * time.Hour)

		rows := sqlmock.NewRows([]string{"id", "job_id", "backup_id", "occurred_at", "error_message"}).
			AddRow("1", "job-123", backupID.String(), occurredAt1, "First error").
			AddRow("2", "job-456", backupID.String(), occurredAt2, "Second error")

		mockDB.ExpectQuery(`SELECT id, job_id, backup_id, occurred_at, error_message FROM backup_errors WHERE backup_id = \$1 ORDER BY occurred_at DESC`).
			WithArgs(backupID.String()).
			WillReturnRows(rows)

		errors, err := repo.FindByBackupID(context.Background(), backupID)
		require.NoError(t, err)
		require.Len(t, errors, 2)

		assert.Equal(t, "1", errors[0].ID)
		assert.Equal(t, "job-123", errors[0].JobID)
		assert.Equal(t, backupID.String(), errors[0].BackupID.String())
		assert.Equal(t, "First error", errors[0].ErrorMessage)

		assert.Equal(t, "2", errors[1].ID)
		assert.Equal(t, "job-456", errors[1].JobID)
		assert.Equal(t, backupID.String(), errors[1].BackupID.String())
		assert.Equal(t, "Second error", errors[1].ErrorMessage)
	})

	t.Run("success with empty result", func(t *testing.T) {
		backupID := valueobjects.NewBackupID()

		rows := sqlmock.NewRows([]string{"id", "job_id", "backup_id", "occurred_at", "error_message"})

		mockDB.ExpectQuery(`SELECT id, job_id, backup_id, occurred_at, error_message FROM backup_errors WHERE backup_id = \$1 ORDER BY occurred_at DESC`).
			WithArgs(backupID.String()).
			WillReturnRows(rows)

		errors, err := repo.FindByBackupID(context.Background(), backupID)
		require.NoError(t, err)
		assert.Empty(t, errors)
	})

	t.Run("database error on query", func(t *testing.T) {
		backupID := valueobjects.NewBackupID()

		mockDB.ExpectQuery(`SELECT id, job_id, backup_id, occurred_at, error_message FROM backup_errors WHERE backup_id = \$1 ORDER BY occurred_at DESC`).
			WithArgs(backupID.String()).
			WillReturnError(sql.ErrConnDone)

		errors, err := repo.FindByBackupID(context.Background(), backupID)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to query backup errors")
		assert.Contains(t, err.Error(), sql.ErrConnDone.Error())
		assert.Nil(t, errors)
	})

	t.Run("error scanning row", func(t *testing.T) {
		backupID := valueobjects.NewBackupID()

		rows := sqlmock.NewRows([]string{"id", "job_id", "backup_id", "occurred_at", "error_message"}).
			AddRow(nil, "job-123", backupID.String(), time.Now(), "Error message")

		mockDB.ExpectQuery(`SELECT id, job_id, backup_id, occurred_at, error_message FROM backup_errors WHERE backup_id = \$1 ORDER BY occurred_at DESC`).
			WithArgs(backupID.String()).
			WillReturnRows(rows)

		errors, err := repo.FindByBackupID(context.Background(), backupID)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to scan backup error")
		assert.Nil(t, errors)
	})

	t.Run("error parsing backup ID", func(t *testing.T) {
		backupID := valueobjects.NewBackupID()

		rows := sqlmock.NewRows([]string{"id", "job_id", "backup_id", "occurred_at", "error_message"}).
			AddRow(int64(1), "job-123", "invalid-backup-id", time.Now(), "Error message")

		mockDB.ExpectQuery(`SELECT id, job_id, backup_id, occurred_at, error_message FROM backup_errors WHERE backup_id = \$1 ORDER BY occurred_at DESC`).
			WithArgs(backupID.String()).
			WillReturnRows(rows)

		errors, err := repo.FindByBackupID(context.Background(), backupID)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to parse backup ID")
		assert.Nil(t, errors)
	})

	t.Run("error on rows iteration", func(t *testing.T) {
		backupID := valueobjects.NewBackupID()

		// Create rows that will cause an iteration error after the first row
		rows := sqlmock.NewRows([]string{"id", "job_id", "backup_id", "occurred_at", "error_message"}).
			AddRow("1", "job-123", backupID.String(), time.Now(), "Error message")
		// Add an error after the first row
		rows.AddRow("2", "job-456", backupID.String(), time.Now(), "Second error").
			RowError(1, sql.ErrTxDone)

		mockDB.ExpectQuery(`SELECT id, job_id, backup_id, occurred_at, error_message FROM backup_errors WHERE backup_id = \$1 ORDER BY occurred_at DESC`).
			WithArgs(backupID.String()).
			WillReturnRows(rows)

		errors, err := repo.FindByBackupID(context.Background(), backupID)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "error iterating backup errors")
		assert.Nil(t, errors)
	})
}

func TestBackupErrorRepositoryPostgres_DeleteByBackupID(t *testing.T) {
	db, mockDB, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := postgres.NewBackupErrorRepositoryPostgres(db)

	t.Run("success", func(t *testing.T) {
		backupID := valueobjects.NewBackupID()

		mockDB.ExpectExec(`DELETE FROM backup_errors WHERE backup_id = \$1`).
			WithArgs(backupID.String()).
			WillReturnResult(sqlmock.NewResult(2, 2)) // 2 rows deleted

		err := repo.DeleteByBackupID(context.Background(), backupID)
		assert.NoError(t, err)
	})

	t.Run("success with no rows deleted", func(t *testing.T) {
		backupID := valueobjects.NewBackupID()

		mockDB.ExpectExec(`DELETE FROM backup_errors WHERE backup_id = \$1`).
			WithArgs(backupID.String()).
			WillReturnResult(sqlmock.NewResult(0, 0)) // 0 rows deleted

		err := repo.DeleteByBackupID(context.Background(), backupID)
		assert.NoError(t, err)
	})

	t.Run("database error", func(t *testing.T) {
		backupID := valueobjects.NewBackupID()

		mockDB.ExpectExec(`DELETE FROM backup_errors WHERE backup_id = \$1`).
			WithArgs(backupID.String()).
			WillReturnError(sql.ErrConnDone)

		err := repo.DeleteByBackupID(context.Background(), backupID)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to delete backup errors")
		assert.Contains(t, err.Error(), sql.ErrConnDone.Error())
	})
}
