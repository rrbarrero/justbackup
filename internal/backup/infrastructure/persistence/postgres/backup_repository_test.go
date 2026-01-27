package postgres_test

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/rrbarrero/justbackup/internal/backup/domain/entities"
	"github.com/rrbarrero/justbackup/internal/backup/domain/valueobjects"
	"github.com/rrbarrero/justbackup/internal/backup/infrastructure/persistence/postgres"
	shared "github.com/rrbarrero/justbackup/internal/shared/domain"
)

// MockEncryptionService is a mock implementation of EncryptionService
type MockEncryptionService struct {
	mock.Mock
}

func (m *MockEncryptionService) Encrypt(data []byte) ([]byte, error) {
	args := m.Called(data)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]byte), args.Error(1)
}

func (m *MockEncryptionService) Decrypt(data []byte) ([]byte, error) {
	args := m.Called(data)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]byte), args.Error(1)
}

func TestBackupRepositoryPostgres_Save_HookErrors(t *testing.T) {
	db, mockDB, err := sqlmock.New()
	require.NoError(t, err)
	defer func() { _ = db.Close() }()

	mockEnc := new(MockEncryptionService)
	repo := postgres.NewBackupRepositoryPostgres(db, mockEnc)

	// Create valid backup entity
	backupID := valueobjects.NewBackupID()
	hostID := entities.NewHostID()
	schedule := entities.NewBackupSchedule("0 0 * * *")
	backup := entities.RestoreBackup(
		backupID,
		hostID,
		"/source",
		"/dest",
		valueobjects.BackupStatusPending,
		schedule,
		time.Now(),
		time.Now(),
		nil,
		[]string{},
		true,
		false,
		"",
		0,
		false,
	)

	t.Run("fails when encryption fails", func(t *testing.T) {
		// Expect initial backup insert mock
		mockDB.ExpectExec("INSERT INTO backups").WillReturnResult(sqlmock.NewResult(1, 1))
		mockDB.ExpectExec("DELETE FROM backup_hooks").WillReturnResult(sqlmock.NewResult(1, 1))

		// Valid params
		hook := entities.BackupHook{
			ID:       uuid.New(),
			BackupID: uuid.MustParse(backupID.String()),
			Name:     "encrypt_fail_hook",
			Params:   map[string]string{"foo": "bar"},
		}
		backup.SetHooks([]*entities.BackupHook{&hook})

		// Mock encryption failure
		mockEnc.On("Encrypt", mock.Anything).Return(nil, errors.New("encryption error")).Once()

		err := repo.Save(context.Background(), backup)
		assert.Error(t, err)
		assert.Equal(t, "encryption error", err.Error())
		mockEnc.AssertExpectations(t)
	})
}

func TestBackupRepositoryPostgres_LoadHooks_Errors(t *testing.T) {
	db, mockDB, err := sqlmock.New()
	require.NoError(t, err)
	defer func() { _ = db.Close() }()

	mockEnc := new(MockEncryptionService)
	repo := postgres.NewBackupRepositoryPostgres(db, mockEnc)

	backupID := valueobjects.NewBackupID()

	t.Run("fails when hook ID is invalid", func(t *testing.T) {
		// Mock FindByID basic query
		rows := sqlmock.NewRows([]string{
			"id", "host_id", "path", "destination", "status", "schedule",
			"created_at", "updated_at", "last_run", "next_run_at", "excludes",
			"enabled", "incremental", "size", "retention", "encrypted",
		}).AddRow(
			backupID.String(), entities.NewHostID().String(), "/src", "/dst", "pending", "0 0 * * *",
			time.Now(), time.Now(), nil, nil, "{}", true, false, nil, 0, false,
		)
		mockDB.ExpectQuery("SELECT .* FROM backups WHERE id =").WillReturnRows(rows)

		// Mock Hook Load with invalid UUID
		hookRows := sqlmock.NewRows([]string{
			"id", "backup_id", "name", "phase", "enabled", "params", "created_at", "updated_at",
		}).AddRow(
			"invalid-uuid", backupID.String(), "bad_id_hook", "pre", true, "{}", time.Now(), time.Now(),
		)
		mockDB.ExpectQuery("SELECT .* FROM backup_hooks WHERE backup_id =").WillReturnRows(hookRows)

		_, err := repo.FindByID(context.Background(), backupID)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to parse hook ID")
	})

	t.Run("fails when backup ID in hook is invalid", func(t *testing.T) {
		// Mock FindByID basic query
		rows := sqlmock.NewRows([]string{
			"id", "host_id", "path", "destination", "status", "schedule",
			"created_at", "updated_at", "last_run", "next_run_at", "excludes",
			"enabled", "incremental", "size", "retention", "encrypted",
		}).AddRow(
			backupID.String(), entities.NewHostID().String(), "/src", "/dst", "pending", "0 0 * * *",
			time.Now(), time.Now(), nil, nil, "{}", true, false, nil, 0, false,
		)
		mockDB.ExpectQuery("SELECT .* FROM backups WHERE id =").WillReturnRows(rows)

		// Mock Hook Load with invalid Backup UUID in the hook table (corruption)
		hookRows := sqlmock.NewRows([]string{
			"id", "backup_id", "name", "phase", "enabled", "params", "created_at", "updated_at",
		}).AddRow(
			uuid.New().String(), "invalid-backup-uuid", "bad_bid_hook", "pre", true, "{}", time.Now(), time.Now(),
		)
		mockDB.ExpectQuery("SELECT .* FROM backup_hooks WHERE backup_id =").WillReturnRows(hookRows)

		_, err := repo.FindByID(context.Background(), backupID)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to parse backup ID")
	})
}

func TestBackupRepositoryPostgres_Save_Success(t *testing.T) {
	db, mockDB, err := sqlmock.New()
	require.NoError(t, err)
	defer func() { _ = db.Close() }()

	mockEnc := new(MockEncryptionService)
	repo := postgres.NewBackupRepositoryPostgres(db, mockEnc)

	backupID := valueobjects.NewBackupID()
	hostID := entities.NewHostID()
	schedule := entities.NewBackupSchedule("0 0 * * *")
	backup := entities.RestoreBackup(
		backupID,
		hostID,
		"/source",
		"/dest",
		valueobjects.BackupStatusPending,
		schedule,
		time.Now(),
		time.Now(),
		nil,
		[]string{"*.tmp"},
		true,
		false,
		"1GB",
		5,
		true,
	)

	// Add a hook
	hook := entities.BackupHook{
		ID:       uuid.New(),
		BackupID: uuid.MustParse(backupID.String()),
		Name:     "test_hook",
		Phase:    entities.HookPhasePre,
		Enabled:  true,
		Params:   map[string]string{"key": "value"},
	}
	backup.SetHooks([]*entities.BackupHook{&hook})

	// Expectations
	mockDB.ExpectExec("INSERT INTO backups").
		WithArgs(
			backup.ID().String(),
			backup.HostID().String(),
			backup.Path(),
			backup.Destination(),
			backup.Status().String(),
			backup.Schedule().CronExpression,
			sqlmock.AnyArg(), // CreatedAt
			sqlmock.AnyArg(), // UpdatedAt
			nil,              // LastRun
			nil,              // NextRunAt
			sqlmock.AnyArg(), // Excludes (pq.Array)
			backup.Enabled(),
			backup.Incremental(),
			backup.Size(),
			backup.Retention(),
			backup.Encrypted(),
		).
		WillReturnResult(sqlmock.NewResult(1, 1))

	// Hook delete expectation
	mockDB.ExpectExec("DELETE FROM backup_hooks").
		WithArgs(backup.ID().String()).
		WillReturnResult(sqlmock.NewResult(1, 1))

	// Encryption expectation
	// Note: We'll have Encrypt return "encrypted_data" (raw bytes)
	// But JSON.Marshal checks result in base64 string
	encryptedBytes := []byte("encrypted_data")
	mockEnc.On("Encrypt", mock.Anything).Return(encryptedBytes, nil).Once()

	// Calculate what we expect in the DB
	// json.Marshal([]byte("encrypted_data")) -> "ZW5jcnlwdGVkX2RhdGE=" (with quotes)
	expectedJSONBytes, _ := json.Marshal(encryptedBytes)

	// Hook insert expectation
	mockDB.ExpectExec("INSERT INTO backup_hooks").
		WithArgs(
			hook.ID.String(),
			hook.BackupID.String(),
			hook.Name,
			string(hook.Phase),
			hook.Enabled,
			string(expectedJSONBytes), // Corrected expectation
			sqlmock.AnyArg(),          // CreatedAt
			sqlmock.AnyArg(),          // UpdatedAt
		).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err = repo.Save(context.Background(), backup)
	assert.NoError(t, err)
	mockEnc.AssertExpectations(t)
}

func TestBackupRepositoryPostgres_FindByID_Success(t *testing.T) {
	db, mockDB, err := sqlmock.New()
	require.NoError(t, err)
	defer func() { _ = db.Close() }()

	mockEnc := new(MockEncryptionService)
	repo := postgres.NewBackupRepositoryPostgres(db, mockEnc)

	backupID := valueobjects.NewBackupID()
	hostID := entities.NewHostID()

	// Backup Row Mock
	rows := sqlmock.NewRows([]string{
		"id", "host_id", "path", "destination", "status", "schedule",
		"created_at", "updated_at", "last_run", "next_run_at", "excludes",
		"enabled", "incremental", "size", "retention", "encrypted",
	}).AddRow(
		backupID.String(), hostID.String(), "/src", "/dst", "pending", "0 0 * * *",
		time.Now(), time.Now(), nil, nil, "{*.log}", true, false, "500MB", 3, false,
	)
	mockDB.ExpectQuery("SELECT .* FROM backups WHERE id =").
		WithArgs(backupID.String()).
		WillReturnRows(rows)

	// Hook Row Mock
	hookID := uuid.New()

	// Prepare the encrypted content as it sits in the DB jsonb/blob/string
	rawEncryptedContent := []byte("encrypted_content")
	// json.Marshal(rawEncryptedContent) -> "ZW5jcnlwdGVkX2NvbnRlbnQ="
	// This is what we put in the DB column so when 'paramsJSON' is scanned, it returns these bytes
	dbStoredJSON, _ := json.Marshal(rawEncryptedContent)

	hookRows := sqlmock.NewRows([]string{
		"id", "backup_id", "name", "phase", "enabled", "params", "created_at", "updated_at",
	}).AddRow(
		hookID.String(), backupID.String(), "my_hook", "post", true, dbStoredJSON, time.Now(), time.Now(),
	)
	mockDB.ExpectQuery("SELECT .* FROM backup_hooks WHERE backup_id =").
		WithArgs(backupID.String()).
		WillReturnRows(hookRows)

	// Decryption expectation - repo calls Decrypt with the raw bytes after JSON unmarshal
	mockEnc.On("Decrypt", rawEncryptedContent).
		Return([]byte(`{"key":"decrypted_value"}`), nil).
		Once()

	start := time.Now()
	backup, err := repo.FindByID(context.Background(), backupID)
	require.NoError(t, err)
	require.NotNil(t, backup)
	assert.Equal(t, backupID.String(), backup.ID().String())
	assert.Len(t, backup.Hooks(), 1)
	assert.Equal(t, "decrypted_value", backup.Hooks()[0].Params["key"])

	if time.Since(start) > 2*time.Second {
		t.Log("Warning: Test took longer than expected")
	}
}

func TestBackupRepositoryPostgres_Delete(t *testing.T) {
	db, mockDB, err := sqlmock.New()
	require.NoError(t, err)
	defer func() { _ = db.Close() }()

	repo := postgres.NewBackupRepositoryPostgres(db, nil)
	backupID := valueobjects.NewBackupID()

	t.Run("success", func(t *testing.T) {
		mockDB.ExpectExec("DELETE FROM backups").
			WithArgs(backupID.String()).
			WillReturnResult(sqlmock.NewResult(1, 1))

		err := repo.Delete(context.Background(), backupID)
		assert.NoError(t, err)
	})

	t.Run("not found", func(t *testing.T) {
		mockDB.ExpectExec("DELETE FROM backups").
			WithArgs(backupID.String()).
			WillReturnResult(sqlmock.NewResult(1, 0))

		err := repo.Delete(context.Background(), backupID)
		assert.Error(t, err)
		assert.Equal(t, shared.ErrNotFound, err)
	})
}

func TestBackupRepositoryPostgres_GetStats(t *testing.T) {
	db, mockDB, err := sqlmock.New()
	require.NoError(t, err)
	defer func() { _ = db.Close() }()

	repo := postgres.NewBackupRepositoryPostgres(db, nil)

	// Count query mock
	rows := sqlmock.NewRows([]string{"total", "pending", "completed", "failed"}).
		AddRow(10, 2, 5, 3)
	mockDB.ExpectQuery("SELECT .* FROM backups").WillReturnRows(rows)

	// Size query mock
	sizeRows := sqlmock.NewRows([]string{"size"}).
		AddRow("1 GB").
		AddRow("500 MB")
	mockDB.ExpectQuery("SELECT size FROM backups").WillReturnRows(sizeRows)

	stats, err := repo.GetStats(context.Background())
	assert.NoError(t, err)
	assert.NotNil(t, stats)
	assert.Equal(t, 10, stats.Total)
	assert.Equal(t, 2, stats.Pending)
	assert.Equal(t, 5, stats.Completed)
	assert.Equal(t, 3, stats.Failed)
	assert.NotEmpty(t, stats.TotalSize)
}
