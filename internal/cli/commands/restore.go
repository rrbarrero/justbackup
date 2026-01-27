package commands

import (
	"flag"
	"fmt"
	"os"

	"github.com/rrbarrero/justbackup/internal/cli/client"
	"github.com/rrbarrero/justbackup/internal/cli/config"
)

type restoreOptions struct {
	backupID     string
	isLocal      bool
	isRemote     bool
	remotePath   string
	localDest    string
	targetHostID string
	targetPath   string
	addr         string
}

func RestoreCommand() {
	opts, err := parseRestoreFlags()
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	svc, err := initRestoreService()
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	if opts.isLocal {
		executeLocalRestore(svc, opts)
		return
	}

	if opts.isRemote {
		executeRemoteRestore(svc, opts)
		return
	}
}

func parseRestoreFlags() (*restoreOptions, error) {
	fs := flag.NewFlagSet("restore", flag.ExitOnError)
	opts := &restoreOptions{}

	fs.BoolVar(&opts.isLocal, "local", false, "")
	fs.BoolVar(&opts.isRemote, "remote", false, "")
	fs.StringVar(&opts.remotePath, "path", "", "")
	fs.StringVar(&opts.localDest, "dest", ".", "")
	fs.StringVar(&opts.targetHostID, "to-host", "", "")
	fs.StringVar(&opts.targetPath, "to-path", "", "")
	fs.StringVar(&opts.addr, "addr", "", "")

	if len(os.Args) < 3 {
		printRestoreUsage()
		return nil, fmt.Errorf("missing backup ID")
	}

	opts.backupID = os.Args[2]
	if err := fs.Parse(os.Args[3:]); err != nil {
		fmt.Printf("Error parsing flags: %v\n", err)
		os.Exit(1)
	}

	if opts.remotePath == "" {
		fs.Usage()
		return nil, fmt.Errorf("--path is required")
	}

	if opts.isLocal && opts.isRemote {
		return nil, fmt.Errorf("--local and --remote are mutually exclusive")
	}

	if !opts.isLocal && !opts.isRemote {
		printRestoreUsage()
		return nil, fmt.Errorf("either --local or --remote must be specified")
	}

	return opts, nil
}

func initRestoreService() (*RestoreService, error) {
	cfg, err := config.LoadConfig()
	if err != nil {
		return nil, fmt.Errorf("loading config: %w", err)
	}

	apiClient := client.NewClient(cfg)
	return NewRestoreService(NewAPIService(apiClient), NewNetService()), nil
}

func executeLocalRestore(svc *RestoreService, opts *restoreOptions) {
	params := LocalRestoreParams{
		BackupID:   opts.backupID,
		Path:       opts.remotePath,
		LocalDest:  opts.localDest,
		CustomAddr: opts.addr,
	}
	if err := svc.ExecuteLocal(params); err != nil {
		fmt.Printf("Local restoration failed: %v\n", err)
		return
	}
	fmt.Printf("Restoration completed successfully to %s\n", opts.localDest)
}

func executeRemoteRestore(svc *RestoreService, opts *restoreOptions) {
	if opts.targetPath == "" {
		fmt.Println("Error: --to-path is required for remote restoration")
		return
	}
	params := RemoteRestoreParams{
		BackupID:     opts.backupID,
		Path:         opts.remotePath,
		TargetHostID: opts.targetHostID,
		TargetPath:   opts.targetPath,
	}
	taskID, err := svc.ExecuteRemote(params)
	if err != nil {
		fmt.Printf("Remote restoration failed: %v\n", err)
		return
	}

	fmt.Printf("Remote restore task triggered successfully. Task ID: %s\n", taskID)
	fmt.Println("The worker will now rsync the files to the destination host.")
}

func printRestoreUsage() {
	fmt.Println("Usage: justbackup restore <backup-id> [options]")
	fmt.Println("Options for Local Restore:")
	fmt.Println("  --local              Restore to the local machine")
	fmt.Println("  --path <path>        Path inside the backup (required)")
	fmt.Println("  --dest <dir>         Local destination directory (default: .)")
	fmt.Println("\nOptions for Remote Restore:")
	fmt.Println("  --remote             Restore to a remote host")
	fmt.Println("  --path <path>        Path inside the backup (required)")
	fmt.Println("  --to-path <path>     Target path on the remote host (required)")
	fmt.Println("  --to-host <id>       Target host ID (optional, defaults to original host)")
}
