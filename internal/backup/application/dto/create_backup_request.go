package dto

type CreateBackupRequest struct {
	HostID      string              `json:"host_id"`
	Path        string              `json:"path"`
	Destination string              `json:"destination"`
	Schedule    string              `json:"schedule"` // Cron expression
	Excludes    []string            `json:"excludes"`
	Incremental bool                `json:"incremental"`
	Retention   int                 `json:"retention"`
	Encrypted   bool                `json:"encrypted"`
	Hooks       []CreateHookRequest `json:"hooks"`
}
