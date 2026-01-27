package valueobjects

import (
	"github.com/google/uuid"
	shared "github.com/rrbarrero/justbackup/internal/shared/domain"
)

type BackupID struct {
	value string
}

func NewBackupID() BackupID {
	return BackupID{value: uuid.New().String()}
}

func NewBackupIDFromString(id string) (BackupID, error) {
	if _, err := uuid.Parse(id); err != nil {
		return BackupID{}, shared.ErrInvalidID
	}
	return BackupID{value: id}, nil
}

func (id BackupID) String() string {
	return id.value
}

func (id BackupID) Equals(other shared.ValueObject) bool {
	otherID, ok := other.(BackupID)
	return ok && id.value == otherID.value
}
