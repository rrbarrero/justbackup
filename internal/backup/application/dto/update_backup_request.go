package dto

type UpdateBackupRequest struct {
	ID          string              `json:"id"`
	Path        string              `json:"path"`
	Destination string              `json:"destination"`
	Schedule    string              `json:"schedule"`
	Excludes    []string            `json:"excludes"`
	Incremental bool                `json:"incremental"`
	Retention   int                 `json:"retention"`
	Encrypted   bool                `json:"encrypted"`
	Hooks       []CreateHookRequest `json:"hooks"`
}
