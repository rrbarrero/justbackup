package rsync

import (
	"context"
	"fmt"
	"log"
	"os/exec"
)

type RsyncExecutor struct{}

func NewRsyncExecutor() *RsyncExecutor {
	return &RsyncExecutor{}
}

// Sync implements domain.SyncEngine
func (e *RsyncExecutor) Sync(ctx context.Context, source, destination string, excludes []string, dryRun bool) error {
	args := []string{"-avz", "--delete", "--no-owner", "--no-group"}
	for _, exclude := range excludes {
		if exclude == "" {
			continue
		}
		args = append(args, "--exclude", exclude)
	}
	if dryRun {
		args = append(args, "--dry-run")
	}
	args = append(args, source, destination)

	cmd := exec.CommandContext(ctx, "rsync", args...)

	cmd.Stdout = log.Writer()
	cmd.Stderr = log.Writer()

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("rsync command failed: %w", err)
	}

	return nil
}
