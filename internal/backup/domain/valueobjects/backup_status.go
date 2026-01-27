package valueobjects

import "errors"

type BackupStatus string

const (
	BackupStatusPending   BackupStatus = "pending"
	BackupStatusRunning   BackupStatus = "running"
	BackupStatusCompleted BackupStatus = "completed"
	BackupStatusFailed    BackupStatus = "failed"
)

var ErrInvalidStatus = errors.New("invalid backup status")

func NewBackupStatus(status string) (BackupStatus, error) {
	switch BackupStatus(status) {
	case BackupStatusPending, BackupStatusRunning, BackupStatusCompleted, BackupStatusFailed:
		return BackupStatus(status), nil
	default:
		return "", ErrInvalidStatus
	}
}

func (s BackupStatus) String() string {
	return string(s)
}
