package dto

type RestoreRequest struct {
	BackupID     string `json:"backup_id"`
	Path         string `json:"path"`          // Path inside the backup
	RestoreType  string `json:"restore_type"`  // "local" or "remote"
	RestoreAddr  string `json:"restore_addr"`  // For local: CLI address
	RestoreToken string `json:"restore_token"` // For local: Auth token
	TargetHostID string `json:"target_host_id"`
	TargetPath   string `json:"target_path"`
}
