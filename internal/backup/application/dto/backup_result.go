package dto

type BackupResult struct {
	BackupID string `json:"backup_id"`
	Status   string `json:"status"` // "completed", "failed"
	Message  string `json:"message"`
	Path     string `json:"path,omitempty"`
}
