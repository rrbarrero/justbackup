package commands

import (
	"encoding/json"
	"fmt"

	"github.com/rrbarrero/justbackup/internal/cli/client"
	"github.com/rrbarrero/justbackup/internal/cli/config"
)

type RunResponse struct {
	TaskID string `json:"task_id"`
}

func RunBackupCommand(backupID string) {
	cfg, err := config.LoadConfig()
	if err != nil {
		fmt.Printf("Error loading config: %v\nRun 'justbackup config' to configure the CLI.\n", err)
		return
	}

	apiClient := client.NewClient(cfg)

	path := fmt.Sprintf("/backups/%s/run", backupID)

	// Empty body for this request
	data, err := apiClient.Post(path, nil)
	if err != nil {
		fmt.Printf("Error triggering backup: %v\n", err)
		return
	}

	var resp RunResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		fmt.Printf("Error parsing response: %v\n", err)
		return
	}

	fmt.Printf("Backup triggered successfully. Task ID: %s\n", resp.TaskID)
	fmt.Println("You can check the status in the dashboard.")
}
