package commands

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/rrbarrero/justbackup/internal/cli/client"
	"github.com/rrbarrero/justbackup/internal/cli/config"
)

type BackupFile struct {
	Name  string `json:"name"`
	IsDir bool   `json:"is_dir"`
	Size  int64  `json:"size"`
}

func FilesCommand() {
	filesCmd := flag.NewFlagSet("files", flag.ExitOnError)
	path := filesCmd.String("path", "", "Subpath to list (optional)")

	if len(os.Args) < 3 {
		fmt.Println("Usage: justbackup files <backup-id> [--path <subpath>]")
		return
	}

	backupID := os.Args[2]
	if err := filesCmd.Parse(os.Args[3:]); err != nil {
		fmt.Printf("Error parsing flags: %v\n", err)
		os.Exit(1)
	}

	cfg, err := config.LoadConfig()
	if err != nil {
		fmt.Printf("Error loading config: %v\n", err)
		return
	}

	apiClient := client.NewClient(cfg)
	apiUrl := fmt.Sprintf("/backups/%s/files", backupID)
	if *path != "" {
		apiUrl = fmt.Sprintf("%s?path=%s", apiUrl, *path)
	}

	data, err := apiClient.Get(apiUrl)
	if err != nil {
		fmt.Printf("Error fetching files: %v\n", err)
		return
	}

	var files []BackupFile
	if err := json.Unmarshal(data, &files); err != nil {
		fmt.Printf("Error parsing response: %v\n", err)
		return
	}

	if len(files) == 0 {
		fmt.Println("No files found or empty directory.")
		return
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
	_, _ = fmt.Fprintln(w, "NAME\tTYPE\tSIZE")
	for _, f := range files {
		fType := "FILE"
		if f.IsDir {
			fType = "DIR"
		}

		sizeStr := formatSize(f.Size)
		if f.IsDir {
			sizeStr = "-"
		}

		_, _ = fmt.Fprintf(w, "%s\t%s\t%s\n", f.Name, fType, sizeStr)
	}
	_ = w.Flush()
}

func formatSize(size int64) string {
	const unit = 1024
	if size < unit {
		return fmt.Sprintf("%d B", size)
	}
	div, exp := int64(unit), 0
	for n := size / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(size)/float64(div), "KMGTPE"[exp])
}
