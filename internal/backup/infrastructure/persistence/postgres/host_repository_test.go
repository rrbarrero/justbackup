package postgres_test

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/rrbarrero/justbackup/internal/backup/domain/entities"
	"github.com/rrbarrero/justbackup/internal/backup/infrastructure/persistence/postgres"
	shared "github.com/rrbarrero/justbackup/internal/shared/domain"
)

func TestHostRepositoryPostgres_Save(t *testing.T) {
	db, mockDB, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := postgres.NewHostRepositoryPostgres(db)

	t.Run("success - insert new host", func(t *testing.T) {
		hostID := entities.NewHostID()
		host := entities.RestoreHost(
			hostID,
			"test-host",
			"hostname.example.com",
			"testuser",
			22,
			"/backup/path",
			false,
			time.Now(),
		)

		mockDB.ExpectExec(`INSERT INTO hosts \(id, name, hostname, "user", port, host_path, is_workstation, created_at\)`).
			WithArgs(
				hostID.String(),
				"test-host",
				"hostname.example.com",
				"testuser",
				22,
				"/backup/path",
				false,
				sqlmock.AnyArg(), // created_at
			).
			WillReturnResult(sqlmock.NewResult(1, 1))

		err := repo.Save(context.Background(), host)
		assert.NoError(t, err)
	})

	t.Run("success - update existing host", func(t *testing.T) {
		hostID := entities.NewHostID()
		host := entities.RestoreHost(
			hostID,
			"updated-host",
			"updated.example.com",
			"updateduser",
			2222,
			"/updated/path",
			true,
			time.Now(),
		)

		mockDB.ExpectExec(`INSERT INTO hosts .* ON CONFLICT \(id\) DO UPDATE SET`).
			WithArgs(
				hostID.String(),
				"updated-host",
				"updated.example.com",
				"updateduser",
				2222,
				"/updated/path",
				true,
				sqlmock.AnyArg(), // created_at
			).
			WillReturnResult(sqlmock.NewResult(1, 1))

		err := repo.Save(context.Background(), host)
		assert.NoError(t, err)
	})

	t.Run("database error", func(t *testing.T) {
		hostID := entities.NewHostID()
		host := entities.RestoreHost(
			hostID,
			"error-host",
			"error.example.com",
			"erroruser",
			22,
			"/error/path",
			false,
			time.Now(),
		)

		mockDB.ExpectExec(`INSERT INTO hosts`).
			WithArgs(
				hostID.String(),
				"error-host",
				"error.example.com",
				"erroruser",
				22,
				"/error/path",
				false,
				sqlmock.AnyArg(),
			).
			WillReturnError(sql.ErrConnDone)

		err := repo.Save(context.Background(), host)
		assert.Error(t, err)
		assert.Equal(t, sql.ErrConnDone, err)
	})
}

func TestHostRepositoryPostgres_Get(t *testing.T) {
	db, mockDB, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := postgres.NewHostRepositoryPostgres(db)

	t.Run("success", func(t *testing.T) {
		hostID := entities.NewHostID()
		createdAt := time.Now()

		rows := sqlmock.NewRows([]string{
			"id", "name", "hostname", "user", "port", "host_path", "is_workstation", "created_at",
		}).AddRow(
			hostID.String(),
			"test-host",
			"hostname.example.com",
			"testuser",
			22,
			"/backup/path",
			false,
			createdAt,
		)

		mockDB.ExpectQuery(`SELECT id, name, hostname, "user", port, host_path, is_workstation, created_at FROM hosts WHERE id =`).
			WithArgs(hostID.String()).
			WillReturnRows(rows)

		host, err := repo.Get(context.Background(), hostID)
		require.NoError(t, err)
		require.NotNil(t, host)

		assert.Equal(t, hostID.String(), host.ID().String())
		assert.Equal(t, "test-host", host.Name())
		assert.Equal(t, "hostname.example.com", host.Hostname())
		assert.Equal(t, "testuser", host.User())
		assert.Equal(t, 22, host.Port())
		assert.Equal(t, "/backup/path", host.Path())
		assert.Equal(t, false, host.IsWorkstation())
	})

	t.Run("not found", func(t *testing.T) {
		hostID := entities.NewHostID()

		mockDB.ExpectQuery(`SELECT id, name, hostname, "user", port, host_path, is_workstation, created_at FROM hosts WHERE id =`).
			WithArgs(hostID.String()).
			WillReturnError(sql.ErrNoRows)

		host, err := repo.Get(context.Background(), hostID)
		assert.Error(t, err)
		assert.Equal(t, shared.ErrNotFound, err)
		assert.Nil(t, host)
	})

	t.Run("database error", func(t *testing.T) {
		hostID := entities.NewHostID()

		mockDB.ExpectQuery(`SELECT id, name, hostname, "user", port, host_path, is_workstation, created_at FROM hosts WHERE id =`).
			WithArgs(hostID.String()).
			WillReturnError(sql.ErrConnDone)

		host, err := repo.Get(context.Background(), hostID)
		assert.Error(t, err)
		assert.Equal(t, sql.ErrConnDone, err)
		assert.Nil(t, host)
	})
}

func TestHostRepositoryPostgres_GetByIDs(t *testing.T) {
	db, mockDB, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := postgres.NewHostRepositoryPostgres(db)

	t.Run("success with multiple hosts", func(t *testing.T) {
		hostID1 := entities.NewHostID()
		hostID2 := entities.NewHostID()
		createdAt := time.Now()

		rows := sqlmock.NewRows([]string{
			"id", "name", "hostname", "user", "port", "host_path", "is_workstation", "created_at",
		}).AddRow(
			hostID1.String(),
			"host-1",
			"host1.example.com",
			"user1",
			22,
			"/path1",
			false,
			createdAt,
		).AddRow(
			hostID2.String(),
			"host-2",
			"host2.example.com",
			"user2",
			2222,
			"/path2",
			true,
			createdAt,
		)

		mockDB.ExpectQuery(`SELECT id, name, hostname, "user", port, host_path, is_workstation, created_at FROM hosts WHERE id = ANY\(\$1\)`).
			WithArgs(pq.Array([]string{hostID1.String(), hostID2.String()})).
			WillReturnRows(rows)

		hosts, err := repo.GetByIDs(context.Background(), []entities.HostID{hostID1, hostID2})
		require.NoError(t, err)
		require.Len(t, hosts, 2)

		assert.Equal(t, hostID1.String(), hosts[0].ID().String())
		assert.Equal(t, "host-1", hosts[0].Name())
		assert.Equal(t, hostID2.String(), hosts[1].ID().String())
		assert.Equal(t, "host-2", hosts[1].Name())
	})

	t.Run("success with empty IDs", func(t *testing.T) {
		hosts, err := repo.GetByIDs(context.Background(), []entities.HostID{})
		require.NoError(t, err)
		assert.Empty(t, hosts)
	})

	t.Run("database error", func(t *testing.T) {
		hostID := entities.NewHostID()

		mockDB.ExpectQuery(`SELECT id, name, hostname, "user", port, host_path, is_workstation, created_at FROM hosts WHERE id = ANY\(\$1\)`).
			WithArgs(pq.Array([]string{hostID.String()})).
			WillReturnError(sql.ErrConnDone)

		hosts, err := repo.GetByIDs(context.Background(), []entities.HostID{hostID})
		assert.Error(t, err)
		assert.Equal(t, sql.ErrConnDone, err)
		assert.Nil(t, hosts)
	})
}

func TestHostRepositoryPostgres_List(t *testing.T) {
	db, mockDB, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := postgres.NewHostRepositoryPostgres(db)

	t.Run("success", func(t *testing.T) {
		hostID1 := entities.NewHostID()
		hostID2 := entities.NewHostID()
		createdAt := time.Now()

		rows := sqlmock.NewRows([]string{
			"id", "name", "hostname", "user", "port", "host_path", "is_workstation", "created_at",
		}).AddRow(
			hostID1.String(),
			"list-host-1",
			"list1.example.com",
			"listuser1",
			22,
			"/listpath1",
			false,
			createdAt,
		).AddRow(
			hostID2.String(),
			"list-host-2",
			"list2.example.com",
			"listuser2",
			2222,
			"/listpath2",
			true,
			createdAt,
		)

		mockDB.ExpectQuery(`SELECT id, name, hostname, "user", port, host_path, is_workstation, created_at FROM hosts`).
			WillReturnRows(rows)

		hosts, err := repo.List(context.Background())
		require.NoError(t, err)
		require.Len(t, hosts, 2)

		assert.Equal(t, "list-host-1", hosts[0].Name())
		assert.Equal(t, "list-host-2", hosts[1].Name())
	})

	t.Run("success with empty result", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{
			"id", "name", "hostname", "user", "port", "host_path", "is_workstation", "created_at",
		})

		mockDB.ExpectQuery(`SELECT id, name, hostname, "user", port, host_path, is_workstation, created_at FROM hosts`).
			WillReturnRows(rows)

		hosts, err := repo.List(context.Background())
		require.NoError(t, err)
		assert.Empty(t, hosts)
	})

	t.Run("database error", func(t *testing.T) {
		mockDB.ExpectQuery(`SELECT id, name, hostname, "user", port, host_path, is_workstation, created_at FROM hosts`).
			WillReturnError(sql.ErrConnDone)

		hosts, err := repo.List(context.Background())
		assert.Error(t, err)
		assert.Equal(t, sql.ErrConnDone, err)
		assert.Nil(t, hosts)
	})
}

func TestHostRepositoryPostgres_Update(t *testing.T) {
	db, mockDB, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := postgres.NewHostRepositoryPostgres(db)

	t.Run("success", func(t *testing.T) {
		hostID := entities.NewHostID()
		host := entities.RestoreHost(
			hostID,
			"updated-host",
			"updated.example.com",
			"updateduser",
			2222,
			"/updated/path",
			true,
			time.Now(),
		)

		mockDB.ExpectExec(`UPDATE hosts SET name = \$2, hostname = \$3, "user" = \$4, port = \$5, host_path = \$6, is_workstation = \$7 WHERE id = \$1`).
			WithArgs(
				hostID.String(),
				"updated-host",
				"updated.example.com",
				"updateduser",
				2222,
				"/updated/path",
				true,
			).
			WillReturnResult(sqlmock.NewResult(1, 1))

		err := repo.Update(context.Background(), host)
		assert.NoError(t, err)
	})

	t.Run("not found", func(t *testing.T) {
		hostID := entities.NewHostID()
		host := entities.RestoreHost(
			hostID,
			"not-found-host",
			"notfound.example.com",
			"notfounduser",
			22,
			"/notfound/path",
			false,
			time.Now(),
		)

		mockDB.ExpectExec(`UPDATE hosts SET name = \$2, hostname = \$3, "user" = \$4, port = \$5, host_path = \$6, is_workstation = \$7 WHERE id = \$1`).
			WithArgs(
				hostID.String(),
				"not-found-host",
				"notfound.example.com",
				"notfounduser",
				22,
				"/notfound/path",
				false,
			).
			WillReturnResult(sqlmock.NewResult(0, 0))

		err := repo.Update(context.Background(), host)
		assert.Error(t, err)
		assert.Equal(t, shared.ErrNotFound, err)
	})

	t.Run("database error", func(t *testing.T) {
		hostID := entities.NewHostID()
		host := entities.RestoreHost(
			hostID,
			"error-host",
			"error.example.com",
			"erroruser",
			22,
			"/error/path",
			false,
			time.Now(),
		)

		mockDB.ExpectExec(`UPDATE hosts SET name = \$2, hostname = \$3, "user" = \$4, port = \$5, host_path = \$6, is_workstation = \$7 WHERE id = \$1`).
			WithArgs(
				hostID.String(),
				"error-host",
				"error.example.com",
				"erroruser",
				22,
				"/error/path",
				false,
			).
			WillReturnError(sql.ErrConnDone)

		err := repo.Update(context.Background(), host)
		assert.Error(t, err)
		assert.Equal(t, sql.ErrConnDone, err)
	})
}

func TestHostRepositoryPostgres_Delete(t *testing.T) {
	db, mockDB, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := postgres.NewHostRepositoryPostgres(db)

	t.Run("success", func(t *testing.T) {
		hostID := entities.NewHostID()

		mockDB.ExpectExec(`DELETE FROM hosts WHERE id = \$1`).
			WithArgs(hostID.String()).
			WillReturnResult(sqlmock.NewResult(1, 1))

		err := repo.Delete(context.Background(), hostID)
		assert.NoError(t, err)
	})

	t.Run("not found", func(t *testing.T) {
		hostID := entities.NewHostID()

		mockDB.ExpectExec(`DELETE FROM hosts WHERE id = \$1`).
			WithArgs(hostID.String()).
			WillReturnResult(sqlmock.NewResult(0, 0))

		err := repo.Delete(context.Background(), hostID)
		assert.Error(t, err)
		assert.Equal(t, shared.ErrNotFound, err)
	})

	t.Run("database error", func(t *testing.T) {
		hostID := entities.NewHostID()

		mockDB.ExpectExec(`DELETE FROM hosts WHERE id = \$1`).
			WithArgs(hostID.String()).
			WillReturnError(sql.ErrConnDone)

		err := repo.Delete(context.Background(), hostID)
		assert.Error(t, err)
		assert.Equal(t, sql.ErrConnDone, err)
	})
}

func TestHostRepositoryPostgres_Count(t *testing.T) {
	db, mockDB, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := postgres.NewHostRepositoryPostgres(db)

	t.Run("success", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"count"}).AddRow(int64(5))

		mockDB.ExpectQuery(`SELECT COUNT\(\*\) FROM hosts`).
			WillReturnRows(rows)

		count, err := repo.Count(context.Background())
		assert.NoError(t, err)
		assert.Equal(t, int64(5), count)
	})

	t.Run("success with zero", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"count"}).AddRow(int64(0))

		mockDB.ExpectQuery(`SELECT COUNT\(\*\) FROM hosts`).
			WillReturnRows(rows)

		count, err := repo.Count(context.Background())
		assert.NoError(t, err)
		assert.Equal(t, int64(0), count)
	})

	t.Run("database error", func(t *testing.T) {
		mockDB.ExpectQuery(`SELECT COUNT\(\*\) FROM hosts`).
			WillReturnError(sql.ErrConnDone)

		count, err := repo.Count(context.Background())
		assert.Error(t, err)
		assert.Equal(t, sql.ErrConnDone, err)
		assert.Equal(t, int64(0), count)
	})
}
