package commands

import (
	"encoding/json"
	"fmt"
	"os"
	"text/tabwriter"
	"time"

	"github.com/rrbarrero/justbackup/internal/cli/client"
	"github.com/rrbarrero/justbackup/internal/cli/config"
)

type Backup struct {
	ID          string    `json:"id"`
	HostID      string    `json:"host_id"`
	Path        string    `json:"path"`
	Destination string    `json:"destination"`
	Status      string    `json:"status"`
	Schedule    string    `json:"schedule"`
	LastRun     time.Time `json:"last_run"`
	Excludes    []string  `json:"excludes"`
	Incremental bool      `json:"incremental"`
}

func BackupsCommand(hostID string) {
	cfg, err := config.LoadConfig()
	if err != nil {
		fmt.Printf("Error loading config: %v\nRun 'justbackup config' to configure the CLI.\n", err)
		return
	}

	apiClient := client.NewClient(cfg)

	path := "/backups"
	if hostID != "" {
		path = fmt.Sprintf("/backups?host_id=%s", hostID)
	}

	data, err := apiClient.Get(path)
	if err != nil {
		fmt.Printf("Error fetching backups: %v\n", err)
		return
	}

	var backups []Backup
	if err := json.Unmarshal(data, &backups); err != nil {
		fmt.Printf("Error parsing response: %v\n", err)
		return
	}

	if len(backups) == 0 {
		fmt.Println("No backups found.")
		return
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
	_, _ = fmt.Fprintln(w, "ID\tHOST ID\tPATH\tDESTINATION\tSTATUS\tSCHEDULE\tLAST RUN")
	for _, b := range backups {
		lastRun := "Never"
		if !b.LastRun.IsZero() {
			lastRun = b.LastRun.Format(time.RFC3339)
		}
		_, _ = fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\t%s\n", b.ID, b.HostID, b.Path, b.Destination, b.Status, b.Schedule, lastRun)
	}
	_ = w.Flush()
}
