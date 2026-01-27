package valueobjects

import (
	"testing"

	"github.com/google/uuid"
	shared "github.com/rrbarrero/justbackup/internal/shared/domain"
	"github.com/stretchr/testify/assert"
)

func TestNewBackupID(t *testing.T) {
	id := NewBackupID()
	assert.NotEmpty(t, id.value)
	_, err := uuid.Parse(id.value)
	assert.NoError(t, err, "NewBackupID should generate a valid UUID")
}

func TestNewBackupIDFromString(t *testing.T) {
	t.Run("valid UUID string", func(t *testing.T) {
		validUUID := uuid.New().String()
		id, err := NewBackupIDFromString(validUUID)
		assert.NoError(t, err)
		assert.Equal(t, validUUID, id.value)
	})

	t.Run("invalid UUID string", func(t *testing.T) {
		invalidUUID := "not-a-uuid"
		id, err := NewBackupIDFromString(invalidUUID)
		assert.Error(t, err)
		assert.Equal(t, shared.ErrInvalidID, err)
		assert.Empty(t, id.value)
	})
}

func TestBackupID_String(t *testing.T) {
	expectedUUID := uuid.New().String()
	id := BackupID{value: expectedUUID}
	assert.Equal(t, expectedUUID, id.String())
}

type AnotherValueObject struct{}

func (a AnotherValueObject) Equals(other shared.ValueObject) bool { return false }

func TestBackupID_Equals(t *testing.T) {
	uuid1 := uuid.New().String()
	uuid2 := uuid.New().String()

	id1 := BackupID{value: uuid1}
	id1Copy := BackupID{value: uuid1}
	id2 := BackupID{value: uuid2}

	t.Run("equal IDs", func(t *testing.T) {
		assert.True(t, id1.Equals(id1Copy))
	})

	t.Run("unequal IDs", func(t *testing.T) {
		assert.False(t, id1.Equals(id2))
	})

	t.Run("different type", func(t *testing.T) {
		assert.False(t, id1.Equals(AnotherValueObject{}))
	})
}
