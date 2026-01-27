package commands

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/rrbarrero/justbackup/internal/backup/application/dto"
	"github.com/rrbarrero/justbackup/internal/cli/client"
	"github.com/rrbarrero/justbackup/internal/cli/config"
)

func AddBackupCommand() {
	addCmd := flag.NewFlagSet("add-backup", flag.ExitOnError)
	hostID := addCmd.String("host-id", "", "ID of the host (required)")
	path := addCmd.String("path", "", "Path to backup (required)")
	destination := addCmd.String("dest", "", "Destination folder name (required)")
	schedule := addCmd.String("schedule", "0 0 * * *", "Cron schedule (default: daily at midnight)")
	excludes := addCmd.String("excludes", "", "Comma-separated list of exclude patterns")
	incremental := addCmd.Bool("incremental", true, "Whether the backup is incremental")

	if len(os.Args) < 2 {
		printAddBackupUsage()
		os.Exit(1)
	}

	if err := addCmd.Parse(os.Args[2:]); err != nil {
		fmt.Printf("Error parsing flags: %v\n", err)
		os.Exit(1)
	}

	if *hostID == "" || *path == "" || *destination == "" {
		fmt.Println("Error: --host-id, --path, and --dest are required.")
		printAddBackupUsage()
		os.Exit(1)
	}

	var excludeList []string
	if *excludes != "" {
		excludeList = strings.Split(*excludes, ",")
	}

	cfg, err := config.LoadConfig()
	if err != nil {
		fmt.Printf("Error loading config: %v\nRun 'justbackup config' to configure the CLI.\n", err)
		return
	}

	apiClient := client.NewClient(cfg)

	req := dto.CreateBackupRequest{
		HostID:      *hostID,
		Path:        *path,
		Destination: *destination,
		Schedule:    *schedule,
		Excludes:    excludeList,
		Incremental: *incremental,
	}

	body, err := json.Marshal(req)
	if err != nil {
		fmt.Printf("Error marshaling request: %v\n", err)
		return
	}

	fmt.Println("Creating backup task...")
	_, err = apiClient.Post("/backups", bytes.NewBuffer(body))
	if err != nil {
		fmt.Printf("Error creating backup: %v\n", err)
		return
	}

	fmt.Println("Backup task created successfully!")
}

func printAddBackupUsage() {
	fmt.Println("Usage: justbackup add-backup --host-id <id> --path <path> --dest <name> [options]")
	fmt.Println("Options:")
	fmt.Println("  --host-id <id>    ID of the host (required)")
	fmt.Println("  --path <path>     Source path to backup (required)")
	fmt.Println("  --dest <name>     Destination folder name (required)")
	fmt.Println("  --schedule <cron> Cron schedule expression (default: '0 0 * * *')")
	fmt.Println("  --excludes <p1,p2> Comma-separated exclude patterns")
	fmt.Println("  --incremental      Enable incremental backups (default: true)")
}
