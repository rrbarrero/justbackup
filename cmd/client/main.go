package main

import (
	"fmt"
	"os"

	"github.com/rrbarrero/justbackup/internal/cli/commands"
)

// DefaultBackendURL is injected at compile time via -ldflags
var DefaultBackendURL string

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	command := os.Args[1]
	switch command {
	case "config":
		commands.ConfigCommand(DefaultBackendURL)
	case "hosts":
		commands.HostsCommand()
	case "backups":
		hostID := ""
		if len(os.Args) > 2 {
			hostID = os.Args[2]
		}
		commands.BackupsCommand(hostID)
	case "run":
		if len(os.Args) < 3 {
			fmt.Println("Error: Backup ID is required")
			printUsage()
			os.Exit(1)
		}
		backupID := os.Args[2]
		commands.RunBackupCommand(backupID)
	case "bootstrap":
		commands.BootstrapCommand()
	case "search":
		if len(os.Args) < 3 {
			fmt.Println("Error: Search pattern is required")
			printUsage()
			os.Exit(1)
		}
		pattern := os.Args[2]
		commands.SearchCommand(pattern)
	case "restore":
		commands.RestoreCommand()
	case "decrypt":
		commands.DecryptCommand()
	case "files":
		commands.FilesCommand()
	case "add-backup":
		commands.AddBackupCommand()
	default:
		fmt.Printf("Unknown command: %s\n", command)
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println("Usage: justbackup <command> [args]")
	fmt.Println("Commands:")
	fmt.Println("  config       Configure the CLI (URL and Token)")
	fmt.Println("  hosts        List all hosts")
	fmt.Println("  backups      List backups (optional: [host-id])")
	fmt.Println("  add-backup   Create a new backup task")
	fmt.Println("  run          Trigger a backup immediately (required: <backup-id>)")
	fmt.Println("  bootstrap    Bootstrap a new host (args: --host, --user, --name, [--port])")
	fmt.Println("  search       Search for files in backups (required: <pattern>)")
	fmt.Println("  restore      Restore files or directories (required: <backup-id>)")
	fmt.Println("  files        List files in a backup (required: <backup-id>, optional: --path <subpath>)")
	fmt.Println("  decrypt      Decrypt a backup file offline (args: --file, --out, --id, --key)")
}
