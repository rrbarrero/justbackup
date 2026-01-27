package dto

import "time"

type BackupResponse struct {
	ID          string    `json:"id"`
	HostID      string    `json:"host_id"`
	HostName    string    `json:"host_name"`
	HostAddress string    `json:"host_address"`
	Path        string    `json:"path"`
	Destination string    `json:"destination"`
	Status      string    `json:"status"`
	Schedule    string    `json:"schedule"`
	LastRun     time.Time `json:"last_run"`
	Excludes    []string  `json:"excludes"`
	Incremental bool      `json:"incremental"`
	Size        string    `json:"size"`
	Retention   int       `json:"retention"`
	Encrypted   bool      `json:"encrypted"`
	Hooks       []HookDTO `json:"hooks"`
}

type FileSearchResult struct {
	Path   string          `json:"path"`
	Backup *BackupResponse `json:"backup,omitempty"`
}
