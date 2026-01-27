package dto

type WorkerResult struct {
	Type    TaskType    `json:"type"`
	TaskID  string      `json:"task_id"`
	JobID   string      `json:"job_id"`
	Status  string      `json:"status"` // "completed", "failed"
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

type SearchFilesResult struct {
	Files []string `json:"files"`
}

type FileListItem struct {
	Name  string `json:"name"`
	IsDir bool   `json:"is_dir"`
	Size  int64  `json:"size"`
}

type ListFilesResult struct {
	Files []FileListItem `json:"files"`
}
