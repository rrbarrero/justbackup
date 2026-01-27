package entities

type BackupStats struct {
	Total     int    `json:"total"`
	Pending   int    `json:"pending"`
	Completed int    `json:"completed"`
	Failed    int    `json:"failed"`
	TotalSize string `json:"total_size"`
}
