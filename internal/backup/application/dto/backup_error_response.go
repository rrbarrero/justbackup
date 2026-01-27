package dto

import "time"

type BackupErrorResponse struct {
	ID           string    `json:"id"`
	JobID        string    `json:"job_id"`
	BackupID     string    `json:"backup_id"`
	OccurredAt   time.Time `json:"occurred_at"`
	ErrorMessage string    `json:"error_message"`
}
