package commands

import (
	"encoding/json"
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/rrbarrero/justbackup/internal/cli/client"
	"github.com/rrbarrero/justbackup/internal/cli/config"
)

type Host struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Host string `json:"host"`
	Port int    `json:"port"`
	User string `json:"user"`
}

func HostsCommand() {
	cfg, err := config.LoadConfig()
	if err != nil {
		fmt.Printf("Error loading config: %v\nRun 'justbackup config' to configure the CLI.\n", err)
		return
	}

	apiClient := client.NewClient(cfg)

	data, err := apiClient.Get("/hosts")
	if err != nil {
		fmt.Printf("Error fetching hosts: %v\n", err)
		return
	}

	var hosts []Host
	if err := json.Unmarshal(data, &hosts); err != nil {
		fmt.Printf("Error parsing response: %v\n", err)
		return
	}

	if len(hosts) == 0 {
		fmt.Println("No hosts found.")
		return
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
	_, _ = fmt.Fprintln(w, "ID\tNAME\tHOST\tPORT\tUSER")
	for _, h := range hosts {
		_, _ = fmt.Fprintf(w, "%s\t%s\t%s\t%d\t%s\n", h.ID, h.Name, h.Host, h.Port, h.User)
	}
	_ = w.Flush()
}
