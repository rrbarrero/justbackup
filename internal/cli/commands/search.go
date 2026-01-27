package commands

import (
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"text/tabwriter"

	"github.com/rrbarrero/justbackup/internal/cli/client"
	"github.com/rrbarrero/justbackup/internal/cli/config"
)

type FileSearchResult struct {
	Path   string         `json:"path"`
	Backup BackupResponse `json:"backup"`
}

type BackupResponse struct {
	ID          string `json:"id"`
	HostName    string `json:"host_name"`
	Destination string `json:"destination"`
	Status      string `json:"status"`
}

func SearchCommand(pattern string) {
	cfg, err := config.LoadConfig()
	if err != nil {
		fmt.Printf("Error loading config: %v\nRun 'justbackup config' to configure the CLI.\n", err)
		return
	}

	apiClient := client.NewClient(cfg)

	path := fmt.Sprintf("/files/search?pattern=%s", url.QueryEscape(pattern))
	data, err := apiClient.Get(path)
	if err != nil {
		fmt.Printf("Error searching files: %v\n", err)
		return
	}

	var results []FileSearchResult
	if err := json.Unmarshal(data, &results); err != nil {
		fmt.Printf("Error parsing response: %v\n", err)
		return
	}

	if len(results) == 0 {
		fmt.Printf("No files found matching pattern: %s\n", pattern)
		return
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
	fmt.Fprintln(w, "FILE PATH\tHOST\tDESTINATION\tID")
	for _, res := range results {
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\n",
			res.Path,
			res.Backup.HostName,
			res.Backup.Destination,
			res.Backup.ID,
		)
	}
	w.Flush()
}
