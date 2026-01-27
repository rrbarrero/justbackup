package entities

import "time"

// BackupSchedule defines the temporal orchestration rules for a backup process.
type BackupSchedule struct {
	CronExpression string
	LastRun        time.Time
	NextRun        time.Time
}

func NewBackupSchedule(cron string) BackupSchedule {
	return BackupSchedule{
		CronExpression: cron,
	}
}
