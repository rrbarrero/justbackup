package entities_test

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/rrbarrero/justbackup/internal/backup/domain/entities"
	shared "github.com/rrbarrero/justbackup/internal/shared/domain"
	"github.com/stretchr/testify/assert"
)

func TestHostID(t *testing.T) {
	t.Run("NewHostID should generate a valid UUID", func(t *testing.T) {
		hostID := entities.NewHostID()
		_, err := uuid.Parse(hostID.String())
		assert.NoError(t, err)
		assert.NotEmpty(t, hostID.String())
	})

	t.Run("NewHostIDFromString should create HostID from valid string", func(t *testing.T) {
		validUUID := uuid.New().String()
		hostID, err := entities.NewHostIDFromString(validUUID)
		assert.NoError(t, err)
		assert.Equal(t, validUUID, hostID.String())
	})

	t.Run("NewHostIDFromString should return error for invalid string", func(t *testing.T) {
		invalidUUID := "invalid-uuid"
		_, err := entities.NewHostIDFromString(invalidUUID)
		assert.Error(t, err)
		assert.Equal(t, shared.ErrInvalidID, err)
	})

	t.Run("HostID.String should return the underlying value", func(t *testing.T) {
		idValue := uuid.New().String()
		hostID, _ := entities.NewHostIDFromString(idValue)
		assert.Equal(t, idValue, hostID.String())
	})
}

func TestHost(t *testing.T) {
	originalNowFunc := entities.NowFunc
	defer func() { entities.NowFunc = originalNowFunc }()

	t.Run("NewHost should create a host with generated ID and current time", func(t *testing.T) {
		name := "TestHost"
		hostname := "testhost.com"
		user := "testuser"
		port := 22
		path := "test-host"

		// Mock time.Now for deterministic test
		fixedTime := time.Date(2023, time.January, 1, 12, 0, 0, 0, time.UTC)
		entities.NowFunc = func() time.Time { return fixedTime }

		host := entities.NewHost(name, hostname, user, port, path, false)

		assert.NotNil(t, host)
		assert.NotEmpty(t, host.ID().String())
		_, err := uuid.Parse(host.ID().String())
		assert.NoError(t, err)
		assert.Equal(t, name, host.Name())
		assert.Equal(t, hostname, host.Hostname())
		assert.Equal(t, user, host.User())
		assert.Equal(t, port, host.Port())
		assert.Equal(t, path, host.Path())
		assert.False(t, host.IsWorkstation())
		assert.Equal(t, fixedTime, host.CreatedAt())
	})

	t.Run("RestoreHost should create a host with provided values", func(t *testing.T) {
		id := entities.NewHostID()
		name := "RestoredHost"
		hostname := "restoredhost.com"
		user := "restoreuser"
		port := 2222
		path := "restored-host"
		isWorkstation := true
		createdAt := time.Date(2022, time.December, 31, 23, 59, 59, 0, time.UTC)

		host := entities.RestoreHost(id, name, hostname, user, port, path, isWorkstation, createdAt)

		assert.NotNil(t, host)
		assert.Equal(t, id, host.ID())
		assert.Equal(t, name, host.Name())
		assert.Equal(t, hostname, host.Hostname())
		assert.Equal(t, user, host.User())
		assert.Equal(t, port, host.Port())
		assert.Equal(t, path, host.Path())
		assert.True(t, host.IsWorkstation())
		assert.Equal(t, createdAt, host.CreatedAt())
	})

	t.Run("Host getter methods should return correct values", func(t *testing.T) {
		id := entities.NewHostID()
		name := "GetterHost"
		hostname := "getterhost.com"
		user := "getteruser"
		port := 8080
		path := "getter-host"
		isWorkstation := false
		createdAt := time.Date(2021, time.November, 1, 10, 0, 0, 0, time.UTC)

		host := entities.RestoreHost(id, name, hostname, user, port, path, isWorkstation, createdAt)

		assert.Equal(t, id, host.ID())
		assert.Equal(t, name, host.Name())
		assert.Equal(t, hostname, host.Hostname())
		assert.Equal(t, user, host.User())
		assert.Equal(t, port, host.Port())
		assert.Equal(t, path, host.Path())
		assert.Equal(t, isWorkstation, host.IsWorkstation())
		assert.Equal(t, createdAt, host.CreatedAt())
	})
}
